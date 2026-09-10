package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"taygant_backend/internal/models"
)

// Tasks — задачи проекта и вычисляемые поля реестра/диаграммы Ганта.
type Tasks struct {
	db *gorm.DB
}

// NewTasks собирает сервис задач.
func NewTasks(db *gorm.DB) *Tasks {
	return &Tasks{db: db}
}

// ProjectID возвращает проект, которому принадлежит задача. Нужен хендлерам
// вложенных ресурсов (dependencies, checklist, comments, history, simulation),
// адресуемым taskId без projectId в пути.
func (s *Tasks) ProjectID(ctx context.Context, taskID uuid.UUID) (uuid.UUID, error) {
	return projectIDOfTask(ctx, s.db, taskID)
}

// Filter — параметры выборки задач из GET /projects/{projectId}/tasks.
//
// criticalPathOnly из спецификации здесь не реализован: критический путь
// требует CPM-движка (internal/schedule), которого пока нет — фильтр вернул бы
// либо пустой список, либо все задачи, и оба варианта вводят в заблуждение.
type Filter struct {
	SprintID   *uuid.UUID
	AssigneeID *uuid.UUID
	Status     models.TaskStatus
	RisksOnly  bool
	Search     string
}

// List возвращает задачи проекта с вычисленными статусами и длительностями.
//
// Фильтр по статусу применяется уже после вычисления effective-статуса
// (overdue/blocked пользователь в БД не хранит — см. resolveStatuses), поэтому
// он не переносится в SQL WHERE и выполняется в Go после загрузки.
func (s *Tasks) List(ctx context.Context, projectID uuid.UUID, calendar models.Project, f Filter) ([]models.Task, error) {
	query := s.db.WithContext(ctx).
		Preload("Assignee").
		Where("project_id = ?", projectID).
		Order("order_index")

	if f.SprintID != nil {
		query = query.Where("sprint_id = ?", *f.SprintID)
	}
	if f.AssigneeID != nil {
		query = query.Where("assignee_id = ?", *f.AssigneeID)
	}
	if search := strings.TrimSpace(f.Search); search != "" {
		pattern := "%" + escapeLikePattern(search) + "%"
		query = query.Where("title ILIKE ? OR code ILIKE ?", pattern, pattern)
	}

	var tasks []models.Task
	if err := query.Find(&tasks).Error; err != nil {
		return nil, fmt.Errorf("выбрать задачи: %w", err)
	}

	deps, err := dependenciesForProject(ctx, s.db, projectID)
	if err != nil {
		return nil, err
	}
	resolveStatuses(tasks, statusMap(tasks), deps)
	applyDerivedFields(tasks, calendar)

	filtered := tasks[:0]
	for _, t := range tasks {
		if f.Status != "" && t.Status != f.Status {
			continue
		}
		if f.RisksOnly && t.PlanVsActualDeviation() >= 0 {
			continue
		}
		filtered = append(filtered, t)
	}
	return filtered, nil
}

// resolveStatuses выставляет overdue/blocked — статусы, которые пользователь
// не может задать напрямую (см. models.TaskStatus). Задача просрочена, если её
// плановый конец в прошлом и она не завершена; заблокирована — если у неё есть
// незавершённый предшественник. Просрочка приоритетнее блокировки: сорванный
// срок важнее показать, чем причину, по которой задача не началась.
//
// statusByID должен содержать хранимый статус КАЖДОЙ задачи проекта, а не только
// тех, что перечислены в tasks: иначе для предшественника, отсутствующего в tasks
// (например, при запросе одной задачи через Get), лукап вернёт нулевое значение,
// которое не равно Done, и задача навсегда останется заблокированной — баг,
// пойманный на живом сценарии «завершили задачу — она осталась blocked».
func resolveStatuses(tasks []models.Task, statusByID map[uuid.UUID]models.TaskStatus, deps []models.TaskDependency) {
	blockedBy := make(map[uuid.UUID]bool)
	for _, d := range deps {
		if predStatus := statusByID[d.PredecessorTaskID]; predStatus != models.TaskStatusDone {
			blockedBy[d.SuccessorTaskID] = true
		}
	}

	today := models.Today()
	for i := range tasks {
		switch {
		case tasks[i].Status == models.TaskStatusDone:
			// финальный статус, ничего не пересчитываем
		case today.After(tasks[i].EndDate):
			tasks[i].Status = models.TaskStatusOverdue
		case blockedBy[tasks[i].ID]:
			tasks[i].Status = models.TaskStatusBlocked
		}
	}
}

// statusMap строит карту статусов из уже загруженного среза задач.
func statusMap(tasks []models.Task) map[uuid.UUID]models.TaskStatus {
	m := make(map[uuid.UUID]models.TaskStatus, len(tasks))
	for _, t := range tasks {
		m[t.ID] = t.Status
	}
	return m
}

