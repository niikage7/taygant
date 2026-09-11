package models

// Перечисления повторяют enum'ы из api-spec.yml дословно: значения уезжают
// в JSON как есть и приходят оттуда же, поэтому переименовывать их нельзя.
//
// Валидность проверяется в Go, а не CHECK-ограничением в БД: набор значений
// меняется вместе со спецификацией, и держать его в одном месте дешевле,
// чем догонять миграциями.

// ProjectStatus — жизненный цикл проекта.
type ProjectStatus string

const (
	ProjectStatusDraft     ProjectStatus = "draft"     // черновик, не запущен
	ProjectStatusActive    ProjectStatus = "active"    // выполняется
	ProjectStatusOnHold    ProjectStatus = "on_hold"   // приостановлен
	ProjectStatusCompleted ProjectStatus = "completed" // завершён
	ProjectStatusArchived  ProjectStatus = "archived"  // архивирован (удалён логически)
)

// Valid сообщает, что значение входит в перечисление спецификации.
func (s ProjectStatus) Valid() bool {
	switch s {
	case ProjectStatusDraft, ProjectStatusActive, ProjectStatusOnHold,
		ProjectStatusCompleted, ProjectStatusArchived:
		return true
	}
	return false
}

// ProjectRole — роль участника в проекте.
//
// Роль описывает, чем человек занимается, и не даёт прав сама по себе:
// доступ определяется отдельным полем AccessLevel.
type ProjectRole string

const (
	ProjectRoleProjectManager ProjectRole = "project_manager"
	ProjectRoleAnalyst        ProjectRole = "analyst"
	ProjectRoleDeveloper      ProjectRole = "developer"
	ProjectRoleCurator        ProjectRole = "curator"
	ProjectRoleOther          ProjectRole = "other"
)

// Valid сообщает, что значение входит в перечисление спецификации.
func (r ProjectRole) Valid() bool {
	switch r {
	case ProjectRoleProjectManager, ProjectRoleAnalyst, ProjectRoleDeveloper,
		ProjectRoleCurator, ProjectRoleOther:
		return true
	}
	return false
}

// AccessLevel — уровень доступа участника к проекту. Именно от него,
// а не от ProjectRole, вычисляются права в internal/access.
type AccessLevel string

const (
	AccessLevelFull AccessLevel = "full" // управление проектом и командой
	AccessLevelEdit AccessLevel = "edit" // редактирование задач
	AccessLevelView AccessLevel = "view" // только просмотр
)

// Valid сообщает, что значение входит в перечисление спецификации.
func (l AccessLevel) Valid() bool {
	switch l {
	case AccessLevelFull, AccessLevelEdit, AccessLevelView:
		return true
	}
	return false
}

// TaskStatus — состояние задачи.
//
// overdue и blocked система выставляет сама (просрочка по датам и незакрытый
// предшественник), поэтому пользовательские переходы в них не принимаются.
type TaskStatus string

const (
	TaskStatusPlanned    TaskStatus = "planned"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusDone       TaskStatus = "done"
	TaskStatusOverdue    TaskStatus = "overdue"
	TaskStatusBlocked    TaskStatus = "blocked"
)

// Valid сообщает, что значение входит в перечисление спецификации.
func (s TaskStatus) Valid() bool {
	switch s {
	case TaskStatusPlanned, TaskStatusInProgress, TaskStatusDone,
		TaskStatusOverdue, TaskStatusBlocked:
		return true
	}
	return false
}

// DependencyType — тип связи между задачами по методу критического пути (CPM).
type DependencyType string

const (
	// DependencyFS — Finish-to-Start: последователь стартует после финиша предшественника.
	DependencyFS DependencyType = "FS"
	// DependencySS — Start-to-Start: последователь стартует не раньше старта предшественника.
	DependencySS DependencyType = "SS"
	// DependencyFF — Finish-to-Finish: последователь финиширует не раньше финиша предшественника.
	DependencyFF DependencyType = "FF"
	// DependencySF — Start-to-Finish: последователь финиширует не раньше старта предшественника.
	DependencySF DependencyType = "SF"
)

// Valid сообщает, что значение входит в перечисление спецификации.
func (t DependencyType) Valid() bool {
	switch t {
	case DependencyFS, DependencySS, DependencyFF, DependencySF:
		return true
	}
	return false
}

// WorkingCalendarType — рабочий календарь проекта.
type WorkingCalendarType string

const (
	Calendar52 WorkingCalendarType = "5/2" // пятидневка, выходные — сб и вс
	Calendar61 WorkingCalendarType = "6/1" // шестидневка, выходной — вс
)

// Valid сообщает, что значение входит в перечисление спецификации.
func (c WorkingCalendarType) Valid() bool {
	switch c {
	case Calendar52, Calendar61:
		return true
	}
	return false
}

// MilestoneStatus — состояние контрольной точки.
type MilestoneStatus string

const (
	MilestoneStatusDone    MilestoneStatus = "done"
	MilestoneStatusCurrent MilestoneStatus = "current"
	MilestoneStatusPlanned MilestoneStatus = "planned"
	MilestoneStatusFinal   MilestoneStatus = "final"
)

// Valid сообщает, что значение входит в перечисление спецификации.
func (s MilestoneStatus) Valid() bool {
	switch s {
	case MilestoneStatusDone, MilestoneStatusCurrent, MilestoneStatusPlanned, MilestoneStatusFinal:
		return true
	}
	return false
}
