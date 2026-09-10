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
	taskID, err := pathUUID(c, "taskId")
	if err != nil {
		return err
	}
	if _, err := a.requireMemberByTask(c, taskID); err != nil {
		return err
	}

	comments, err := a.comments.List(c.Context(), taskID)
	if err != nil {
		return fail(err)
	}
	out := make([]commentDTO, 0, len(comments))
	for _, cm := range comments {
		out = append(out, newCommentDTO(cm))
	}
	return c.JSON(out)
}

// commentCreateRequest — тело POST /tasks/{taskId}/comments.
type commentCreateRequest struct {
	Text string `json:"text"`
}

// POST /tasks/{taskId}/comments — опубликовать комментарий от имени текущего пользователя.
// Body: { text }. 201 -> Comment.
//
// Комментировать может любой участник проекта, а не только те, у кого есть
// право редактировать план: обсуждение задачи — не то же самое, что её правка.
func (a *API) commentsCreate(c *fiber.Ctx) error {
	taskID, err := pathUUID(c, "taskId")
	if err != nil {
		return err
	}
	if _, err := a.requireMemberByTask(c, taskID); err != nil {
		return err
	}

	body, err := parseBody[commentCreateRequest](c)
	if err != nil {
		return err
	}
	comment, err := a.comments.Add(c.Context(), taskID, currentUserID(c), body.Text)
	if err != nil {
		return fail(err)
	}
	return c.Status(fiber.StatusCreated).JSON(newCommentDTO(comment))
}
