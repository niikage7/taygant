package mcpserver

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"taygant_backend/internal/access"
	"taygant_backend/internal/models"
	"taygant_backend/internal/service"
)

// projectInput — как инструменты уровня проекта принимают ссылку на проект.
type projectInput struct {
	Project string `json:"project,omitempty" jsonschema:"проект: id, код или часть названия; можно не указывать, если активный проект у вас один"`
}

// projectSnapshot — всё, что нужно для ответов о проекте: доска задач и
// сводка дашборда. Цифры те же, что на экране «Обзор и Аналитика».
type projectSnapshot struct {
	membership access.Membership
	board      *board
	dashboard  service.DashboardResult
	today      models.Date
}

func (s *Server) snapshot(ctx context.Context, req *mcp.CallToolRequest, projectRef string) (projectSnapshot, error) {
	c, err := callerFrom(req)
	if err != nil {
		return projectSnapshot{}, err
	}
	all, err := s.memberships(ctx, c.User.ID)
	if err != nil {
		return projectSnapshot{}, toolError(err)
	}
	m, err := pickProject(all, projectRef)
	if err != nil {
		return projectSnapshot{}, toolError(err)
	}
	b, err := s.loadBoard(ctx, m.Project)
	if err != nil {
		return projectSnapshot{}, toolError(err)
	}
	dash, err := s.dashboard.Get(ctx, m.Project.ID)
	if err != nil {
		return projectSnapshot{}, toolError(err)
	}
	return projectSnapshot{membership: m, board: b, dashboard: dash, today: models.Today()}, nil
}

// forecastFinish — когда проект закончится по текущему плану; пусто, если
// задач нет и прогнозировать нечего.
func (p projectSnapshot) forecastFinish() string {
	if len(p.board.tasks) == 0 {
		return ""
	}
	return p.membership.Project.Deadline.AddDays(p.dashboard.ScheduleAdherence.ForecastDeviationDays).String()
}

// ---------- project_status ----------

type taskBreakdown struct {
	Total      int `json:"total"`
	Done       int `json:"done"`
	InProgress int `json:"inProgress"`
	Planned    int `json:"planned"`
	Blocked    int `json:"blocked" jsonschema:"ждут незавершённые задачи"`
	Overdue    int `json:"overdue"`
}

type scheduleOut struct {
	Status         string `json:"status" jsonschema:"on_track — успевает; at_risk — запас почти исчерпан; delayed — прогноз позже дедлайна"`
	StatusLabel    string `json:"statusLabel"`
	ForecastFinish string `json:"forecastFinish,omitempty" jsonschema:"когда проект закончится по текущему плану с учётом связей (критический путь)"`
	DeviationDays  int    `json:"deviationDays" jsonschema:"на сколько дней прогноз позже дедлайна; отрицательное число — раньше"`
	BufferDays     int    `json:"bufferDays" jsonschema:"сколько дней проект может потерять и всё ещё успеть к дедлайну"`
}

type milestoneOut struct {
	Name        string `json:"name"`
	PlannedDate string `json:"plannedDate"`
	Status      string `json:"status"`
	StatusLabel string `json:"statusLabel"`
	OverdueDays int    `json:"overdueDays,omitempty" jsonschema:"на сколько дней плановая дата уже прошла"`
	TasksDone   int    `json:"tasksDone"`
	TasksTotal  int    `json:"tasksTotal"`
}

type statusUpdatesOut struct {
	PeriodDays   int `json:"periodDays"`
	Total        int `json:"total" jsonschema:"сколько раз за период меняли статусы задач"`
	ViaAssistant int `json:"viaAssistant" jsonschema:"из них — через ассистента"`
}

