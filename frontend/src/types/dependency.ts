import type { DependencyType } from "./common";

/** Связь (зависимость) между двумя задачами. */
export interface TaskDependency {
  /** Идентификатор связи. */
  id: string;
  /** Идентификатор задачи-предшественника. */
  predecessorTaskId: string;
  /** Идентификатор задачи-последователя. */
  successorTaskId: string;
  /** Название задачи-предшественника (для отображения без дополнительного запроса). */
  predecessorTitle?: string;
  /** Название задачи-последователя (для отображения без дополнительного запроса). */
  successorTitle?: string;
  /** Тип связи между задачами. */
  type: DependencyType;
  /** Задержка (лаг) в днях между задачами; может быть отрицательной (опережение). */
  lagDays: number;
  /** Признак того, что связь лежит на критическом пути проекта. */
  isCritical: boolean;
}

/** Направление связи относительно текущей задачи. */
export type TaskDependencyDirection = "predecessor" | "successor";

/** Данные для создания связи между текущей задачей и другой задачей. */
export interface TaskDependencyCreateRequest {
  /** Идентификатор задачи, с которой создаётся связь. */
  relatedTaskId: string;
  /**
   * Направление связи относительно текущей задачи: predecessor — связываемая задача
   * является предшественником, successor — последователем.
   */
  direction: TaskDependencyDirection;
  /** Тип создаваемой связи. */
  type: DependencyType;
  /** Задержка (лаг) в днях между задачами. */
  lagDays?: number;
}

/** Входящие и исходящие связи задачи. */
export interface TaskDependenciesResponse {
  /** Задачи, от которых зависит текущая задача. */
  predecessors: TaskDependency[];
  /** Задачи, которые зависят от текущей задачи. */
  successors: TaskDependency[];
}
