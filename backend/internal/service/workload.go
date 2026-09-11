package service

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"taygant_backend/internal/models"
)

// Workload — загрузка участников команды часами относительно недельного лимита.
type Workload struct {
	db *gorm.DB
}

// NewWorkload собирает сервис загрузки команды.
func NewWorkload(db *gorm.DB) *Workload {
	return &Workload{db: db}
}

// MemberWorkload — загрузка одного участника в пределах выбранного периода
// (спринта или, если ни один спринт не выбран и не идёт сейчас, всего проекта).
type MemberWorkload struct {
	Member               models.ProjectMember
	SprintID             *uuid.UUID
	AssignedHoursPerWeek float64
	UtilizationPercent   int
}

// hoursPerWorkingDay — предполагаемая полная загрузка рабочего дня.
//
// В схеме Task нет оценки трудозатрат (часов на задачу) — только даты начала
// и конца, см. комментарий к models.Task. Без неё единственная считаемая из
// данных проекта эвристика — предположить, что на каждый активный (не done)
// назначенный день задачи исполнитель тратит полный рабочий день. Если у
// участника несколько параллельных задач, их часы складываются — это и даёт
// utilizationPercent выше 100%: реальный сигнал о том, что человек назначен
// на пересекающиеся по срокам задачи и не может быть полностью занят обеими.
const hoursPerWorkingDay = 8.0

// List считает загрузку всех участников проекта за период. sprintID задаёт
// период явно; nil означает «текущий спринт» (даты которого содержат сегодня),
// а если такого спринта нет — весь срок проекта (StartDate..Deadline).
func (s *Workload) List(ctx context.Context, projectID uuid.UUID, sprintID *uuid.UUID) ([]MemberWorkload, error) {
	var project models.Project
	if err := s.db.WithContext(ctx).First(&project, "id = ?", projectID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("найти проект: %w", err)
	}

	scopeStart, scopeEnd, resolvedSprintID, err := s.resolveScope(ctx, projectID, project, sprintID)
	if err != nil {
		return nil, err
	}

	var members []models.ProjectMember
	if err := s.db.WithContext(ctx).Preload("User").Where("project_id = ?", projectID).Order("created_at").Find(&members).Error; err != nil {
		return nil, fmt.Errorf("выбрать участников проекта: %w", err)
	}

	var tasks []models.Task
	err = s.db.WithContext(ctx).
		Select("id", "assignee_id", "start_date", "end_date").
		Where("project_id = ? AND is_milestone = false AND status <> ? AND assignee_id IS NOT NULL", projectID, models.TaskStatusDone).
		Where("start_date <= ? AND end_date >= ?", scopeEnd, scopeStart).
		Find(&tasks).Error
	if err != nil {
		return nil, fmt.Errorf("выбрать активные задачи для загрузки: %w", err)
	}

	workingDaysByAssignee := make(map[uuid.UUID]int, len(members))
	for _, t := range tasks {
		overlapStart := models.MaxDate(t.StartDate, scopeStart)
		overlapEnd := models.MinDate(t.EndDate, scopeEnd)
		if overlapEnd.Before(overlapStart) {
			continue
		}
		workingDaysByAssignee[*t.AssigneeID] += workingDaysBetween(overlapStart, overlapEnd, project.WorkingCalendarType)
	}

	weeksInScope := float64(scopeStart.DaysUntil(scopeEnd)+1) / 7.0

	result := make([]MemberWorkload, 0, len(members))
	for _, m := range members {
		hours := float64(workingDaysByAssignee[m.UserID]) * hoursPerWorkingDay / weeksInScope
		utilization := 0
		if m.WeeklyHoursLimit > 0 {
			utilization = int(math.Round(hours / float64(m.WeeklyHoursLimit) * 100))
		}
		result = append(result, MemberWorkload{
			Member:               m,
			SprintID:             resolvedSprintID,
			AssignedHoursPerWeek: math.Round(hours*10) / 10,
			UtilizationPercent:   utilization,
		})
	}
	return result, nil
}

// resolveScope находит границы периода расчёта: явно переданный спринт (404,
// если он не найден в этом проекте), иначе спринт, чьи даты содержат сегодня,
// иначе — весь срок проекта.
func (s *Workload) resolveScope(ctx context.Context, projectID uuid.UUID, project models.Project, sprintID *uuid.UUID) (start, end models.Date, resolved *uuid.UUID, err error) {
	if sprintID != nil {
		var sprint models.Sprint
		err := s.db.WithContext(ctx).First(&sprint, "id = ? AND project_id = ?", *sprintID, projectID).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Date{}, models.Date{}, nil, ErrNotFound
		}
		if err != nil {
			return models.Date{}, models.Date{}, nil, fmt.Errorf("найти спринт: %w", err)
		}
		return sprint.StartDate, sprint.EndDate, &sprint.ID, nil
	}

	today := models.Today()
	var current models.Sprint
	err = s.db.WithContext(ctx).
		Where("project_id = ? AND start_date <= ? AND end_date >= ?", projectID, today, today).
		Order("start_date").
		First(&current).Error
	switch {
	case err == nil:
		return current.StartDate, current.EndDate, &current.ID, nil
	case errors.Is(err, gorm.ErrRecordNotFound):
		return project.StartDate, project.Deadline, nil, nil
	default:
		return models.Date{}, models.Date{}, nil, fmt.Errorf("найти текущий спринт: %w", err)
	}
}
