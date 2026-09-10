package api

import (
	"github.com/google/uuid"

	"taygant_backend/internal/models"
)

// taskDTO — схема Task.
type taskDTO struct {
	ID                        uuid.UUID         `json:"id"`
	ProjectID                 uuid.UUID         `json:"projectId"`
	SprintID                  *uuid.UUID        `json:"sprintId"`
	ParentTaskID              *uuid.UUID        `json:"parentTaskId"`
	Code                      string            `json:"code"`
	WBSNumber                 string            `json:"wbsNumber"`
	Title                     string            `json:"title"`
	Assignee                  *userDTO          `json:"assignee"`
	Status                    models.TaskStatus `json:"status"`
	IsMilestone               bool              `json:"isMilestone"`
	StartDate                 models.Date       `json:"startDate"`
	EndDate                   models.Date       `json:"endDate"`
	DurationCalendarDays      int               `json:"durationCalendarDays"`
	DurationWorkingDays       int               `json:"durationWorkingDays"`
	ProgressPercent           int               `json:"progressPercent"`
	WeightPercent             float64           `json:"weightPercent"`
	IsCriticalPath            bool              `json:"isCriticalPath"`
	BufferDays                int               `json:"bufferDays"`
	PlanVsActualDeviationDays int               `json:"planVsActualDeviationDays"`
}

func newTaskDTO(t models.Task) taskDTO {
	return taskDTO{
		ID:                        t.ID,
		ProjectID:                 t.ProjectID,
		SprintID:                  t.SprintID,
		ParentTaskID:              t.ParentTaskID,
		Code:                      t.Code,
		WBSNumber:                 t.WBSNumber,
		Title:                     t.Title,
		Assignee:                  newUserDTOPtr(t.Assignee),
		Status:                    t.Status,
		IsMilestone:               t.IsMilestone,
		StartDate:                 t.StartDate,
		EndDate:                   t.EndDate,
		DurationCalendarDays:      t.DurationCalendarDays,
		DurationWorkingDays:       t.DurationWorkingDays,
		ProgressPercent:           t.ProgressPercent,
		WeightPercent:             t.WeightPercent,
		IsCriticalPath:            t.IsCriticalPath,
		BufferDays:                t.BufferDays,
		PlanVsActualDeviationDays: t.PlanVsActualDeviation(),
	}
}

// externalLinkDTO — вложенный объект TaskDetail.externalLink.
type externalLinkDTO struct {
	Provider    string `json:"provider"`
	URL         string `json:"url"`
	ReferenceID string `json:"referenceId"`
	Synced      bool   `json:"synced"`
}

// taskDetailDTO — схема TaskDetail.
type taskDetailDTO struct {
	taskDTO
	Description   string              `json:"description"`
	ExternalLink  *externalLinkDTO    `json:"externalLink"`
	Checklist     []checklistItemDTO  `json:"checklist"`
	Predecessors  []taskDependencyDTO `json:"predecessors"`
	Successors    []taskDependencyDTO `json:"successors"`
	CommentsCount int                 `json:"commentsCount"`
}

func newTaskDetailDTO(t models.Task, checklist []models.ChecklistItem, predecessors, successors []models.TaskDependency, commentsCount int) taskDetailDTO {
	var link *externalLinkDTO
	if t.ExternalProvider != "" {
		link = &externalLinkDTO{
			Provider:    t.ExternalProvider,
			URL:         t.ExternalURL,
			ReferenceID: t.ExternalReferenceID,
			Synced:      t.ExternalSynced,
		}
	}

	checklistDTOs := make([]checklistItemDTO, 0, len(checklist))
	for _, c := range checklist {
		checklistDTOs = append(checklistDTOs, newChecklistItemDTO(c))
	}
	predDTOs := make([]taskDependencyDTO, 0, len(predecessors))
	for _, d := range predecessors {
		predDTOs = append(predDTOs, newTaskDependencyDTO(d))
	}
	succDTOs := make([]taskDependencyDTO, 0, len(successors))
	for _, d := range successors {
		succDTOs = append(succDTOs, newTaskDependencyDTO(d))
	}

	return taskDetailDTO{
		taskDTO:       newTaskDTO(t),
		Description:   t.Description,
		ExternalLink:  link,
		Checklist:     checklistDTOs,
		Predecessors:  predDTOs,
		Successors:    succDTOs,
		CommentsCount: commentsCount,
	}
}

