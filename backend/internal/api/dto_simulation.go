package api

import (
	"github.com/google/uuid"

	"taygant_backend/internal/models"
	"taygant_backend/internal/service"
)

// shiftAffectedTaskDTO — схема ShiftAffectedTask.
type shiftAffectedTaskDTO struct {
	TaskID            uuid.UUID             `json:"taskId"`
	Title             string                `json:"title"`
	IsCriticalPath    bool                  `json:"isCriticalPath"`
	OriginalStartDate models.Date           `json:"originalStartDate"`
	OriginalEndDate   models.Date           `json:"originalEndDate"`
	NewStartDate      models.Date           `json:"newStartDate"`
	NewEndDate        models.Date           `json:"newEndDate"`
	DependencyType    models.DependencyType `json:"dependencyType"`
}

// shiftSimulationDTO — схема ShiftSimulation.
type shiftSimulationDTO struct {
	SourceTaskID          uuid.UUID              `json:"sourceTaskId"`
	ShiftDays             int                    `json:"shiftDays"`
	Algorithm             string                 `json:"algorithm"`
	AffectedTasks         []shiftAffectedTaskDTO `json:"affectedTasks"`
	ProjectDeadlineImpact struct {
		OriginalDeadline models.Date `json:"originalDeadline"`
		NewDeadline      models.Date `json:"newDeadline"`
		DeltaDays        int         `json:"deltaDays"`
	} `json:"projectDeadlineImpact"`
	BufferAvailableDays int  `json:"bufferAvailableDays"`
	Applied             bool `json:"applied"`
}

func newShiftSimulationDTO(r service.Result) shiftSimulationDTO {
	affected := make([]shiftAffectedTaskDTO, 0, len(r.AffectedTasks))
	for _, t := range r.AffectedTasks {
		affected = append(affected, shiftAffectedTaskDTO{
			TaskID:            t.TaskID,
			Title:             t.Title,
			IsCriticalPath:    t.IsCriticalPath,
			OriginalStartDate: t.OriginalStart,
			OriginalEndDate:   t.OriginalEnd,
			NewStartDate:      t.NewStart,
			NewEndDate:        t.NewEnd,
			DependencyType:    t.DependencyType,
		})
	}

	dto := shiftSimulationDTO{
		SourceTaskID:        r.SourceTaskID,
		ShiftDays:           r.ShiftDays,
		Algorithm:           r.Algorithm,
		AffectedTasks:       affected,
		BufferAvailableDays: r.BufferAvailableDays,
		Applied:             r.Applied,
	}
	dto.ProjectDeadlineImpact.OriginalDeadline = r.OriginalDeadline
	dto.ProjectDeadlineImpact.NewDeadline = r.NewDeadline
	dto.ProjectDeadlineImpact.DeltaDays = r.DeadlineDeltaDays
	return dto
}
