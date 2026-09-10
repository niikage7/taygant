package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"taygant_backend/internal/models"
	"taygant_backend/internal/service"
)

// registerTasksRoutes — тег Tasks: задачи проекта и диаграмма Ганта.
//
// GET /projects/{projectId}/gantt здесь не зарегистрирован: критический путь
// требует CPM-движка (internal/schedule), которого в проекте пока нет.
func (a *API) registerTasksRoutes(r fiber.Router) {
	projectTasks := r.Group("/projects/:projectId")
	projectTasks.Get("/tasks", a.tasksList)
	projectTasks.Post("/tasks", a.tasksCreate)

	tasks := r.Group("/tasks/:taskId")
	tasks.Get("/", a.tasksGet)
	tasks.Patch("/", a.tasksUpdate)
	tasks.Delete("/", a.tasksDelete)
}

// GET /projects/{projectId}/tasks — реестр задач проекта / данные для Ганта.
// Query: sprintId, assigneeId, status, criticalPathOnly, risksOnly, myTasksOnly, search.
//
// criticalPathOnly принимается, но не применяется: критический путь ещё не
// считается (internal/schedule). myTasksOnly подставляет ID текущего
// пользователя в тот же фильтр, что и assigneeId.
// 200 -> []Task.
func (a *API) tasksList(c *fiber.Ctx) error {
	projectID, err := pathUUID(c, "projectId")
	if err != nil {
		return err
	}
	m, err := a.requireMember(c, projectID)
	if err != nil {
		return err
	}

	sprintID, err := queryUUID(c, "sprintId")
	if err != nil {
		return err
	}
	assigneeID, err := queryUUID(c, "assigneeId")
	if err != nil {
		return err
	}
	if queryBool(c, "myTasksOnly") {
		me := currentUserID(c)
		assigneeID = &me
	}
	status := models.TaskStatus(c.Query("status"))
	if status != "" && !status.Valid() {
		return fiber.NewError(fiber.StatusBadRequest, "недопустимый статус задачи")
	}

	tasks, err := a.tasks.List(c.Context(), projectID, m.Project, service.Filter{
		SprintID:   sprintID,
		AssigneeID: assigneeID,
		Status:     status,
		RisksOnly:  queryBool(c, "risksOnly"),
		Search:     c.Query("search"),
	})
	if err != nil {
		return fail(err)
	}

	out := make([]taskDTO, 0, len(tasks))
	for _, t := range tasks {
		out = append(out, newTaskDTO(t))
	}
	return c.JSON(out)
}

// taskPredecessorRequest — элемент TaskCreateRequest.predecessors.
type taskPredecessorRequest struct {
	TaskID  uuid.UUID             `json:"taskId"`
	Type    models.DependencyType `json:"type"`
	LagDays int                   `json:"lagDays"`
}

// taskCreateRequest — тело POST /projects/{projectId}/tasks.
type taskCreateRequest struct {
	SprintID      *uuid.UUID               `json:"sprintId"`
	ParentTaskID  *uuid.UUID               `json:"parentTaskId"`
	Title         string                   `json:"title"`
	Description   string                   `json:"description"`
	AssigneeID    *uuid.UUID               `json:"assigneeId"`
	Status        *models.TaskStatus       `json:"status"`
	IsMilestone   *bool                    `json:"isMilestone"`
	StartDate     models.Date              `json:"startDate"`
	EndDate       models.Date              `json:"endDate"`
	WeightPercent *float64                 `json:"weightPercent"`
	Predecessors  []taskPredecessorRequest `json:"predecessors"`
}