// projectStatusMap подгружает хранимые статусы всех задач проекта одним
// лёгким запросом (без Preload). Нужен там, где resolveStatuses вызывается
// не для полного списка задач проекта, а для одной задачи (см. Get) —
// иначе проверка предшественника не сможет узнать его статус.
func projectStatusMap(ctx context.Context, db *gorm.DB, projectID uuid.UUID) (map[uuid.UUID]models.TaskStatus, error) {
	var rows []models.Task
	err := db.WithContext(ctx).
		Select("id", "status").
		Where("project_id = ?", projectID).
		Find(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("выбрать статусы задач проекта: %w", err)
	}
	return statusMap(rows), nil
}

// applyDerivedFields считает длительности, которые не хранятся в таблице.
// Критический путь (IsCriticalPath) требует CPM-движка (internal/schedule)
// и пока не реализован — до его появления поле остаётся false.
func applyDerivedFields(tasks []models.Task, project models.Project) {
	for i := range tasks {
		tasks[i].DurationCalendarDays = tasks[i].StartDate.DaysUntil(tasks[i].EndDate)
		tasks[i].DurationWorkingDays = workingDaysBetween(tasks[i].StartDate, tasks[i].EndDate, project.WorkingCalendarType)
	}
}

// workingDaysBetween считает рабочие дни календаря проекта между двумя датами
// (включительно). Государственные праздники РФ не вычитаются: их календаря
// в проекте нет, а зашивать фиксированный список в код означало бы, что он
// устареет к следующему году. IncludePublicHolidays учитывается, когда
// такой календарь появится.
func workingDaysBetween(start, end models.Date, calendar models.WorkingCalendarType) int {
	if end.Before(start) {
		return 0
	}
	days := 0
	for d := start; !d.After(end); d = d.AddDays(1) {
		weekday := d.Weekday()
		isWorking := weekday != 0 && weekday != 6 // вс всегда выходной
		if calendar == models.Calendar52 && weekday == 6 {
			isWorking = false
		}
		if isWorking {
			days++
		}
	}
	return days
}

// TaskInput — общие поля создания и редактирования задачи.
type TaskInput struct {
	SprintID        *uuid.UUID
	ParentTaskID    *uuid.UUID
	Title           *string
	Description     *string
	AssigneeID      *uuid.UUID
	Status          *models.TaskStatus
	IsMilestone     *bool
	StartDate       *models.Date
	EndDate         *models.Date
	ProgressPercent *int
	WeightPercent   *float64
}

// PredecessorInput — связь-предшественник, создаваемая вместе с задачей.
type PredecessorInput struct {
	TaskID  uuid.UUID
	Type    models.DependencyType
	LagDays int
}

// Create создаёт задачу в проекте, опционально сразу со связями-предшественниками.
func (s *Tasks) Create(ctx context.Context, projectID, actorID uuid.UUID, in TaskInput, predecessors []PredecessorInput) (models.Task, error) {
	if in.Title == nil || strings.TrimSpace(*in.Title) == "" {
		return models.Task{}, Invalid("укажите название задачи")
	}
	if in.StartDate == nil || in.EndDate == nil {
		return models.Task{}, Invalid("укажите сроки задачи")
	}
	if in.EndDate.Before(*in.StartDate) {
		return models.Task{}, Invalid("дата окончания не может быть раньше даты начала")
	}
	status := models.TaskStatusPlanned
	if in.Status != nil {
		if !userSettableStatus(*in.Status) {
			return models.Task{}, Invalid("статус %q выставляется системой автоматически", *in.Status)
		}
		status = *in.Status
	}

	task := models.Task{
		ProjectID:       projectID,
		SprintID:        in.SprintID,
		ParentTaskID:    in.ParentTaskID,
		Title:           strings.TrimSpace(*in.Title),
		Status:          status,
		IsMilestone:     in.IsMilestone != nil && *in.IsMilestone,
		StartDate:       *in.StartDate,
		EndDate:         *in.EndDate,
		ProgressPercent: 0,
	}
	if in.Description != nil {
		task.Description = strings.TrimSpace(*in.Description)
	}
	if in.AssigneeID != nil {
		task.AssigneeID = in.AssigneeID
	}
	if in.WeightPercent != nil {
		task.WeightPercent = *in.WeightPercent
	}

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		code, orderIndex, err := nextTaskCode(tx, projectID)
		if err != nil {
			return err
		}
		task.Code = code
		task.OrderIndex = orderIndex
		task.WBSNumber, err = nextWBSNumber(tx, projectID, in.ParentTaskID)
		if err != nil {
			return err
		}

		if err := validateSprintAndParent(tx, projectID, in.SprintID, in.ParentTaskID); err != nil {
			return err
		}
		if in.AssigneeID != nil {
			if err := validateAssignee(tx, projectID, *in.AssigneeID); err != nil {
				return err
			}
		}

		if err := tx.Create(&task).Error; err != nil {
			return fmt.Errorf("создать задачу: %w", err)
		}

		for _, p := range predecessors {
			if err := createDependency(ctx, tx, projectID, p.TaskID, task.ID, p.Type, p.LagDays); err != nil {
				return err
			}
		}

		return recordHistory(tx, task.ID, actorID, models.HistoryActionCreated, nil, nil, nil)
	})
	if err != nil {
		return models.Task{}, err
	}
	return s.Get(ctx, task.ID)
}

