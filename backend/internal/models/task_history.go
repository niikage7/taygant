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

// Каналы, через которые сделано изменение. Тоже часть контракта с фронтом:
// по ним карточка задачи показывает пометку «через ассистента».
const (
	// HistorySourceApp — правка в веб-приложении.
	HistorySourceApp = "app"
	// HistorySourceAssistant — правка нейронкой пользователя через MCP,
	// с его подтверждением. Доля таких смен статуса — метрика подключения.
	HistorySourceAssistant = "assistant"
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

	// Source — канал изменения (HistorySource*). Значение по умолчанию в БД
	// нужно уже существующим записям: при миграции колонка добавляется к
	// заполненной таблице, и все старые правки сделаны в приложении.
	Source string `gorm:"type:varchar(20);not null;default:app"`
	// Via — чем именно воспользовались: для ассистента это имя личного ключа.
	// Хранится строкой, а не ссылкой на ключ: отзыв ключа не должен
	// обезличивать уже сделанные им записи.
	Via string `gorm:"type:varchar(120);not null;default:''"`
}

// TableName фиксирует имя таблицы.
func (TaskHistoryEntry) TableName() string { return "task_history" }
