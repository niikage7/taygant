package mcpserver

import (
	"context"
	"fmt"
	"sort"

	"github.com/google/uuid"

	"taygant_backend/internal/models"
	"taygant_backend/internal/service"
)

// projectRef — проект в ответах инструментов.
type projectRef struct {
	ID   string `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

func newProjectRef(p models.Project) projectRef {
	return projectRef{ID: p.ID.String(), Code: p.Code, Name: p.Name}
}

// taskItem — задача в ответах инструментов. Даты — строками YYYY-MM-DD:
// так их проще читать нейронке, и так они описаны в схеме ответа.
type taskItem struct {
	ID    string `json:"id"`
	Code  string `json:"code" jsonschema:"код задачи, уникален в пределах проекта"`
	Title string `json:"title"`
	// Project есть только в списках из нескольких проектов: в ответах про
	// один проект он повторял бы одно и то же в каждой задаче.
	Project         *projectRef `json:"project,omitempty"`
	Status          string      `json:"status" jsonschema:"planned, in_progress, done, overdue (срок прошёл, не закрыта) или blocked (ждёт незавершённую задачу)"`
	StatusLabel     string      `json:"statusLabel" jsonschema:"статус так, как он подписан в приложении"`
	StartDate       string      `json:"startDate"`
	EndDate         string      `json:"endDate"`
	DaysLeft        int         `json:"daysLeft" jsonschema:"дней до планового окончания; отрицательное число — на столько дней срок уже прошёл"`
	ProgressPercent int         `json:"progressPercent"`
	OnCriticalPath  bool        `json:"onCriticalPath" jsonschema:"задача на критическом пути: её задержка сдвигает срок всего проекта"`
	Assignee        string      `json:"assignee,omitempty" jsonschema:"исполнитель; пусто — не назначен"`
	BlockedBy       []string    `json:"blockedBy,omitempty" jsonschema:"незавершённые задачи, которые по плану должны закончиться раньше этой"`
}

// board — все задачи проекта с вычисленными статусами и связи между ними.
//
// Задачи берутся без фильтров, а отбор (например, «только мои») делается уже
// здесь: overdue/blocked и критический путь корректны только на полном графе
// проекта (см. комментарий к service.Tasks.Get).
type board struct {
	project models.Project
	tasks   []models.Task
	byID    map[uuid.UUID]models.Task
	// preds — предшественники каждой задачи: successor → []predecessor.
	preds map[uuid.UUID][]uuid.UUID
}

func (s *Server) loadBoard(ctx context.Context, project models.Project) (*board, error) {
	tasks, err := s.tasks.List(ctx, project.ID, project, service.Filter{})
	if err != nil {
		return nil, err
	}
	deps, err := s.dependencies.ListForProject(ctx, project.ID)
	if err != nil {
		return nil, err
	}

	b := &board{
		project: project,
		tasks:   tasks,
		byID:    make(map[uuid.UUID]models.Task, len(tasks)),
		preds:   make(map[uuid.UUID][]uuid.UUID, len(deps)),
	}
	for _, t := range tasks {
		b.byID[t.ID] = t
	}
	for _, d := range deps {
		b.preds[d.SuccessorTaskID] = append(b.preds[d.SuccessorTaskID], d.PredecessorTaskID)
	}
	return b, nil
}

// blockedBy перечисляет незавершённых предшественников задачи.
func (b *board) blockedBy(t models.Task) []string {
	if t.Status == models.TaskStatusDone {
		return nil
	}
	var out []string
	for _, id := range b.preds[t.ID] {
		p, ok := b.byID[id]
		if !ok || p.Status == models.TaskStatusDone {
			continue
		}
		line := fmt.Sprintf("%s «%s» — %s", p.Code, p.Title, taskStatusLabel(p.Status))
		if p.Assignee != nil {
			line += ", исполнитель " + p.Assignee.FullName
		}
		out = append(out, line)
	}
	return out
}

// itemWithProject — задача для ответа вместе с проектом, в котором она лежит.
func (b *board) itemWithProject(t models.Task, today models.Date) taskItem {
	item := b.item(t, today)
	project := newProjectRef(b.project)
	item.Project = &project
	return item
}

// item собирает задачу для ответа о проекте, где проект и так известен.
func (b *board) item(t models.Task, today models.Date) taskItem {
	item := taskItem{
		ID:              t.ID.String(),
		Code:            t.Code,
		Title:           t.Title,
		Status:          string(t.Status),
		StatusLabel:     taskStatusLabel(t.Status),
		StartDate:       t.StartDate.String(),
		EndDate:         t.EndDate.String(),
		DaysLeft:        today.DaysUntil(t.EndDate),
		ProgressPercent: t.ProgressPercent,
		OnCriticalPath:  t.IsCriticalPath,
		BlockedBy:       b.blockedBy(t),
	}
	if t.Assignee != nil {
		item.Assignee = t.Assignee.FullName
	}
	return item
}

// urgencyRank упорядочивает задачи «что делать первым»: сорванные сроки,
// потом начатое, потом то, что ждёт других, потом будущее, в конце готовое.
func urgencyRank(s models.TaskStatus) int {
	switch s {
	case models.TaskStatusOverdue:
		return 0
	case models.TaskStatusInProgress:
		return 1
	case models.TaskStatusBlocked:
		return 2
	case models.TaskStatusPlanned:
		return 3
	default:
		return 4
	}
}

// boardTask — задача вместе с доской её проекта: списки «мои задачи» собираются
// из нескольких проектов, а для ответа нужны связи именно её проекта.
type boardTask struct {
	board *board
	task  models.Task
}

// sortByUrgency сортирует по срочности, внутри одной группы — по плановому
// окончанию (что раньше заканчивается, то выше), затем по коду.
func sortByUrgency(tasks []boardTask) {
	sort.SliceStable(tasks, func(i, j int) bool {
		a, b := tasks[i].task, tasks[j].task
		if ra, rb := urgencyRank(a.Status), urgencyRank(b.Status); ra != rb {
			return ra < rb
		}
		if !a.EndDate.Equal(b.EndDate) {
			return a.EndDate.Before(b.EndDate)
		}
		return a.Code < b.Code
	})
}
