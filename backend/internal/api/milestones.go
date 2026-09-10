package api

import (
	"github.com/gofiber/fiber/v2"

	"taygant_backend/internal/models"
	"taygant_backend/internal/service"
)

// registerMilestonesRoutes — тег Milestones: контрольные точки (вехи) проекта.
func (a *API) registerMilestonesRoutes(r fiber.Router) {
	projectMilestones := r.Group("/projects/:projectId/milestones")
	projectMilestones.Get("/", a.milestonesList)
	projectMilestones.Post("/", a.milestonesCreate)

	milestones := r.Group("/milestones/:milestoneId")
	milestones.Patch("/", a.milestonesUpdate)
	milestones.Delete("/", a.milestonesDelete)
}

// milestoneRequest — тело создания/изменения вехи.
type milestoneRequest struct {
	Code        string      `json:"code"`
	Name        string      `json:"name"`
	PlannedDate models.Date `json:"plannedDate"`
	// ActualDate — только для PATCH: не передано — не менять, null — снять
	// отметку о достижении, значение — отметить веху достигнутой (см. Nullable).
	ActualDate Nullable[models.Date] `json:"actualDate"`
}

func (r milestoneRequest) toInput() service.MilestoneInput {
	return service.MilestoneInput{
		Code:          r.Code,
		Name:          r.Name,
		PlannedDate:   r.PlannedDate,
		ActualDate:    r.ActualDate.Value,
		ActualDateSet: r.ActualDate.Set,
	}
}

// GET /projects/{projectId}/milestones — все вехи проекта в хронологическом порядке.
// 200 -> []Milestone.
func (a *API) milestonesList(c *fiber.Ctx) error {
	projectID, err := pathUUID(c, "projectId")
	if err != nil {
		return err
	}
	if _, err := a.requireMember(c, projectID); err != nil {
		return err
	}

	milestones, err := a.milestones.List(c.Context(), projectID)
	if err != nil {
		return fail(err)
	}
	out := make([]milestoneDTO, 0, len(milestones))
	for _, m := range milestones {
		out = append(out, newMilestoneDTO(m))
	}
	return c.JSON(out)
}

// POST /projects/{projectId}/milestones — создать контрольную точку с плановой датой.
// Body: MilestoneCreateRequest. 201 -> Milestone.
func (a *API) milestonesCreate(c *fiber.Ctx) error {
	projectID, err := pathUUID(c, "projectId")
	if err != nil {
		return err
	}
	if err := a.requireEdit(c, projectID); err != nil {
		return err
	}

	body, err := parseBody[milestoneRequest](c)
	if err != nil {
		return err
	}
	milestone, err := a.milestones.Create(c.Context(), projectID, body.toInput())
	if err != nil {
		return fail(err)
	}
	return c.Status(fiber.StatusCreated).JSON(newMilestoneDTO(milestone))
}

// PATCH /milestones/{milestoneId} — изменить название, код, плановую дату вехи
// или отметить её достигнутой через actualDate (null снимает отметку).
// Body: MilestoneCreateRequest. 200 -> Milestone.
func (a *API) milestonesUpdate(c *fiber.Ctx) error {
	milestoneID, err := pathUUID(c, "milestoneId")
	if err != nil {
		return err
	}
	projectID, err := a.milestones.ProjectID(c.Context(), milestoneID)
	if err != nil {
		return fail(err)
	}
	if err := a.requireEdit(c, projectID); err != nil {
		return err
	}

	body, err := parseBody[milestoneRequest](c)
	if err != nil {
		return err
	}
	milestone, err := a.milestones.Update(c.Context(), milestoneID, body.toInput())
	if err != nil {
		return fail(err)
	}
	return c.JSON(newMilestoneDTO(milestone))
}

// DELETE /milestones/{milestoneId} — удалить веху из проекта. 204.
func (a *API) milestonesDelete(c *fiber.Ctx) error {
	milestoneID, err := pathUUID(c, "milestoneId")
	if err != nil {
		return err
	}
	projectID, err := a.milestones.ProjectID(c.Context(), milestoneID)
	if err != nil {
		return fail(err)
	}
	if err := a.requireEdit(c, projectID); err != nil {
		return err
	}

	if err := a.milestones.Delete(c.Context(), milestoneID); err != nil {
		return fail(err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}
