import { httpClient } from "./http-client";
import type {
  ChecklistItem,
  ChecklistItemCreateRequest,
  ChecklistItemUpdateRequest,
} from "@/types";

export const checklistService = {
  /** Список критериев готовности (Definition of Done) для задачи. */
  list(taskId: string): Promise<ChecklistItem[]> {
    return httpClient.get<ChecklistItem[]>(`/tasks/${taskId}/checklist`);
  },

  /** Добавляет новый пункт в чек-лист Definition of Done задачи. */
  add(taskId: string, payload: ChecklistItemCreateRequest): Promise<ChecklistItem> {
    return httpClient.post<ChecklistItem>(`/tasks/${taskId}/checklist`, payload);
  },

  /** Обновляет текст критерия и/или отмечает его выполненным. */
  update(itemId: string, payload: ChecklistItemUpdateRequest): Promise<ChecklistItem> {
    return httpClient.patch<ChecklistItem>(`/checklist/${itemId}`, payload);
  },

  /** Удаляет пункт чек-листа DoD из задачи. */
  remove(itemId: string): Promise<void> {
    return httpClient.delete<void>(`/checklist/${itemId}`);
  },
};
