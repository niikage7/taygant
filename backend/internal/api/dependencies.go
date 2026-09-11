package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"taygant_backend/internal/models"
)

// registerDependenciesRoutes — тег Dependencies: зависимости (связи) между задачами.
func (a *API) registerDependenciesRoutes(r fiber.Router) {
	taskDependencies := r.Group("/tasks/:taskId/dependencies")
	taskDependencies.Get("/", a.dependenciesList)
	taskDependencies.Post("/", a.dependenciesCreate)

	r.Delete("/dependencies/:dependencyId", a.dependenciesDelete)
}

// GET /tasks/{taskId}/dependencies — входящие и исходящие связи задачи для сетевого графа.
// 200 -> { predecessors: []TaskDependency, successors: []TaskDependency }.
func (a *API) dependenciesList(c *fiber.Ctx) error {
	taskID, err := pathUUID(c, "taskId")
	if err != nil {
		return err
	}
	projectID, err := a.requireMemberByTask(c, taskID)
	if err != nil {
		return err
	}

	predecessors, successors, err := a.dependencies.List(c.Context(), taskID)
	if err != nil {
		return fail(err)
	}
	critical, err := a.tasks.CriticalTaskIDs(c.Context(), projectID)
	if err != nil {
		return fail(err)
	}

	predDTOs := make([]taskDependencyDTO, 0, len(predecessors))
	for _, d := range predecessors {
		predDTOs = append(predDTOs, newTaskDependencyDTO(d, critical))
	}
	succDTOs := make([]taskDependencyDTO, 0, len(successors))
	for _, d := range successors {
		succDTOs = append(succDTOs, newTaskDependencyDTO(d, critical))
	}
	return c.JSON(fiber.Map{"predecessors": predDTOs, "successors": succDTOs})
}

// dependencyCreateRequest — тело POST /tasks/{taskId}/dependencies.
type dependencyCreateRequest struct {
	RelatedTaskID uuid.UUID             `json:"relatedTaskId"`
	Direction     string                `json:"direction"`
	Type          models.DependencyType `json:"type"`
	LagDays       int                   `json:"lagDays"`
}

// POST /tasks/{taskId}/dependencies — связать задачу с другой задачей проекта.
// Body: TaskDependencyCreateRequest. 201 -> TaskDependency.
func (a *API) dependenciesCreate(c *fiber.Ctx) error {
	taskID, err := pathUUID(c, "taskId")
	if err != nil {
		return err
	}
	projectID, err := a.requireManageByTask(c, taskID)
	if err != nil {
		return err
	}

	body, err := parseBody[dependencyCreateRequest](c)
	if err != nil {
		return err
	}

	dep, err := a.dependencies.Create(c.Context(), projectID, currentUserID(c), taskID, body.RelatedTaskID, body.Direction, body.Type, body.LagDays)
	if err != nil {
		return fail(err)
	}
	critical, err := a.tasks.CriticalTaskIDs(c.Context(), projectID)
	if err != nil {
		return fail(err)
	}
	return c.Status(fiber.StatusCreated).JSON(newTaskDependencyDTO(dep, critical))
}

// DELETE /dependencies/{dependencyId} — разорвать зависимость между двумя задачами.
// Сами задачи не удаляются. 204.
func (a *API) dependenciesDelete(c *fiber.Ctx) error {
	dependencyID, err := pathUUID(c, "dependencyId")
	if err != nil {
		return err
	}
	projectID, err := a.dependencies.ProjectID(c.Context(), dependencyID)
	if err != nil {
		return fail(err)
	}
	if _, err := a.requireManage(c, projectID); err != nil {
		return err
	}

	if err := a.dependencies.Delete(c.Context(), currentUserID(c), dependencyID); err != nil {
		return fail(err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}
