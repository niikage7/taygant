package api

import "github.com/gofiber/fiber/v2"

// registerChecklistRoutes — тег Checklist: критерии приёмки задачи (Definition of Done).
func registerChecklistRoutes(r fiber.Router) {
	taskChecklist := r.Group("/tasks/:taskId/checklist")
	taskChecklist.Get("/", checklistList)
	taskChecklist.Post("/", checklistCreate)

	items := r.Group("/checklist/:itemId")
	items.Patch("/", checklistUpdate)
	items.Delete("/", checklistDelete)
}

// GET /tasks/{taskId}/checklist — критерии готовности (DoD) задачи.
// 200 -> []ChecklistItem.
func checklistList(c *fiber.Ctx) error {
	return notImplemented(c)
}

// POST /tasks/{taskId}/checklist — добавить пункт в чек-лист DoD задачи.
// Body: { text }. 201 -> ChecklistItem.
func checklistCreate(c *fiber.Ctx) error {
	return notImplemented(c)
}

// PATCH /checklist/{itemId} — отметить критерий выполненным и/или изменить его текст.
// Body: { text, isDone }. 200 -> ChecklistItem.
func checklistUpdate(c *fiber.Ctx) error {
	return notImplemented(c)
}

// DELETE /checklist/{itemId} — удалить пункт чек-листа DoD из задачи. 204.
func checklistDelete(c *fiber.Ctx) error {
	return notImplemented(c)
}