// userSettableStatus запрещает клиенту напрямую выставлять статусы, которые
// система вычисляет сама (см. resolveStatuses).
func userSettableStatus(status models.TaskStatus) bool {
	switch status {
	case models.TaskStatusPlanned, models.TaskStatusInProgress, models.TaskStatusDone:
		return true
	default:
		return false
	}
}

// nextTaskCode и порядковый индекс — по числу существующих задач проекта.
// Код не уникален на уровне БД (см. models.Task): при гонке двух параллельных
// создателей возможно совпадение кода, что для MVP приемлемо.
func nextTaskCode(tx *gorm.DB, projectID uuid.UUID) (string, int, error) {
	var count int64
	if err := tx.Model(&models.Task{}).Where("project_id = ?", projectID).Count(&count).Error; err != nil {
		return "", 0, fmt.Errorf("посчитать задачи проекта: %w", err)
	}
	return fmt.Sprintf("TASK-%03d", count+1), int(count), nil
}

// nextWBSNumber строит номер в иерархии работ: "N" для задач верхнего уровня,
// "родительWBS.N" для подзадач — N считается по числу уже существующих
// соседей с тем же родителем.
func nextWBSNumber(tx *gorm.DB, projectID uuid.UUID, parentID *uuid.UUID) (string, error) {
	query := tx.Model(&models.Task{}).Where("project_id = ?", projectID)
	if parentID == nil {
		query = query.Where("parent_task_id IS NULL")
	} else {
		query = query.Where("parent_task_id = ?", *parentID)
	}
	var siblings int64
	if err := query.Count(&siblings).Error; err != nil {
		return "", fmt.Errorf("посчитать соседние задачи: %w", err)
	}

	if parentID == nil {
		return fmt.Sprintf("%d", siblings+1), nil
	}
	var parent models.Task
	if err := tx.Select("wbs_number").First(&parent, "id = ?", *parentID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", Invalid("родительская задача не найдена")
		}
		return "", fmt.Errorf("найти родительскую задачу: %w", err)
	}
	return fmt.Sprintf("%s.%d", parent.WBSNumber, siblings+1), nil
}

func validateSprintAndParent(tx *gorm.DB, projectID uuid.UUID, sprintID, parentID *uuid.UUID) error {
	if sprintID != nil {
		var count int64
		if err := tx.Model(&models.Sprint{}).Where("id = ? AND project_id = ?", *sprintID, projectID).Count(&count).Error; err != nil {
			return fmt.Errorf("проверить спринт: %w", err)
		}
		if count == 0 {
			return Invalid("спринт не найден в этом проекте")
		}
	}
	if parentID != nil {
		var count int64
		if err := tx.Model(&models.Task{}).Where("id = ? AND project_id = ?", *parentID, projectID).Count(&count).Error; err != nil {
			return fmt.Errorf("проверить родительскую задачу: %w", err)
		}
		if count == 0 {
			return Invalid("родительская задача не найдена в этом проекте")
		}
	}
	return nil
}

func validateAssignee(tx *gorm.DB, projectID, userID uuid.UUID) error {
	var count int64
	err := tx.Model(&models.ProjectMember{}).
		Where("project_id = ? AND user_id = ?", projectID, userID).
		Count(&count).Error
	if err != nil {
		return fmt.Errorf("проверить ответственного: %w", err)
	}
	if count == 0 {
		return Invalid("ответственный должен быть участником проекта")
	}
	return nil
}

// Get возвращает задачу с вычисленными полями.
func (s *Tasks) Get(ctx context.Context, taskID uuid.UUID) (models.Task, error) {
	task, err := s.get(ctx, taskID)
	if err != nil {
		return models.Task{}, err
	}
	var project models.Project
	if err := s.db.WithContext(ctx).Select("working_calendar_type").First(&project, "id = ?", task.ProjectID).Error; err != nil {
		return models.Task{}, fmt.Errorf("найти проект задачи: %w", err)
	}
	deps, err := dependenciesForProject(ctx, s.db, task.ProjectID)
	if err != nil {
		return models.Task{}, err
	}
	statuses, err := projectStatusMap(ctx, s.db, task.ProjectID)
	if err != nil {
		return models.Task{}, err
	}
	tasks := []models.Task{task}
	resolveStatuses(tasks, statuses, deps)
	applyDerivedFields(tasks, project)
	return tasks[0], nil
}

