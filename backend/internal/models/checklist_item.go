package models

import "github.com/google/uuid"

// ChecklistItem — пункт чек-листа критериев приёмки (Definition of Done) задачи.
type ChecklistItem struct {
	Model

	TaskID uuid.UUID `gorm:"type:uuid;not null;index"`
	Task   *Task     `gorm:"foreignKey:TaskID;constraint:OnDelete:CASCADE"`

	Text   string `gorm:"type:text;not null"`
	IsDone bool   `gorm:"not null;default:false"`

	// OrderIndex сохраняет порядок пунктов: чек-лист читают сверху вниз,
	// и перестановка при каждой выборке сбивала бы с толку.
	OrderIndex int `gorm:"not null;default:0"`
}

// TableName фиксирует имя таблицы.
func (ChecklistItem) TableName() string { return "checklist_items" }
