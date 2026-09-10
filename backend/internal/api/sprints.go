package api

import "github.com/gofiber/fiber/v2"

// registerSprintsRoutes — тег Sprints: спринты и релизы внутри проекта.
func (a *API) registerSprintsRoutes(r fiber.Router) {
	projectSprints := r.Group("/projects/:projectId/sprints")
	projectSprints.Get("/", a.sprintsList)
	projectSprints.Post("/", a.sprintsCreate)

	sprints := r.Group("/sprints/:sprintId")
	sprints.Patch("/", a.sprintsUpdate)
	sprints.Delete("/", a.sprintsDelete)
}

// GET /projects/{projectId}/sprints — спринты проекта, используемые для группировки задач.
// 200 -> []Sprint.
func (a *API) sprintsList(c *fiber.Ctx) error {
	return notImplemented(c)
}

// POST /projects/{projectId}/sprints — создать спринт/релиз в рамках проекта.
// Body: SprintCreateRequest. 201 -> Sprint.
func (a *API) sprintsCreate(c *fiber.Ctx) error {
	return notImplemented(c)
}

// PATCH /sprints/{sprintId} — изменить название, версию релиза или даты спринта.
// Body: SprintCreateRequest. 200 -> Sprint.
func (a *API) sprintsUpdate(c *fiber.Ctx) error {
	return notImplemented(c)
}

// DELETE /sprints/{sprintId} — удалить спринт.
// Связанные задачи остаются в проекте без привязки к спринту. 204.
func (a *API) sprintsDelete(c *fiber.Ctx) error {
	return notImplemented(c)
}
