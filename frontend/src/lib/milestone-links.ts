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

/**
 * Все задачи, до которых можно дойти от `taskId` по связям вперёд: его
 * последователи, их последователи и так далее. Сама задача в результат не входит.
 */
export function successorsReachableFrom(taskId: string, dependencies: TaskDependency[]): Set<string> {
  const next = new Map<string, string[]>();
  for (const { predecessorTaskId, successorTaskId } of dependencies) {
    const list = next.get(predecessorTaskId) ?? [];
    list.push(successorTaskId);
    next.set(predecessorTaskId, list);
  }

  const reached = new Set<string>();
  const queue = [taskId];
  while (queue.length > 0) {
    const current = queue.shift() as string;
    for (const successor of next.get(current) ?? []) {
      if (successor === taskId || reached.has(successor)) continue;
      reached.add(successor);
      queue.push(successor);
    }
  }
  return reached;
}

/**
 * Какие задачи можно сделать предшественниками задачи-вехи.
 *
 * Не предлагаем саму веху и уже привязанных предшественников, а задачи, до
 * которых веха дотягивается по связям вперёд, возвращаем отдельно: связь с ними
 * замкнула бы цикл, и бэкенд её отклонит («такая связь создала бы цикл»).
 * Показать их выключенными честнее, чем молча не показывать.
 */
export function predecessorCandidates(
  milestoneTask: Task,
  tasks: Task[],
  dependencies: TaskDependency[],
): { available: Task[]; wouldCycle: Task[] } {
  const existing = new Set(
    dependencies
      .filter((dependency) => dependency.successorTaskId === milestoneTask.id)
      .map((dependency) => dependency.predecessorTaskId),
  );
  const downstream = successorsReachableFrom(milestoneTask.id, dependencies);

  const available: Task[] = [];
  const wouldCycle: Task[] = [];
  for (const task of tasks) {
    if (task.id === milestoneTask.id || existing.has(task.id)) continue;
    if (downstream.has(task.id)) wouldCycle.push(task);
    else available.push(task);
  }
  return { available, wouldCycle };
}
