package models

import "github.com/google/uuid"

// Sprint — итерация или релиз внутри проекта, используется для группировки задач
// и расчёта загрузки участников.
type Sprint struct {
	Model

	ProjectID uuid.UUID `gorm:"type:uuid;not null;index"`
	Project   *Project  `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE"`

	Name string `gorm:"type:varchar(255);not null"`
	// ReleaseVersion — версия релиза, связанная со спринтом («2.1»).
	ReleaseVersion string `gorm:"type:varchar(50)"`

	StartDate Date `gorm:"type:date;not null"`
	EndDate   Date `gorm:"type:date;not null"`
}

// TableName фиксирует имя таблицы.
func (Sprint) TableName() string { return "sprints" }
