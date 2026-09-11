package api

import (
	"github.com/gofiber/fiber/v2"

	"taygant_backend/internal/service"
)

// registerChecklistRoutes — тег Checklist: критерии приёмки задачи (Definition of Done).
func (a *API) registerChecklistRoutes(r fiber.Router) {
	taskChecklist := r.Group("/tasks/:taskId/checklist")
	taskChecklist.Get("/", a.checklistList)
	taskChecklist.Post("/", a.checklistCreate)

	items := r.Group("/checklist/:itemId")
	items.Patch("/", a.checklistUpdate)
	items.Delete("/", a.checklistDelete)
}

// GET /tasks/{taskId}/checklist — критерии готовности (DoD) задачи.
// 200 -> []ChecklistItem.
func (a *API) checklistList(c *fiber.Ctx) error {
	taskID, err := pathUUID(c, "taskId")
	if err != nil {
		return err
	}
	if _, err := a.requireMemberByTask(c, taskID); err != nil {
		return err
	}

	items, err := a.checklist.List(c.Context(), taskID)
	if err != nil {
		return fail(err)
	}
	out := make([]checklistItemDTO, 0, len(items))
	for _, i := range items {
		out = append(out, newChecklistItemDTO(i))
	}
	return c.JSON(out)
}

// checklistCreateRequest — тело POST /tasks/{taskId}/checklist.
type checklistCreateRequest struct {
	Text string `json:"text"`
}

// POST /tasks/{taskId}/checklist — добавить пункт в чек-лист DoD задачи.
// Body: { text }. 201 -> ChecklistItem.
func (a *API) checklistCreate(c *fiber.Ctx) error {
	taskID, err := pathUUID(c, "taskId")
	if err != nil {
		return err
	}
	if _, err := a.requireManageByTask(c, taskID); err != nil {
		return err
	}

	body, err := parseBody[checklistCreateRequest](c)
	if err != nil {
		return err
	}
	item, err := a.checklist.Add(c.Context(), taskID, currentUserID(c), body.Text)
	if err != nil {
		return fail(err)
	}
	return c.Status(fiber.StatusCreated).JSON(newChecklistItemDTO(item))
}

// checklistUpdateRequest — тело PATCH /checklist/{itemId}.
type checklistUpdateRequest struct {
	Text   *string `json:"text"`
	IsDone *bool   `json:"isDone"`
}

// PATCH /checklist/{itemId} — отметить критерий выполненным и/или изменить его текст.
// Body: { text, isDone }. 200 -> ChecklistItem.
func (a *API) checklistUpdate(c *fiber.Ctx) error {
	itemID, err := pathUUID(c, "itemId")
	if err != nil {
		return err
	}
	projectID, _, err := a.checklist.ProjectID(c.Context(), itemID)
	if err != nil {
		return fail(err)
	}
	if _, err := a.requireManage(c, projectID); err != nil {
		return err
	}

	body, err := parseBody[checklistUpdateRequest](c)
	if err != nil {
		return err
	}
	item, err := a.checklist.Update(c.Context(), itemID, currentUserID(c), service.ChecklistUpdateInput{
		Text:   body.Text,
		IsDone: body.IsDone,
	})
	if err != nil {
		return fail(err)
	}
	return c.JSON(newChecklistItemDTO(item))
}

// DELETE /checklist/{itemId} — удалить пункт чек-листа DoD из задачи. 204.
func (a *API) checklistDelete(c *fiber.Ctx) error {
	itemID, err := pathUUID(c, "itemId")
	if err != nil {
		return err
	}
	projectID, _, err := a.checklist.ProjectID(c.Context(), itemID)
	if err != nil {
		return fail(err)
	}
	if _, err := a.requireManage(c, projectID); err != nil {
		return err
	}

	if err := a.checklist.Delete(c.Context(), itemID); err != nil {
		return fail(err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}
