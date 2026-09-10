package api

import (
	"github.com/gofiber/fiber/v2"

	"taygant_backend/internal/models"
	"taygant_backend/internal/service"
)

// registerSprintsRoutes — тег Sprints: спринты и релизы внутри проекта.
func (a *API) registerSprintsRoutes(r fiber.Router) {
	projectSprints := r.Group("/projects/:projectId/sprints")
	projectSprints.Get("/", a.sprintsList)
	projectSprints.Post("/", a.sprintsCreate)

	sprints := r.Group("/sprints/:sprintId")
	sprints.Patch("/", a.sprintsUpdate)
	sprints.Delete("/", a.sprintsDelete)
}

// sprintRequest — тело создания/изменения спринта (общий формат для POST и PATCH).
type sprintRequest struct {
	Name           string      `json:"name"`
	ReleaseVersion string      `json:"releaseVersion"`
	StartDate      models.Date `json:"startDate"`
	EndDate        models.Date `json:"endDate"`
}

func (r sprintRequest) toInput() service.SprintInput {
	return service.SprintInput{
		Name:           r.Name,
		ReleaseVersion: r.ReleaseVersion,
		StartDate:      r.StartDate,
		EndDate:        r.EndDate,
	}
}

// GET /projects/{projectId}/sprints — спринты проекта, используемые для группировки задач.
// 200 -> []Sprint.
func (a *API) sprintsList(c *fiber.Ctx) error {
	projectID, err := pathUUID(c, "projectId")
	if err != nil {
		return err
	}
	if _, err := a.requireMember(c, projectID); err != nil {
		return err
	}

	sprints, err := a.sprints.List(c.Context(), projectID)
	if err != nil {
		return fail(err)
	}
	out := make([]sprintDTO, 0, len(sprints))
	for _, s := range sprints {
		out = append(out, newSprintDTO(s))
	}
	return c.JSON(out)
}

// POST /projects/{projectId}/sprints — создать спринт/релиз в рамках проекта.
// Body: SprintCreateRequest. 201 -> Sprint.
func (a *API) sprintsCreate(c *fiber.Ctx) error {
	projectID, err := pathUUID(c, "projectId")
	if err != nil {
		return err
	}
	if err := a.requireEdit(c, projectID); err != nil {
		return err
	}

	body, err := parseBody[sprintRequest](c)
	if err != nil {
		return err
	}
	sprint, err := a.sprints.Create(c.Context(), projectID, body.toInput())
	if err != nil {
		return fail(err)
	}
	return c.Status(fiber.StatusCreated).JSON(newSprintDTO(sprint))
}

// PATCH /sprints/{sprintId} — изменить название, версию релиза или даты спринта.
// Body: SprintCreateRequest. 200 -> Sprint.
func (a *API) sprintsUpdate(c *fiber.Ctx) error {
	sprintID, err := pathUUID(c, "sprintId")
	if err != nil {
		return err
	}
	projectID, err := a.sprints.ProjectID(c.Context(), sprintID)
	if err != nil {
		return fail(err)
	}
	if err := a.requireEdit(c, projectID); err != nil {
		return err
	}

	body, err := parseBody[sprintRequest](c)
	if err != nil {
		return err
	}
	sprint, err := a.sprints.Update(c.Context(), sprintID, body.toInput())
	if err != nil {
		return fail(err)
	}
	return c.JSON(newSprintDTO(sprint))
}

// DELETE /sprints/{sprintId} — удалить спринт.
// Связанные задачи остаются в проекте без привязки к спринту. 204.
func (a *API) sprintsDelete(c *fiber.Ctx) error {
	sprintID, err := pathUUID(c, "sprintId")
	if err != nil {
		return err
	}
	projectID, err := a.sprints.ProjectID(c.Context(), sprintID)
	if err != nil {
		return fail(err)
	}
	if err := a.requireEdit(c, projectID); err != nil {
		return err
	}

	if err := a.sprints.Delete(c.Context(), sprintID); err != nil {
		return fail(err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}
