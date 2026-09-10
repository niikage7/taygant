"use client";

import { CircleCheck, CircleDot, Circle, Flag, Gem } from "lucide-react";
import Link from "next/link";

import { ROW_HEIGHT } from "@/components/gantt/timeline";
import { Avatar } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { formatDayMonth } from "@/lib/format";
import { TASK_STATUS_META } from "@/lib/task-status";
import { cn } from "@/lib/utils";
import type { Task, TaskDependency } from "@/types";

/** Левая часть диаграммы — реестр задач, построчно совпадающий с таймлайном. */
export function TaskTable({
  tasks,
  allTasks,
  dependencies,
  highlightCriticalPath,
}: {
  /** Отображаемые строки (после фильтров). */
  tasks: Task[];
  /** Полный список задач проекта — нужен, чтобы найти номер предшественника,
   *  даже когда сам предшественник отфильтрован и в таблице не показан. */
  allTasks: Task[];
  dependencies: TaskDependency[];
  highlightCriticalPath: boolean;
}) {
  // Критический путь проходит через несколько задач, но в макете красным
  // выделена только та, что реально отстаёт: подсветка означает «требует
  // вмешательства», а не «лежит на критическом пути».
  const isAtRisk = (task: Task) =>
    highlightCriticalPath && task.isCriticalPath && task.planVsActualDeviationDays < 0;

  const predecessorOf = new Map(
    dependencies.map((d) => [d.successorTaskId, d.predecessorTaskId]),
  );
  const wbsOf = new Map(allTasks.map((task) => [task.id, task.wbsNumber]));

  return (
    <div className="w-[600px] shrink-0 border-r border-line">
      <div className="sticky top-0 z-10 flex h-12 items-end border-b border-line bg-surface px-3 pb-2 text-[11px] tracking-wider text-ink-faint uppercase">
        <span className="w-8 shrink-0 font-semibold">#</span>
        <span className="min-w-0 flex-1 truncate font-semibold">Наименование задачи</span>
        <span className="w-30 shrink-0 truncate pl-2 font-semibold">Исполнитель</span>
        <span className="w-20 shrink-0 font-semibold">Сроки</span>
        <span className="w-10 shrink-0 text-right font-semibold">Дней</span>
        <span className="w-24 shrink-0 pl-3 font-semibold">Статус</span>
        <span className="w-12 shrink-0 text-right font-semibold">Пред.</span>
      </div>

      <ul>
        {tasks.map((task) => {
          const critical = isAtRisk(task) && !task.isMilestone;
          const status = TASK_STATUS_META[task.status];
          const predecessor = predecessorOf.get(task.id);
          const StatusIcon =
            task.isMilestone
              ? Gem
              : task.status === "done"
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
              )}
              style={{ height: ROW_HEIGHT }}
            >
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
                ) : (
                  <StatusIcon
                    className={cn(
                      "size-4 shrink-0",
                      task.isMilestone && "text-brand",
                      !task.isMilestone && task.status === "done" && "text-success",
                      !task.isMilestone && task.status === "in_progress" && "text-warning",
                      !task.isMilestone && task.status === "planned" && "text-ink-faint",
                    )}
                  />
                )}
                <span
                  className={cn(
                    "truncate text-[13px]",
                    critical
                      ? "font-semibold text-danger"
                      : task.isMilestone
                        ? "font-semibold text-brand"
                        : "text-ink",
                  )}
                >
                  {task.title}
                </span>
              </Link>

              <span className="flex w-30 shrink-0 items-center gap-1.5">
                {task.assignee ? (
                  <>
                    <Avatar fullName={task.assignee.fullName} className="size-5 text-[9px]" />
                    <span className="truncate text-[13px] text-ink-muted">
                      {task.assignee.fullName}
                    </span>
                  </>
                ) : null}
              </span>

              <span
                className={cn(
                  "w-20 shrink-0 font-mono text-[11px] leading-tight",
                  critical ? "text-danger" : "text-ink-muted",
                )}
              >
                {task.isMilestone ? (
                  formatDayMonth(task.startDate)
                ) : (
                  <>
                    {formatDayMonth(task.startDate)} –<br />
                    {formatDayMonth(task.endDate)}
                  </>
                )}
              </span>

              <span className="w-10 shrink-0 text-right font-mono text-xs text-ink-muted">
                {task.durationCalendarDays}
              </span>

              <span className="w-24 shrink-0 pl-3">
                {task.isMilestone ? (
                  <Badge tone="brand" size="sm">Веха</Badge>
                ) : critical ? (
                  <Badge tone="danger" size="sm" dot>Крит. путь</Badge>
                ) : (
                  <Badge tone={status.tone} size="sm" dot={task.status !== "planned"}>
                    {status.label}
                  </Badge>
                )}
              </span>

              <span className="w-12 shrink-0 text-right font-mono text-xs text-brand">
                {predecessor ? `#${wbsOf.get(predecessor) ?? "?"}` : "—"}
              </span>
            </li>
          );
        })}
      </ul>
    </div>
  );
}
