import { httpClient } from "./http-client";
import type { HistoryEntry } from "@/types";

export const historyService = {
  /** Хронологический список изменений полей задачи с указанием автора действия. */
  list(taskId: string): Promise<HistoryEntry[]> {
    return httpClient.get<HistoryEntry[]>(`/tasks/${taskId}/history`);
  },
};
