package api

import (
	"github.com/gofiber/fiber/v2"

	"taygant_backend/internal/models"
	"taygant_backend/internal/service"
)

// registerSimulationRoutes — тег Simulation: what-if моделирование и применение сдвига сроков (CPM).
func (a *API) registerSimulationRoutes(r fiber.Router) {
	tasks := r.Group("/tasks/:taskId")

	tasks.Post("/simulate-shift", a.simulationSimulateShift)
	tasks.Post("/apply-shift", a.simulationApplyShift)
}

// simulateShiftRequest — тело POST /tasks/{taskId}/simulate-shift.
type simulateShiftRequest struct {
	NewStartDate *models.Date `json:"newStartDate"`
	ShiftDays    *int         `json:"shiftDays"`
}

// POST /tasks/{taskId}/simulate-shift — смоделировать сдвиг сроков задачи (what-if анализ).
// Рассчитывает каскадное влияние переноса дат на зависимые задачи и срок сдачи проекта,
// НЕ применяя изменения (applied = false). Доступно любому участнику проекта —
// это просмотр сценария, а не изменение плана.
// Body: { newStartDate, shiftDays }. 200 -> ShiftSimulation.
func (a *API) simulationSimulateShift(c *fiber.Ctx) error {
	taskID, err := pathUUID(c, "taskId")
	if err != nil {
		return err
	}
	if _, err := a.requireMemberByTask(c, taskID); err != nil {
		return err
	}

	body, err := parseBody[simulateShiftRequest](c)
	if err != nil {
		return err
	}
	result, err := a.simulation.Simulate(c.Context(), taskID, service.ShiftInput{
		NewStartDate: body.NewStartDate,
		ShiftDays:    body.ShiftDays,
	})
	if err != nil {
		return fail(err)
	}
	return c.JSON(newShiftSimulationDTO(result))
}

// applyShiftRequest — тело POST /tasks/{taskId}/apply-shift.
type applyShiftRequest struct {
	NewStartDate         *models.Date `json:"newStartDate"`
	ShiftDays            *int         `json:"shiftDays"`
	CompensateFromBuffer bool         `json:"compensateFromBuffer"`
	NotifyAssignees      bool         `json:"notifyAssignees"`
}

// POST /tasks/{taskId}/apply-shift — применить сдвиг цепочки зависимых задач.
// Фактически переносит сроки задачи и каскадно связанных с ней задач по алгоритму CPM
// (applied = true). Требует полного доступа: переносится не одна задача, а цепочка.
// Body: { shiftDays, compensateFromBuffer, notifyAssignees }. 200 -> ShiftSimulation.
func (a *API) simulationApplyShift(c *fiber.Ctx) error {
	taskID, err := pathUUID(c, "taskId")
	if err != nil {
		return err
	}
	if _, err := a.requireManageByTask(c, taskID); err != nil {
		return err
	}

	body, err := parseBody[applyShiftRequest](c)
	if err != nil {
		return err
	}
	result, err := a.simulation.Apply(c.Context(), currentUserID(c), taskID, service.ShiftInput{
		NewStartDate: body.NewStartDate,
		ShiftDays:    body.ShiftDays,
	}, service.ApplyOptions{
		CompensateFromBuffer: body.CompensateFromBuffer,
		NotifyAssignees:      body.NotifyAssignees,
	})
	if err != nil {
		return fail(err)
	}
	return c.JSON(newShiftSimulationDTO(result))
}
