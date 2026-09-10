package api

import (
	"fmt"

	"github.com/gofiber/fiber/v2"

	"taygant_backend/internal/service"
)

// registerExportRoutes — тег Export: экспорт данных проекта во внешние форматы.
func (a *API) registerExportRoutes(r fiber.Router) {
	r.Get("/projects/:projectId/export", a.exportProject)
}

// GET /projects/{projectId}/export — файл выгрузки проекта (реестр задач, Гант, сводка).
// Query: format (pdf|xlsx, обязательный). 200 -> бинарное содержимое файла.
//
// Отсутствующий или недопустимый format ловит service.Export.Generate (пустая
// строка — не валидный ExportFormat) и превращается в 400 через fail(), так
// что обязательность параметра не приходится проверять отдельно здесь же.
func (a *API) exportProject(c *fiber.Ctx) error {
	projectID, err := pathUUID(c, "projectId")
	if err != nil {
		return err
	}
	if _, err := a.requireMember(c, projectID); err != nil {
		return err
	}

	format := service.ExportFormat(c.Query("format"))
	file, err := a.export.Generate(c.Context(), projectID, format)
	if err != nil {
		return fail(err)
	}

	c.Set(fiber.HeaderContentType, file.ContentType)
	c.Set(fiber.HeaderContentDisposition, fmt.Sprintf(`attachment; filename="%s"`, file.FileName))
	return c.Send(file.Content)
}
