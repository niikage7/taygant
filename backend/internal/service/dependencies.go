package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"taygant_backend/internal/models"
)

// Dependencies — связи (зависимости) между задачами.
type Dependencies struct {
	db *gorm.DB
}

// NewDependencies собирает сервис связей.
func NewDependencies(db *gorm.DB) *Dependencies {
	return &Dependencies{db: db}
}

// ProjectID возвращает проект, которому принадлежит связь.
func (s *Dependencies) ProjectID(ctx context.Context, dependencyID uuid.UUID) (uuid.UUID, error) {
	projectID, _, err := projectIDOfDependency(ctx, s.db, dependencyID)
	return projectID, err
}

// ListForProject возвращает все связи проекта (для диаграммы Ганта — там
// показывается весь граф сразу, а не связи одной задачи).
func (s *Dependencies) ListForProject(ctx context.Context, projectID uuid.UUID) ([]models.TaskDependency, error) {
	var deps []models.TaskDependency
	err := s.db.WithContext(ctx).
		Preload("PredecessorTask").
		Preload("SuccessorTask").
		Joins("JOIN tasks ON tasks.id = task_dependencies.successor_task_id").
		Where("tasks.project_id = ?", projectID).
		Find(&deps).Error
	if err != nil {
		return nil, fmt.Errorf("выбрать связи проекта: %w", err)
	}
	return deps, nil
}

// List возвращает предшественников и последователей задачи вместе с названиями
// связанных задач — так фронту не нужен отдельный запрос ради подписи на графе.
func (s *Dependencies) List(ctx context.Context, taskID uuid.UUID) (predecessors, successors []models.TaskDependency, err error) {
	err = s.db.WithContext(ctx).
		Preload("PredecessorTask").
		Where("successor_task_id = ?", taskID).
		Find(&predecessors).Error
	if err != nil {
		return nil, nil, fmt.Errorf("выбрать предшественников: %w", err)
	}

	err = s.db.WithContext(ctx).
		Preload("SuccessorTask").
		Where("predecessor_task_id = ?", taskID).
		Find(&successors).Error
	if err != nil {
		return nil, nil, fmt.Errorf("выбрать последователей: %w", err)
	}
	return predecessors, successors, nil
}

// Create связывает currentTaskID с relatedTaskID согласно direction: "predecessor"
// значит relatedTaskID предшествует текущей задаче, "successor" — наоборот.
func (s *Dependencies) Create(ctx context.Context, projectID, currentTaskID, relatedTaskID uuid.UUID, direction string, depType models.DependencyType, lagDays int) (models.TaskDependency, error) {
	predecessorID, successorID := currentTaskID, relatedTaskID
	if direction == "predecessor" {
		predecessorID, successorID = relatedTaskID, currentTaskID
	} else if direction != "successor" {
		return models.TaskDependency{}, Invalid("направление связи должно быть predecessor или successor")
	}

	var dep models.TaskDependency
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var err error
		dep, err = createDependencyRow(ctx, tx, projectID, predecessorID, successorID, depType, lagDays)
		return err
	})
	if err != nil {
		return models.TaskDependency{}, err
	}
	return s.get(ctx, dep.ID)
}

// createDependency — вариант createDependencyRow для случая, когда направление
// уже известно (создание задачи сразу со списком предшественников).
func createDependency(ctx context.Context, tx *gorm.DB, projectID, predecessorID, successorID uuid.UUID, depType models.DependencyType, lagDays int) error {
	_, err := createDependencyRow(ctx, tx, projectID, predecessorID, successorID, depType, lagDays)
	return err
}

func createDependencyRow(ctx context.Context, tx *gorm.DB, projectID, predecessorID, successorID uuid.UUID, depType models.DependencyType, lagDays int) (models.TaskDependency, error) {
	if predecessorID == successorID {
		return models.TaskDependency{}, Invalid("задача не может зависеть сама от себя")
	}
	if !depType.Valid() {
		return models.TaskDependency{}, Invalid("недопустимый тип связи %q", depType)
	}

	var count int64
	if err := tx.Model(&models.Task{}).Where("id IN ? AND project_id = ?", []uuid.UUID{predecessorID, successorID}, projectID).Count(&count).Error; err != nil {
		return models.TaskDependency{}, fmt.Errorf("проверить задачи связи: %w", err)
	}
	if count != 2 {
		return models.TaskDependency{}, Invalid("обе задачи должны принадлежать одному проекту")
	}

	edges, err := dependenciesForProject(ctx, tx, projectID)
	if err != nil {
		return models.TaskDependency{}, err
	}
	if wouldCreateCycle(edges, predecessorID, successorID) {
		return models.TaskDependency{}, Invalid("такая связь создала бы цикл зависимостей")
	}

	dep := models.TaskDependency{
		PredecessorTaskID: predecessorID,
		SuccessorTaskID:   successorID,
		Type:              depType,
		LagDays:           lagDays,
	}
	if err := tx.Create(&dep).Error; err != nil {
		if isUniqueViolation(err) {
			return models.TaskDependency{}, fmt.Errorf("%w: такая связь уже существует", ErrConflict)
		}
		return models.TaskDependency{}, fmt.Errorf("создать связь: %w", err)
	}
	return dep, nil
}

// wouldCreateCycle сообщает, есть ли уже путь successor → ... → predecessor.
// Если есть, ребро predecessor → successor замкнёт граф в цикл: последователь
// оказался бы одновременно предком своего предшественника.
func wouldCreateCycle(edges []models.TaskDependency, predecessorID, successorID uuid.UUID) bool {
	adjacency := make(map[uuid.UUID][]uuid.UUID, len(edges))
	for _, e := range edges {
		adjacency[e.PredecessorTaskID] = append(adjacency[e.PredecessorTaskID], e.SuccessorTaskID)
	}

	visited := map[uuid.UUID]bool{successorID: true}
	queue := []uuid.UUID{successorID}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if current == predecessorID {
			return true
		}
		for _, next := range adjacency[current] {
			if !visited[next] {
				visited[next] = true
				queue = append(queue, next)
			}
		}
	}
	return false
}

// Delete разрывает связь между задачами.
func (s *Dependencies) Delete(ctx context.Context, dependencyID uuid.UUID) error {
	res := s.db.WithContext(ctx).Delete(&models.TaskDependency{}, "id = ?", dependencyID)
	if res.Error != nil {
		return fmt.Errorf("удалить связь: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Dependencies) get(ctx context.Context, dependencyID uuid.UUID) (models.TaskDependency, error) {
	var dep models.TaskDependency
	err := s.db.WithContext(ctx).
		Preload("PredecessorTask").
		Preload("SuccessorTask").
		First(&dep, "id = ?", dependencyID).Error
	if err != nil {
		return models.TaskDependency{}, fmt.Errorf("найти связь: %w", err)
	}
	return dep, nil
}
