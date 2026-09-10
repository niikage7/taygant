package api

import "github.com/gofiber/fiber/v2"

// registerHistoryRoutes — тег History: журнал изменений (аудит) задач.
func registerHistoryRoutes(r fiber.Router) {
	r.Get("/tasks/:taskId/history", historyList)
}

// GET /tasks/{taskId}/history — хронологический список изменений полей задачи с автором действия.
// 200 -> []HistoryEntry.
func historyList(c *fiber.Ctx) error {
	return notImplemented(c)
}
