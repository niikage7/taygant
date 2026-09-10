package api

import "github.com/gofiber/fiber/v2"

// registerMilestonesRoutes — тег Milestones: контрольные точки (вехи) проекта.
func (a *API) registerMilestonesRoutes(r fiber.Router) {
	projectMilestones := r.Group("/projects/:projectId/milestones")
	projectMilestones.Get("/", a.milestonesList)
	projectMilestones.Post("/", a.milestonesCreate)

	milestones := r.Group("/milestones/:milestoneId")
	milestones.Patch("/", a.milestonesUpdate)
	milestones.Delete("/", a.milestonesDelete)
}

// GET /projects/{projectId}/milestones — все вехи проекта в хронологическом порядке.
// 200 -> []Milestone.
func (a *API) milestonesList(c *fiber.Ctx) error {
	return notImplemented(c)
}

// POST /projects/{projectId}/milestones — создать контрольную точку с плановой датой.
// Body: MilestoneCreateRequest. 201 -> Milestone.
func (a *API) milestonesCreate(c *fiber.Ctx) error {
	return notImplemented(c)
}

// PATCH /milestones/{milestoneId} — изменить название, код или плановую дату вехи.
// Body: MilestoneCreateRequest. 200 -> Milestone.
func (a *API) milestonesUpdate(c *fiber.Ctx) error {
	return notImplemented(c)
}

// DELETE /milestones/{milestoneId} — удалить веху из проекта. 204.
func (a *API) milestonesDelete(c *fiber.Ctx) error {
	return notImplemented(c)
}
