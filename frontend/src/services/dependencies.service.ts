import { httpClient } from "./http-client";
import type {
  TaskDependenciesResponse,
  TaskDependency,
  TaskDependencyCreateRequest,
} from "@/types";

export const dependenciesService = {
  /** Возвращает предшественников и последователей задачи для сетевого графа связей. */
  list(taskId: string): Promise<TaskDependenciesResponse> {
    return httpClient.get<TaskDependenciesResponse>(`/tasks/${taskId}/dependencies`);
  },

  /** Создаёт связь (зависимость) между текущей задачей и другой задачей проекта. */
  create(
    taskId: string,
    payload: TaskDependencyCreateRequest,
  ): Promise<TaskDependency> {
    return httpClient.post<TaskDependency>(`/tasks/${taskId}/dependencies`, payload);
  },

  /** Разрывает зависимость между двумя задачами; сами задачи не удаляются. */
  remove(dependencyId: string): Promise<void> {
    return httpClient.delete<void>(`/dependencies/${dependencyId}`);
  },
};
