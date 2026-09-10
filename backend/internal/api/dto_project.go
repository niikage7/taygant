package api

import (
	"github.com/google/uuid"

	"taygant_backend/internal/models"
	"taygant_backend/internal/service"
)

// projectSummaryDTO — схема ProjectSummary.
type projectSummaryDTO struct {
	ID                 uuid.UUID            `json:"id"`
	Code               string               `json:"code"`
	Name               string               `json:"name"`
	Status             models.ProjectStatus `json:"status"`
	Phase              string               `json:"phase"`
	StartDate          models.Date          `json:"startDate"`
	Deadline           models.Date          `json:"deadline"`
	ProgressPercent    float64              `json:"progressPercent"`
	HealthIndex        int                  `json:"healthIndex"`
	CriticalRisksCount int                  `json:"criticalRisksCount"`
}

func newProjectSummaryDTO(p models.Project, m service.Metrics) projectSummaryDTO {
	return projectSummaryDTO{
		ID:                 p.ID,
		Code:               p.Code,
		Name:               p.Name,
		Status:             p.Status,
		Phase:              p.Phase,
		StartDate:          p.StartDate,
		Deadline:           p.Deadline,
		ProgressPercent:    m.ProgressPercent,
		HealthIndex:        m.HealthIndex,
		CriticalRisksCount: m.CriticalRisksCount,
	}
}

// projectDTO — схема Project (ProjectSummary + параметры календаря и Ганта).
type projectDTO struct {
	projectSummaryDTO
	Description               string                     `json:"description"`
	CustomerOrg               string                     `json:"customerOrg"`
	DurationCalendarDays      int                        `json:"durationCalendarDays"`
	WorkingCalendarType       models.WorkingCalendarType `json:"workingCalendarType"`
	IncludePublicHolidays     bool                       `json:"includePublicHolidays"`
	AutoRecalculateDependents bool                       `json:"autoRecalculateDependents"`
	HighlightCriticalPath     bool                       `json:"highlightCriticalPath"`
	CreatedBy                 *userDTO                   `json:"createdBy"`
	CreatedAt                 string                     `json:"createdAt"`
	UpdatedAt                 string                     `json:"updatedAt"`
}

func newProjectDTO(p models.Project, m service.Metrics) projectDTO {
	return projectDTO{
		projectSummaryDTO:         newProjectSummaryDTO(p, m),
		Description:               p.Description,
		CustomerOrg:               p.CustomerOrg,
		DurationCalendarDays:      p.StartDate.DaysUntil(p.Deadline),
		WorkingCalendarType:       p.WorkingCalendarType,
		IncludePublicHolidays:     p.IncludePublicHolidays,
		AutoRecalculateDependents: p.AutoRecalculateDependents,
		HighlightCriticalPath:     p.HighlightCriticalPath,
		CreatedBy:                 newUserDTOPtr(p.CreatedBy),
		CreatedAt:                 p.CreatedAt.Format(rfc3339),
		UpdatedAt:                 p.UpdatedAt.Format(rfc3339),
	}
}

// projectMemberDTO — схема ProjectMember.
type projectMemberDTO struct {
	ID               uuid.UUID          `json:"id"`
	User             userDTO            `json:"user"`
	ProjectRole      models.ProjectRole `json:"projectRole"`
	SprintRole       string             `json:"sprintRole"`
	AccessLevel      models.AccessLevel `json:"accessLevel"`
	WeeklyHoursLimit int                `json:"weeklyHoursLimit"`
}

func newProjectMemberDTO(m models.ProjectMember) projectMemberDTO {
	var user userDTO
	if m.User != nil {
		user = newUserDTO(*m.User)
	}
	return projectMemberDTO{
		ID:               m.ID,
		User:             user,
		ProjectRole:      m.ProjectRole,
		SprintRole:       m.SprintRole,
		AccessLevel:      m.AccessLevel,
		WeeklyHoursLimit: m.WeeklyHoursLimit,
	}
}

// sprintDTO — схема Sprint.
type sprintDTO struct {
	ID             uuid.UUID   `json:"id"`
	ProjectID      uuid.UUID   `json:"projectId"`
	Name           string      `json:"name"`
	ReleaseVersion string      `json:"releaseVersion"`
	StartDate      models.Date `json:"startDate"`
	EndDate        models.Date `json:"endDate"`
}

func newSprintDTO(s models.Sprint) sprintDTO {
	return sprintDTO{
		ID:             s.ID,
		ProjectID:      s.ProjectID,
		Name:           s.Name,
		ReleaseVersion: s.ReleaseVersion,
		StartDate:      s.StartDate,
		EndDate:        s.EndDate,
	}
}

// milestoneDTO — схема Milestone.
type milestoneDTO struct {
	ID          uuid.UUID              `json:"id"`
	ProjectID   uuid.UUID              `json:"projectId"`
	Code        string                 `json:"code"`
	Name        string                 `json:"name"`
	PlannedDate models.Date            `json:"plannedDate"`
	ActualDate  *models.Date           `json:"actualDate"`
	Status      models.MilestoneStatus `json:"status"`
	RiskDays    *int                   `json:"riskDays"`
}

func newMilestoneDTO(m models.Milestone) milestoneDTO {
	return milestoneDTO{
		ID:          m.ID,
		ProjectID:   m.ProjectID,
		Code:        m.Code,
		Name:        m.Name,
		PlannedDate: m.PlannedDate,
		ActualDate:  m.ActualDate,
		Status:      m.Status,
		RiskDays:    m.RiskDays,
	}
}

// rfc3339 — формат date-time из api-spec.yml.
const rfc3339 = "2006-01-02T15:04:05Z07:00"
