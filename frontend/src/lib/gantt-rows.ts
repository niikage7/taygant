import type { Milestone, Task } from "@/types";

/**
 * Строка реестра.
 *
 * Заголовок группы — контрольная точка проекта (`Milestone`), а НЕ задача с
 * `isMilestone`. Это разные сущности: `Task.milestoneId` ссылается на запись в
 * таблице вех, что проверяет `validateMilestone` на бэкенде. Задача с
 * `isMilestone` — просто элемент графика нулевой длительности, она сама может
 * принадлежать вехе.
 */
export type GanttRow =
  | { kind: "milestone"; milestone: Milestone; childCount: number }
  | { kind: "task"; task: Task; depth: number };

/**
 * Раскладывает задачи по контрольным точкам проекта.
 *
 * @param collapsed идентификаторы свёрнутых вех — их задачи не попадают в результат
 */
export function buildGanttRows(
  tasks: Task[],
  milestones: Milestone[],
  collapsed: ReadonlySet<string>,
): GanttRow[] {
  const known = new Set(milestones.map((milestone) => milestone.id));

  const childrenOf = new Map<string, Task[]>();
  for (const task of tasks) {
    if (!task.milestoneId || !known.has(task.milestoneId)) continue;
    const list = childrenOf.get(task.milestoneId) ?? [];
    list.push(task);
    childrenOf.set(task.milestoneId, list);
  }

  const rows: GanttRow[] = [];

  // Задачи вне вех идут первыми, в исходном порядке.
  for (const task of tasks) {
    if (task.milestoneId && known.has(task.milestoneId)) continue;
    rows.push({ kind: "task", task, depth: 0 });
  }

  for (const milestone of milestones) {
    const children = childrenOf.get(milestone.id) ?? [];
    rows.push({ kind: "milestone", milestone, childCount: children.length });
    if (collapsed.has(milestone.id)) continue;
    for (const child of children) {
      rows.push({ kind: "task", task: child, depth: 1 });
    }
  }

  return rows;
}
