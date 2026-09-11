import type { Task, TaskStatus } from "@/types";

/** Колонка канбан-доски — ровно один из статусов, которые пользователь вправе выставить. */
export type BoardColumnId = Extract<TaskStatus, "planned" | "in_progress" | "done">;

export const BOARD_COLUMNS: { id: BoardColumnId; title: string }[] = [
  { id: "planned", title: "План" },
  { id: "in_progress", title: "В работе" },
  { id: "done", title: "Завершены" },
];

const PLACEMENT_KEY = "taygant.boardPlacement";

type Placements = Record<string, BoardColumnId>;

function readPlacements(): Placements {
  if (typeof window === "undefined") return {};
  try {
    const raw = window.localStorage.getItem(PLACEMENT_KEY);
    return raw ? (JSON.parse(raw) as Placements) : {};
  } catch {
    return {};
  }
}

function writePlacements(placements: Placements): void {
  try {
    window.localStorage.setItem(PLACEMENT_KEY, JSON.stringify(placements));
  } catch {
    // Приватный режим может запрещать запись — перенос тогда живёт до перезагрузки.
  }
}

/**
 * Запоминает, в какую колонку пользователь перенёс задачу.
 *
 * Нужно только задачам с вычисленным статусом (`overdue`/`blocked`): сервер
 * хранит под ними planned или in_progress, но в ответе отдаёт уже вычисленный
 * статус, и какой из двух сохранён — из API не узнать. Без отметки просроченная
 * задача, перенесённая в «В работе», после обновления данных вернулась бы в «План».
 */
export function rememberPlacement(taskId: string, column: BoardColumnId): void {
  const placements = readPlacements();
  placements[taskId] = column;
  writePlacements(placements);
}

/**
 * Колонка задачи на доске.
 *
 * planned/in_progress/done — как есть. Для вычисленных статусов — последний
 * перенос на этой доске, а если его не было — по прогрессу: у начатой задачи
 * (прогресс больше нуля) работа уже идёт, у нетронутой — ещё нет. Заблокированная
 * задача без отметки почти всегда нетронута, так что попадает в «План».
 */
export function columnOf(task: Task, placements: Placements): BoardColumnId {
  switch (task.status) {
    case "planned":
    case "in_progress":
    case "done":
      return task.status;
    default:
      return placements[task.id] ?? (task.progressPercent > 0 ? "in_progress" : "planned");
  }
}

/** Раскладывает задачи по колонкам, сохраняя порядок: сначала ближайший срок. */
export function groupByColumn(tasks: Task[]): Record<BoardColumnId, Task[]> {
  const placements = readPlacements();
  const groups: Record<BoardColumnId, Task[]> = { planned: [], in_progress: [], done: [] };
  const sorted = [...tasks].sort(
    (a, b) => a.endDate.localeCompare(b.endDate) || a.wbsNumber.localeCompare(b.wbsNumber),
  );
  for (const task of sorted) groups[columnOf(task, placements)].push(task);
  return groups;
}
