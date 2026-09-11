import type { User } from "./user";

/** Загруженность участника команды в рамках спринта. */
export interface MemberWorkload {
  /** Участник, для которого рассчитана загрузка. */
  user: User;
  /** Спринт, для которого рассчитана загрузка (null — по всему проекту). */
  sprintId: string | null;
  /** Назначенная нагрузка участника, часов в неделю. */
  assignedHoursPerWeek: number;
  /** Лимит рабочих часов участника в неделю. */
  weeklyHoursLimit: number;
  /** Загрузка участника в процентах от лимита; значение >100 означает перегрузку. */
  utilizationPercent: number;
}

/** Параметры запроса загрузки команды проекта. */
export interface WorkloadParams {
  /** Ограничить расчёт конкретным спринтом (по умолчанию — текущий). */
  sprintId?: string;
}
