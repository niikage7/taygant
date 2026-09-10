package api

import "github.com/gofiber/fiber/v2"

// registerDependenciesRoutes — тег Dependencies: зависимости (связи) между задачами.
func registerDependenciesRoutes(r fiber.Router) {
	taskDependencies := r.Group("/tasks/:taskId/dependencies")
	taskDependencies.Get("/", dependenciesList)
	taskDependencies.Post("/", dependenciesCreate)

	r.Delete("/dependencies/:dependencyId", dependenciesDelete)
}

// GET /tasks/{taskId}/dependencies — входящие и исходящие связи задачи для сетевого графа.
// 200 -> { predecessors: []TaskDependency, successors: []TaskDependency }.
func dependenciesList(c *fiber.Ctx) error {
	return notImplemented(c)
}

// POST /tasks/{taskId}/dependencies — связать задачу с другой задачей проекта.
// Body: TaskDependencyCreateRequest. 201 -> TaskDependency.
func dependenciesCreate(c *fiber.Ctx) error {
	return notImplemented(c)
}

// DELETE /dependencies/{dependencyId} — разорвать зависимость между двумя задачами.
// Сами задачи не удаляются. 204.
func dependenciesDelete(c *fiber.Ctx) error {
	return notImplemented(c)
}
