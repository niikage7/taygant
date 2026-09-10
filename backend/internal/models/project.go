package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Project — проект: контейнер задач, спринтов и вех.
//
// Здесь только то, что задаёт пользователь. Производные показатели
// (progressPercent, healthIndex, durationCalendarDays, criticalRisksCount)
// в таблице не хранятся: они считаются из задач и мгновенно устаревали бы
// после любой правки сроков.
type Project struct {
	Model

	// Code — короткий читаемый код вида TPU-2026-A1B2, генерируется при создании.
	Code        string `gorm:"type:varchar(64);not null;index"`
	Name        string `gorm:"type:varchar(120);not null"`
	Description string `gorm:"type:text"`
	CustomerOrg string `gorm:"type:varchar(255)"`

	Status ProjectStatus `gorm:"type:varchar(20);not null;index"`
	// Phase — текущая фаза реализации, произвольный текст («Фаза 3: Внедрение»).
	Phase string `gorm:"type:varchar(255)"`

	StartDate Date `gorm:"type:date;not null"`
	Deadline  Date `gorm:"type:date;not null"`

	// Настройки расчёта графика — влияют на пересчёт сроков и отрисовку Ганта.
	WorkingCalendarType       WorkingCalendarType `gorm:"type:varchar(10);not null"`
	IncludePublicHolidays     bool                `gorm:"not null;default:true"`
	AutoRecalculateDependents bool                `gorm:"not null;default:true"`
	HighlightCriticalPath     bool                `gorm:"not null;default:true"`

	CreatedByID uuid.UUID `gorm:"type:uuid;not null;index"`
	CreatedBy   *User     `gorm:"foreignKey:CreatedByID"`

	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// TableName фиксирует имя таблицы.
func (Project) TableName() string { return "projects" }

// ProjectMember — участник команды проекта.
//
// ProjectRole и AccessLevel разделены сознательно: роль отвечает на вопрос
// «чем человек занимается», уровень доступа — «что ему разрешено». Аналитик
// может иметь полный доступ, а руководитель со стороны заказчика — только просмотр.
type ProjectMember struct {
	Model

	// Пара (ProjectID, UserID) уникальна: одного человека нельзя добавить
	// в проект дважды. Ограничение стоит в БД, потому что проверка в сервисе
	// не переживает параллельных запросов.
	ProjectID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_project_members_project_user"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_project_members_project_user"`

	Project *Project `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE"`
	User    *User    `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`

	ProjectRole ProjectRole `gorm:"type:varchar(30);not null"`
	// SprintRole — роль в рамках спринтов, свободный текст («Разработчик»).
	SprintRole  string      `gorm:"type:varchar(120)"`
	AccessLevel AccessLevel `gorm:"type:varchar(10);not null"`

	// WeeklyHoursLimit — лимит часов в неделю, база для расчёта загрузки в /workload.
	WeeklyHoursLimit int `gorm:"not null;default:40"`
}

// TableName фиксирует имя таблицы.
func (ProjectMember) TableName() string { return "project_members" }
