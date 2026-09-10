package models

import "github.com/google/uuid"

// Task — задача проекта, основной элемент реестра и диаграммы Ганта.
//
// В таблице лежит только то, что задал пользователь или пересчитал сервис сдвига.
// Производные величины — длительность, признак критического пути, отклонение
// план/факт — считаются на лету в internal/schedule: хранить их значило бы
// поддерживать их в согласованном виде после каждой правки любой связанной задачи.
type Task struct {
	Model

	ProjectID uuid.UUID `gorm:"type:uuid;not null;index"`
	Project   *Project  `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE"`

	// SprintID — привязка к спринту, необязательная.
	SprintID *uuid.UUID `gorm:"type:uuid;index"`
	Sprint   *Sprint    `gorm:"foreignKey:SprintID;constraint:OnDelete:SET NULL"`

	// ParentTaskID — родитель в иерархии работ (WBS). Ссылка на ту же таблицу:
	// удаление родительской задачи уносит её подзадачи.
	ParentTaskID *uuid.UUID `gorm:"type:uuid;index"`
	ParentTask   *Task      `gorm:"foreignKey:ParentTaskID;constraint:OnDelete:CASCADE"`

	// Code — читаемый код задачи («TASK-004»), уникален в пределах проекта.
	Code string `gorm:"type:varchar(32);not null;index"`
	// WBSNumber — номер в иерархической структуре работ («4.2»).
	WBSNumber string `gorm:"type:varchar(32)"`
	// OrderIndex задаёт стабильный порядок задач в реестре и на диаграмме:
	// без него postgres возвращает строки в произвольном порядке и Гант «прыгает».
	OrderIndex int `gorm:"not null;default:0"`

	Title       string `gorm:"type:varchar(255);not null"`
	Description string `gorm:"type:text"`

	AssigneeID *uuid.UUID `gorm:"type:uuid;index"`
	Assignee   *User      `gorm:"foreignKey:AssigneeID;constraint:OnDelete:SET NULL"`

	Status TaskStatus `gorm:"type:varchar(20);not null;index"`
	// IsMilestone — веха на диаграмме: точка нулевой длительности, StartDate == EndDate.
	IsMilestone bool `gorm:"not null;default:false"`

	StartDate Date `gorm:"type:date;not null"`
	EndDate   Date `gorm:"type:date;not null"`

	// Фактические даты заполняются при переходе задачи в работу и при закрытии.
	// Из них считается planVsActualDeviationDays для режима план/факт.
	ActualStartDate *Date `gorm:"type:date"`
	ActualEndDate   *Date `gorm:"type:date"`

	ProgressPercent int     `gorm:"not null;default:0"`
	WeightPercent   float64 `gorm:"not null;default:0"`

	// Связь с внешним трекером хранится плоскими колонками: вложенный объект
	// в ответе собирает DTO, а отдельная таблица ради четырёх полей избыточна.
	ExternalProvider    string `gorm:"type:varchar(50)"`
	ExternalURL         string `gorm:"type:text"`
	ExternalReferenceID string `gorm:"type:varchar(64)"`
	ExternalSynced      bool   `gorm:"not null;default:false"`

	// Ниже — производные поля из комментария выше. gorm:"-" исключает их из
	// миграции и запросов: сервис задач (internal/service, CPM-движок в
	// internal/schedule) заполняет их при чтении, а не хранит.
	//
	// BufferDays раньше был обычной хранимой колонкой со значением по
	// умолчанию 0; она осталась в БД неиспользуемой (AutoMigrate не удаляет
	// колонки) — это ожидаемо и безвредно, не полагаться на её содержимое.
	DurationCalendarDays int  `gorm:"-"`
	DurationWorkingDays  int  `gorm:"-"`
	IsCriticalPath       bool `gorm:"-"`
	BufferDays           int  `gorm:"-"`
}

// TableName фиксирует имя таблицы.
func (Task) TableName() string { return "tasks" }

// PlanVsActualDeviation — отклонение факта от плана в днях; положительное
// значение — опережение графика, отрицательное — отставание.
//
// Полноценный расчёт (earned value) требует CPM-движка, которого пока нет,
// поэтому для задач без фактических дат используется упрощённая эвристика:
// просрочка планового конца без ActualEndDate уже читается как отставание.
func (t Task) PlanVsActualDeviation() int {
	if t.ActualEndDate != nil {
		return t.ActualEndDate.DaysUntil(t.EndDate)
	}
	if today := Today(); today.After(t.EndDate) {
		return -t.EndDate.DaysUntil(today)
	}
	return 0
}

// TaskDependency — связь между двумя задачами по методу критического пути.
//
// Направление задаётся парой полей, а не флагом: предшественник и последователь
// не взаимозаменяемы, и для типов SS/FF/SF порядок определяет смысл связи.
type TaskDependency struct {
	Model

	// Пара уникальна: дублирующие связи между теми же задачами запрещены.
	// Запрет самозависимости и циклов проверяет сервис — выразить их
	// ограничением таблицы нельзя.
	PredecessorTaskID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_task_dependencies_pair"`
	SuccessorTaskID   uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_task_dependencies_pair"`

	PredecessorTask *Task `gorm:"foreignKey:PredecessorTaskID;constraint:OnDelete:CASCADE"`
	SuccessorTask   *Task `gorm:"foreignKey:SuccessorTaskID;constraint:OnDelete:CASCADE"`

	Type DependencyType `gorm:"type:varchar(2);not null"`
	// LagDays — задержка между задачами; отрицательное значение означает опережение.
	LagDays int `gorm:"not null;default:0"`
}

// TableName фиксирует имя таблицы.
func (TaskDependency) TableName() string { return "task_dependencies" }
