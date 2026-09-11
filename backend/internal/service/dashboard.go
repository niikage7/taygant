package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"taygant_backend/internal/models"
)

// Dashboard — сводные показатели состояния и рисков проекта.
type Dashboard struct {
	db         *gorm.DB
	tasks      *Tasks
	milestones *Milestones
	workload   *Workload
}

// NewDashboard собирает сервис дашборда поверх уже готовых сервисов задач,
// вех и загрузки команды: их логика (эффективные статусы, CPM, статус вехи,
// часы/неделю по спринту) не дублируется здесь.
func NewDashboard(db *gorm.DB, tasks *Tasks, milestones *Milestones, workload *Workload) *Dashboard {
	return &Dashboard{db: db, tasks: tasks, milestones: milestones, workload: workload}
}

// ScheduleAdherence — соблюдение сроков проекта, посчитанное CPM-движком.
type ScheduleAdherence struct {
	Status                string
	ForecastDeviationDays int
	ProjectBufferDays     int
}

// TaskStatusBreakdown — разбивка задач проекта по эффективным статусам.
type TaskStatusBreakdown struct {
	Total, Done, InProgress, Planned, Overdue int
}

// Risk — активный риск проекта.
type Risk struct {
	Severity      string
	Title         string
	Description   string
	RelatedTaskID *uuid.UUID
	ImpactDays    *int
}

// AttentionTask — задача, отстающая от плана.
type AttentionTask struct {
	Task          models.Task
	DeviationDays int
}

// Result — данные экрана «Обзор и Аналитика».
//
// TeamWorkload считается тем же сервисом, что и GET /projects/{id}/workload,
// за текущий спринт (см. Workload.List и её комментарий про эвристику часов).
// ProgressDeltaPercent всегда 0: сравнение с предыдущим периодом требует
// исторических снимков прогресса, которых в БД не хранится ни одного —
// это отдельный, не CPM-related пробел (нужна таблица снапшотов и job,
// который их пишет).
type DashboardResult struct {
	Project              models.Project
	Metrics              Metrics
	ProgressDeltaPercent float64
	MilestonesClosed     int
	MilestonesTotal      int
	ScheduleAdherence    ScheduleAdherence
	TaskStatusBreakdown  TaskStatusBreakdown
	TasksPerWeek         float64
	NearestMilestones    []models.Milestone
	Risks                []Risk
	AttentionTasks       []AttentionTask
	TeamWorkload         []MemberWorkload
	// StatusChanges — сколько смен статуса за последние 30 дней сделано
	// через ассистента (MCP): метрика подключения к нейронкам из docs/mcp.md.
	StatusChanges StatusChangeStats
}

// bufferCriticalThresholdDays — ниже этого остатка резерва проект считается
// «в зоне риска», даже если формально ещё укладывается в срок. Число выбрано
// как разумный порог для MVP, а не выведено из данных — с реальной историей
// проектов его стоит пересмотреть.
const bufferCriticalThresholdDays = 3

// Get собирает сводку по проекту.
func (s *Dashboard) Get(ctx context.Context, projectID uuid.UUID) (DashboardResult, error) {
	var project models.Project
	if err := s.db.WithContext(ctx).First(&project, "id = ?", projectID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return DashboardResult{}, ErrNotFound
		}
		return DashboardResult{}, fmt.Errorf("найти проект: %w", err)
	}

	metrics, err := ComputeMetrics(ctx, s.db, project)
	if err != nil {
		return DashboardResult{}, err
	}

	tasks, err := s.tasks.List(ctx, projectID, project, Filter{})
	if err != nil {
		return DashboardResult{}, err
	}
	milestones, err := s.milestones.List(ctx, projectID)
	if err != nil {
		return DashboardResult{}, err
	}
	summary, err := s.tasks.ComputeSchedule(ctx, projectID)
	if err != nil {
		return DashboardResult{}, err
	}
	workload, err := s.workload.List(ctx, projectID, nil)
	if err != nil {
		return DashboardResult{}, err
	}
	statusChanges, err := statusChangeStats(ctx, s.db, projectID)
	if err != nil {
		return DashboardResult{}, err
	}

	breakdown := TaskStatusBreakdown{Total: len(tasks)}
	var attention []AttentionTask
	var risks []Risk
	for _, t := range tasks {
		switch t.Status {
		case models.TaskStatusDone:
			breakdown.Done++
		case models.TaskStatusInProgress:
			breakdown.InProgress++
		case models.TaskStatusPlanned, models.TaskStatusBlocked:
			breakdown.Planned++
		case models.TaskStatusOverdue:
			breakdown.Overdue++
		}

		if deviation := t.PlanVsActualDeviation(); deviation < 0 {
			attention = append(attention, AttentionTask{Task: t, DeviationDays: deviation})
		}

		if t.Status == models.TaskStatusOverdue {
			risks = append(risks, overdueRisk(t))
		}
	}

	forecastDeviation, projectBuffer := 0, 0
	if !summary.ComputedFinish.IsZero() {
		forecastDeviation = project.Deadline.DaysUntil(summary.ComputedFinish)
		projectBuffer = projectBufferDays(forecastDeviation)
	}

	adherence := ScheduleAdherence{
		ForecastDeviationDays: forecastDeviation,
		ProjectBufferDays:     projectBuffer,
		Status:                adherenceStatus(forecastDeviation, projectBuffer),
	}

	var nearest []models.Milestone
	for _, m := range milestones {
		if m.Status != models.MilestoneStatusDone {
			nearest = append(nearest, m)
		}
	}
	if len(nearest) > 3 {
		nearest = nearest[:3]
	}
	milestonesClosed := 0
	for _, m := range milestones {
		if m.Status == models.MilestoneStatusDone {
			milestonesClosed++
		}
	}

	if forecastDeviation > 0 {
		risks = append(risks, deadlineRisk(forecastDeviation))
	}

	return DashboardResult{
		Project:              project,
		Metrics:              metrics,
		ProgressDeltaPercent: 0,
		MilestonesClosed:     milestonesClosed,
		MilestonesTotal:      len(milestones),
		ScheduleAdherence:    adherence,
		TaskStatusBreakdown:  breakdown,
		TasksPerWeek:         tasksPerWeek(tasks),
		NearestMilestones:    nearest,
		Risks:                risks,
		AttentionTasks:       attention,
		TeamWorkload:         workload,
		StatusChanges:        statusChanges,
	}, nil
}