type projectStatusOutput struct {
	Project         projectRef       `json:"project"`
	Today           string           `json:"today"`
	Summary         string           `json:"summary" jsonschema:"состояние проекта в двух фразах"`
	Status          string           `json:"status" jsonschema:"draft, active, on_hold, completed или archived"`
	StatusLabel     string           `json:"statusLabel"`
	Phase           string           `json:"phase,omitempty"`
	StartDate       string           `json:"startDate"`
	Deadline        string           `json:"deadline"`
	DaysToDeadline  int              `json:"daysToDeadline"`
	ProgressPercent int              `json:"progressPercent"`
	HealthIndex     int              `json:"healthIndex" jsonschema:"здоровье 0–100: 100 — без просрочек, минус 10 за каждую просроченную задачу"`
	Tasks           taskBreakdown    `json:"tasks"`
	Schedule        scheduleOut      `json:"schedule"`
	Milestones      []milestoneOut   `json:"milestones" jsonschema:"ближайшие несданные контрольные точки"`
	InProgress      []taskItem       `json:"inProgress" jsonschema:"что сейчас в работе, ближайшие сроки первыми"`
	TasksPerWeek    float64          `json:"tasksPerWeek" jsonschema:"сколько задач в неделю команда закрывала за последние две недели"`
	StatusUpdates   statusUpdatesOut `json:"statusUpdates" jsonschema:"как часто обновляют статусы задач и какая доля обновлений — через ассистента"`
}

// inProgressLimit — сколько задач «в работе» перечислять в сводке проекта.
const inProgressLimit = 10

func (s *Server) projectStatus(ctx context.Context, req *mcp.CallToolRequest, in projectInput) (*mcp.CallToolResult, projectStatusOutput, error) {
	snap, err := s.snapshot(ctx, req, in.Project)
	if err != nil {
		return nil, projectStatusOutput{}, err
	}
	project, dash, today := snap.membership.Project, snap.dashboard, snap.today

	breakdown := taskBreakdown{Total: len(snap.board.tasks)}
	var inProgress []boardTask
	for _, t := range snap.board.tasks {
		switch t.Status {
		case models.TaskStatusDone:
			breakdown.Done++
		case models.TaskStatusInProgress:
			breakdown.InProgress++
			inProgress = append(inProgress, boardTask{board: snap.board, task: t})
		case models.TaskStatusPlanned:
			breakdown.Planned++
		case models.TaskStatusBlocked:
			breakdown.Blocked++
		case models.TaskStatusOverdue:
			breakdown.Overdue++
		}
	}
	sortByUrgency(inProgress)

	schedule := scheduleOut{
		Status:         dash.ScheduleAdherence.Status,
		StatusLabel:    scheduleStatusLabel(dash.ScheduleAdherence.Status),
		ForecastFinish: snap.forecastFinish(),
		DeviationDays:  dash.ScheduleAdherence.ForecastDeviationDays,
		BufferDays:     dash.ScheduleAdherence.ProjectBufferDays,
	}

	out := projectStatusOutput{
		Project:         newProjectRef(project),
		Today:           today.String(),
		Status:          string(project.Status),
		StatusLabel:     projectStatusLabel(project.Status),
		Phase:           project.Phase,
		StartDate:       project.StartDate.String(),
		Deadline:        project.Deadline.String(),
		DaysToDeadline:  today.DaysUntil(project.Deadline),
		ProgressPercent: int(math.Round(dash.Metrics.ProgressPercent)),
		HealthIndex:     dash.Metrics.HealthIndex,
		Tasks:           breakdown,
		Schedule:        schedule,
		Milestones:      make([]milestoneOut, 0, len(dash.NearestMilestones)),
		InProgress:      make([]taskItem, 0, min(len(inProgress), inProgressLimit)),
		TasksPerWeek:    math.Round(dash.TasksPerWeek*10) / 10,
		StatusUpdates: statusUpdatesOut{
			PeriodDays:   dash.StatusChanges.PeriodDays,
			Total:        dash.StatusChanges.Total,
			ViaAssistant: dash.StatusChanges.ViaAssistant,
		},
	}
	out.Summary = projectSummary(project, breakdown, schedule)

	for _, m := range dash.NearestMilestones {
		item := milestoneOut{
			Name:        m.Name,
			PlannedDate: m.PlannedDate.String(),
			Status:      string(m.Status),
			StatusLabel: milestoneStatusLabel(m.Status),
			TasksDone:   m.TasksDone,
			TasksTotal:  m.TasksTotal,
		}
		if m.RiskDays != nil {
			item.OverdueDays = *m.RiskDays
		}
		out.Milestones = append(out.Milestones, item)
	}
	for i, bt := range inProgress {
		if i == inProgressLimit {
			break
		}
		out.InProgress = append(out.InProgress, bt.board.item(bt.task, today))
	}
	return nil, out, nil
}