// Update частично обновляет задачу, ведёт журнал изменений по каждому
// изменённому полю и синхронизирует фактические даты со статусом:
// переход в in_progress проставляет ActualStartDate, переход в done — ActualEndDate,
// если они ещё не заданы. Без этого плановое/фактическое отклонение (см. Task
// в api-spec.yml) никогда бы не заполнялось для задач, которые ведут через статус,
// а не отдельным полем фактических дат.
func (s *Tasks) Update(ctx context.Context, taskID, actorID uuid.UUID, in TaskInput) (models.Task, error) {
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var task models.Task
		if err := tx.First(&task, "id = ?", taskID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNotFound
			}
			return fmt.Errorf("найти задачу: %w", err)
		}

		if in.SprintID != nil {
			if err := validateSprintAndParent(tx, task.ProjectID, in.SprintID, nil); err != nil {
				return err
			}
			task.SprintID = in.SprintID
		}
		if in.Title != nil {
			title := strings.TrimSpace(*in.Title)
			if title == "" {
				return Invalid("название задачи не может быть пустым")
			}
			task.Title = title
		}
		if in.Description != nil {
			task.Description = strings.TrimSpace(*in.Description)
		}
		if in.AssigneeID != nil {
			if err := validateAssignee(tx, task.ProjectID, *in.AssigneeID); err != nil {
				return err
			}
			old := task.AssigneeID
			task.AssigneeID = in.AssigneeID
			if !uuidPtrEqual(old, in.AssigneeID) {
				if err := recordHistory(tx, taskID, actorID, models.HistoryActionAssigned, strPtr("assigneeId"), uuidPtrString(old), uuidPtrString(in.AssigneeID)); err != nil {
					return err
				}
			}
		}
		if in.StartDate != nil {
			task.StartDate = *in.StartDate
		}
		if in.EndDate != nil {
			task.EndDate = *in.EndDate
		}
		if task.EndDate.Before(task.StartDate) {
			return Invalid("дата окончания не может быть раньше даты начала")
		}
		if in.ProgressPercent != nil {
			if *in.ProgressPercent < 0 || *in.ProgressPercent > 100 {
				return Invalid("процент выполнения должен быть от 0 до 100")
			}
			task.ProgressPercent = *in.ProgressPercent
		}
		if in.WeightPercent != nil {
			task.WeightPercent = *in.WeightPercent
		}
		if in.Status != nil {
			if !userSettableStatus(*in.Status) {
				return Invalid("статус %q выставляется системой автоматически", *in.Status)
			}
			if task.Status != *in.Status {
				old := string(task.Status)
				if err := recordHistory(tx, taskID, actorID, models.HistoryActionStatusChanged, strPtr("status"), &old, strPtr(string(*in.Status))); err != nil {
					return err
				}
				task.Status = *in.Status
				today := models.Today()
				if task.Status == models.TaskStatusInProgress && task.ActualStartDate == nil {
					task.ActualStartDate = &today
				}
				if task.Status == models.TaskStatusDone && task.ActualEndDate == nil {
					task.ActualEndDate = &today
					if task.ActualStartDate == nil {
						task.ActualStartDate = &today
					}
				}
			}
		}

		if err := tx.Save(&task).Error; err != nil {
			return fmt.Errorf("сохранить задачу: %w", err)
		}
		return nil
	})
	if err != nil {
		return models.Task{}, err
	}
	return s.Get(ctx, taskID)
}

// Delete удаляет задачу вместе со связями, чек-листом, комментариями и историей —
// все они настроены в моделях на OnDelete:CASCADE.
func (s *Tasks) Delete(ctx context.Context, taskID uuid.UUID) error {
	res := s.db.WithContext(ctx).Delete(&models.Task{}, "id = ?", taskID)
	if res.Error != nil {
		return fmt.Errorf("удалить задачу: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Tasks) get(ctx context.Context, taskID uuid.UUID) (models.Task, error) {
	var task models.Task
	err := s.db.WithContext(ctx).Preload("Assignee").First(&task, "id = ?", taskID).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return models.Task{}, ErrNotFound
	case err != nil:
		return models.Task{}, fmt.Errorf("найти задачу: %w", err)
	}
	return task, nil
}

func uuidPtrEqual(a, b *uuid.UUID) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

func uuidPtrString(id *uuid.UUID) *string {
	if id == nil {
		return nil
	}
	s := id.String()
	return &s
}
