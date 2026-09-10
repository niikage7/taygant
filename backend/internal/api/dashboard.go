package api

import "github.com/gofiber/fiber/v2"

// registerDashboardRoutes — тег Dashboard: сводные показатели состояния и рисков проекта.
func registerDashboardRoutes(r fiber.Router) {
	r.Get("/projects/:projectId/dashboard", dashboardGet)
}

// GET /projects/{projectId}/dashboard — агрегированные метрики для экрана «Обзор и Аналитика».
// 200 -> ProjectDashboard.
func dashboardGet(c *fiber.Ctx) error {
	return notImplemented(c)
}
