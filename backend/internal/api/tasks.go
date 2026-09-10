package api

import "github.com/gofiber/fiber/v2"

// registerTasksRoutes — тег Tasks: задачи проекта и диаграмма Ганта.
func registerTasksRoutes(r fiber.Router) {
	projectTasks := r.Group("/projects/:projectId")
	projectTasks.Get("/tasks", tasksList)
	projectTasks.Post("/tasks", tasksCreate)
	projectTasks.Get("/gantt", tasksGantt)

	tasks := r.Group("/tasks/:taskId")
	tasks.Get("/", tasksGet)
	tasks.Patch("/", tasksUpdate)
	tasks.Delete("/", tasksDelete)
}

// GET /projects/{projectId}/tasks — реестр задач проекта / данные для Ганта.
// Query: sprintId, assigneeId, status, criticalPathOnly, risksOnly, myTasksOnly, search.
// 200 -> []Task.
func tasksList(c *fiber.Ctx) error {
	return notImplemented(c)
}

// POST /projects/{projectId}/tasks — добавить задачу, опционально сразу со связями-предшественниками.
// Body: TaskCreateRequest. 201 -> Task.
func tasksCreate(c *fiber.Ctx) error {
	return notImplemented(c)
}

// GET /projects/{projectId}/gantt — данные для диаграммы Ганта: задачи, связи и рассчитанный
// критический путь (CPM), готовые для отрисовки таймлайна.
// Query: scale (days|weeks|months, по умолчанию weeks). 200 -> GanttChart.
func tasksGantt(c *fiber.Ctx) error {
	return notImplemented(c)
}

// GET /tasks/{taskId} — карточка задачи со связями, чек-листом DoD и числом комментариев.
// 200 -> TaskDetail, 404 -> не найдена.
func tasksGet(c *fiber.Ctx) error {
	return notImplemented(c)
}

// PATCH /tasks/{taskId} — редактировать задачу (сроки, статус, ответственный, прогресс и т.д.).
// Body: TaskUpdateRequest. 200 -> Task.
func tasksUpdate(c *fiber.Ctx) error {
	return notImplemented(c)
}

// DELETE /tasks/{taskId} — удалить задачу вместе со связями, чек-листом и комментариями. 204.
func tasksDelete(c *fiber.Ctx) error {
	return notImplemented(c)
}