// taskDependencyDTO — схема TaskDependency.
//
// IsCritical пока всегда false: признак критического пути связи требует
// CPM-движка (internal/schedule), которого в проекте ещё нет.
type taskDependencyDTO struct {
	ID                uuid.UUID             `json:"id"`
	PredecessorTaskID uuid.UUID             `json:"predecessorTaskId"`
	SuccessorTaskID   uuid.UUID             `json:"successorTaskId"`
	PredecessorTitle  string                `json:"predecessorTitle"`
	SuccessorTitle    string                `json:"successorTitle"`
	Type              models.DependencyType `json:"type"`
	LagDays           int                   `json:"lagDays"`
	IsCritical        bool                  `json:"isCritical"`
}

func newTaskDependencyDTO(d models.TaskDependency) taskDependencyDTO {
	dto := taskDependencyDTO{
		ID:                d.ID,
		PredecessorTaskID: d.PredecessorTaskID,
		SuccessorTaskID:   d.SuccessorTaskID,
		Type:              d.Type,
		LagDays:           d.LagDays,
	}
	if d.PredecessorTask != nil {
		dto.PredecessorTitle = d.PredecessorTask.Title
	}
	if d.SuccessorTask != nil {
		dto.SuccessorTitle = d.SuccessorTask.Title
	}
	return dto
}

// checklistItemDTO — схема ChecklistItem.
type checklistItemDTO struct {
	ID     uuid.UUID `json:"id"`
	TaskID uuid.UUID `json:"taskId"`
	Text   string    `json:"text"`
	IsDone bool      `json:"isDone"`
}

func newChecklistItemDTO(c models.ChecklistItem) checklistItemDTO {
	return checklistItemDTO{ID: c.ID, TaskID: c.TaskID, Text: c.Text, IsDone: c.IsDone}
}

// commentDTO — схема Comment.
type commentDTO struct {
	ID        uuid.UUID `json:"id"`
	TaskID    uuid.UUID `json:"taskId"`
	Author    userDTO   `json:"author"`
	Text      string    `json:"text"`
	CreatedAt string    `json:"createdAt"`
}

func newCommentDTO(c models.TaskComment) commentDTO {
	var author userDTO
	if c.Author != nil {
		author = newUserDTO(*c.Author)
	}
	return commentDTO{
		ID:        c.ID,
		TaskID:    c.TaskID,
		Author:    author,
		Text:      c.Text,
		CreatedAt: c.CreatedAt.Format(rfc3339),
	}
}

// historyEntryDTO — схема HistoryEntry.
type historyEntryDTO struct {
	ID        uuid.UUID `json:"id"`
	TaskID    uuid.UUID `json:"taskId"`
	Actor     userDTO   `json:"actor"`
	Action    string    `json:"action"`
	Field     *string   `json:"field"`
	OldValue  *string   `json:"oldValue"`
	NewValue  *string   `json:"newValue"`
	CreatedAt string    `json:"createdAt"`
}

func newHistoryEntryDTO(h models.TaskHistoryEntry) historyEntryDTO {
	var actor userDTO
	if h.Actor != nil {
		actor = newUserDTO(*h.Actor)
	}
	return historyEntryDTO{
		ID:        h.ID,
		TaskID:    h.TaskID,
		Actor:     actor,
		Action:    h.Action,
		Field:     h.Field,
		OldValue:  h.OldValue,
		NewValue:  h.NewValue,
		CreatedAt: h.CreatedAt.Format(rfc3339),
	}
}
