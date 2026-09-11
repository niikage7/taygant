"use client";

import { ChevronDown, ChevronRight, Circle, CircleCheck, CircleDot, Flag } from "lucide-react";
import Link from "next/link";

import { ROW_HEIGHT } from "@/components/gantt/timeline";
import { TASK_TABLE_COMPACT_WIDTH, TASK_TABLE_WIDTH } from "@/lib/gantt";
import { UserHoverCard } from "@/components/user/user-card";
import { Avatar } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { formatDayMonth } from "@/lib/format";
import { TASK_STATUS_META } from "@/lib/task-status";
import { cn } from "@/lib/utils";
import type { GanttRow } from "@/lib/gantt-rows";
import type { Task, TaskDependency } from "@/types";

/** Левая часть диаграммы — реестр задач, построчно совпадающий с таймлайном. */
export function TaskTable({
  rows,
  allTasks,
  dependencies,
  collapsed,
  onToggleCollapse,
  highlightCriticalPath,
  compact = false,
}: {
  /** Готовые строки в порядке отрисовки — тот же массив получает таймлайн. */
  rows: GanttRow[];
  /** Полный список задач проекта — нужен, чтобы найти номер предшественника,
   *  даже когда сам предшественник отфильтрован и в таблице не показан. */
  allTasks: Task[];
  dependencies: TaskDependency[];
  collapsed: ReadonlySet<string>;
  onToggleCollapse: (milestoneId: string) => void;
  highlightCriticalPath: boolean;
  /** Узкий экран: только номер и название, остальное видно в карточке задачи. */
  compact?: boolean;
}) {
  // Критический путь проходит через несколько задач, но в макете красным
  // выделена только та, что реально отстаёт: подсветка означает «требует
  // вмешательства», а не «лежит на критическом пути».
  const isAtRisk = (task: Task) =>
    highlightCriticalPath && task.isCriticalPath && task.planVsActualDeviationDays < 0;

  const predecessorOf = new Map(dependencies.map((d) => [d.successorTaskId, d.predecessorTaskId]));
  const wbsOf = new Map(allTasks.map((task) => [task.id, task.wbsNumber]));

  return (
    // Реестр приморожен слева: при горизонтальной прокрутке таймлайна названия
    // задач остаются на месте, иначе по широкому графику непонятно, чей отрезок
    // сейчас видно. Непрозрачный фон обязателен — под ним проезжает таймлайн.
    <div
      className="sticky left-0 z-20 shrink-0 border-r border-line bg-surface"
      style={{ width: compact ? TASK_TABLE_COMPACT_WIDTH : TASK_TABLE_WIDTH }}
    >
      <div className="sticky top-0 z-10 flex h-12 items-end border-b border-line bg-surface px-3 pb-2 text-2xs tracking-wider text-ink-faint uppercase">
        <span className="w-8 shrink-0 font-semibold">#</span>
        <span className="min-w-0 flex-1 truncate font-semibold">
          {compact ? "Задача" : "Наименование задачи"}
        </span>
        {compact ? null : (
          <>
            <span className="w-30 shrink-0 truncate pl-2 font-semibold">Исполнитель</span>
            <span className="w-20 shrink-0 font-semibold">Сроки</span>
            <span className="w-10 shrink-0 text-right font-semibold">Дней</span>
            <span className="w-24 shrink-0 pl-3 font-semibold">Статус</span>
            <span className="w-12 shrink-0 text-right font-semibold">Пред.</span>
          </>
        )}
      </div>

      <ul>
        {rows.map((row) => {
          if (row.kind === "milestone") {
            const { milestone, childCount } = row;
            const isCollapsed = collapsed.has(milestone.id);
            const complete =
              milestone.tasksTotal > 0 && milestone.tasksDone === milestone.tasksTotal;

            return (
              <li
                key={milestone.id}
                className="flex items-center gap-2 border-y border-accent/30 bg-accent-tint px-3"
                style={{ height: ROW_HEIGHT }}
              >
                <button
                  type="button"
                  onClick={() => onToggleCollapse(milestone.id)}
                  disabled={childCount === 0}
                  aria-expanded={!isCollapsed}
                  aria-label={
                    isCollapsed
                      ? `Развернуть веху «${milestone.name}»`
                      : `Свернуть веху «${milestone.name}»`
                  }
                  className="flex size-5 shrink-0 items-center justify-center rounded-control text-accent transition-colors hover:bg-accent/10 focus-visible:focus-ring disabled:opacity-30"
                >
                  {childCount === 0 ? (
                    <span className="size-1.5 rounded-full bg-current" />
                  ) : isCollapsed ? (
                    <ChevronRight className="size-4" />
                  ) : (
                    <ChevronDown className="size-4" />
                  )}
                </button>

                <span className="size-3 shrink-0 rotate-45 rounded-[2px] bg-accent" />

                <span className="min-w-0 flex-1 truncate text-13 font-semibold text-accent">
                  {milestone.code ? (
                    <span className="font-mono opacity-70">{milestone.code} </span>
                  ) : null}
                  {milestone.name}
                </span>

                {compact ? null : (
                  <Badge
                    tone={complete ? "success" : childCount > 0 ? "accent" : "neutral"}
                    size="sm"
                  >
                    {milestone.tasksTotal > 0
                      ? `${milestone.tasksDone} / ${milestone.tasksTotal}`
                      : "нет задач"}
                  </Badge>
                )}

                <span className="shrink-0 font-mono text-2xs text-accent">
                  {formatDayMonth(milestone.plannedDate)}
                </span>
              </li>
            );
          }

          const { task, depth } = row;
          const critical = isAtRisk(task);
          // Задача-веха — событие нулевой длительности: у неё нет «плана» и
          // «работы», есть только «ещё не достигнута» и «достигнута». Вычисленные
          // сервером overdue/blocked показываем как есть — это сигнал о риске.
          const isMilestoneTask = task.isMilestone;
          const status =
            isMilestoneTask && task.status === "done"
              ? { label: "Достигнута", tone: "success" as const }
              : isMilestoneTask && (task.status === "planned" || task.status === "in_progress")
                ? { label: "Веха", tone: "brand" as const }
                : TASK_STATUS_META[task.status];
          const predecessor = predecessorOf.get(task.id);
          const StatusIcon =
            task.status === "done"
              ? CircleCheck
              : task.status === "in_progress"
                ? CircleDot
                : Circle;

          return (
            <li
              key={task.id}
              className={cn(
                "flex items-center border-b border-line px-3",
                critical && "bg-danger-tint",
                isMilestoneTask && !critical && "bg-brand-tint/40",
              )}
              style={{ height: ROW_HEIGHT }}
            >
              {/* Вложенность показываем отступом: задача читается как часть вехи. */}
              {depth > 0 ? (
                <span aria-hidden className="mr-1 h-full w-4 shrink-0 border-l border-accent/40" />
              ) : null}
              <span
                className={cn(
                  "w-8 shrink-0 font-mono text-xs",
                  critical ? "font-bold text-danger" : "text-ink-faint",
                )}
              >
                {task.wbsNumber}
              </span>

              <Link
                href={`/tasks/${task.id}`}
                className="flex min-w-0 flex-1 items-center gap-2 rounded-control pr-2 focus-visible:focus-ring"
              >
                {critical ? (
                  <Flag className="size-4 shrink-0 text-danger" />
                ) : isMilestoneTask ? (
                  // Тот же ромб, что на таймлайне и в легенде: строку реестра и
                  // точку на графике глаз связывает по форме.
                  <span
                    aria-hidden
                    className={cn(
                      "mx-[3px] size-2.5 shrink-0 rotate-45 rounded-[1px]",
                      task.status === "done" ? "bg-success" : "bg-brand",
                    )}
                  />
                ) : (
                  <StatusIcon
                    className={cn(
                      "size-4 shrink-0",
                      task.status === "done" && "text-success",
                      task.status === "in_progress" && "text-warning",
                      task.status === "planned" && "text-ink-faint",
                    )}
                  />
                )}
                <span
                  className={cn(
                    "truncate text-13",
                    critical
                      ? "font-semibold text-danger"
                      : isMilestoneTask
                        ? "font-semibold text-brand"
                        : "text-ink",
                  )}
                >
                  {isMilestoneTask ? <span className="sr-only">Веха: </span> : null}
                  {task.title}
                </span>
              </Link>

              {compact ? null : (
              <>
              <span className="flex w-30 shrink-0 items-center gap-1.5">
                {task.assignee ? (
                  <UserHoverCard user={task.assignee}>
                    <span className="flex min-w-0 items-center gap-1.5">
                      <Avatar fullName={task.assignee.fullName} className="size-5 text-2xs" />
                      <span className="truncate text-13 text-ink-muted">
                        {task.assignee.fullName}
                      </span>
                    </span>
                  </UserHoverCard>
                ) : null}
              </span>

              <span
                className={cn(
                  "w-20 shrink-0 font-mono text-2xs leading-tight",
                  critical ? "text-danger" : "text-ink-muted",
                )}
              >
                {isMilestoneTask ? (
                  formatDayMonth(task.startDate)
                ) : (
                  <>
                    {formatDayMonth(task.startDate)} –<br />
                    {formatDayMonth(task.endDate)}
                  </>
                )}
              </span>

              <span className="w-10 shrink-0 text-right font-mono text-xs text-ink-muted">
                {isMilestoneTask ? "—" : task.durationCalendarDays}
              </span>

              <span className="w-24 shrink-0 pl-3">
                {critical ? (
                  <Badge tone="danger" size="sm" dot>
                    Крит. путь
                  </Badge>
                ) : (
                  <Badge tone={status.tone} size="sm" dot={status.tone !== "brand"}>
                    {status.label}
                  </Badge>
                )}
              </span>

              <span className="w-12 shrink-0 text-right font-mono text-xs text-brand">
                {predecessor ? `#${wbsOf.get(predecessor) ?? "?"}` : "—"}
              </span>
              </>
              )}
            </li>
          );
        })}
      </ul>
    </div>
  );
}
