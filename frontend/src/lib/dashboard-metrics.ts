import { differenceInCalendarDays, parseISO } from "date-fns";

import type { Milestone, Project, Task, TaskStatus } from "@/types";

/**
 * Метрики обзора, посчитанные из данных, которые бэкенд уже отдаёт:
 * `GET /projects/{id}`, `/tasks` и `/milestones`.
 *
 * Отдельный эндпоинт `/dashboard` пока возвращает 501, но большая часть его
 * показателей — производные от задач и вех, поэтому считаем их на клиенте
 * вместо демонстрационных заглушек. Показатели, у которых источника нет вовсе
 * (прогноз отставания, резерв проекта, скорость команды, риски, загрузка в
 * часах), здесь намеренно отсутствуют: выдуманные цифры на экране мониторинга
 * хуже, чем их отсутствие.
 */
export type TaskStatusBreakdown = Record<TaskStatus, number> & { total: number };

export type DashboardMetrics = {
  progressPercent: number;
  milestonesClosed: number;
  milestonesTotal: number;
  daysToDeadline: number;
  breakdown: TaskStatusBreakdown;
  /** Задачи с отставанием от плана, худшие первыми. */
  laggingTasks: { task: Task; deviationDays: number }[];
  /** Наибольшее отставание среди задач, дней (0 — отставаний нет). */
  worstLagDays: number;
};

export function buildDashboardMetrics({
  project,
  tasks,
  milestones,
  today,
}: {
  project: Project;
  tasks: Task[];
  milestones: Milestone[];
  today: Date;
}): DashboardMetrics {
  const breakdown = tasks.reduce<TaskStatusBreakdown>(
    (acc, task) => {
      acc[task.status] += 1;
      acc.total += 1;
      return acc;
    },
    { planned: 0, in_progress: 0, done: 0, overdue: 0, blocked: 0, total: 0 },
  );

  const laggingTasks = tasks
    .filter((task) => task.planVsActualDeviationDays < 0)
    .map((task) => ({ task, deviationDays: task.planVsActualDeviationDays }))
    .sort((a, b) => a.deviationDays - b.deviationDays);

  return {
    progressPercent: project.progressPercent,
    milestonesClosed: milestones.filter((m) => m.status === "done").length,
    milestonesTotal: milestones.length,
    daysToDeadline: differenceInCalendarDays(parseISO(project.deadline), today),
    breakdown,
    laggingTasks,
    worstLagDays: laggingTasks.length > 0 ? Math.abs(laggingTasks[0].deviationDays) : 0,
  };
}
