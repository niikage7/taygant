import { httpClient } from "./http-client";
import type { Milestone, MilestoneCreateRequest } from "@/types";

export const milestonesService = {
  /** Возвращает все вехи проекта в хронологическом порядке. */
  list(projectId: string): Promise<Milestone[]> {
    return httpClient.get<Milestone[]>(`/projects/${projectId}/milestones`);
  },

  /** Добавляет новую веху проекта с плановой датой. */
  create(projectId: string, payload: MilestoneCreateRequest): Promise<Milestone> {
    return httpClient.post<Milestone>(`/projects/${projectId}/milestones`, payload);
  },

  /** Обновляет название, код или плановую дату вехи. */
  update(milestoneId: string, payload: MilestoneCreateRequest): Promise<Milestone> {
    return httpClient.patch<Milestone>(`/milestones/${milestoneId}`, payload);
  },

  /** Удаляет веху из проекта. */
  remove(milestoneId: string): Promise<void> {
    return httpClient.delete<void>(`/milestones/${milestoneId}`);
  },
};
