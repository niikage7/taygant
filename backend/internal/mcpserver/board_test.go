package mcpserver

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"

	"taygant_backend/internal/models"
	"taygant_backend/internal/service"
)

func boardTaskOf(code string, status models.TaskStatus, endInDays int) boardTask {
	task := models.Task{Code: code, Status: status, EndDate: models.Today().AddDays(endInDays)}
	task.ID = uuid.New()
	return boardTask{task: task}
}

func TestSortByUrgency(t *testing.T) {
	tasks := []boardTask{
		boardTaskOf("TASK-001", models.TaskStatusDone, -5),
		boardTaskOf("TASK-002", models.TaskStatusPlanned, 3),
		boardTaskOf("TASK-003", models.TaskStatusBlocked, 9),
		boardTaskOf("TASK-004", models.TaskStatusInProgress, 7),
		boardTaskOf("TASK-005", models.TaskStatusOverdue, -1),
		boardTaskOf("TASK-006", models.TaskStatusInProgress, 2),
		boardTaskOf("TASK-007", models.TaskStatusPlanned, 3),
	}
	sortByUrgency(tasks)

	var got []string
	for _, bt := range tasks {
		got = append(got, bt.task.Code)
	}
	// Просроченная → в работе (по сроку) → ждёт → план (по сроку, затем по коду) → готовая.
	want := "TASK-005 TASK-006 TASK-004 TASK-003 TASK-002 TASK-007 TASK-001"
	if strings.Join(got, " ") != want {
		t.Fatalf("порядок = %v, ожидался %s", got, want)
	}
}

func TestBlockedByListsOnlyUnfinishedPredecessors(t *testing.T) {
	newTask := func(code, title string, status models.TaskStatus, assignee string) models.Task {
		task := models.Task{Code: code, Title: title, Status: status}
		task.ID = uuid.New()
		if assignee != "" {
			task.Assignee = &models.User{FullName: assignee}
		}
		return task
	}
	design := newTask("TASK-001", "Дизайн", models.TaskStatusDone, "")
	api := newTask("TASK-002", "API", models.TaskStatusInProgress, "Дмитрий Петров")
	front := newTask("TASK-003", "Фронтенд", models.TaskStatusBlocked, "")
	b := &board{
		byID: map[uuid.UUID]models.Task{design.ID: design, api.ID: api, front.ID: front},
		// Третий предшественник — задача, которой уже нет на доске (удалена).
		preds: map[uuid.UUID][]uuid.UUID{front.ID: {design.ID, api.ID, uuid.New()}},
	}

	got := b.blockedBy(front)
	if len(got) != 1 || got[0] != "TASK-002 «API» — В работе, исполнитель Дмитрий Петров" {
		t.Fatalf("blockedBy = %q", got)
	}
	front.Status = models.TaskStatusDone
	if got := b.blockedBy(front); got != nil {
		t.Fatalf("у готовой задачи нечего ждать, а blockedBy = %q", got)
	}
}

func TestItemShowsProjectOnlyWhenAsked(t *testing.T) {
	project := models.Project{Code: "PRJ-1", Name: "Сайт"}
	project.ID = uuid.New()
	task := models.Task{Code: "TASK-001", Title: "Задача", Status: models.TaskStatusPlanned, EndDate: models.Today().AddDays(4)}
	task.ID = uuid.New()
	b := &board{project: project, byID: map[uuid.UUID]models.Task{task.ID: task}}

	if item := b.item(task, models.Today()); item.Project != nil || item.DaysLeft != 4 || item.StatusLabel != "План" {
		t.Fatalf("item = %+v", item)
	}
	if item := b.itemWithProject(task, models.Today()); item.Project == nil || item.Project.Name != "Сайт" {
		t.Fatalf("itemWithProject без проекта: %+v", item.Project)
	}
}

func TestToolErrorHidesInternalDetails(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want string
	}{
		{"текст для ассистента проходит как есть", failf("уточните проект"), "уточните проект"},
		{"не найдено", fmt.Errorf("найти: %w", service.ErrNotFound), "не найдено"},
		{"нет прав", service.ErrForbidden, "недостаточно прав"},
		{"ошибка ввода", service.Invalid("дата окончания раньше начала"), "дата окончания раньше начала"},
		{"сбой БД", errors.New(`pq: relation "tasks" does not exist`), "внутренняя ошибка"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := toolError(tc.err)
			if got == nil || !strings.Contains(got.Error(), tc.want) {
				t.Fatalf("toolError = %v, ожидалось с %q", got, tc.want)
			}
			if strings.Contains(got.Error(), "relation") {
				t.Fatalf("наружу утёк текст SQL: %v", got)
			}
		})
	}
	if toolError(nil) != nil {
		t.Fatal("toolError(nil) не nil")
	}
}