// projectSummary — две фразы о проекте: успевает ли он и что с задачами.
// Готовая формулировка избавляет нейронку от пересчёта цифр и делает ответы
// разных ассистентов одинаковыми по смыслу.
func projectSummary(p models.Project, tasks taskBreakdown, schedule scheduleOut) string {
	if tasks.Total == 0 {
		return fmt.Sprintf("В проекте пока нет задач, прогноза срока нет. Дедлайн — %s.", p.Deadline.String())
	}

	var first string
	switch schedule.Status {
	case "delayed":
		first = fmt.Sprintf("Проект отстаёт: по текущему плану закончится %s — на %d дн. позже дедлайна %s.",
			schedule.ForecastFinish, schedule.DeviationDays, p.Deadline.String())
	case "at_risk":
		first = fmt.Sprintf("Проект пока успевает к дедлайну %s, но запас почти исчерпан: %d дн.",
			p.Deadline.String(), schedule.BufferDays)
	default:
		first = fmt.Sprintf("Проект идёт по плану: прогноз окончания %s, до дедлайна %s остаётся запас %d дн.",
			schedule.ForecastFinish, p.Deadline.String(), schedule.BufferDays)
	}

	second := fmt.Sprintf("Готово %d из %d задач, в работе %d", tasks.Done, tasks.Total, tasks.InProgress)
	if tasks.Overdue > 0 {
		second += fmt.Sprintf(", просрочено %d", tasks.Overdue)
	}
	if tasks.Blocked > 0 {
		second += fmt.Sprintf(", ждут других %d", tasks.Blocked)
	}
	return first + " " + second + "."
}

// ---------- project_risks ----------

type riskItem struct {
	Severity string    `json:"severity" jsonschema:"critical, high или medium"`
	Kind     string    `json:"kind" jsonschema:"deadline — прогноз срыва срока проекта; overdue — просрочена задача; blocked — задача должна идти, но ждёт незавершённую; late_start — задача не начата в срок; milestone — не сдана контрольная точка"`
	Title    string    `json:"title"`
	Details  string    `json:"details"`
	Task     *taskItem `json:"task,omitempty"`
}

type projectRisksOutput struct {
	Project projectRef `json:"project"`
	Today   string     `json:"today"`
	Summary string     `json:"summary"`
	Items   []riskItem `json:"items" jsonschema:"самое срочное первым"`
}

// risksLimit — сколько пунктов «что горит» отдавать: дальше двадцатого
// пункта список перестаёт отвечать на вопрос «что горит».
const risksLimit = 20

var severityRank = map[string]int{"critical": 0, "high": 1, "medium": 2}

