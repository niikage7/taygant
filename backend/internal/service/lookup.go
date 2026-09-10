package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"taygant_backend/internal/models"
)

// Вложенные ресурсы (задача, спринт, веха, ...) адресуются собственным ID без
// projectId в пути — так устроена спецификация. Проверка доступа всё равно нужна
// на уровне проекта, поэтому прежде чем звать access.Load, сервисы находят его
// этими функциями.

func projectIDOfTask(ctx context.Context, db *gorm.DB, taskID uuid.UUID) (uuid.UUID, error) {
	var task models.Task
	err := db.WithContext(ctx).Select("id", "project_id").First(&task, "id = ?", taskID).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return uuid.Nil, ErrNotFound
	case err != nil:
		return uuid.Nil, fmt.Errorf("найти задачу: %w", err)
	}
	return task.ProjectID, nil
}

func projectIDOfSprint(ctx context.Context, db *gorm.DB, sprintID uuid.UUID) (uuid.UUID, error) {
	var sprint models.Sprint
	err := db.WithContext(ctx).Select("id", "project_id").First(&sprint, "id = ?", sprintID).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return uuid.Nil, ErrNotFound
	case err != nil:
		return uuid.Nil, fmt.Errorf("найти спринт: %w", err)
	}
	return sprint.ProjectID, nil
}

func projectIDOfMilestone(ctx context.Context, db *gorm.DB, milestoneID uuid.UUID) (uuid.UUID, error) {
	var milestone models.Milestone
	err := db.WithContext(ctx).Select("id", "project_id").First(&milestone, "id = ?", milestoneID).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return uuid.Nil, ErrNotFound
	case err != nil:
		return uuid.Nil, fmt.Errorf("найти веху: %w", err)
	}
	return milestone.ProjectID, nil
}

func projectIDOfChecklistItem(ctx context.Context, db *gorm.DB, itemID uuid.UUID) (uuid.UUID, uuid.UUID, error) {
	var item models.ChecklistItem
	err := db.WithContext(ctx).Select("id", "task_id").First(&item, "id = ?", itemID).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return uuid.Nil, uuid.Nil, ErrNotFound
	case err != nil:
		return uuid.Nil, uuid.Nil, fmt.Errorf("найти пункт чек-листа: %w", err)
	}
	projectID, err := projectIDOfTask(ctx, db, item.TaskID)
	return projectID, item.TaskID, err
}

// dependenciesForProject загружает все связи между задачами проекта одним запросом.
// Используется и для отображения (список/детали задачи), и для проверки циклов
// при добавлении новой связи — в обоих случаях нужен весь граф проекта, а не
// связи одной задачи.
func dependenciesForProject(ctx context.Context, db *gorm.DB, projectID uuid.UUID) ([]models.TaskDependency, error) {
	var deps []models.TaskDependency
	err := db.WithContext(ctx).
		Joins("JOIN tasks ON tasks.id = task_dependencies.successor_task_id").
		Where("tasks.project_id = ?", projectID).
		Find(&deps).Error
	if err != nil {
		return nil, fmt.Errorf("выбрать связи проекта: %w", err)
	}
	return deps, nil
}

func projectIDOfDependency(ctx context.Context, db *gorm.DB, dependencyID uuid.UUID) (uuid.UUID, models.TaskDependency, error) {
	var dep models.TaskDependency
	err := db.WithContext(ctx).First(&dep, "id = ?", dependencyID).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return uuid.Nil, models.TaskDependency{}, ErrNotFound
	case err != nil:
		return uuid.Nil, models.TaskDependency{}, fmt.Errorf("найти связь: %w", err)
	}
	projectID, err := projectIDOfTask(ctx, db, dep.PredecessorTaskID)
	return projectID, dep, err
}