// POST /projects/{projectId}/tasks — добавить задачу, опционально сразу со связями-предшественниками.
// Body: TaskCreateRequest. 201 -> Task.
func (a *API) tasksCreate(c *fiber.Ctx) error {
	projectID, err := pathUUID(c, "projectId")
	if err != nil {
		return err
	}
	if err := a.requireEdit(c, projectID); err != nil {
		return err
	}

	body, err := parseBody[taskCreateRequest](c)
	if err != nil {
		return err
	}

	title := body.Title
	predecessors := make([]service.PredecessorInput, 0, len(body.Predecessors))
	for _, p := range body.Predecessors {
		predecessors = append(predecessors, service.PredecessorInput{TaskID: p.TaskID, Type: p.Type, LagDays: p.LagDays})
	}
	description := body.Description

	task, err := a.tasks.Create(c.Context(), projectID, currentUserID(c), service.TaskInput{
		SprintID:      body.SprintID,
		ParentTaskID:  body.ParentTaskID,
		Title:         &title,
		Description:   &description,
		AssigneeID:    body.AssigneeID,
		Status:        body.Status,
		IsMilestone:   body.IsMilestone,
		StartDate:     &body.StartDate,
		EndDate:       &body.EndDate,
		WeightPercent: body.WeightPercent,
	}, predecessors)
	if err != nil {
		return fail(err)
	}
	return c.Status(fiber.StatusCreated).JSON(newTaskDTO(task))
}

// GET /tasks/{taskId} — карточка задачи со связями, чек-листом DoD и числом комментариев.
// 200 -> TaskDetail, 404 -> не найдена.
func (a *API) tasksGet(c *fiber.Ctx) error {
	taskID, err := pathUUID(c, "taskId")
	if err != nil {
		return err
	}
	if _, err := a.requireMemberByTask(c, taskID); err != nil {
		return err
	}

	task, err := a.tasks.Get(c.Context(), taskID)
	if err != nil {
		return fail(err)
	}
	checklist, err := a.checklist.List(c.Context(), taskID)
	if err != nil {
		return fail(err)
	}
	predecessors, successors, err := a.dependencies.List(c.Context(), taskID)
	if err != nil {
		return fail(err)
	}
	comments, err := a.comments.List(c.Context(), taskID)
	if err != nil {
		return fail(err)
	}

	return c.JSON(newTaskDetailDTO(task, checklist, predecessors, successors, len(comments)))
}

// taskUpdateRequest — тело PATCH /tasks/{taskId}.
type taskUpdateRequest struct {
	SprintID        *uuid.UUID         `json:"sprintId"`
	Title           *string            `json:"title"`
	Description     *string            `json:"description"`
	AssigneeID      *uuid.UUID         `json:"assigneeId"`
	Status          *models.TaskStatus `json:"status"`
	StartDate       *models.Date       `json:"startDate"`
	EndDate         *models.Date       `json:"endDate"`
	ProgressPercent *int               `json:"progressPercent"`
	WeightPercent   *float64           `json:"weightPercent"`
}

// PATCH /tasks/{taskId} — редактировать задачу (сроки, статус, ответственный, прогресс и т.д.).
// Body: TaskUpdateRequest. 200 -> Task.
func (a *API) tasksUpdate(c *fiber.Ctx) error {
	taskID, err := pathUUID(c, "taskId")
	if err != nil {
		return err
	}
	if _, err := a.requireEditByTask(c, taskID); err != nil {
		return err
	}

	body, err := parseBody[taskUpdateRequest](c)
	if err != nil {
		return err
	}

	task, err := a.tasks.Update(c.Context(), taskID, currentUserID(c), service.TaskInput{
		SprintID:        body.SprintID,
		Title:           body.Title,
		Description:     body.Description,
		AssigneeID:      body.AssigneeID,
		Status:          body.Status,
		StartDate:       body.StartDate,
		EndDate:         body.EndDate,
		ProgressPercent: body.ProgressPercent,
		WeightPercent:   body.WeightPercent,
	})
	if err != nil {
		return fail(err)
	}
	return c.JSON(newTaskDTO(task))
}

// DELETE /tasks/{taskId} — удалить задачу вместе со связями, чек-листом и комментариями. 204.
func (a *API) tasksDelete(c *fiber.Ctx) error {
	taskID, err := pathUUID(c, "taskId")
	if err != nil {
		return err
	}
	if _, err := a.requireEditByTask(c, taskID); err != nil {
		return err
	}

	if err := a.tasks.Delete(c.Context(), taskID); err != nil {
		return fail(err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}
