import type { AccessLevel, ProjectRole } from "./common";
import type { User } from "./user";

/** Участник команды проекта с назначенной ролью и уровнем доступа. */
export interface ProjectMember {
  /** Идентификатор записи участника проекта. */
  id: string;
  /** Пользователь, добавленный в проект. */
  user: User;
  /** Роль участника в проекте. */
  projectRole: ProjectRole;
  /** Роль участника в рамках спринтов (свободный текст, например «Разработчик»). */
  sprintRole?: string;
  /** Уровень доступа участника к проекту. */
  accessLevel: AccessLevel;
  /** Лимит часов участника в неделю, используемый для расчёта загрузки. */
  weeklyHoursLimit?: number;
}

/** Данные для добавления участника в проект. */
export interface ProjectMemberCreateRequest {
  /** Идентификатор добавляемого пользователя. */
  userId: string;
  /** Назначаемая роль участника в проекте. */
  projectRole: ProjectRole;
  /** Роль участника в рамках спринтов. */
  sprintRole?: string;
  /** Назначаемый уровень доступа. */
  accessLevel?: AccessLevel;
}

/** Данные для изменения роли/уровня доступа участника. */
export interface ProjectMemberUpdateRequest {
  /** Новая роль участника в проекте. */
  projectRole?: ProjectRole;
  /** Новая роль участника в рамках спринтов (например, "Аналитик"). */
  sprintRole?: string;
  /** Новый уровень доступа участника к проекту. */
  accessLevel?: AccessLevel;
}
