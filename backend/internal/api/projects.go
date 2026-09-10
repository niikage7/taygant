package api

import "github.com/gofiber/fiber/v2"

// registerProjectsRoutes — тег Projects: проекты и их параметры.
func (a *API) registerProjectsRoutes(r fiber.Router) {
	projects := r.Group("/projects")

	projects.Get("/", a.projectsList)
	projects.Post("/", a.projectsCreate)
	projects.Get("/:projectId", a.projectsGet)
	projects.Patch("/:projectId", a.projectsUpdate)
	projects.Delete("/:projectId", a.projectsDelete)
}

// GET /projects — список проектов пользователя с краткими сводными метриками.
// Query: status (ProjectStatus), search. 200 -> []ProjectSummary.
func (a *API) projectsList(c *fiber.Ctx) error {
	return notImplemented(c)
}

// POST /projects — создать проект (шаг 1 мастера инициации).
// Создаёт проект вместе с параметрами Ганта и опциональным начальным составом команды.
// Body: ProjectCreateRequest. 201 -> Project.
func (a *API) projectsCreate(c *fiber.Ctx) error {
	return notImplemented(c)
}

// GET /projects/{projectId} — полная карточка проекта, включая календарь и настройки Ганта.
// 200 -> Project, 404 -> не найден.
func (a *API) projectsGet(c *fiber.Ctx) error {
	return notImplemented(c)
}

// PATCH /projects/{projectId} — изменить параметры проекта (сроки, статус, описание, календарь).
// Body: ProjectUpdateRequest. 200 -> Project.
func (a *API) projectsUpdate(c *fiber.Ctx) error {
	return notImplemented(c)
}

// DELETE /projects/{projectId} — удалить (архивировать) проект.
// Переводит проект в статус archived; данные не удаляются физически. 204.
func (a *API) projectsDelete(c *fiber.Ctx) error {
	return notImplemented(c)
}
