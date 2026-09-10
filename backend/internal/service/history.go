package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"taygant_backend/internal/models"
)

// History — журнал изменений задач.
type History struct {
	db *gorm.DB
}

// NewHistory собирает сервис истории.
func NewHistory(db *gorm.DB) *History {
	return &History{db: db}
}

// List возвращает журнал задачи в хронологическом порядке.
func (s *History) List(ctx context.Context, taskID uuid.UUID) ([]models.TaskHistoryEntry, error) {
	var entries []models.TaskHistoryEntry
	err := s.db.WithContext(ctx).
		Preload("Actor").
		Where("task_id = ?", taskID).
		Order("created_at").
		Find(&entries).Error
	if err != nil {
		return nil, fmt.Errorf("выбрать историю задачи: %w", err)
	}
	return entries, nil
}

// recordHistory добавляет запись в журнал задачи в рамках переданной транзакции.
//
// Общая функция для всех сервисов, меняющих задачи (tasks, dependencies, checklist,
// comments): без неё правило «одно изменение — одна запись» пришлось бы повторять
// в каждом месте, где меняется задача, и они неизбежно разошлись бы.
func recordHistory(tx *gorm.DB, taskID, actorID uuid.UUID, action string, field, oldValue, newValue *string) error {
	entry := models.TaskHistoryEntry{
		TaskID:   taskID,
		ActorID:  actorID,
		Action:   action,
		Field:    field,
		OldValue: oldValue,
		NewValue: newValue,
	}
	return tx.Create(&entry).Error
}

// strPtr — удобный конструктор *string для литералов в местах вызова recordHistory.
func strPtr(s string) *string { return &s }
