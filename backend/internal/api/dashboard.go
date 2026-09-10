package api

import "github.com/gofiber/fiber/v2"

// registerDashboardRoutes — тег Dashboard: сводные показатели состояния и рисков проекта.
func (a *API) registerDashboardRoutes(r fiber.Router) {
	r.Get("/projects/:projectId/dashboard", a.dashboardGet)
}

// GET /projects/{projectId}/dashboard — агрегированные метрики для экрана «Обзор и Аналитика».
//
// progressDeltaPercent всегда 0 (нет исторических снимков прогресса) — см.
// комментарий в internal/service/dashboard.go. Остальные поля посчитаны
// по-настоящему, включая scheduleAdherence через CPM (internal/schedule) и
// teamWorkload за текущий спринт (internal/service/workload.go).
// 200 -> ProjectDashboard.
func (a *API) dashboardGet(c *fiber.Ctx) error {
	projectID, err := pathUUID(c, "projectId")
	if err != nil {
		return err
	}
	if _, err := a.requireMember(c, projectID); err != nil {
		return err
	}

	result, err := a.dashboard.Get(c.Context(), projectID)
	if err != nil {
		return fail(err)
	}
	return c.JSON(newProjectDashboardDTO(result))
}
