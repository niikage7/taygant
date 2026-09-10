import type { DependencyType } from "./common";

/** Задача, затронутая сдвигом, с датами до и после пересчёта. */
export interface ShiftAffectedTask {
  /** Идентификатор затронутой задачи. */
  taskId: string;
  /** Название затронутой задачи. */
  title: string;
  /** Признак того, что задача лежит на критическом пути. */
  isCriticalPath: boolean;
  /** Дата начала задачи до пересчёта. */
  originalStartDate: string;
  /** Дата окончания задачи до пересчёта. */
  originalEndDate: string;
  /** Дата начала задачи после пересчёта. */
  newStartDate: string;
  /** Дата окончания задачи после пересчёта. */
  newEndDate: string;
  /** Тип связи, через которую сдвиг передался на эту задачу. */
  dependencyType: DependencyType;
}

/** Влияние сдвига на общий срок сдачи проекта. */
export interface ShiftProjectDeadlineImpact {
  /** Плановый дедлайн проекта до сдвига. */
  originalDeadline: string;
  /** Прогнозируемый дедлайн проекта после сдвига. */
  newDeadline: string;
  /** Изменение срока сдачи проекта в днях (положительное — задержка). */
  deltaDays: number;
}

/** Результат what-if моделирования либо фактического применения сдвига сроков задачи. */
export interface ShiftSimulation {
  /** Идентификатор задачи, инициировавшей сдвиг. */
  sourceTaskId: string;
  /** Величина сдвига в днях, применённая к исходной задаче. */
  shiftDays: number;
  /** Алгоритм расчёта каскадного сдвига сроков. */
  algorithm: string;
  /** Список задач, затронутых сдвигом, с датами до и после пересчёта. */
  affectedTasks: ShiftAffectedTask[];
  /** Влияние сдвига на общий срок сдачи проекта. */
  projectDeadlineImpact: ShiftProjectDeadlineImpact;
  /** Количество дней резерва, доступных для компенсации сдвига без переноса дедлайна. */
  bufferAvailableDays: number;
  /** Признак того, что изменения фактически применены — false для /simulate-shift, true для /apply-shift. */
  applied: boolean;
}

/** Данные для моделирования сдвига сроков задачи (What-if анализ). */
export interface SimulateShiftRequest {
  /** Новая дата начала задачи, от которой рассчитывается сдвиг. */
  newStartDate?: string;
  /** Альтернативный способ задать сдвиг — количество дней смещения (может быть отрицательным). */
  shiftDays?: number;
}

/** Данные для применения сдвига цепочки зависимых задач. */
export interface ApplyShiftRequest {
  /** Количество дней смещения задачи (может быть отрицательным). */
  shiftDays?: number;
  /** Если true — сначала пытаться компенсировать сдвиг за счёт резерва (буфера) задачи/проекта. */
  compensateFromBuffer?: boolean;
  /** Если true — отправить уведомления ответственным за все затронутые сдвигом задачи. */
  notifyAssignees?: boolean;
}
