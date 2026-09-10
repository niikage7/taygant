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

	// RiskDays — не хранится: пересчитывается при каждом чтении из PlannedDate
	// и текущей даты, как и Status. gorm:"-" исключает поле из миграции и запросов.
	RiskDays *int `gorm:"-"`

	// TasksTotal/TasksDone — сколько задач ведёт к вехе и сколько из них
	// завершено. Не хранится: считается сервисом по Task.MilestoneID при
	// каждом чтении списка вех, как и RiskDays.
	TasksTotal int `gorm:"-"`
	TasksDone  int `gorm:"-"`
}

// TableName фиксирует имя таблицы.
func (Milestone) TableName() string { return "milestones" }
