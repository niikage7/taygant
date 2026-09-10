package api

import "github.com/gofiber/fiber/v2"

// registerProjectsRoutes — тег Projects: проекты и их параметры.
func registerProjectsRoutes(r fiber.Router) {
	projects := r.Group("/projects")

	projects.Get("/", projectsList)
	projects.Post("/", projectsCreate)
	projects.Get("/:projectId", projectsGet)
	projects.Patch("/:projectId", projectsUpdate)
	projects.Delete("/:projectId", projectsDelete)
}

// GET /projects — список проектов пользователя с краткими сводными метриками.
// Query: status (ProjectStatus), search. 200 -> []ProjectSummary.
func projectsList(c *fiber.Ctx) error {
	return notImplemented(c)
}

// POST /projects — создать проект (шаг 1 мастера инициации).
// Создаёт проект вместе с параметрами Ганта и опциональным начальным составом команды.
// Body: ProjectCreateRequest. 201 -> Project.
func projectsCreate(c *fiber.Ctx) error {
	return notImplemented(c)
}

// GET /projects/{projectId} — полная карточка проекта, включая календарь и настройки Ганта.
// 200 -> Project, 404 -> не найден.
func projectsGet(c *fiber.Ctx) error {
	return notImplemented(c)
}

// PATCH /projects/{projectId} — изменить параметры проекта (сроки, статус, описание, календарь).
// Body: ProjectUpdateRequest. 200 -> Project.
func projectsUpdate(c *fiber.Ctx) error {
	return notImplemented(c)
}

// DELETE /projects/{projectId} — удалить (архивировать) проект.
// Переводит проект в статус archived; данные не удаляются физически. 204.
func projectsDelete(c *fiber.Ctx) error {
	return notImplemented(c)
}
