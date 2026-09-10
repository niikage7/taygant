import type { ProjectSummary } from "./project";
import type { Milestone } from "./milestone";
import type { Task } from "./task";
import type { MemberWorkload } from "./workload";

/** Статус соблюдения сроков проекта. */
export type ScheduleAdherenceStatus = "on_track" | "at_risk" | "delayed";

/** Соблюдение сроков проекта. */
export interface ScheduleAdherence {
  /**
   * Статус соблюдения сроков: on_track — в графике, at_risk — в зоне риска,
   * delayed — сорван срок.
   */
  status: ScheduleAdherenceStatus;
  /** Прогнозируемое отклонение от плановых сроков, дней. */
  forecastDeviationDays: number;
  /** Остаток общего резерва (буфера) сроков проекта, дней. */
  projectBufferDays: number;
}

/** Разбивка задач проекта по статусам. */
export interface TaskStatusBreakdown {
  /** Общее количество задач проекта. */
  total: number;
  /** Количество выполненных задач. */
  done: number;
  /** Количество задач в работе. */
  inProgress: number;
  /** Количество предстоящих (запланированных) задач. */
  planned: number;
  /** Количество просроченных задач. */
  overdue: number;
}

/** Скорость выполнения задач командой. */
export interface TeamVelocity {
  /** Среднее количество завершаемых задач в неделю. */
  tasksPerWeek: number;
}

/** Серьёзность риска. */
export type RiskSeverity = "critical" | "high" | "medium" | "low";

/** Активный риск проекта. */
export interface ProjectRisk {
  /** Серьёзность риска. */
  severity: RiskSeverity;
  /** Краткое название риска. */
  title: string;
  /** Подробное описание риска и его причины. */
  description?: string;
  /** Задача, с которой связан риск (если применимо). */
  relatedTaskId: string | null;
  /** Оценка влияния риска на срок проекта, дней. */
  impactDays: number | null;
}

/** Задача, требующая внимания (отклонение от плана). */
export interface AttentionTask {
  /** Задача, требующая внимания. */
  task: Task;
  /** Отклонение задачи от плана в днях (отрицательное — отставание). */
  deviationDays: number;
}

/** Сводные показатели состояния и рисков проекта для экрана «Обзор и Аналитика». */
export interface ProjectDashboard {
  /** Краткая карточка проекта. */
  project: ProjectSummary;
  /** Общий процент выполнения проекта. */
  overallProgressPercent: number;
  /** Изменение процента выполнения по сравнению с предыдущим периодом. */
  progressDeltaPercent: number;
  /** Количество закрытых контрольных точек. */
  milestonesClosed: number;
  /** Общее количество контрольных точек проекта. */
  milestonesTotal: number;
  /** Соблюдение сроков проекта. */
  scheduleAdherence: ScheduleAdherence;
  /** Разбивка задач проекта по статусам. */
  taskStatusBreakdown: TaskStatusBreakdown;
  /** Скорость выполнения задач командой. */
  teamVelocity: TeamVelocity;
  /** Ближайшие контрольные точки проекта. */
  nearestMilestones: Milestone[];
  /** Активные риски проекта. */
  risks: ProjectRisk[];
  /** Задачи, требующие внимания (отклонение от плана). */
  attentionTasks: AttentionTask[];
  /** Загрузка участников команды текущего спринта. */
  teamWorkload: MemberWorkload[];
}
