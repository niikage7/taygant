// Package api содержит HTTP-слой приложения: объявление и реализацию всех
// эндпоинтов, описанных в api-spec.yml. Файлы разбиты по тегам спецификации.
package api

import "github.com/gofiber/fiber/v2"

// Register монтирует все маршруты API на базовый путь /api/v1.
// Переданные middleware применяются ко всей группе /api/v1, но не к health-check.
func Register(app *fiber.App, middlewares ...fiber.Handler) {
	v1 := app.Group("/api/v1", middlewares...)

	registerAuthRoutes(v1)
	registerUsersRoutes(v1)
	registerProjectsRoutes(v1)
	registerMembersRoutes(v1)
	registerSprintsRoutes(v1)
	registerMilestonesRoutes(v1)
	registerTasksRoutes(v1)
	registerDependenciesRoutes(v1)
	registerChecklistRoutes(v1)
	registerCommentsRoutes(v1)
	registerHistoryRoutes(v1)
	registerSimulationRoutes(v1)
	registerWorkloadRoutes(v1)
	registerDashboardRoutes(v1)
	registerExportRoutes(v1)
}

// notImplemented — временная заглушка для эндпоинтов, реализация которых ждёт
// готовности БД и аутентификации.
func notImplemented(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": "not implemented",
		"path":  c.Path(),
	})
}
