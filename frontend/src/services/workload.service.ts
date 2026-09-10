import { httpClient } from "./http-client";
import type { MemberWorkload, WorkloadParams } from "@/types";

export const workloadService = {
  /** Возвращает загруженность участников проекта в часах относительно их недельного лимита. */
  list(projectId: string, params?: WorkloadParams): Promise<MemberWorkload[]> {
    return httpClient.get<MemberWorkload[]>(`/projects/${projectId}/workload`, {
      query: { sprintId: params?.sprintId },
    });
  },
};
