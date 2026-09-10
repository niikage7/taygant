"use client";

import { CircleCheck, Link2, Unlink } from "lucide-react";
import Link from "next/link";

import { LinkTasksDialog } from "@/components/task/link-tasks-dialog";
import { Button } from "@/components/ui/button";
import { Card, CardBody, CardHeader, CardTitle } from "@/components/ui/card";
import { Progress } from "@/components/ui/progress";
import { useAssignTasksToMilestone } from "@/data/queries";
import { formatDayMonth } from "@/lib/format";
import { TASK_STATUS_META } from "@/lib/task-status";
import { cn } from "@/lib/utils";
import type { Task } from "@/types";

/**
 * Работы, ведущие к вехе.
 *
 * Принадлежность — поле `milestoneId` задачи, а не связь: веха агрегирует
 * работы, а зависимости описывают порядок их выполнения. Это разные вещи, и
 * живут они в разных блоках карточки.
 */
export function MilestoneTasksCard({
  milestoneTaskId,
  projectTasks,
}: {
  milestoneTaskId: string;
  projectTasks: Task[];
}) {
  const assignTasks = useAssignTasksToMilestone();
  const tasks = projectTasks.filter((task) => task.milestoneId === milestoneTaskId);
  const done = tasks.filter((task) => task.status === "done").length;

  return (
    <Card>
      <CardHeader className="items-start">
        <div>
          <CardTitle>Задачи вехи</CardTitle>
          <p className="mt-0.5 text-[13px] text-ink-muted">
            {tasks.length > 0
              ? `Выполнено ${done} из ${tasks.length}`
              : "К вехе пока не привязано ни одной задачи"}
          </p>
        </div>
        <LinkTasksDialog
          milestoneTaskId={milestoneTaskId}
          candidates={projectTasks}
          trigger={
            <Button size="sm">
              <Link2 />
              Привязать задачи
            </Button>
          }
        />
      </CardHeader>

      <CardBody>
        {tasks.length > 0 ? (
          <>
            <Progress
              value={(done / tasks.length) * 100}
              className="mb-3 h-1.5"
              barClassName={done === tasks.length ? "bg-success" : undefined}
              label="Прогресс вехи"
            />
            <ul className="space-y-1.5">
              {tasks.map((task) => {
                const status = TASK_STATUS_META[task.status];
                return (
                  <li
                    key={task.id}
                    className="flex items-center gap-2.5 rounded-control bg-surface-subtle px-3 py-2"
                  >
                    <CircleCheck
                      className={cn(
                        "size-4 shrink-0",
                        task.status === "done" ? "text-success" : "text-ink-faint",
                      )}
                    />
                    <Link
                      href={`/tasks/${task.id}`}
                      className="min-w-0 flex-1 rounded-control focus-visible:focus-ring"
                    >
                      <span className="truncate text-[13px] text-ink hover:text-brand">
                        <span className="font-mono text-ink-faint">#{task.wbsNumber}</span>{" "}
                        {task.title}
                      </span>
                    </Link>
                    <span className="shrink-0 font-mono text-[11px] text-ink-faint">
                      {formatDayMonth(task.endDate)}
                    </span>
                    <span className="shrink-0 text-[11px] text-ink-muted">
                      {status.label}
                    </span>
                    <button
                      type="button"
                      onClick={() =>
                        assignTasks.mutate({ taskIds: [task.id], milestoneId: null })
                      }
                      disabled={assignTasks.isPending}
                      aria-label={`Отвязать «${task.title}» от вехи`}
                      className="shrink-0 rounded-control p-1 text-ink-faint transition-colors hover:text-danger focus-visible:focus-ring disabled:opacity-50"
                    >
                      <Unlink className="size-3.5" />
                    </button>
                  </li>
                );
              })}
            </ul>
          </>
        ) : (
          <p className="py-6 text-center text-[13px] text-ink-muted">
            Отметьте работы, завершение которых означает достижение этой вехи.
          </p>
        )}
      </CardBody>
    </Card>
  );
}
