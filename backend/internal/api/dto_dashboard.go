package api

import (
	"github.com/google/uuid"

	"taygant_backend/internal/service"
)

type scheduleAdherenceDTO struct {
	Status                string  `json:"status"`
	ForecastDeviationDays int     `json:"forecastDeviationDays"`
	ProjectBufferDays     float64 `json:"projectBufferDays"`
}

type taskStatusBreakdownDTO struct {
	Total      int `json:"total"`
	Done       int `json:"done"`
	InProgress int `json:"inProgress"`
	Planned    int `json:"planned"`
	Overdue    int `json:"overdue"`
}

type teamVelocityDTO struct {
	TasksPerWeek float64 `json:"tasksPerWeek"`
}

type projectRiskDTO struct {
	Severity      string     `json:"severity"`
	Title         string     `json:"title"`
	Description   string     `json:"description"`
	RelatedTaskID *uuid.UUID `json:"relatedTaskId"`
	ImpactDays    *int       `json:"impactDays"`
}

type attentionTaskDTO struct {
	Task          taskDTO `json:"task"`
	DeviationDays int     `json:"deviationDays"`
}

// memberWorkloadDTO — схема MemberWorkload.
type memberWorkloadDTO struct {
	User                 userDTO    `json:"user"`
	SprintID             *uuid.UUID `json:"sprintId"`
	AssignedHoursPerWeek float64    `json:"assignedHoursPerWeek"`
	WeeklyHoursLimit     int        `json:"weeklyHoursLimit"`
	UtilizationPercent   int        `json:"utilizationPercent"`
}

func newMemberWorkloadDTO(w service.MemberWorkload) memberWorkloadDTO {
	var user userDTO
	if w.Member.User != nil {
		user = newUserDTO(*w.Member.User)
	}
	return memberWorkloadDTO{
		User:                 user,
		SprintID:             w.SprintID,
		AssignedHoursPerWeek: w.AssignedHoursPerWeek,
		WeeklyHoursLimit:     w.Member.WeeklyHoursLimit,
		UtilizationPercent:   w.UtilizationPercent,
	}
}

// statusChangesDTO — схема StatusChangeStats: смены статусов за период и
// сколько из них сделано через ассистента (MCP).
type statusChangesDTO struct {
	PeriodDays   int `json:"periodDays"`
	Total        int `json:"total"`
	ViaAssistant int `json:"viaAssistant"`
}

// projectDashboardDTO — схема ProjectDashboard.
type projectDashboardDTO struct {
	Project                projectSummaryDTO      `json:"project"`
	OverallProgressPercent float64                `json:"overallProgressPercent"`
	ProgressDeltaPercent   float64                `json:"progressDeltaPercent"`
	MilestonesClosed       int                    `json:"milestonesClosed"`
	MilestonesTotal        int                    `json:"milestonesTotal"`
	ScheduleAdherence      scheduleAdherenceDTO   `json:"scheduleAdherence"`
	TaskStatusBreakdown    taskStatusBreakdownDTO `json:"taskStatusBreakdown"`
	TeamVelocity           teamVelocityDTO        `json:"teamVelocity"`
	NearestMilestones      []milestoneDTO         `json:"nearestMilestones"`
	Risks                  []projectRiskDTO       `json:"risks"`
	AttentionTasks         []attentionTaskDTO     `json:"attentionTasks"`
	TeamWorkload           []memberWorkloadDTO    `json:"teamWorkload"`
	StatusChanges          statusChangesDTO       `json:"statusChanges"`
}

func newProjectDashboardDTO(r service.DashboardResult) projectDashboardDTO {
	risks := make([]projectRiskDTO, 0, len(r.Risks))
	for _, risk := range r.Risks {
		risks = append(risks, projectRiskDTO{
			Severity:      risk.Severity,
			Title:         risk.Title,
			Description:   risk.Description,
			RelatedTaskID: risk.RelatedTaskID,
			ImpactDays:    risk.ImpactDays,
		})
	}

	attention := make([]attentionTaskDTO, 0, len(r.AttentionTasks))
	for _, a := range r.AttentionTasks {
		attention = append(attention, attentionTaskDTO{
			Task:          newTaskDTO(a.Task),
			DeviationDays: a.DeviationDays,
		})
	}

	milestones := make([]milestoneDTO, 0, len(r.NearestMilestones))
	for _, m := range r.NearestMilestones {
		milestones = append(milestones, newMilestoneDTO(m))
	}

	workload := make([]memberWorkloadDTO, 0, len(r.TeamWorkload))
	for _, w := range r.TeamWorkload {
		workload = append(workload, newMemberWorkloadDTO(w))
	}

	return projectDashboardDTO{
		Project:                newProjectSummaryDTO(r.Project, r.Metrics),
		OverallProgressPercent: r.Metrics.ProgressPercent,
		ProgressDeltaPercent:   r.ProgressDeltaPercent,
		MilestonesClosed:       r.MilestonesClosed,
		MilestonesTotal:        r.MilestonesTotal,
		ScheduleAdherence: scheduleAdherenceDTO{
			Status:                r.ScheduleAdherence.Status,
			ForecastDeviationDays: r.ScheduleAdherence.ForecastDeviationDays,
			ProjectBufferDays:     float64(r.ScheduleAdherence.ProjectBufferDays),
		},
		TaskStatusBreakdown: taskStatusBreakdownDTO{
			Total:      r.TaskStatusBreakdown.Total,
			Done:       r.TaskStatusBreakdown.Done,
			InProgress: r.TaskStatusBreakdown.InProgress,
			Planned:    r.TaskStatusBreakdown.Planned,
			Overdue:    r.TaskStatusBreakdown.Overdue,
		},
		TeamVelocity:      teamVelocityDTO{TasksPerWeek: r.TasksPerWeek},
		NearestMilestones: milestones,
		Risks:             risks,
		AttentionTasks:    attention,
		TeamWorkload:      workload,
		StatusChanges: statusChangesDTO{
			PeriodDays:   r.StatusChanges.PeriodDays,
			Total:        r.StatusChanges.Total,
			ViaAssistant: r.StatusChanges.ViaAssistant,
		},
	}
}
