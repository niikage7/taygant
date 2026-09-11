"use client";

import {
  Check,
  ChevronRight,
  History,
  MessageSquare,
  Trash2,
  Waypoints,
  X,
} from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useCallback, useState } from "react";
import { use } from "react";

import { PageError, PageLoading } from "@/components/app/page-state";
import { DependencyGraph } from "@/components/task/dependency-graph";
import { ShareTaskButton } from "@/components/task/share-task-button";
import { ShiftSimulator } from "@/components/task/shift-simulator";
import { TaskComments } from "@/components/task/task-comments";
import { TaskHistory } from "@/components/task/task-history";
import { TaskParams } from "@/components/task/task-params";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardBody } from "@/components/ui/card";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";
import { useCurrentProject } from "@/data/current-project";
import { useProjectAccess } from "@/data/project-access";
import { Alert } from "@/components/ui/alert";
import { useDeleteTask, useProject, useTaskDetail, useTasks, useUpdateTask } from "@/data/queries";
import { toUserMessage } from "@/lib/api-error-message";
import type { TaskUpdateRequest } from "@/types";
import { TASK_STATUS_META } from "@/lib/task-status";
import { cn } from "@/lib/utils";

export default function TaskDetailPage({ params }: PageProps<"/tasks/[taskId]">) {
  const { taskId } = use(params);
  const { projectId } = useCurrentProject();
  const project = useProject(projectId);
  const task = useTaskDetail(taskId);
  const projectTasks = useTasks(projectId);
  const updateTask = useUpdateTask(taskId);
  const access = useProjectAccess();
  const deleteTask = useDeleteTask(taskId);
  const router = useRouter();

  const [tab, setTab] = useState<"links" | "history" | "comments">("links");
  const [draft, setDraft] = useState<TaskUpdateRequest | null>(null);
  // Колбэк стабилен, иначе эффект в TaskParams пересчитывался бы на каждый рендер.
  const handleDraftChange = useCallback((next: TaskUpdateRequest | null) => {
    setDraft(next);
  }, []);

  if (task.error) return <PageError error={task.error} />;
  if (!task.data) return <PageLoading label="Загружаем карточку задачи…" />;

  const detail = task.data;
  const status = TASK_STATUS_META[detail.status];
  const canEdit = access.canEditTask(detail);

  return (
    <div className="mx-auto max-w-[1400px] space-y-4">
      <Card>
        <CardBody className="flex flex-wrap items-center justify-between gap-3 py-3">
          <nav
            aria-label="Хлебные крошки"
            className="flex min-w-0 flex-wrap items-center gap-1.5 text-13 text-ink-muted"
          >
            <Link
              href="/overview"
              className="rounded-control hover:text-brand focus-visible:focus-ring"
            >
              {project.data?.name ?? "Проект"}
            </Link>
            <ChevronRight className="size-3.5 shrink-0 text-ink-faint" />
            <span className="font-mono font-semibold text-brand">{detail.code}</span>
          </nav>

          <div className="flex items-center gap-2">
            {detail.externalLink?.synced ? (
              <Badge tone="success" dot size="sm">
                Синхронизировано ({detail.externalLink.provider} #{detail.externalLink.referenceId})
              </Badge>
            ) : null}
            <Link
              href="/gantt"
              aria-label="Закрыть редактор"
              className="rounded-control p-1 text-ink-faint hover:text-ink focus-visible:focus-ring"
            >
              <X className="size-4" />
            </Link>
          </div>
        </CardBody>
      </Card>

      <Card>
        <CardBody className="space-y-4">
          <div className="flex flex-wrap items-start justify-between gap-4">
            <div className="min-w-0">
              <div className="flex flex-wrap items-center gap-2">
                {detail.wbsNumber ? (
                  <Badge tone="neutral" size="sm" className="font-mono">
                    #{detail.wbsNumber}
                  </Badge>
                ) : null}
                <Badge tone={status.tone} size="sm" dot>
                  {status.label}
                </Badge>
                {detail.isCriticalPath ? (
                  <Badge tone="danger" size="sm" className="font-semibold tracking-wide uppercase">
                    <Waypoints className="size-3" />
                    Критический путь (CPM)
                  </Badge>
                ) : null}
                <Badge tone="neutral" size="sm">
                  Вес задачи: {detail.weightPercent}%
                </Badge>
              </div>
              <h1 className="mt-3 max-w-2xl text-2xl leading-tight font-bold tracking-tight text-ink">
                {detail.title}
              </h1>
            </div>

            <div className="flex items-center gap-2">
              <ShareTaskButton />
              {canEdit ? (
                <ConfirmDialog
                  title="Удалить задачу?"
                  description={
                    <>
                      Задача «{detail.title}» будет удалена вместе со связями,
                      комментариями и критериями приёмки. Действие необратимо.
                    </>
                  }
                  confirmLabel="Удалить задачу"
                  pendingLabel="Удаляем…"
                  isPending={deleteTask.isPending}
                  error={deleteTask.error}
                  onConfirm={() =>
                    deleteTask.mutate(undefined, {
                      onSuccess: () => router.push("/gantt"),
                    })
                  }
                  trigger={
                    <Button variant="secondary" aria-label="Удалить задачу">
                      <Trash2 />
                    </Button>
                  }
                />
              ) : null}
              {canEdit ? (
                <Button
                  onClick={() => draft && updateTask.mutate(draft)}
                  disabled={!draft || updateTask.isPending}
                >
                  <Check />
                  {updateTask.isPending
                    ? "Сохраняем…"
                    : draft
                      ? "Сохранить изменения"
                      : "Изменений нет"}
                </Button>
              ) : null}
            </div>
          </div>

          <div
            role="tablist"
            aria-label="Разделы задачи"
            className="flex flex-wrap items-center gap-2 border-t border-line pt-3"
          >
            <TabButton
              active={tab === "links"}
              onClick={() => setTab("links")}
              icon={<Waypoints className="size-3.5" />}
              label="Связи и влияние (Что, если…)"
            />
            <TabButton
              active={tab === "history"}
              onClick={() => setTab("history")}
              icon={<History className="size-3.5" />}
              label="История и журнал аудита"
            />
            <TabButton
              active={tab === "comments"}
              onClick={() => setTab("comments")}
              icon={<MessageSquare className="size-3.5" />}
              label="Комментарии команды"
              counter={detail.commentsCount}
            />
          </div>
        </CardBody>
      </Card>

      <div className="grid gap-4 lg:grid-cols-[minmax(0,360px)_minmax(0,1fr)]">
        {project.data ? (
          <TaskParams
            task={detail}
            project={project.data}
            plannedProgressPercent={Math.max(detail.progressPercent - 5, 0)}
            onDraftChange={handleDraftChange}
          />
        ) : (
          <PageLoading label="Загружаем проект…" />
        )}

        <div className="space-y-4">
          {updateTask.isError ? (
            <Alert tone="danger">
              {toUserMessage(
                updateTask.error,
                { 403: "Недостаточно прав для редактирования задачи" },
                "Не удалось сохранить изменения",
              )}
            </Alert>
          ) : null}

          {tab === "history" ? <TaskHistory taskId={taskId} /> : null}
          {tab === "comments" ? <TaskComments taskId={taskId} /> : null}

          {tab === "links" ? (
            <>

              <DependencyGraph
                taskId={taskId}
                taskNumber={detail.wbsNumber ?? ""}
            isMilestone={detail.isMilestone}
                predecessors={detail.predecessors}
                successors={detail.successors}
                projectTasks={projectTasks.data ?? []}
                canManage={access.isFull}
              />

              <ShiftSimulator
                taskId={taskId}
                projectTasks={projectTasks.data ?? []}
                canApply={access.isFull}
              />
            </>
          ) : null}
        </div>
      </div>
    </div>
  );
}

function TabButton({
  active,
  onClick,
  icon,
  label,
  counter,
}: {
  active: boolean;
  onClick: () => void;
  icon: React.ReactNode;
  label: string;
  counter?: number;
}) {
  return (
    <button
      type="button"
      role="tab"
      aria-selected={active}
      onClick={onClick}
      className={cn(
        "flex items-center gap-2 rounded-control px-3 py-2 text-13 transition-colors focus-visible:focus-ring",
        active
          ? "bg-brand-tint font-medium text-brand"
          : "text-ink-muted hover:bg-surface-muted hover:text-ink",
      )}
    >
      {icon}
      {label}
      {counter !== undefined ? (
        <span className="font-mono text-xs text-ink-faint">{counter}</span>
      ) : null}
    </button>
  );
}
