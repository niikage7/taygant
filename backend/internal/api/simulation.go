package api

import "github.com/gofiber/fiber/v2"

// registerSimulationRoutes — тег Simulation: what-if моделирование и применение сдвига сроков (CPM).
func registerSimulationRoutes(r fiber.Router) {
	tasks := r.Group("/tasks/:taskId")

	tasks.Post("/simulate-shift", simulationSimulateShift)
	tasks.Post("/apply-shift", simulationApplyShift)
}

// POST /tasks/{taskId}/simulate-shift — смоделировать сдвиг сроков задачи (what-if анализ).
// Рассчитывает каскадное влияние переноса дат на зависимые задачи и срок сдачи проекта,
// НЕ применяя изменения (applied = false).
// Body: { newStartDate, shiftDays }. 200 -> ShiftSimulation.
func simulationSimulateShift(c *fiber.Ctx) error {
	return notImplemented(c)
}

// POST /tasks/{taskId}/apply-shift — применить сдвиг цепочки зависимых задач.
// Фактически переносит сроки задачи и каскадно связанных с ней задач по алгоритму CPM
// (applied = true).
// Body: { shiftDays, compensateFromBuffer, notifyAssignees }. 200 -> ShiftSimulation.
func simulationApplyShift(c *fiber.Ctx) error {
	return notImplemented(c)
}
