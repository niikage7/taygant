import type { MilestoneStatus, TaskStatus } from "@/types";

type Tone = "neutral" | "brand" | "success" | "warning" | "danger" | "accent";

/** Подпись и тональность статуса задачи для бейджей и таблиц. */
export const TASK_STATUS_META: Record<TaskStatus, { label: string; tone: Tone }> = {
  planned: { label: "План", tone: "brand" },
  in_progress: { label: "В работе", tone: "warning" },
  done: { label: "Готово", tone: "success" },
  overdue: { label: "Просрочена", tone: "danger" },
  blocked: { label: "Заблокирована", tone: "danger" },
};

/**
 * Статусы, которые пользователь вправе выставить сам.
 *
 * `overdue` и `blocked` бэкенд вычисляет из дат и незакрытых предшественников
 * (`resolveStatuses` в `service/tasks.go`) и отклоняет в запросе: «статус
 * выставляется системой автоматически», 400. Поэтому в выборе их быть не должно.
 */
export const ASSIGNABLE_TASK_STATUSES: TaskStatus[] = ["planned", "in_progress", "done"];

/** Вычисляемый сервером статус нельзя выбрать — только увидеть. */
export function isAssignableStatus(status: TaskStatus): boolean {
  return ASSIGNABLE_TASK_STATUSES.includes(status);
}

export const MILESTONE_STATUS_META: Record<MilestoneStatus, { label: string; tone: Tone }> = {
  done: { label: "Сдано", tone: "success" },
  current: { label: "Текущая", tone: "warning" },
  planned: { label: "План", tone: "neutral" },
  final: { label: "Финал", tone: "brand" },
};
