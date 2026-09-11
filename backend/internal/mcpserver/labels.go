package mcpserver

import (
	"fmt"

	"taygant_backend/internal/models"
)

// Подписи совпадают с теми, что человек видит в приложении
// (frontend/src/lib/task-status.ts): ассистент говорит на языке диаграммы.

var taskStatusLabels = map[models.TaskStatus]string{
	models.TaskStatusPlanned:    "План",
	models.TaskStatusInProgress: "В работе",
	models.TaskStatusDone:       "Готово",
	models.TaskStatusOverdue:    "Просрочена",
	models.TaskStatusBlocked:    "Заблокирована",
}

func taskStatusLabel(s models.TaskStatus) string {
	if label, ok := taskStatusLabels[s]; ok {
		return label
	}
	return string(s)
}

var projectStatusLabels = map[models.ProjectStatus]string{
	models.ProjectStatusDraft:     "Черновик",
	models.ProjectStatusActive:    "Идёт",
	models.ProjectStatusOnHold:    "Приостановлен",
	models.ProjectStatusCompleted: "Завершён",
	models.ProjectStatusArchived:  "В архиве",
}

func projectStatusLabel(s models.ProjectStatus) string {
	if label, ok := projectStatusLabels[s]; ok {
		return label
	}
	return string(s)
}

var milestoneStatusLabels = map[models.MilestoneStatus]string{
	models.MilestoneStatusDone:    "Сдано",
	models.MilestoneStatusCurrent: "Текущая",
	models.MilestoneStatusPlanned: "План",
	models.MilestoneStatusFinal:   "Финал",
}

func milestoneStatusLabel(s models.MilestoneStatus) string {
	if label, ok := milestoneStatusLabels[s]; ok {
		return label
	}
	return string(s)
}

// scheduleStatusLabels — статусы соблюдения сроков из service.adherenceStatus.
var scheduleStatusLabels = map[string]string{
	"on_track": "Успевает к дедлайну",
	"at_risk":  "Под угрозой: запас почти исчерпан",
	"delayed":  "Отстаёт: прогноз позже дедлайна",
}

func scheduleStatusLabel(s string) string {
	if label, ok := scheduleStatusLabels[s]; ok {
		return label
	}
	return s
}

var dependencyTypeLabels = map[models.DependencyType]string{
	models.DependencyFS: "начать после окончания",
	models.DependencySS: "начать после начала",
	models.DependencyFF: "закончить после окончания",
	models.DependencySF: "закончить после начала",
}

// fieldLabels — поля задачи в журнале изменений (TaskHistoryEntry.Field).
var fieldLabels = map[string]string{
	"title":           "название",
	"description":     "описание",
	"sprintId":        "спринт",
	"milestoneId":     "контрольную точку",
	"weightPercent":   "вес задачи",
	"checklist":       "чек-лист",
	"dependency":      "связь",
	"assigneeId":      "ответственного",
	"startDate":       "начало",
	"endDate":         "окончание",
	"progressPercent": "прогресс",
}

// describeChange пересказывает запись журнала одной фразой для ассистента.
func describeChange(e models.TaskHistoryEntry) string {
	field := ""
	if e.Field != nil {
		field = *e.Field
		if label, ok := fieldLabels[field]; ok {
			field = label
		}
	}
	oldValue, newValue := deref(e.OldValue), deref(e.NewValue)

	switch e.Action {
	case models.HistoryActionCreated:
		return "создал(а) задачу"
	case models.HistoryActionStatusChanged:
		return fmt.Sprintf("статус: «%s» → «%s»",
			taskStatusLabel(models.TaskStatus(oldValue)), taskStatusLabel(models.TaskStatus(newValue)))
	case models.HistoryActionAssigned:
		return "сменил(а) ответственного"
	case models.HistoryActionDateShifted, models.HistoryActionProgressSet:
		return fmt.Sprintf("%s: %s → %s", field, oldValue, newValue)
	case models.HistoryActionDependencyAdd:
		return "добавил(а) связь с другой задачей"
	case models.HistoryActionDependencyDel:
		return "удалил(а) связь с другой задачей"
	case models.HistoryActionCommented:
		return "оставил(а) комментарий"
	case models.HistoryActionUpdated:
		if field != "" {
			return "изменил(а) " + field
		}
		return "изменил(а) задачу"
	}
	return e.Action
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
