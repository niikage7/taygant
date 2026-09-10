import { httpClient } from "./http-client";
import type { ProjectDashboard } from "@/types";

export const dashboardService = {
  /** Агрегированные метрики для экрана «Обзор и Аналитика». */
  get(projectId: string): Promise<ProjectDashboard> {
    return httpClient.get<ProjectDashboard>(`/projects/${projectId}/dashboard`);
  },
};
