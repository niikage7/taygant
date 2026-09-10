import { httpClient } from "./http-client";
import type {
  Project,
  ProjectCreateRequest,
  ProjectListParams,
  ProjectSummary,
  ProjectUpdateRequest,
} from "@/types";

export const projectsService = {
  /** Возвращает проекты, доступные текущему пользователю, с краткими сводными метриками. */
  list(params?: ProjectListParams): Promise<ProjectSummary[]> {
    return httpClient.get<ProjectSummary[]>("/projects", {
      query: { status: params?.status, search: params?.search },
    });
  },

  /** Создаёт проект вместе с базовыми параметрами Ганта и начальным составом команды. */
  create(payload: ProjectCreateRequest): Promise<Project> {
    return httpClient.post<Project>("/projects", payload);
  },

  /** Возвращает полную карточку проекта, включая параметры календаря и настройки Ганта. */
  getById(projectId: string): Promise<Project> {
    return httpClient.get<Project>(`/projects/${projectId}`);
  },

  /** Частичное обновление проекта — передаются только изменяемые поля. */
  update(projectId: string, payload: ProjectUpdateRequest): Promise<Project> {
    return httpClient.patch<Project>(`/projects/${projectId}`, payload);
  },

  /** Переводит проект в статус archived; данные проекта не удаляются физически. */
  remove(projectId: string): Promise<void> {
    return httpClient.delete<void>(`/projects/${projectId}`);
  },
};
