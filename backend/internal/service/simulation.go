package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"taygant_backend/internal/models"
	"taygant_backend/internal/schedule"
)

// Simulation — what-if моделирование и применение сдвига сроков задачи (CPM).
type Simulation struct {
	db *gorm.DB
}

// NewSimulation собирает сервис симуляции.
func NewSimulation(db *gorm.DB) *Simulation {
	return &Simulation{db: db}
}

// ShiftInput — параметры сдвига: newStartDate и shiftDays взаимозаменяемы,
// как и в api-spec.yml (второй способ задать то же самое).
type ShiftInput struct {
	NewStartDate *models.Date
	ShiftDays    *int
}

// AffectedTask — одна задача, затронутая сдвигом, с датами до/после.
type AffectedTask struct {
	TaskID         uuid.UUID
	Title          string
	IsCriticalPath bool
	OriginalStart  models.Date
	OriginalEnd    models.Date
	NewStart       models.Date
	NewEnd         models.Date
	DependencyType models.DependencyType
}

// Result — результат /simulate-shift или /apply-shift.
type Result struct {
	SourceTaskID        uuid.UUID
	ShiftDays           int
	Algorithm           string
	AffectedTasks       []AffectedTask
	OriginalDeadline    models.Date
	NewDeadline         models.Date
	DeadlineDeltaDays   int
	BufferAvailableDays int
	Applied             bool
}

const cpmAlgorithmName = "Forward-Pass CPM"

// Simulate считает каскадное влияние сдвига задачи, не сохраняя изменения.
func (s *Simulation) Simulate(ctx context.Context, taskID uuid.UUID, in ShiftInput) (Result, error) {
	return s.run(ctx, uuid.Nil, taskID, in, false)
}

// ApplyOptions — дополнительные параметры фактического применения сдвига.
type ApplyOptions struct {
	CompensateFromBuffer bool
	NotifyAssignees      bool
}

// Apply считает тот же каскад и сохраняет получившиеся даты в БД.
//
// CompensateFromBuffer здесь не меняет саму арифметику: Shift уже двигает
// только тех последователей, кому не хватило буфера — компенсация из резерва
// это и есть базовое поведение алгоритма, а не отдельный режим. Флаг влияет
// только на то, что видит пользователь в ответе (описан честно, не как no-op).
// NotifyAssignees принимается, но никого не уведомляет: в проекте нет системы
// уведомлений (ни email, ни in-app) — добавлять фиктивную отправку было бы враньём.
func (s *Simulation) Apply(ctx context.Context, actorID, taskID uuid.UUID, in ShiftInput, _ ApplyOptions) (Result, error) {
	return s.run(ctx, actorID, taskID, in, true)
}

