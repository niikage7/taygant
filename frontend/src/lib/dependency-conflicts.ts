import { addDays, differenceInCalendarDays, format, parseISO } from "date-fns";

import { addWorkdays } from "@/lib/working-calendar";
import type { DependencyType, Task, TaskDependency, WorkingCalendarType } from "@/types";

type TaskDates = Pick<Task, "id" | "startDate" | "endDate">;

/** Связь, которую нарушат новые даты задачи. */
export type DependencyConflict = {
  dependency: TaskDependency;
  /** Вторая задача связи — та, с которой возник конфликт. */
  otherTaskId: string;
  /** Кем вторая задача приходится проверяемой. */
  otherRole: "predecessor" | "successor";
};

const shiftIso = (iso: string, days: number) => format(addDays(parseISO(iso), days), "yyyy-MM-dd");

/**
 * Самое раннее допустимое начало последователя — как `requiredSuccessorStart`
 * на бэкенде: шаг по связи в рабочих днях, длительность — в календарных.
 */
export function requiredSuccessorStart(
  calendar: WorkingCalendarType,
  predecessor: Pick<Task, "startDate" | "endDate">,
  type: DependencyType,
  lagDays: number,
  successorDurationDays: number,
): string {
  switch (type) {
    case "SS":
      return addWorkdays(calendar, predecessor.startDate, lagDays);
    case "FF":
      return shiftIso(addWorkdays(calendar, predecessor.endDate, lagDays), -successorDurationDays);
    case "SF":
      return shiftIso(addWorkdays(calendar, predecessor.startDate, lagDays), -successorDurationDays);
    default:
      return addWorkdays(calendar, predecessor.endDate, 1 + lagDays);
  }
}

/**
 * Какие связи задачи нарушат её новые даты: она начнётся раньше, чем позволяет
 * предшественник, или её последователь окажется раньше, чем позволяет она.
 *
 * `checkSuccessors: false` — для каскадного переноса: последователей сервер
 * отодвинет сам, нарушить можно только связь с предшественниками.
 */
export function findDependencyConflicts({
  taskId,
  startDate,
  endDate,
  dependencies,
  tasks,
  calendar,
  checkSuccessors = true,
}: {
  taskId: string;
  startDate: string;
  endDate: string;
  dependencies: TaskDependency[];
  tasks: TaskDates[];
  calendar: WorkingCalendarType;
  checkSuccessors?: boolean;
}): DependencyConflict[] {
  const byId = new Map(tasks.map((task) => [task.id, task]));
  const moved = { startDate, endDate };
  const conflicts: DependencyConflict[] = [];

  for (const dependency of dependencies) {
    const { predecessorTaskId, successorTaskId, type, lagDays } = dependency;
    if (successorTaskId === taskId) {
      const predecessor = byId.get(predecessorTaskId);
      if (!predecessor) continue;
      const duration = differenceInCalendarDays(parseISO(endDate), parseISO(startDate));
      if (startDate < requiredSuccessorStart(calendar, predecessor, type, lagDays, duration)) {
        conflicts.push({ dependency, otherTaskId: predecessorTaskId, otherRole: "predecessor" });
      }
    } else if (checkSuccessors && predecessorTaskId === taskId) {
      const successor = byId.get(successorTaskId);
      if (!successor) continue;
      const duration = differenceInCalendarDays(parseISO(successor.endDate), parseISO(successor.startDate));
      if (successor.startDate < requiredSuccessorStart(calendar, moved, type, lagDays, duration)) {
        conflicts.push({ dependency, otherTaskId: successorTaskId, otherRole: "successor" });
      }
    }
  }
  return conflicts;
}

/** Текст предупреждения: «Задача начнётся раньше, чем закончится #4». */
export function describeConflict(
  conflict: DependencyConflict,
  numberOf: (taskId: string) => string | undefined,
): string {
  const other = `#${numberOf(conflict.otherTaskId) ?? "?"}`;
  const { type } = conflict.dependency;
  if (conflict.otherRole === "predecessor") {
    return type === "FS"
      ? `Задача начнётся раньше, чем закончится ${other}`
      : `Задача нарушит связь ${type} с предшественником ${other}`;
  }
  return type === "FS"
    ? `Задача закончится позже, чем начнётся ${other}`
    : `Задача нарушит связь ${type} с последователем ${other}`;
}
