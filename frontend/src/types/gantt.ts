import type { Milestone } from "./milestone";
import type { TaskDependency } from "./dependency";
import type { Task } from "./task";

/** Масштаб шкалы времени диаграммы Ганта. */
export type GanttScale = "days" | "weeks" | "months";

/** Агрегированные данные для отрисовки диаграммы Ганта проекта. */
export interface GanttChart {
  /** Идентификатор проекта. */
  projectId: string;
  /** Масштаб шкалы времени, использованный при формировании ответа. */
  scale: GanttScale;
  /** Начало отображаемого диапазона дат. */
  rangeStart: string;
  /** Конец отображаемого диапазона дат. */
  rangeEnd: string;
  /** Идентификаторы задач, входящих в критический путь проекта. */
  criticalPathTaskIds: string[];
  /** Задачи проекта, попадающие в отображаемый диапазон. */
  tasks: Task[];
  /** Связи между задачами, отображаемые на диаграмме. */
  dependencies: TaskDependency[];
  /** Контрольные точки проекта, отображаемые на диаграмме. */
  milestones: Milestone[];
}

/** Параметры запроса данных диаграммы Ганта. */
export interface GanttParams {
  /** Масштаб шкалы времени на диаграмме. */
  scale?: GanttScale;
}
