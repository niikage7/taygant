package api

import "github.com/gofiber/fiber/v2"

// registerWorkloadRoutes — тег Workload: загрузка участников команды по часам.
func (a *API) registerWorkloadRoutes(r fiber.Router) {
	r.Get("/projects/:projectId/workload", a.workloadList)
}

// GET /projects/{projectId}/workload — загрузка команды по спринту.
// Возвращает загруженность участников в часах относительно недельного лимита.
// Query: sprintId (по умолчанию — текущий спринт). 200 -> []MemberWorkload.
func (a *API) workloadList(c *fiber.Ctx) error {
	projectID, err := pathUUID(c, "projectId")
	if err != nil {
		return err
	}
	if _, err := a.requireMember(c, projectID); err != nil {
		return err
	}

	sprintID, err := queryUUID(c, "sprintId")
	if err != nil {
		return err
	}

	workload, err := a.workload.List(c.Context(), projectID, sprintID)
	if err != nil {
		return fail(err)
	}

	out := make([]memberWorkloadDTO, 0, len(workload))
	for _, w := range workload {
		out = append(out, newMemberWorkloadDTO(w))
	}
	return c.JSON(out)
}
