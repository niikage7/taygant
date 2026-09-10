package api

import "github.com/gofiber/fiber/v2"

// registerWorkloadRoutes — тег Workload: загрузка участников команды по часам.
func (a *API) registerWorkloadRoutes(r fiber.Router) {
	r.Get("/projects/:projectId/workload", a.workloadList)
}

// GET /projects/{projectId}/workload — загрузка команды по спринту.
// Возвращает загруженность участников в часах относительно недельного лимита.
// Query: sprintId (по умолчанию — текущий спринт). 200 -> []MemberWorkload.
func (a *API) workloadList(c *fiber.Ctx) error {
	return notImplemented(c)
}
