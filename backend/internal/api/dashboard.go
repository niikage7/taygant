package api

import "github.com/gofiber/fiber/v2"

// registerDashboardRoutes — тег Dashboard: сводные показатели состояния и рисков проекта.
func (a *API) registerDashboardRoutes(r fiber.Router) {
	r.Get("/projects/:projectId/dashboard", a.dashboardGet)
}

// GET /projects/{projectId}/dashboard — агрегированные метрики для экрана «Обзор и Аналитика».
// 200 -> ProjectDashboard.
func (a *API) dashboardGet(c *fiber.Ctx) error {
	return notImplemented(c)
}
