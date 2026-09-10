package api

import "github.com/gofiber/fiber/v2"

// registerCommentsRoutes — тег Comments: комментарии команды к задачам.
func (a *API) registerCommentsRoutes(r fiber.Router) {
	comments := r.Group("/tasks/:taskId/comments")

	comments.Get("/", a.commentsList)
	comments.Post("/", a.commentsCreate)
}

// GET /tasks/{taskId}/comments — лента комментариев задачи в хронологическом порядке.
// 200 -> []Comment.
func (a *API) commentsList(c *fiber.Ctx) error {
	return notImplemented(c)
}

// POST /tasks/{taskId}/comments — опубликовать комментарий от имени текущего пользователя.
// Body: { text }. 201 -> Comment.
func (a *API) commentsCreate(c *fiber.Ctx) error {
	return notImplemented(c)
}