func (s *Server) projectRisks(ctx context.Context, req *mcp.CallToolRequest, in projectInput) (*mcp.CallToolResult, projectRisksOutput, error) {
	snap, err := s.snapshot(ctx, req, in.Project)
	if err != nil {
		return nil, projectRisksOutput{}, err
	}
	b, today, dash := snap.board, snap.today, snap.dashboard

	var items []riskItem

	// Срыв срока проекта и просроченные задачи — те же риски, что на дашборде,
	// с теми же уровнями серьёзности.
	for _, r := range dash.Risks {
		item := riskItem{Severity: r.Severity, Kind: "deadline", Title: r.Title, Details: r.Description}
		if r.RelatedTaskID != nil {
			item.Kind = "overdue"
			if t, ok := b.byID[*r.RelatedTaskID]; ok {
				task := b.item(t, today)
				item.Task = &task
				item.Title = fmt.Sprintf("%s «%s» просрочена на %d дн.", t.Code, t.Title, -task.DaysLeft)
				item.Details = overdueDetails(t, task)
			}
		}
		items = append(items, item)
	}

	for _, t := range b.tasks {
		switch {
		// Заблокированная задача горит, только если ей уже пора идти.
		case t.Status == models.TaskStatusBlocked && !t.StartDate.After(today):
			task := b.item(t, today)
			items = append(items, riskItem{
				Severity: criticalOr(t, "high", "medium"),
				Kind:     "blocked",
				Title:    fmt.Sprintf("%s «%s» не может начаться", t.Code, t.Title),
				Details:  "Плановый старт уже наступил, но задача ждёт: " + strings.Join(task.BlockedBy, "; ") + ".",
				Task:     &task,
			})
		case t.Status == models.TaskStatusPlanned && t.StartDate.Before(today):
			late := t.StartDate.DaysUntil(today)
			task := b.item(t, today)
			severity := criticalOr(t, "high", "medium")
			if late >= 7 {
				severity = "high"
			}
			details := fmt.Sprintf("Плановый старт был %s, до срока окончания %d дн.", t.StartDate.String(), task.DaysLeft)
			if task.Assignee == "" {
				details += " Исполнитель не назначен."
			}
			items = append(items, riskItem{
				Severity: severity,
				Kind:     "late_start",
				Title:    fmt.Sprintf("%s «%s» не начата, хотя должна идти уже %d дн.", t.Code, t.Title, late),
				Details:  details,
				Task:     &task,
			})
		}
	}

	for _, m := range dash.NearestMilestones {
		if m.RiskDays == nil || *m.RiskDays <= 0 {
			continue
		}
		severity := "medium"
		if *m.RiskDays >= 7 {
			severity = "high"
		}
		items = append(items, riskItem{
			Severity: severity,
			Kind:     "milestone",
			Title:    fmt.Sprintf("Контрольная точка «%s» не сдана, плановая дата прошла %d дн. назад", m.Name, *m.RiskDays),
			Details:  fmt.Sprintf("Плановая дата — %s, готово задач %d из %d.", m.PlannedDate.String(), m.TasksDone, m.TasksTotal),
		})
	}

	sort.SliceStable(items, func(i, j int) bool {
		return severityRank[items[i].Severity] < severityRank[items[j].Severity]
	})

	out := projectRisksOutput{
		Project: newProjectRef(snap.membership.Project),
		Today:   today.String(),
		Items:   make([]riskItem, 0, min(len(items), risksLimit)),
	}
	for i, item := range items {
		if i == risksLimit {
			break
		}
		out.Items = append(out.Items, item)
	}
	out.Summary = risksSummary(items)
	return nil, out, nil
}

// criticalOr выбирает серьёзность: задача на критическом пути сдвигает срок
// всего проекта, поэтому её проблема на уровень серьёзнее.
func criticalOr(t models.Task, critical, other string) string {
	if t.IsCriticalPath {
		return critical
	}
	return other
}

func overdueDetails(t models.Task, task taskItem) string {
	details := fmt.Sprintf("Плановое окончание было %s, прогресс %d%%.", t.EndDate.String(), t.ProgressPercent)
	if task.Assignee != "" {
		details += " Исполнитель — " + task.Assignee + "."
	} else {
		details += " Исполнитель не назначен."
	}
	if t.IsCriticalPath {
		details += " Задача на критическом пути: её задержка сдвигает срок всего проекта."
	}
	return details
}

// risksSummary — одна фраза: сколько всего горит и насколько сильно.
func risksSummary(items []riskItem) string {
	if len(items) == 0 {
		return "Ничего не горит: просрочек, заблокированных и не начатых вовремя задач нет, прогноз укладывается в дедлайн."
	}
	counts := map[string]int{}
	for _, item := range items {
		counts[item.Severity]++
	}
	var parts []string
	for _, sev := range []struct{ key, label string }{
		{"critical", "критичных"}, {"high", "серьёзных"}, {"medium", "умеренных"},
	} {
		if n := counts[sev.key]; n > 0 {
			parts = append(parts, fmt.Sprintf("%d %s", n, sev.label))
		}
	}
	return fmt.Sprintf("Проблем: %d (%s). Самые срочные — первыми в списке.", len(items), strings.Join(parts, ", "))
}
