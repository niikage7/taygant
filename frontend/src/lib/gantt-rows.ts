import type { Task } from "@/types";

/** Строка реестра: задача плюс сведения о её месте в группировке. */
export type GanttRow = {
  task: Task;
  /** 0 — веха или задача вне вех, 1 — задача внутри вехи. */
  depth: number;
  /** Для вехи — сколько задач в неё входит. */
  childCount: number;
};

/**
 * Раскладывает задачи по вехам для реестра и таймлайна.
 *
 * Принадлежность берётся из `task.milestoneId` — настоящего поля модели. Раньше
 * её приходилось выводить из связей, потому что поля не было; эвристика убрана.
 *
 * Веха здесь — задача с `isMilestone`. Вехи проекта (`Milestone`) живут отдельно
 * и на реестр не влияют, они показаны отметками на таймлайне.
 *
 * @param collapsed идентификаторы свёрнутых вех — их дети не попадают в результат
 */
export function buildGanttRows(
  tasks: Task[],
  collapsed: ReadonlySet<string>,
): GanttRow[] {
  const milestones = tasks.filter((task) => task.isMilestone);
  const milestoneIds = new Set(milestones.map((task) => task.id));

  const childrenOf = new Map<string, Task[]>();
  for (const task of tasks) {
    // Веху, привязанную к другой вехе, не вкладываем: она сама заголовок группы.
    if (task.isMilestone || !task.milestoneId) continue;
    if (!milestoneIds.has(task.milestoneId)) continue;
    const list = childrenOf.get(task.milestoneId) ?? [];
    list.push(task);
    childrenOf.set(task.milestoneId, list);
  }

  const rows: GanttRow[] = [];

  // Задачи вне вех идут первыми, в исходном порядке.
  for (const task of tasks) {
    if (task.isMilestone) continue;
    if (task.milestoneId && milestoneIds.has(task.milestoneId)) continue;
    rows.push({ task, depth: 0, childCount: 0 });
  }

  for (const milestone of milestones) {
    const children = childrenOf.get(milestone.id) ?? [];
    rows.push({ task: milestone, depth: 0, childCount: children.length });
    if (collapsed.has(milestone.id)) continue;
    for (const child of children) {
      rows.push({ task: child, depth: 1, childCount: 0 });
    }
  }

  return rows;
}
