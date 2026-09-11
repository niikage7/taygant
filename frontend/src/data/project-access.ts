"use client";

import { useCurrentProject } from "@/data/current-project";
import { useProjectMembers } from "@/data/queries";
import { getStoredUser } from "@/lib/session";
import type { AccessLevel, Task } from "@/types";

/** Минимальные данные задачи, нужные для проверки прав на её правку. */
type TaskRef = Pick<Task, "assignee"> | { assignee?: Task["assignee"] };

export type ProjectAccess = {
  /** Уровень доступа текущего пользователя в выбранном проекте. */
  accessLevel: AccessLevel | undefined;
  /** Полный доступ: план, параметры проекта и команда. */
  isFull: boolean;
  /** Уровень edit: статус и сроки своих задач. */
  isEdit: boolean;
  /** Только просмотр. */
  isView: boolean;
  /** Идентификатор текущего пользователя (для проверки «моя задача»). */
  currentUserId: string | undefined;
  /** Участник команды ещё не загружен. */
  isLoading: boolean;
  /** Можно ли открыть задачу на редактирование. */
  canEditTask: (task: TaskRef) => boolean;
  /** Правка задачи ограничена статусом и сроками (уровень edit). */
  isRestrictedTask: (task: TaskRef) => boolean;
};

/**
 * Права текущего пользователя в выбранном проекте.
 *
 * Уровень доступа лежит в записи участника, а не в профиле пользователя:
 * один и тот же человек может быть full в одном проекте и view в другом.
 * Поэтому список команды служит источником истины, а профиль из сессии
 * подсказывает, какая именно запись принадлежит текущему пользователю.
 */
export function useProjectAccess(): ProjectAccess {
  const { projectId } = useCurrentProject();
  const members = useProjectMembers(projectId);
  const currentUserId = getStoredUser()?.id;

  const me = members.data?.find((member) => member.user.id === currentUserId);
  const accessLevel = me?.accessLevel;

  const isFull = accessLevel === "full";
  const isEdit = accessLevel === "edit";
  const isView = accessLevel === "view";

  const canEditTask = (task: TaskRef) =>
    isFull || (isEdit && Boolean(currentUserId) && task.assignee?.id === currentUserId);

  const isRestrictedTask = (task: TaskRef) => isEdit && canEditTask(task);

  return {
    accessLevel,
    isFull,
    isEdit,
    isView,
    currentUserId,
    isLoading: members.isPending,
    canEditTask,
    isRestrictedTask,
  };
}
