package models

import "github.com/google/uuid"

// Milestone — контрольная точка проекта: значимое событие с плановой датой.
//
// Вехи существуют отдельно от задач с флагом IsMilestone: первые описывают
// обязательства проекта перед заказчиком, вторые — нулевые по длительности
// элементы графика. На диаграмме Ганта показываются и те, и другие.
type Milestone struct {
	Model

	ProjectID uuid.UUID `gorm:"type:uuid;not null;index"`
	Project   *Project  `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE"`

	// Code — короткий код вехи («КТ-2»).
	Code string `gorm:"type:varchar(32)"`
	Name string `gorm:"type:varchar(255);not null"`

	PlannedDate Date `gorm:"type:date;not null"`
	// ActualDate заполняется по факту достижения; до этого момента nil.
	ActualDate *Date `gorm:"type:date"`

	Status MilestoneStatus `gorm:"type:varchar(20);not null"`
}

// TableName фиксирует имя таблицы.
func (Milestone) TableName() string { return "milestones" }