// overdueRisk строит риск из просроченной задачи. ImpactDays заполняется,
// только если задача лежит на критическом пути: только тогда её просрочка
// напрямую сдвигает срок всего проекта, а не поглощается чьим-то буфером —
// это уже реальный вывод CPM, а не догадка.
func overdueRisk(t models.Task) Risk {
	severity := "medium"
	overdueDays := -t.PlanVsActualDeviation()
	switch {
	case overdueDays >= 7:
		severity = "critical"
	case overdueDays >= 3:
		severity = "high"
	}

	risk := Risk{
		Severity:      severity,
		Title:         fmt.Sprintf("Задача «%s» просрочена", t.Title),
		Description:   fmt.Sprintf("Задача %s: плановый срок истёк %d %s назад.", t.Code, overdueDays, pluralizeDaysRu(overdueDays)),
		RelatedTaskID: &t.ID,
	}
	if t.IsCriticalPath {
		impact := overdueDays
		risk.ImpactDays = &impact
	}
	return risk
}

// deadlineRisk — риск, не привязанный к конкретной задаче: прогнозируемый
// финиш проекта (по CPM) уже позже дедлайна.
func deadlineRisk(forecastDeviationDays int) Risk {
	impact := forecastDeviationDays
	return Risk{
		Severity:    deadlineRiskSeverity(forecastDeviationDays),
		Title:       "Прогноз срока сдачи хуже дедлайна",
		Description: fmt.Sprintf("По текущему плану проект завершится на %d %s позже дедлайна.", forecastDeviationDays, pluralizeDaysRu(forecastDeviationDays)),
		ImpactDays:  &impact,
	}
}

// pluralizeDaysRu возвращает верную форму слова «день» для числа n
// (учитывает и отрицательные значения — берёт модуль).
func pluralizeDaysRu(n int) string {
	abs := n
	if abs < 0 {
		abs = -abs
	}
	abs %= 100
	tail := abs % 10
	switch {
	case abs > 10 && abs < 20:
		return "дней"
	case tail == 1:
		return "день"
	case tail >= 2 && tail <= 4:
		return "дня"
	default:
		return "дней"
	}
}

func deadlineRiskSeverity(days int) string {
	switch {
	case days >= 7:
		return "critical"
	case days >= 3:
		return "high"
	default:
		return "medium"
	}
}

func adherenceStatus(forecastDeviationDays, projectBufferDays int) string {
	switch {
	case forecastDeviationDays > 0:
		return "delayed"
	case projectBufferDays <= bufferCriticalThresholdDays:
		return "at_risk"
	default:
		return "on_track"
	}
}

// projectBufferDays — «остаток общего резерва сроков проекта» из спецификации:
// сколько дней лежит между прогнозом окончания (CPM) и дедлайном. Проект,
// который уже не успевает, резерва не имеет — 0, отставание показывает
// forecastDeviationDays.
//
// Раньше здесь брался наименьший slack среди задач, но CPM считает slack от
// прогнозного финиша, а не от дедлайна, — у задач критического пути он всегда
// 0. Резерв выходил нулевым у любого проекта, и даже проект с месячным запасом
// попадал в at_risk.
func projectBufferDays(forecastDeviationDays int) int {
	return max(0, -forecastDeviationDays)
}

// tasksPerWeek — грубая оценка скорости команды: сколько задач в среднем
// завершается за неделю, по фактическим датам закрытия за последние 14 дней.
// Точнее без исторических данных о трудозатратах не посчитать — это оценка,
// а не измерение.
func tasksPerWeek(tasks []models.Task) float64 {
	const windowDays = 14
	today := models.Today()
	windowStart := today.AddDays(-windowDays)

	count := 0
	for _, t := range tasks {
		if t.ActualEndDate == nil {
			continue
		}
		if !t.ActualEndDate.Before(windowStart) && !t.ActualEndDate.After(today) {
			count++
		}
	}
	return float64(count) / (float64(windowDays) / 7)
}
