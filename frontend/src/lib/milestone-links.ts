import type { Task, TaskDependency } from "@/types";

/**
 * Какие работы привязать к контрольной точке вместе с задачей-вехой.
 *
 * Работы, которые «ведут» к задаче-вехе, — это её предшественники по связям:
 * у самой вехи нулевая длительность, работы в ней нет. Поэтому, привязывая веху
 * к КТ, естественно привязать и их. Берём только прямых предшественников:
 * транзитивный обход по цепочке затянул бы в КТ полпроекта.
 *
 * Предшественники, уже привязанные к другой КТ, не переназначаются, а
 * возвращаются отдельно — чтобы показать, что их не тронули. Уже привязанные к
 * этой же КТ пропускаются: делать с ними нечего.
 */
export function predecessorsLeadingTo(
  milestoneTask: Task,
  dependencies: TaskDependency[],
  tasks: Task[],
  targetMilestoneId: string,
): { toLink: Task[]; inOtherMilestones: Task[] } {
  const predecessorIds = new Set(
    dependencies
      .filter((dependency) => dependency.successorTaskId === milestoneTask.id)
      .map((dependency) => dependency.predecessorTaskId),
  );
  predecessorIds.delete(milestoneTask.id);

  const toLink: Task[] = [];
  const inOtherMilestones: Task[] = [];
  // Обходим задачи, а не связи: порядок совпадёт с реестром, и дубликаты связей
  // к одной задаче не дадут повторов.
  for (const task of tasks) {
    if (!predecessorIds.has(task.id)) continue;
    if (!task.milestoneId) toLink.push(task);
    else if (task.milestoneId !== targetMilestoneId) inOtherMilestones.push(task);
  }
  return { toLink, inOtherMilestones };
}
