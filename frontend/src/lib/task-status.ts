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

export const MILESTONE_STATUS_META: Record<
  MilestoneStatus,
  { label: string; tone: Tone }
> = {
  done: { label: "Сдано", tone: "success" },
  current: { label: "Текущая", tone: "warning" },
  planned: { label: "План", tone: "neutral" },
  final: { label: "Финал", tone: "brand" },
};
