import type { Task, TaskDependency } from "@/types";

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
 * Поля «задача принадлежит вехе» в модели нет (см.
 * `docs/backend-request-milestones.md`), поэтому принадлежность выводится из
 * связей: задача входит в веху, если она её прямой предшественник. Задача может
 * вести к нескольким вехам — тогда она относится к первой по порядку реестра,
 * иначе одна и та же строка задваивалась бы и ломала соответствие с таймлайном.
 *
 * @param collapsed идентификаторы свёрнутых вех — их дети не попадают в результат
 */
export function buildGanttRows(
  tasks: Task[],
  dependencies: TaskDependency[],
  collapsed: ReadonlySet<string>,
): GanttRow[] {
  const milestones = tasks.filter((task) => task.isMilestone);
  const milestoneIds = new Set(milestones.map((task) => task.id));

  const predecessorsOf = new Map<string, string[]>();
  for (const dependency of dependencies) {
    if (!milestoneIds.has(dependency.successorTaskId)) continue;
    const list = predecessorsOf.get(dependency.successorTaskId) ?? [];
    list.push(dependency.predecessorTaskId);
    predecessorsOf.set(dependency.successorTaskId, list);
  }

  const taskById = new Map(tasks.map((task) => [task.id, task]));
  const claimed = new Set<string>();
  const childrenOf = new Map<string, Task[]>();

  for (const milestone of milestones) {
    const children: Task[] = [];
    for (const predecessorId of predecessorsOf.get(milestone.id) ?? []) {
      const child = taskById.get(predecessorId);
      // Веху, ведущую к другой вехе, не вкладываем: она сама заголовок группы.
      if (!child || child.isMilestone || claimed.has(child.id)) continue;
      claimed.add(child.id);
      children.push(child);
    }
    childrenOf.set(milestone.id, children);
  }

  const rows: GanttRow[] = [];

  // Задачи, не относящиеся ни к одной вехе, идут первыми — в исходном порядке.
  for (const task of tasks) {
    if (task.isMilestone || claimed.has(task.id)) continue;
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
