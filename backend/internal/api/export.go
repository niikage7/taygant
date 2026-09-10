package api

import "github.com/gofiber/fiber/v2"

// registerExportRoutes — тег Export: экспорт данных проекта во внешние форматы.
func (a *API) registerExportRoutes(r fiber.Router) {
	r.Get("/projects/:projectId/export", a.exportProject)
}

// GET /projects/{projectId}/export — файл выгрузки проекта (реестр задач, Гант, сводка).
// Query: format (pdf|xlsx, обязательный). 200 -> бинарное содержимое файла.
func (a *API) exportProject(c *fiber.Ctx) error {
	return notImplemented(c)
}
