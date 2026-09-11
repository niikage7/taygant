import { httpClient } from "./http-client";
import type { Sprint, SprintCreateRequest } from "@/types";

export const sprintsService = {
  /** Возвращает спринты проекта, используемые для группировки задач. */
  list(projectId: string): Promise<Sprint[]> {
    return httpClient.get<Sprint[]>(`/projects/${projectId}/sprints`);
  },

  /** Создаёт новый спринт/релиз в рамках проекта. */
  create(projectId: string, payload: SprintCreateRequest): Promise<Sprint> {
    return httpClient.post<Sprint>(`/projects/${projectId}/sprints`, payload);
  },

  /** Обновляет название, версию релиза или даты спринта. */
  update(sprintId: string, payload: SprintCreateRequest): Promise<Sprint> {
    return httpClient.patch<Sprint>(`/sprints/${sprintId}`, payload);
  },

  /** Удаляет спринт; связанные задачи остаются в проекте без привязки к спринту. */
  remove(sprintId: string): Promise<void> {
    return httpClient.delete<void>(`/sprints/${sprintId}`);
  },
};
