package api

import "github.com/gofiber/fiber/v2"

// registerHistoryRoutes — тег History: журнал изменений (аудит) задач.
func (a *API) registerHistoryRoutes(r fiber.Router) {
	r.Get("/tasks/:taskId/history", a.historyList)
}

// GET /tasks/{taskId}/history — хронологический список изменений полей задачи с автором действия.
// 200 -> []HistoryEntry.
func (a *API) historyList(c *fiber.Ctx) error {
	taskID, err := pathUUID(c, "taskId")
	if err != nil {
		return err
	}
	if _, err := a.requireMemberByTask(c, taskID); err != nil {
		return err
	}

	entries, err := a.history.List(c.Context(), taskID)
	if err != nil {
		return fail(err)
	}
	out := make([]historyEntryDTO, 0, len(entries))
	for _, e := range entries {
		out = append(out, newHistoryEntryDTO(e))
	}
	return c.JSON(out)
}
