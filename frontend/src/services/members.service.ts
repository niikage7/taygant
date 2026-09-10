import { httpClient } from "./http-client";
import type {
  ProjectMember,
  ProjectMemberCreateRequest,
  ProjectMemberUpdateRequest,
} from "@/types";

export const membersService = {
  /** Список участников проекта с их ролями и уровнем доступа. */
  list(projectId: string): Promise<ProjectMember[]> {
    return httpClient.get<ProjectMember[]>(`/projects/${projectId}/members`);
  },

  /** Добавляет существующего пользователя платформы в команду проекта с указанной ролью. */
  add(projectId: string, payload: ProjectMemberCreateRequest): Promise<ProjectMember> {
    return httpClient.post<ProjectMember>(`/projects/${projectId}/members`, payload);
  },

  /** Обновляет роль в проекте, роль в спринтах и/или уровень доступа участника. */
  update(
    projectId: string,
    memberId: string,
    payload: ProjectMemberUpdateRequest,
  ): Promise<ProjectMember> {
    return httpClient.patch<ProjectMember>(
      `/projects/${projectId}/members/${memberId}`,
      payload,
    );
  },

  /** Исключает участника из команды проекта; ранее назначенные ему задачи не удаляются. */
  remove(projectId: string, memberId: string): Promise<void> {
    return httpClient.delete<void>(`/projects/${projectId}/members/${memberId}`);
  },
};
