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

// Projects — проекты и их сводные показатели.
type Projects struct {
	db *gorm.DB
}

// NewProjects собирает сервис проектов.
func NewProjects(db *gorm.DB) *Projects {
	return &Projects{db: db}
}

// Metrics — показатели проекта, которые не хранятся в таблице, а считаются
// из его задач. Хранить их в колонках значило бы пересчитывать их при каждой
// правке любой задачи проекта — дешевле посчитать один раз при чтении.
//
// Формула упрощённая: полноценный расчёт (критический путь, резерв, риск срыва
// дедлайна) требует CPM-движка (internal/schedule), которого в проекте пока нет.
// До его появления риском считается сам факт просрочки задачи.
type Metrics struct {
	ProgressPercent    float64
	HealthIndex        int
	CriticalRisksCount int
}

// ComputeMetrics считает сводные показатели проекта по его задачам.
func ComputeMetrics(ctx context.Context, db *gorm.DB, project models.Project) (Metrics, error) {
	var tasks []models.Task
	err := db.WithContext(ctx).
		Select("id", "status", "progress_percent", "weight_percent", "end_date").
		Where("project_id = ?", project.ID).
		Find(&tasks).Error
	if err != nil {
		return Metrics{}, fmt.Errorf("выбрать задачи для расчёта показателей: %w", err)
	}
	if len(tasks) == 0 {
		return Metrics{ProgressPercent: 0, HealthIndex: 100, CriticalRisksCount: 0}, nil
	}

	today := models.Today()
	var weightSum, weightedProgress, plainProgress float64
	overdueCount := 0
	for _, t := range tasks {
		plainProgress += float64(t.ProgressPercent)
		weightSum += t.WeightPercent
		weightedProgress += float64(t.ProgressPercent) * t.WeightPercent
		if t.Status != models.TaskStatusDone && today.After(t.EndDate) {
			overdueCount++
		}
	}

	progress := plainProgress / float64(len(tasks))
	if weightSum > 0 {
		progress = weightedProgress / weightSum
	}

	health := 100 - overdueCount*10
	if project.Status == models.ProjectStatusActive && today.After(project.Deadline) && progress < 100 {
		health -= 20
	}
	if health < 0 {
		health = 0
	}
	if health > 100 {
		health = 100
	}

	return Metrics{ProgressPercent: progress, HealthIndex: health, CriticalRisksCount: overdueCount}, nil
}

// List возвращает проекты, в команде которых состоит пользователь.
func (s *Projects) List(ctx context.Context, userID uuid.UUID, status models.ProjectStatus, search string) ([]models.Project, error) {
	query := s.db.WithContext(ctx).
		Model(&models.Project{}).
		Joins("JOIN project_members ON project_members.project_id = projects.id").
		Where("project_members.user_id = ?", userID).
		Order("projects.created_at DESC")

	if status != "" {
		query = query.Where("projects.status = ?", status)
	}
	if search = strings.TrimSpace(search); search != "" {
		pattern := "%" + escapeLikePattern(search) + "%"
		query = query.Where("projects.name ILIKE ? OR projects.code ILIKE ?", pattern, pattern)
	}

	var projects []models.Project
	if err := query.Find(&projects).Error; err != nil {
		return nil, fmt.Errorf("выбрать проекты: %w", err)
	}
	return projects, nil
}

// CreateInput — данные для создания проекта (шаг 1 мастера инициации).
type CreateInput struct {
	Name                      string
	Description               string
	CustomerOrg               string
	StartDate                 models.Date
	Deadline                  models.Date
	WorkingCalendarType       models.WorkingCalendarType
	IncludePublicHolidays     *bool
	AutoRecalculateDependents *bool
	HighlightCriticalPath     *bool
	Members                   []MemberInput
	SaveAsDraft               bool
}

// MemberInput — участник, добавляемый на шаге 2 мастера инициации.
type MemberInput struct {
	UserID      uuid.UUID
	ProjectRole models.ProjectRole
	SprintRole  string
	AccessLevel models.AccessLevel
}

