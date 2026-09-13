import { ArrowRight, CircleAlert } from "lucide-react";
import Link from "next/link";

import { UserHoverCard } from "@/components/user/user-card";
import { Avatar } from "@/components/ui/avatar";
import { Card, CardBody, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { formatSigned, plural } from "@/lib/format";
import { cn } from "@/lib/utils";
import type { AttentionTask } from "@/types";

/** Реестр задач с отклонением от плана. */
export function AttentionTasks({ tasks }: { tasks: AttentionTask[] }) {
  return (
    <Card className="flex flex-col">
      <CardHeader>
        <CardTitle>
          <CircleAlert className="size-4 text-danger" />
          Задачи, требующие внимания
        </CardTitle>
        <span className="rounded-control bg-surface-muted px-2 py-1 font-mono text-2xs text-ink-muted">
          {tasks.length} {plural(tasks.length, ["задача", "задачи", "задач"])} в зоне риска
        </span>
      </CardHeader>
      <CardBody className="flex-1 pt-3">
        {tasks.length === 0 ? (
          <p className="py-8 text-center text-13 text-ink-muted">
            Отстающих задач нет — проект идёт по плану.
          </p>
        ) : (
          <table className="w-full text-left">
            <thead>
              <tr className="border-b border-line text-2xs tracking-wider text-ink-faint uppercase">
                <th scope="col" className="pb-2 font-semibold">
                  WBS / Задача
                </th>
                <th scope="col" className="hidden pb-2 font-semibold sm:table-cell">
                  Исполнитель
                </th>
                <th scope="col" className="pb-2 text-right font-semibold">
                  Отклонение
                </th>
              </tr>
            </thead>
            <tbody>
              {tasks.map(({ task, deviationDays }) => {
                const late = deviationDays < 0;
                return (
                  <tr key={task.id} className="border-b border-line last:border-0">
                    <td className="py-3 pr-3">
                      <Link
                        href={`/tasks/${task.id}`}
                        className="group flex items-baseline gap-2 rounded-control focus-visible:focus-ring"
                      >
                        <span
                          className={cn(
                            "font-mono text-xs",
                            task.isCriticalPath ? "text-danger" : "text-warning",
                          )}
                        >
                          #{task.wbsNumber}
                        </span>
                        <span className="min-w-0">
                          <span className="text-13 text-ink group-hover:text-brand">
                            {task.title}
                          </span>
                          {/* На телефоне колонки исполнителя нет — имя под названием. */}
                          {task.assignee ? (
                            <span className="mt-0.5 block text-xs text-ink-muted sm:hidden">
                              {task.assignee.fullName}
                            </span>
                          ) : null}
                        </span>
                      </Link>
                    </td>
                    <td className="hidden py-3 pr-3 sm:table-cell">
                      {task.assignee ? (
                        <UserHoverCard user={task.assignee}>
                          <span className="flex items-center gap-1.5">
                            <Avatar fullName={task.assignee.fullName} className="size-5 text-2xs" />
                            <span className="text-13 whitespace-nowrap text-ink-muted">
                              {task.assignee.fullName}
                            </span>
                          </span>
                        </UserHoverCard>
                      ) : (
                        <span className="text-13 text-ink-faint">—</span>
                      )}
                    </td>
                    <td
                      className={cn(
                        "py-3 text-right font-mono text-xs whitespace-nowrap",
                        late ? "text-danger" : "text-ink-muted",
                      )}
                    >
                      {formatSigned(deviationDays, Number.isInteger(deviationDays) ? 0 : 1)}{" "}
                      {plural(deviationDays, ["день", "дня", "дн."])}
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        )}
      </CardBody>
      <CardFooter>
        <span className="text-ink-muted">Показаны критические и отстающие позиции</span>
        <Link
          href="/gantt"
          className="flex items-center gap-1.5 rounded-control font-medium text-brand hover:text-brand-hover focus-visible:focus-ring"
        >
          Открыть реестр всех задач
          <ArrowRight className="size-3.5" />
        </Link>
      </CardFooter>
    </Card>
  );
}
