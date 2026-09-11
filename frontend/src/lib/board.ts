import type { AssignableTaskStatus, Task } from "@/types";

/** Колонка канбан-доски — ровно один из статусов, которые пользователь вправе выставить. */
export type BoardColumnId = AssignableTaskStatus;

export const BOARD_COLUMNS: { id: BoardColumnId; title: string }[] = [
  { id: "planned", title: "План" },
  { id: "in_progress", title: "В работе" },
  { id: "done", title: "Завершены" },
];

/**
 * Колонка задачи на доске.
 *
 * Берётся из `baseStatus` — статуса, выставленного человеком. Поле `status`
 * для этого не годится: сервер подмешивает в него вычисленные `overdue` и
 * `blocked` (`resolveStatuses` в `service/tasks.go`), и просроченную задачу по
 * нему не отличить от просроченной-и-начатой. Раньше доска помнила перенос
 * таких задач в localStorage — из-за этого у каждого браузера была своя
 * раскладка, а неудавшийся перенос всё равно двигал карточку, потому что
 * откатывать локальную отметку было некому.
 */
export function columnOf(task: Task): BoardColumnId {
  return task.baseStatus;
}

/** Раскладывает задачи по колонкам, сохраняя порядок: сначала ближайший срок. */
export function groupByColumn(tasks: Task[]): Record<BoardColumnId, Task[]> {
  const groups: Record<BoardColumnId, Task[]> = { planned: [], in_progress: [], done: [] };
  const sorted = [...tasks].sort(
    (a, b) => a.endDate.localeCompare(b.endDate) || a.wbsNumber.localeCompare(b.wbsNumber),
  );
  for (const task of sorted) groups[columnOf(task)].push(task);
  return groups;
}
