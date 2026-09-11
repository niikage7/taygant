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

// Checklist — критерии приёмки задачи (Definition of Done).
type Checklist struct {
	db *gorm.DB
}

// NewChecklist собирает сервис чек-листа.
func NewChecklist(db *gorm.DB) *Checklist {
	return &Checklist{db: db}
}

// ProjectID возвращает проект и задачу, которым принадлежит пункт чек-листа.
func (s *Checklist) ProjectID(ctx context.Context, itemID uuid.UUID) (projectID, taskID uuid.UUID, err error) {
	return projectIDOfChecklistItem(ctx, s.db, itemID)
}

// List возвращает пункты чек-листа задачи по порядку добавления.
func (s *Checklist) List(ctx context.Context, taskID uuid.UUID) ([]models.ChecklistItem, error) {
	var items []models.ChecklistItem
	err := s.db.WithContext(ctx).Where("task_id = ?", taskID).Order("order_index").Find(&items).Error
	if err != nil {
		return nil, fmt.Errorf("выбрать чек-лист: %w", err)
	}
	return items, nil
}

// Add добавляет пункт в чек-лист задачи.
func (s *Checklist) Add(ctx context.Context, taskID, actorID uuid.UUID, text string) (models.ChecklistItem, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return models.ChecklistItem{}, Invalid("укажите текст критерия приёмки")
	}

	var item models.ChecklistItem
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Максимум существующего OrderIndex, а не количество пунктов: удаление
		// пункта не должно освобождать индекс, всё ещё занятый другим пунктом
		// (та же ошибка, что и в nextTaskCode/nextWBSNumber).
		var maxOrder *int
		if err := tx.Model(&models.ChecklistItem{}).Where("task_id = ?", taskID).Select("MAX(order_index)").Scan(&maxOrder).Error; err != nil {
			return fmt.Errorf("посчитать пункты чек-листа: %w", err)
		}
		nextOrder := 0
		if maxOrder != nil {
			nextOrder = *maxOrder + 1
		}
		item = models.ChecklistItem{TaskID: taskID, Text: text, OrderIndex: nextOrder}
		if err := tx.Create(&item).Error; err != nil {
			return fmt.Errorf("создать пункт чек-листа: %w", err)
		}
		return recordHistory(tx, taskID, actorID, models.HistoryActionUpdated, strPtr("checklist"), nil, strPtr(text))
	})
	if err != nil {
		return models.ChecklistItem{}, err
	}
	return item, nil
}

// UpdateInput — поля пункта чек-листа, доступные для частичного обновления.
type ChecklistUpdateInput struct {
	Text   *string
	IsDone *bool
}

// Update меняет текст и/или отметку выполнения пункта.
func (s *Checklist) Update(ctx context.Context, itemID, actorID uuid.UUID, in ChecklistUpdateInput) (models.ChecklistItem, error) {
	var item models.ChecklistItem
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&item, "id = ?", itemID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNotFound
			}
			return fmt.Errorf("найти пункт чек-листа: %w", err)
		}
		if in.Text != nil {
			text := strings.TrimSpace(*in.Text)
			if text == "" {
				return Invalid("текст критерия приёмки не может быть пустым")
			}
			item.Text = text
		}
		if in.IsDone != nil && item.IsDone != *in.IsDone {
			item.IsDone = *in.IsDone
			action := "выполнен"
			if !item.IsDone {
				action = "снят с выполнения"
			}
			if err := recordHistory(tx, item.TaskID, actorID, models.HistoryActionUpdated, strPtr("checklist"), nil, strPtr(item.Text+": "+action)); err != nil {
				return err
			}
		}
		if err := tx.Save(&item).Error; err != nil {
			return fmt.Errorf("сохранить пункт чек-листа: %w", err)
		}
		return nil
	})
	if err != nil {
		return models.ChecklistItem{}, err
	}
	return item, nil
}

// Delete удаляет пункт чек-листа.
func (s *Checklist) Delete(ctx context.Context, itemID uuid.UUID) error {
	res := s.db.WithContext(ctx).Delete(&models.ChecklistItem{}, "id = ?", itemID)
	if res.Error != nil {
		return fmt.Errorf("удалить пункт чек-листа: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
