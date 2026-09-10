import { httpClient } from "./http-client";
import type { ApplyShiftRequest, ShiftSimulation, SimulateShiftRequest } from "@/types";

export const simulationService = {
  /**
   * Рассчитывает каскадное влияние переноса дат задачи на связанные последующие
   * задачи и итоговый срок сдачи проекта, не применяя изменения.
   */
  simulateShift(taskId: string, payload: SimulateShiftRequest): Promise<ShiftSimulation> {
    return httpClient.post<ShiftSimulation>(`/tasks/${taskId}/simulate-shift`, payload);
  },

  /** Фактически переносит сроки задачи и каскадно связанных с ней задач согласно алгоритму CPM. */
  applyShift(taskId: string, payload: ApplyShiftRequest): Promise<ShiftSimulation> {
    return httpClient.post<ShiftSimulation>(`/tasks/${taskId}/apply-shift`, payload);
  },
};
