"use client";

import { ArrowLeftToLine, ArrowRightFromLine, Link2, Trash2 } from "lucide-react";
import Link from "next/link";

import { AddDependencyDialog } from "@/components/task/add-dependency-dialog";
import { LinkPredecessorsDialog } from "@/components/task/link-predecessors-dialog";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
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

/**
 * Сетевой граф связей задачи: предшественники и последователи.
 *
 * У задачи-вехи вместо двух равноправных колонок — «Задачи, ведущие к вехе» с
 * массовой привязкой: у события нулевой длительности главный вопрос «что должно
 * завершиться до него», а выбор направления и типа на каждую связь только мешал.
 * Последователи у вехи законны (после приёмки — запуск), поэтому они видны, но
 * только если есть и без кнопки добавления: такую связь естественно заводить с
 * карточки задачи-последователя.
 */
export function DependencyGraph({
  taskId,
  taskNumber,
  isMilestone,
  predecessors,
  successors,
  projectTasks,
  canManage,
}: {
  taskId: string;
  taskNumber: string;
  isMilestone: boolean;
  predecessors: TaskDependency[];
  successors: TaskDependency[];
  /** Все задачи проекта — источник выбора для новой связи и подписей номеров. */
  projectTasks: Task[];
  /** Создавать и рвать связи может только полный доступ. */
  canManage: boolean;
}) {
  const deleteDependency = useDeleteDependency(taskId);
  const deletingId = deleteDependency.isPending ? deleteDependency.variables : null;
  const linkedTaskIds = [
    ...predecessors.map((item) => item.predecessorTaskId),
    ...successors.map((item) => item.successorTaskId),
  ];

  if (isMilestone) {
    const milestoneTask = projectTasks.find((task) => task.id === taskId);

    return (
      <Card>
        <CardHeader className="items-start">
          <div>
            <CardTitle>Задачи, ведущие к вехе #{taskNumber}</CardTitle>
            <p className="mt-0.5 text-13 text-ink-muted">
              Веха наступит, когда завершатся эти работы; их сдвиг сдвигает и веху (расчёт по
              методу критического пути, CPM)
            </p>
          </div>
          {canManage && milestoneTask ? (
            <LinkPredecessorsDialog
              milestoneTask={milestoneTask}
              projectTasks={projectTasks}
              trigger={
                <Button variant="secondary" size="sm">
                  <Link2 />
                  Привязать задачи
                </Button>
              }
            />
          ) : null}
        </CardHeader>

        <CardBody>
          <div className="space-y-3">
            <DependencyColumn
              icon={<ArrowLeftToLine className="size-3.5" />}
              title="Ведут к вехе"
              items={predecessors}
              relatedIdOf={(item) => item.predecessorTaskId}
              relatedTitleOf={(item) => item.predecessorTitle}
              projectTasks={projectTasks}
              onDelete={(id) => deleteDependency.mutate(id)}
              deletingId={deletingId}
              canManage={canManage}
              emptyText={
                canManage
                  ? "К вехе пока не ведёт ни одна работа — отметьте их кнопкой «Привязать задачи»."
                  : "К вехе пока не ведёт ни одна работа."
              }
            />
            {successors.length > 0 ? (
              <DependencyColumn
                icon={<ArrowRightFromLine className="size-3.5" />}
                title="После вехи"
                items={successors}
                relatedIdOf={(item) => item.successorTaskId}
                relatedTitleOf={(item) => item.successorTitle}
                projectTasks={projectTasks}
                onDelete={(id) => deleteDependency.mutate(id)}
                deletingId={deletingId}
                canManage={canManage}
                emptyText=""
              />
            ) : null}
          </div>
        </CardBody>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader className="items-start">
        <div>
          <CardTitle>Сетевой граф связей задачи #{taskNumber}</CardTitle>
          <p className="mt-0.5 text-13 text-ink-muted">
            Определяет динамический расчёт сроков по алгоритму Critical Path Method (CPM)
          </p>
        </div>
        {canManage ? (
          <AddDependencyDialog
            taskId={taskId}
            candidates={projectTasks}
            linkedTaskIds={linkedTaskIds}
          />
        ) : null}
      </CardHeader>

      <CardBody>
        <div className="grid gap-3 lg:grid-cols-2">
          <DependencyColumn
            icon={<ArrowLeftToLine className="size-3.5" />}
            title="Входящие связи (предшественники)"
            items={predecessors}
            relatedIdOf={(item) => item.predecessorTaskId}
            relatedTitleOf={(item) => item.predecessorTitle}
            projectTasks={projectTasks}
            onDelete={(id) => deleteDependency.mutate(id)}
            deletingId={deletingId}
            canManage={canManage}
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
            deletingId={deletingId}
            canManage={canManage}
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
  canManage,
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
  canManage: boolean;
  emptyText: string;
}) {
  const criticalCount = items.filter((item) => item.isCritical).length;
  const byId = new Map(projectTasks.map((task) => [task.id, task]));

  return (
    <section className="rounded-control border border-line">
      <header className="flex items-start justify-between gap-3 border-b border-line px-3 py-2.5">
        <h4 className="flex items-start gap-1.5 text-2xs font-semibold tracking-wider text-ink-muted uppercase">
          <span className="mt-px text-ink-faint">{icon}</span>
          {title}
        </h4>
        <span
          className={cn(
            "shrink-0 text-right text-2xs font-semibold",
            criticalCount > 0 ? "text-danger" : "text-ink-muted",
          )}
        >
          {items.length} {plural(items.length, ["связь", "связи", "связей"])}
          {criticalCount > 0 ? " · критичные" : ""}
        </span>
      </header>

      {items.length === 0 ? (
        <p className="px-3 py-6 text-center text-13 text-ink-muted">{emptyText}</p>
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
                    <span className="text-13 leading-snug font-medium text-ink group-hover:text-brand">
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
                    {canManage ? (
                      <button
                        type="button"
                        onClick={() => onDelete(item.id)}
                        disabled={deletingId === item.id}
                        aria-label={`Удалить связь с задачей «${title}»`}
                        className="rounded-control p-1 text-ink-faint transition-colors hover:text-danger focus-visible:focus-ring disabled:opacity-50"
                      >
                        <Trash2 className="size-3.5" />
                      </button>
                    ) : null}
                  </span>
                </div>

                <p className="mt-2 flex flex-wrap items-center gap-x-3 gap-y-1">
                  <span className="rounded-control bg-surface-muted px-2 py-1 font-mono text-2xs text-ink-muted">
                    {DEPENDENCY_LABEL[item.type]}
                  </span>
                  <span className="font-mono text-2xs text-ink-faint">
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
