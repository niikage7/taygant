package models

import "github.com/google/uuid"

// TaskComment — комментарий участника команды к задаче.
type TaskComment struct {
	Model

	TaskID uuid.UUID `gorm:"type:uuid;not null;index"`
	Task   *Task     `gorm:"foreignKey:TaskID;constraint:OnDelete:CASCADE"`

	AuthorID uuid.UUID `gorm:"type:uuid;not null;index"`
	Author   *User     `gorm:"foreignKey:AuthorID;constraint:OnDelete:CASCADE"`

	Text string `gorm:"type:text;not null"`
}

// TableName фиксирует имя таблицы.
func (TaskComment) TableName() string { return "task_comments" }
