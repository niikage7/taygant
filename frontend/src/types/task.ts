import type { DependencyType, TaskStatus } from "./common";
import type { ChecklistItem } from "./checklist";
import type { TaskDependency } from "./dependency";
import type { User } from "./user";

/** Задача проекта — основной элемент реестра задач и диаграммы Ганта. */
export interface Task {
  /** Уникальный идентификатор задачи. */
  id: string;
  /** Идентификатор проекта, которому принадлежит задача. */
  projectId: string;
  /** Идентификатор спринта, к которому привязана задача (может отсутствовать). */
  sprintId: string | null;
  /** Идентификатор родительской задачи для построения иерархии WBS (подзадачи). */
  parentTaskId: string | null;
  /** Контрольная точка, к которой ведёт задача; null — задача вне вех. */
  milestoneId: string | null;
  /** Короткий читаемый код задачи. */
  code: string;
  /** Номер задачи в иерархической структуре работ (WBS). */
  wbsNumber: string;
  /** Название задачи. */
  title: string;
  /**
   * Ответственный исполнитель задачи.
   * Бэкенд присылает `null`, если исполнитель не назначен (см. `newUserDTOPtr`
   * в `backend/internal/api/dto.go`), а не опускает поле.
   */
  assignee: User | null;
  /** Текущий статус задачи. */
  status: TaskStatus;
  /** Признак того, что задача является вехой (точкой с нулевой длительностью) на диаграмме Ганта. */
  isMilestone: boolean;
  /** Плановая дата начала задачи. */
  startDate: string;
  /** Плановая дата окончания задачи. */
  endDate: string;
  /** Продолжительность задачи в календарных днях. */
  durationCalendarDays: number;
  /** Продолжительность задачи в рабочих днях согласно календарю проекта. */
  durationWorkingDays: number;
  /** Процент фактического выполнения задачи. */
  progressPercent: number;
  /** Вес задачи в общем прогрессе проекта, %. */
  weightPercent: number;
  /** Признак того, что задача лежит на критическом пути (нулевой резерв времени). */
  isCriticalPath: boolean;
  /** Резерв (буфер) времени задачи в днях, доступный для компенсации сдвигов. */
  bufferDays: number;
  /** Отклонение факта от плана в днях. Положительное — опережение, отрицательное — отставание. */
  planVsActualDeviationDays: number;
}

/** Связь задачи с внешней системой отслеживания (например, GitLab issue). */
export interface TaskExternalLink {
  /** Название внешней системы. */
  provider: string;
  /** Ссылка на объект во внешней системе. */
  url: string;
  /** Идентификатор объекта во внешней системе. */
  referenceId: string;
  /** Признак того, что данные синхронизированы с внешней системой. */
  synced: boolean;
}

/** Полная карточка задачи со связями, чек-листом DoD и метаданными для экрана редактирования. */
export interface TaskDetail extends Task {
  /** Развёрнутое описание задачи. */
  description?: string;
  /** Связь задачи с внешней системой отслеживания (например, GitLab issue). */
  externalLink: TaskExternalLink | null;
  /** Критерии приёмки (Definition of Done) задачи. */
  checklist: ChecklistItem[];
  /** Задачи-предшественники, от которых зависит текущая задача. */
  predecessors: TaskDependency[];
  /** Задачи-последователи, зависящие от текущей задачи. */
  successors: TaskDependency[];
  /** Количество комментариев к задаче. */
  commentsCount: number;
}

/** Связь с задачей-предшественником, создаваемая одновременно с задачей. */
export interface TaskPredecessorInput {
  /** Идентификатор задачи-предшественника. */
  taskId: string;
  /** Тип связи между задачами. */
  type: DependencyType;
  /** Задержка (лаг) в днях между задачами. */
  lagDays?: number;
}

/** Данные для создания новой задачи. */
export interface TaskCreateRequest {
  /** Идентификатор спринта, к которому привязывается задача. */
  sprintId?: string;
  /** Идентификатор родительской задачи (для создания подзадачи в WBS). */
  parentTaskId?: string;
  /** Контрольная точка, к которой ведёт задача. */
  milestoneId?: string;
  /** Название задачи. */
  title: string;
  /** Описание задачи. */
  description?: string;
  /** Идентификатор назначаемого ответственного. */
  assigneeId?: string;
  /** Начальный статус задачи. */
  status?: TaskStatus;
  /** Признак того, что создаваемая задача является вехой. */
  isMilestone?: boolean;
  /** Плановая дата начала задачи. */
  startDate: string;
  /** Плановая дата окончания задачи. */
  endDate: string;
  /** Вес задачи в общем прогрессе проекта, %. */
  weightPercent?: number;
  /** Список связей с задачами-предшественниками, создаваемых одновременно с задачей. */
  predecessors?: TaskPredecessorInput[];
}

/** Набор полей задачи, доступных для частичного обновления (PATCH). */
export interface TaskUpdateRequest {
  /** Новый спринт задачи. */
  sprintId?: string;
  /** Новое название задачи. */
  title?: string;
  /** Новое описание задачи. */
  description?: string;
  /**
   * Идентификатор нового ответственного.
   * `null` снимает ответственного, отсутствие поля оставляет прежнего —
   * на бэкенде это `*uuid.UUID` (см. `taskUpdateRequest` в `tasks.go`).
   */
  assigneeId?: string | null;
  /** Новый статус задачи. */
  status?: TaskStatus;
  /** Новая дата начала задачи. */
  startDate?: string;
  /** Новая дата окончания задачи. */
  endDate?: string;
  /** Новый процент выполнения задачи. */
  progressPercent?: number;
  /** Новый вес задачи в общем прогрессе проекта, %. */
  weightPercent?: number;
  /**
   * Контрольная точка задачи.
   * Поле не передано — не менять, `null` — отвязать от вехи
   * (на бэкенде это различают `MilestoneID` + `MilestoneIDSet`).
   */
  milestoneId?: string | null;
}

/** Параметры фильтрации списка задач проекта. */
export interface TaskListParams {
  /** Ограничить выборку задачами конкретного спринта. */
  sprintId?: string;
  /** Ограничить выборку задачами конкретного ответственного. */
  assigneeId?: string;
  /** Фильтр по статусу задачи. */
  status?: TaskStatus;
  /** Если true — вернуть только задачи, лежащие на критическом пути. */
  criticalPathOnly?: boolean;
  /** Только задачи с отклонением от плана (в зоне риска). */
  risksOnly?: boolean;
  /** Если true — вернуть только задачи текущего пользователя. */
  myTasksOnly?: boolean;
  /** Подстрока для поиска по названию или коду задачи. */
  search?: string;
}
