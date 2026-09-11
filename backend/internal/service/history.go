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
//
// Канал изменения (приложение или ассистент) берётся из контекста транзакции —
// его туда кладёт WithChangeSource; сервисы открывают транзакции через
// db.WithContext(ctx), поэтому контекст запроса доезжает сюда без передачи руками.
func recordHistory(tx *gorm.DB, taskID, actorID uuid.UUID, action string, field, oldValue, newValue *string) error {
	src := changeSourceFrom(tx.Statement.Context)
	entry := models.TaskHistoryEntry{
		TaskID:   taskID,
		ActorID:  actorID,
		Action:   action,
		Field:    field,
		OldValue: oldValue,
		NewValue: newValue,
		Source:   src.Channel,
		Via:      src.Via,
	}
	return tx.Create(&entry).Error
}

// StatusChangeStats — сколько раз за период меняли статусы задач проекта и
// сколько из этих смен сделано через ассистента.
type StatusChangeStats struct {
	PeriodDays   int
	Total        int
	ViaAssistant int
}

// statusChangeStatsWindowDays — окно метрики. Берём последние 30 дней, а не всю
// историю: метрика показывает, пользуются ли подключением сейчас, и старые
// правки (до появления MCP) размывали бы её.
const statusChangeStatsWindowDays = 30

// statusChangeStats считает смены статуса в проекте за последние 30 дней.
// Это метрика из docs/mcp.md: если подключение к нейронкам помогает держать
// статусы актуальными, доля смен «через ассистента» будет расти.
func statusChangeStats(ctx context.Context, db *gorm.DB, projectID uuid.UUID) (StatusChangeStats, error) {
	since := models.Today().AddDays(-statusChangeStatsWindowDays).Time
	var row struct {
		Total        int
		ViaAssistant int
	}
	err := db.WithContext(ctx).
		Table("task_history").
		Select("COUNT(*) AS total, COUNT(*) FILTER (WHERE task_history.source = ?) AS via_assistant", models.HistorySourceAssistant).
		Joins("JOIN tasks ON tasks.id = task_history.task_id").
		Where("tasks.project_id = ? AND task_history.action = ? AND task_history.created_at >= ?",
			projectID, models.HistoryActionStatusChanged, since).
		Scan(&row).Error
	if err != nil {
		return StatusChangeStats{}, fmt.Errorf("посчитать смены статусов: %w", err)
	}
	return StatusChangeStats{
		PeriodDays:   statusChangeStatsWindowDays,
		Total:        row.Total,
		ViaAssistant: row.ViaAssistant,
	}, nil
}

// strPtr — удобный конструктор *string для литералов в местах вызова recordHistory.
func strPtr(s string) *string { return &s }
