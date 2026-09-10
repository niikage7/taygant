"use client";

import { ArrowLeftToLine, ArrowRightFromLine, Trash2 } from "lucide-react";
import Link from "next/link";

import { AddDependencyDialog } from "@/components/task/add-dependency-dialog";
import { Badge } from "@/components/ui/badge";
import { Card, CardBody, CardHeader, CardTitle } from "@/components/ui/card";
import { useDeleteDependency } from "@/data/queries";
import { plural } from "@/lib/format";
import { cn } from "@/lib/utils";
import type { DependencyType, Task, TaskDependency } from "@/types";

const DEPENDENCY_LABEL: Record<DependencyType, string> = {
  FS: "FS (Окончание-к-Началу)",
  SS: "SS (Начало-к-Началу)",
  FF: "FF (Окончание-к-Окончанию)",
  SF: "SF (Начало-к-Окончанию)",
};

/** Сетевой граф связей задачи: предшественники и последователи. */
export function DependencyGraph({
  taskId,
  taskNumber,
  isMilestone,
  predecessors,
  successors,
  projectTasks,
}: {
  taskId: string;
  taskNumber: string;
  /** У вехи входящие связи — это список работ, которые к ней ведут. */
  isMilestone: boolean;
  predecessors: TaskDependency[];
  successors: TaskDependency[];
  /** Все задачи проекта — источник выбора для новой связи и подписей номеров. */
  projectTasks: Task[];
}) {
  const deleteDependency = useDeleteDependency(taskId);
  const linkedTaskIds = [
    ...predecessors.map((item) => item.predecessorTaskId),
    ...successors.map((item) => item.successorTaskId),
  ];

  return (
    <Card>
      <CardHeader className="items-start">
        <div>
          <CardTitle>Сетевой граф связей задачи #{taskNumber}</CardTitle>
          <p className="mt-0.5 text-[13px] text-ink-muted">
            Определяет динамический расчёт сроков по алгоритму Critical Path Method (CPM)
          </p>
        </div>
        <AddDependencyDialog
          taskId={taskId}
          candidates={projectTasks}
          linkedTaskIds={linkedTaskIds}
        />
      </CardHeader>

      <CardBody>
        <div className="grid gap-3 lg:grid-cols-2">
          <DependencyColumn
            icon={<ArrowLeftToLine className="size-3.5" />}
            title={isMilestone ? "Задачи, ведущие к вехе" : "Входящие связи (предшественники)"}
            items={predecessors}
            relatedIdOf={(item) => item.predecessorTaskId}
            relatedTitleOf={(item) => item.predecessorTitle}
            projectTasks={projectTasks}
            onDelete={(id) => deleteDependency.mutate(id)}
            deletingId={deleteDependency.isPending ? deleteDependency.variables : null}
            emptyText="Предшественников нет — задача может начаться сразу."
          />
          <DependencyColumn
            icon={<ArrowRightFromLine className="size-3.5" />}
            title="Исходящие связи (последователи)"
            items={successors}
            relatedIdOf={(item) => item.successorTaskId}
            relatedTitleOf={(item) => item.successorTitle}
            projectTasks={projectTasks}
            onDelete={(id) => deleteDependency.mutate(id)}
            deletingId={deleteDependency.isPending ? deleteDependency.variables : null}
            emptyText="От этой задачи ничего не зависит."
          />
        </div>
      </CardBody>
    </Card>
  );
}

function DependencyColumn({
  icon,
  title,
  items,
  relatedIdOf,
  relatedTitleOf,
  projectTasks,
  onDelete,
  deletingId,
  emptyText,
}: {
  icon: React.ReactNode;
  title: string;
  items: TaskDependency[];
  relatedIdOf: (item: TaskDependency) => string;
  relatedTitleOf: (item: TaskDependency) => string | undefined;
  projectTasks: Task[];
  onDelete: (dependencyId: string) => void;
  deletingId: string | null | undefined;
  emptyText: string;
}) {
  const criticalCount = items.filter((item) => item.isCritical).length;
  const byId = new Map(projectTasks.map((task) => [task.id, task]));

  return (
    <section className="rounded-control border border-line">
      <header className="flex items-start justify-between gap-3 border-b border-line px-3 py-2.5">
        <h4 className="flex items-start gap-1.5 text-[11px] font-semibold tracking-wider text-ink-muted uppercase">
          <span className="mt-px text-ink-faint">{icon}</span>
          {title}
        </h4>
        <span
          className={cn(
            "shrink-0 text-right text-[11px] font-semibold",
            criticalCount > 0 ? "text-danger" : "text-ink-muted",
          )}
        >
          {items.length} {plural(items.length, ["связь", "связи", "связей"])}
          {criticalCount > 0 ? " · критичные" : ""}
        </span>
      </header>

      {items.length === 0 ? (
        <p className="px-3 py-6 text-center text-[13px] text-ink-muted">{emptyText}</p>
      ) : (
        <ul className="space-y-2 p-3">
          {items.map((item) => {
            const relatedId = relatedIdOf(item);
            const related = byId.get(relatedId);
            const title = relatedTitleOf(item) ?? related?.title ?? "Задача";

            return (
              <li key={item.id} className="rounded-control bg-surface-subtle p-3">
                <div className="flex items-start justify-between gap-2">
                  {/* Переход на связанную задачу — граф должен быть проходимым. */}
                  <Link
                    href={`/tasks/${relatedId}`}
                    className="group min-w-0 rounded-control focus-visible:focus-ring"
                  >
                    <span className="text-[13px] leading-snug font-medium text-ink group-hover:text-brand">
                      {related?.wbsNumber ? (
                        <span className="font-mono text-brand">#{related.wbsNumber} </span>
                      ) : null}
                      {title}
                    </span>
                  </Link>

                  <span className="flex shrink-0 items-center gap-1.5">
                    {item.isCritical ? (
                      <Badge tone="danger" size="sm">
                        Критичная
                      </Badge>
                    ) : null}
                    <button
                      type="button"
                      onClick={() => onDelete(item.id)}
                      disabled={deletingId === item.id}
                      aria-label={`Удалить связь с задачей «${title}»`}
                      className="rounded-control p-1 text-ink-faint transition-colors hover:text-danger focus-visible:focus-ring disabled:opacity-50"
                    >
                      <Trash2 className="size-3.5" />
                    </button>
                  </span>
                </div>

                <p className="mt-2 flex flex-wrap items-center gap-x-3 gap-y-1">
                  <span className="rounded-control bg-surface-muted px-2 py-1 font-mono text-[11px] text-ink-muted">
                    {DEPENDENCY_LABEL[item.type]}
                  </span>
                  <span className="font-mono text-[11px] text-ink-faint">
                    Лаг: {item.lagDays} дн.
                  </span>
                </p>
              </li>
            );
          })}
        </ul>
      )}
    </section>
  );
}
