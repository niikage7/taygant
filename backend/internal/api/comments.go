package api

import "github.com/gofiber/fiber/v2"

// registerCommentsRoutes — тег Comments: комментарии команды к задачам.
func registerCommentsRoutes(r fiber.Router) {
	comments := r.Group("/tasks/:taskId/comments")

	comments.Get("/", commentsList)
	comments.Post("/", commentsCreate)
}

// GET /tasks/{taskId}/comments — лента комментариев задачи в хронологическом порядке.
// 200 -> []Comment.
func commentsList(c *fiber.Ctx) error {
	return notImplemented(c)
}

// POST /tasks/{taskId}/comments — опубликовать комментарий от имени текущего пользователя.
// Body: { text }. 201 -> Comment.
func commentsCreate(c *fiber.Ctx) error {
	return notImplemented(c)
}