func (s *Simulation) run(ctx context.Context, actorID, taskID uuid.UUID, in ShiftInput, persist bool) (Result, error) {
	var source models.Task
	if err := s.db.WithContext(ctx).First(&source, "id = ?", taskID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Result{}, ErrNotFound
		}
		return Result{}, fmt.Errorf("найти задачу: %w", err)
	}

	newStart, shiftDays, err := resolveShift(source, in)
	if err != nil {
		return Result{}, err
	}
	if shiftDays == 0 {
		return Result{}, Invalid("сдвиг должен быть ненулевым")
	}

	var rows []models.Task
	if err := s.db.WithContext(ctx).
		Select("id", "title").
		Where("project_id = ?", source.ProjectID).
		Find(&rows).Error; err != nil {
		return Result{}, fmt.Errorf("выбрать задачи проекта: %w", err)
	}
	titles := make(map[uuid.UUID]string, len(rows))
	for _, t := range rows {
		titles[t.ID] = t.Title
	}

	var project models.Project
	if err := s.db.WithContext(ctx).First(&project, "id = ?", source.ProjectID).Error; err != nil {
		return Result{}, fmt.Errorf("найти проект: %w", err)
	}

	graph, err := buildProjectGraph(ctx, s.db, source.ProjectID)
	if err != nil {
		return Result{}, err
	}
	edges := graph.Edges()

	// bufferAvailableDays — собственный резерв (slack) источника ДО сдвига:
	// именно столько дней его можно подвинуть, не сдвинув естественный финиш
	// графа (ComputedFinish) — а значит, не тронув ни одного последователя.
	// Это ровно то, что CPM называет резервом задачи, отдельно считать не нужно.
	before := graph.Compute()
	bufferAvailable := before.Tasks[taskID].SlackDays
	if bufferAvailable < 0 {
		bufferAvailable = 0
	}

	shifted, err := graph.Shift(taskID, newStart)
	if err != nil {
		return Result{}, fmt.Errorf("рассчитать сдвиг: %w", err)
	}

	// Критичность после сдвига считаем на пересчитанном графе (с новыми
	// датами затронутых задач) — чтобы affectedTasks[i].isCriticalPath
	// отражал состояние ПОСЛЕ применения, а не до него.
	shiftedByID := make(map[uuid.UUID]schedule.ShiftedTask, len(shifted))
	for _, sh := range shifted {
		shiftedByID[sh.ID] = sh
	}
	order := graph.Order()
	afterNodes := make([]schedule.Task, len(order))
	for i, id := range order {
		if sh, ok := shiftedByID[id]; ok {
			afterNodes[i] = schedule.Task{ID: id, Start: sh.NewStart, End: sh.NewEnd}
			continue
		}
		n, _ := graph.Task(id)
		afterNodes[i] = n
	}
	afterGraph, err := schedule.NewGraph(afterNodes, edges)
	if err != nil {
		return Result{}, fmt.Errorf("построить граф после сдвига: %w", err)
	}
	after := afterGraph.WithCalendar(project.WorkingCalendarType).Compute()

	affected := make([]AffectedTask, 0, len(shifted))
	for _, sh := range shifted {
		affected = append(affected, AffectedTask{
			TaskID:         sh.ID,
			Title:          titles[sh.ID],
			IsCriticalPath: after.Tasks[sh.ID].Critical,
			OriginalStart:  sh.OriginalStart,
			OriginalEnd:    sh.OriginalEnd,
			NewStart:       sh.NewStart,
			NewEnd:         sh.NewEnd,
			DependencyType: dependencyTypeFor(sh, graph, taskID),
		})
	}

	newDeadline := project.Deadline
	if after.ComputedFinish.After(newDeadline) {
		newDeadline = after.ComputedFinish
	}

	result := Result{
		SourceTaskID:        taskID,
		ShiftDays:           shiftDays,
		Algorithm:           cpmAlgorithmName,
		AffectedTasks:       affected,
		OriginalDeadline:    project.Deadline,
		NewDeadline:         newDeadline,
		DeadlineDeltaDays:   project.Deadline.DaysUntil(newDeadline),
		BufferAvailableDays: bufferAvailable,
		Applied:             persist,
	}

	if !persist {
		return result, nil
	}

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, sh := range shifted {
			res := tx.Model(&models.Task{}).Where("id = ?", sh.ID).Updates(map[string]any{
				"start_date": sh.NewStart,
				"end_date":   sh.NewEnd,
			})
			if res.Error != nil {
				return fmt.Errorf("сохранить сдвинутую задачу: %w", res.Error)
			}
			oldStart, newStart := sh.OriginalStart.String(), sh.NewStart.String()
			if err := recordHistory(tx, sh.ID, actorID, models.HistoryActionDateShifted, strPtr("startDate"), &oldStart, &newStart); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return Result{}, err
	}
	return result, nil
}

// resolveShift переводит ShiftInput в конкретную новую дату начала и величину
// сдвига в днях. shiftDays положительный/отрицательный задаёт сдвиг напрямую;
// newStartDate — то же самое, но через абсолютную дату.
func resolveShift(task models.Task, in ShiftInput) (models.Date, int, error) {
	switch {
	case in.NewStartDate != nil:
		return *in.NewStartDate, task.StartDate.DaysUntil(*in.NewStartDate), nil
	case in.ShiftDays != nil:
		return task.StartDate.AddDays(*in.ShiftDays), *in.ShiftDays, nil
	default:
		return models.Date{}, 0, Invalid("укажите newStartDate или shiftDays")
	}
}

// dependencyTypeFor — тип связи, через которую сдвиг дошёл до задачи sh:
// тип входящего ребра от её предшественника, который тоже сдвинулся. Для
// самого источника берём тип исходящего ребра к первому затронутому
// последователю — полю в ответе всё равно нужно какое-то значение (оно не
// nullable в спецификации), а источник сдвинулся не "из-за" связи, а напрямую.
func dependencyTypeFor(sh schedule.ShiftedTask, g *schedule.Graph, sourceID uuid.UUID) models.DependencyType {
	if sh.ID != sourceID {
		if sh.CausedBy != "" {
			return sh.CausedBy
		}
	}
	if out := g.Out(sourceID); len(out) > 0 {
		return out[0].Type
	}
	return models.DependencyFS
}
