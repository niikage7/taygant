import { httpClient } from "./http-client";
import type {
  GanttChart,
  GanttParams,
  Task,
  TaskCreateRequest,
  TaskDetail,
  TaskListParams,
  TaskUpdateRequest,
} from "@/types";

export const tasksService = {
  /** Возвращает задачи проекта с поддержкой фильтров реестра и диаграммы Ганта. */
  list(projectId: string, params?: TaskListParams): Promise<Task[]> {
    return httpClient.get<Task[]>(`/projects/${projectId}/tasks`, {
      query: {
        sprintId: params?.sprintId,
        assigneeId: params?.assigneeId,
        status: params?.status,
        criticalPathOnly: params?.criticalPathOnly,
        risksOnly: params?.risksOnly,
        myTasksOnly: params?.myTasksOnly,
        search: params?.search,
      },
    });
  },

  /** Создаёт задачу в проекте, опционально сразу со связями-предшественниками. */
  create(projectId: string, payload: TaskCreateRequest): Promise<Task> {
    return httpClient.post<Task>(`/projects/${projectId}/tasks`, payload);
  },

  /**
   * Агрегированные данные диаграммы Ганта.
   *
   * ⚠️ НЕ ВЫЗЫВАТЬ: маршрут `GET /projects/{projectId}/gantt` на бэкенде не
   * зарегистрирован — расчёт критического пути требует CPM-движка, которого
   * пока нет (см. комментарий в `backend/internal/api/tasks.go`). Запрос вернёт
   * 404. Экран Ганта собирает задачи и связи отдельными запросами, см.
   * `useProjectDependencies` в `src/data/queries.ts`.
   */
  getGantt(projectId: string, params?: GanttParams): Promise<GanttChart> {
    return httpClient.get<GanttChart>(`/projects/${projectId}/gantt`, {
      query: { scale: params?.scale },
    });
  },

  /** Полная карточка задачи со связями, чек-листом DoD и количеством комментариев. */
  getById(taskId: string): Promise<TaskDetail> {
    return httpClient.get<TaskDetail>(`/tasks/${taskId}`);
  },

  /** Частичное обновление задачи — передаются только изменяемые поля. */
  update(taskId: string, payload: TaskUpdateRequest): Promise<Task> {
    return httpClient.patch<Task>(`/tasks/${taskId}`, payload);
  },

  /** Удаляет задачу вместе с её связями, чек-листом и комментариями. */
  remove(taskId: string): Promise<void> {
    return httpClient.delete<void>(`/tasks/${taskId}`);
  },
};
