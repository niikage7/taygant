package api

import "github.com/gofiber/fiber/v2"

// registerMilestonesRoutes — тег Milestones: контрольные точки (вехи) проекта.
func registerMilestonesRoutes(r fiber.Router) {
	projectMilestones := r.Group("/projects/:projectId/milestones")
	projectMilestones.Get("/", milestonesList)
	projectMilestones.Post("/", milestonesCreate)

	milestones := r.Group("/milestones/:milestoneId")
	milestones.Patch("/", milestonesUpdate)
	milestones.Delete("/", milestonesDelete)
}

// GET /projects/{projectId}/milestones — все вехи проекта в хронологическом порядке.
// 200 -> []Milestone.
func milestonesList(c *fiber.Ctx) error {
	return notImplemented(c)
}

// POST /projects/{projectId}/milestones — создать контрольную точку с плановой датой.
// Body: MilestoneCreateRequest. 201 -> Milestone.
func milestonesCreate(c *fiber.Ctx) error {
	return notImplemented(c)
}

// PATCH /milestones/{milestoneId} — изменить название, код или плановую дату вехи.
// Body: MilestoneCreateRequest. 200 -> Milestone.
func milestonesUpdate(c *fiber.Ctx) error {
	return notImplemented(c)
}

// DELETE /milestones/{milestoneId} — удалить веху из проекта. 204.
func milestonesDelete(c *fiber.Ctx) error {
	return notImplemented(c)
}
