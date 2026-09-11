package mcpserver

import (
	"context"
	"fmt"
	"math"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"taygant_backend/internal/access"
	"taygant_backend/internal/models"
	"taygant_backend/internal/service"
)

// ---------- list_my_projects ----------

type listProjectsInput struct {
	IncludeArchived bool `json:"includeArchived,omitempty" jsonschema:"показать и архивные проекты"`
}

type projectListItem struct {
	ID              string `json:"id"`
	Code            string `json:"code"`
	Name            string `json:"name"`
	Status          string `json:"status" jsonschema:"draft, active, on_hold, completed или archived"`
	StatusLabel     string `json:"statusLabel"`
	MyAccess        string `json:"myAccess" jsonschema:"full — управляет проектом; edit — меняет статус и сроки своих задач; view — только просмотр"`
	StartDate       string `json:"startDate"`
	Deadline        string `json:"deadline"`
	DaysToDeadline  int    `json:"daysToDeadline" jsonschema:"отрицательное число — дедлайн уже прошёл"`
	ProgressPercent int    `json:"progressPercent"`
	OverdueTasks    int    `json:"overdueTasks"`
	MyOpenTasks     int    `json:"myOpenTasks" jsonschema:"незавершённые задачи, где вы исполнитель"`
}

type listProjectsOutput struct {
	Projects []projectListItem `json:"projects"`
}

func (s *Server) listMyProjects(ctx context.Context, req *mcp.CallToolRequest, in listProjectsInput) (*mcp.CallToolResult, listProjectsOutput, error) {
	c, err := callerFrom(req)
	if err != nil {
		return nil, listProjectsOutput{}, err
	}
	all, err := s.memberships(ctx, c.User.ID)
	if err != nil {
		return nil, listProjectsOutput{}, toolError(err)
	}

	today := models.Today()
	out := listProjectsOutput{Projects: make([]projectListItem, 0, len(all))}
	for _, m := range all {
		if m.Project.Status == models.ProjectStatusArchived && !in.IncludeArchived {
			continue
		}
		metrics, err := service.ComputeMetrics(ctx, s.db, m.Project)
		if err != nil {
			return nil, listProjectsOutput{}, toolError(err)
		}
		var myOpen int64
		err = s.db.WithContext(ctx).Model(&models.Task{}).
			Where("project_id = ? AND assignee_id = ? AND status <> ?", m.Project.ID, c.User.ID, models.TaskStatusDone).
			Count(&myOpen).Error
		if err != nil {
			return nil, listProjectsOutput{}, toolError(fmt.Errorf("посчитать задачи пользователя: %w", err))
		}

		out.Projects = append(out.Projects, projectListItem{
			ID:              m.Project.ID.String(),
			Code:            m.Project.Code,
			Name:            m.Project.Name,
			Status:          string(m.Project.Status),
			StatusLabel:     projectStatusLabel(m.Project.Status),
			MyAccess:        string(m.Member.AccessLevel),
			StartDate:       m.Project.StartDate.String(),
			Deadline:        m.Project.Deadline.String(),
			DaysToDeadline:  today.DaysUntil(m.Project.Deadline),
			ProgressPercent: int(math.Round(metrics.ProgressPercent)),
			OverdueTasks:    metrics.CriticalRisksCount,
			MyOpenTasks:     int(myOpen),
		})
	}
	return nil, out, nil
}

// ---------- my_tasks ----------

type myTasksInput struct {
	Project     string `json:"project,omitempty" jsonschema:"проект: id, код или часть названия; без него — задачи из всех ваших проектов"`
	IncludeDone bool   `json:"includeDone,omitempty" jsonschema:"добавить уже завершённые задачи"`
}

type myTasksSummary struct {
	Open       int `json:"open" jsonschema:"незавершённых задач"`
	Overdue    int `json:"overdue"`
	InProgress int `json:"inProgress"`
	Blocked    int `json:"blocked"`
	Planned    int `json:"planned"`
	DueSoon    int `json:"dueSoon" jsonschema:"незавершённых задач, срок которых наступает в ближайшие 3 дня"`
}

type myTasksOutput struct {
	Today   string         `json:"today"`
	Summary myTasksSummary `json:"summary"`
	Tasks   []taskItem     `json:"tasks" jsonschema:"сначала просроченные, потом начатые, ждущие других и будущие; внутри — по сроку"`
	Note    string         `json:"note,omitempty"`
}

// myTasksLimit — сколько задач отдавать ассистенту за раз. Длинный список
// съедает контекст нейронки, а «что у меня сегодня» — это верх списка.
const myTasksLimit = 50

// dueSoonDays — горизонт «скоро срок» для сводки.
const dueSoonDays = 3

func (s *Server) myTasks(ctx context.Context, req *mcp.CallToolRequest, in myTasksInput) (*mcp.CallToolResult, myTasksOutput, error) {
	c, err := callerFrom(req)
	if err != nil {
		return nil, myTasksOutput{}, err
	}
	all, err := s.memberships(ctx, c.User.ID)
	if err != nil {
		return nil, myTasksOutput{}, toolError(err)
	}

	scope := make([]access.Membership, 0, len(all))
	if in.Project != "" {
		m, err := pickProject(all, in.Project)
		if err != nil {
			return nil, myTasksOutput{}, toolError(err)
		}
		scope = append(scope, m)
	} else {
		for _, m := range all {
			if m.Project.Status != models.ProjectStatusArchived {
				scope = append(scope, m)
			}
		}
	}

	today := models.Today()
	var mine []boardTask
	summary := myTasksSummary{}
	for _, m := range scope {
		b, err := s.loadBoard(ctx, m.Project)
		if err != nil {
			return nil, myTasksOutput{}, toolError(err)
		}
		for _, t := range b.tasks {
			if t.AssigneeID == nil || *t.AssigneeID != c.User.ID {
				continue
			}
			if t.Status != models.TaskStatusDone {
				summary.Open++
				switch t.Status {
				case models.TaskStatusOverdue:
					summary.Overdue++
				case models.TaskStatusInProgress:
					summary.InProgress++
				case models.TaskStatusBlocked:
					summary.Blocked++
				case models.TaskStatusPlanned:
					summary.Planned++
				}
				if days := today.DaysUntil(t.EndDate); days >= 0 && days <= dueSoonDays {
					summary.DueSoon++
				}
			} else if !in.IncludeDone {
				continue
			}
			mine = append(mine, boardTask{board: b, task: t})
		}
	}
	sortByUrgency(mine)

	out := myTasksOutput{Today: today.String(), Summary: summary, Tasks: make([]taskItem, 0, min(len(mine), myTasksLimit))}
	for i, bt := range mine {
		if i == myTasksLimit {
			out.Note = fmt.Sprintf("показаны первые %d задач из %d — сузьте выбор параметром project", myTasksLimit, len(mine))
			break
		}
		out.Tasks = append(out.Tasks, bt.board.itemWithProject(bt.task, today))
	}
	switch {
	case out.Note != "":
	case len(mine) == 0 && in.IncludeDone:
		out.Note = "на вас не назначено ни одной задачи"
	case summary.Open == 0:
		out.Note = "незавершённых задач на вас нет"
	}
	return nil, out, nil
}
