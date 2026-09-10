package models

import "github.com/google/uuid"

// Коды действий в журнале изменений. Фронт различает записи по этому полю,
// поэтому набор кодов — часть контракта, а не свободный текст.
const (
	HistoryActionCreated       = "created"
	HistoryActionUpdated       = "updated"
	HistoryActionStatusChanged = "status_changed"
	HistoryActionAssigned      = "assigned"
	HistoryActionDateShifted   = "date_shifted"
	HistoryActionProgressSet   = "progress_set"
	HistoryActionDependencyAdd = "dependency_added"
	HistoryActionDependencyDel = "dependency_removed"
	HistoryActionCommented     = "commented"
)

// TaskHistoryEntry — запись аудита по задаче.
//
// Хранит одно изменение одного поля: так журнал читается как список
// «что, кто, когда, с чего на что» без разбора вложенных структур.
// Действие, не связанное с конкретным полем (создание задачи), оставляет
// Field, OldValue и NewValue пустыми.
type TaskHistoryEntry struct {
	Model

	TaskID uuid.UUID `gorm:"type:uuid;not null;index"`
	Task   *Task     `gorm:"foreignKey:TaskID;constraint:OnDelete:CASCADE"`

	// ActorID не ссылается на users каскадом: удаление пользователя не должно
	// стирать историю его действий.
	ActorID uuid.UUID `gorm:"type:uuid;not null;index"`
	Actor   *User     `gorm:"foreignKey:ActorID"`

	Action   string  `gorm:"type:varchar(40);not null"`
	Field    *string `gorm:"type:varchar(60)"`
	OldValue *string `gorm:"type:text"`
	NewValue *string `gorm:"type:text"`
}

// TableName фиксирует имя таблицы.
func (TaskHistoryEntry) TableName() string { return "task_history" }