// Create создаёт проект и сразу добавляет создателя в команду с полным доступом:
// без этого автор проекта не прошёл бы собственную проверку доступа на GET /projects/{id}.
func (s *Projects) Create(ctx context.Context, userID uuid.UUID, in CreateInput) (models.Project, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return models.Project{}, Invalid("укажите название проекта")
	}
	if in.StartDate.IsZero() || in.Deadline.IsZero() {
		return models.Project{}, Invalid("укажите сроки проекта")
	}
	if in.Deadline.Before(in.StartDate) {
		return models.Project{}, Invalid("дедлайн не может быть раньше даты начала")
	}
	calendar := in.WorkingCalendarType
	if calendar == "" {
		calendar = models.Calendar52
	}
	if !calendar.Valid() {
		return models.Project{}, Invalid("недопустимый рабочий календарь %q", calendar)
	}

	status := models.ProjectStatusActive
	if in.SaveAsDraft {
		status = models.ProjectStatusDraft
	}

	project := models.Project{
		Code:                      generateCode("PRJ"),
		Name:                      name,
		Description:               strings.TrimSpace(in.Description),
		CustomerOrg:               strings.TrimSpace(in.CustomerOrg),
		Status:                    status,
		StartDate:                 in.StartDate,
		Deadline:                  in.Deadline,
		WorkingCalendarType:       calendar,
		IncludePublicHolidays:     boolOrDefault(in.IncludePublicHolidays, true),
		AutoRecalculateDependents: boolOrDefault(in.AutoRecalculateDependents, true),
		HighlightCriticalPath:     boolOrDefault(in.HighlightCriticalPath, true),
		CreatedByID:               userID,
	}

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&project).Error; err != nil {
			return fmt.Errorf("создать проект: %w", err)
		}

		creator := models.ProjectMember{
			ProjectID:        project.ID,
			UserID:           userID,
			ProjectRole:      models.ProjectRoleProjectManager,
			AccessLevel:      models.AccessLevelFull,
			WeeklyHoursLimit: 40,
		}
		if err := tx.Create(&creator).Error; err != nil {
			return fmt.Errorf("добавить создателя в команду: %w", err)
		}

		for _, m := range in.Members {
			if m.UserID == userID {
				// Создатель уже добавлен с полным доступом; повторная запись
				// упала бы на уникальном индексе (projectId, userId).
				continue
			}
			if !m.ProjectRole.Valid() {
				return Invalid("недопустимая роль участника %q", m.ProjectRole)
			}
			accessLevel := m.AccessLevel
			if accessLevel == "" {
				accessLevel = models.AccessLevelEdit
			}
			if !accessLevel.Valid() {
				return Invalid("недопустимый уровень доступа %q", accessLevel)
			}
			member := models.ProjectMember{
				ProjectID:        project.ID,
				UserID:           m.UserID,
				ProjectRole:      m.ProjectRole,
				SprintRole:       strings.TrimSpace(m.SprintRole),
				AccessLevel:      accessLevel,
				WeeklyHoursLimit: 40,
			}
			if err := tx.Create(&member).Error; err != nil {
				if isUniqueViolation(err) {
					return Invalid("пользователь уже добавлен в команду")
				}
				if isForeignKeyViolation(err) {
					return Invalid("пользователь не найден")
				}
				return fmt.Errorf("добавить участника: %w", err)
			}
		}
		return nil
	})
	if err != nil {
		return models.Project{}, err
	}

	return s.Get(ctx, project.ID)
}

// Get возвращает проект по идентификатору.
func (s *Projects) Get(ctx context.Context, projectID uuid.UUID) (models.Project, error) {
	var project models.Project
	err := s.db.WithContext(ctx).Preload("CreatedBy").First(&project, "id = ?", projectID).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return models.Project{}, ErrNotFound
	case err != nil:
		return models.Project{}, fmt.Errorf("найти проект: %w", err)
	}
	return project, nil
}

// UpdateInput — поля проекта, доступные для частичного обновления.
// Указатели различают «поле не передано» и «поле сброшено в пустое значение».
type UpdateInput struct {
	Name                *string
	Description         *string
	Status              *models.ProjectStatus
	Phase               *string
	StartDate           *models.Date
	Deadline            *models.Date
	WorkingCalendarType *models.WorkingCalendarType
}

// Update частично обновляет параметры проекта.
func (s *Projects) Update(ctx context.Context, projectID uuid.UUID, in UpdateInput) (models.Project, error) {
	project, err := s.Get(ctx, projectID)
	if err != nil {
		return models.Project{}, err
	}

	if in.Name != nil {
		name := strings.TrimSpace(*in.Name)
		if name == "" {
			return models.Project{}, Invalid("название проекта не может быть пустым")
		}
		project.Name = name
	}
	if in.Description != nil {
		project.Description = strings.TrimSpace(*in.Description)
	}
	if in.Status != nil {
		if !in.Status.Valid() {
			return models.Project{}, Invalid("недопустимый статус проекта %q", *in.Status)
		}
		project.Status = *in.Status
	}
	if in.Phase != nil {
		project.Phase = strings.TrimSpace(*in.Phase)
	}
	if in.StartDate != nil {
		project.StartDate = *in.StartDate
	}
	if in.Deadline != nil {
		project.Deadline = *in.Deadline
	}
	if in.WorkingCalendarType != nil {
		if !in.WorkingCalendarType.Valid() {
			return models.Project{}, Invalid("недопустимый рабочий календарь %q", *in.WorkingCalendarType)
		}
		project.WorkingCalendarType = *in.WorkingCalendarType
	}
	if project.Deadline.Before(project.StartDate) {
		return models.Project{}, Invalid("дедлайн не может быть раньше даты начала")
	}

	if err := s.db.WithContext(ctx).Save(&project).Error; err != nil {
		return models.Project{}, fmt.Errorf("сохранить проект: %w", err)
	}
	return s.Get(ctx, projectID)
}

// Archive переводит проект в статус archived. Данные проекта не удаляются —
// спецификация прямо требует логическое, а не физическое удаление.
func (s *Projects) Archive(ctx context.Context, projectID uuid.UUID) error {
	res := s.db.WithContext(ctx).
		Model(&models.Project{}).
		Where("id = ?", projectID).
		Update("status", models.ProjectStatusArchived)
	if res.Error != nil {
		return fmt.Errorf("архивировать проект: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// generateCode собирает короткий читаемый код вида PRJ-2026-A1B2.
// Случайный суффикс вместо инкремента: инкремент потребовал бы блокировки
// таблицы на вставке, а коллизия 16^4 вариантов на масштабе хакатона исключена.
func generateCode(prefix string) string {
	return fmt.Sprintf("%s-%d-%s", prefix, models.Today().Year(), strings.ToUpper(uuid.New().String()[:4]))
}

func boolOrDefault(v *bool, fallback bool) bool {
	if v == nil {
		return fallback
	}
	return *v
}
