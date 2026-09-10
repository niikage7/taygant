"use client";

import {
  Check,
  ChevronRight,
  History,
  Maximize2,
  MessageSquare,
  Share2,
  PlugZap,
  Waypoints,
  X,
} from "lucide-react";
import Link from "next/link";
import { use } from "react";

import { PageError, PageLoading } from "@/components/app/page-state";
import { DependencyGraph } from "@/components/task/dependency-graph";
import { TaskParams } from "@/components/task/task-params";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardBody } from "@/components/ui/card";
import { useCurrentProjectId, useProject, useTaskDetail } from "@/data/queries";
import { TASK_STATUS_META } from "@/lib/task-status";

export default function TaskDetailPage({ params }: PageProps<"/tasks/[taskId]">) {
  const { taskId } = use(params);
  const { projectId } = useCurrentProjectId();
  const project = useProject(projectId);
  const task = useTaskDetail(taskId);

  if (task.error) return <PageError error={task.error} />;
  if (!task.data) return <PageLoading label="Загружаем карточку задачи…" />;

  const detail = task.data;
  const status = TASK_STATUS_META[detail.status];

  return (
    <div className="mx-auto max-w-[1400px] space-y-4">
      <Card>
        <CardBody className="flex flex-wrap items-center justify-between gap-3 py-3">
          <nav
            aria-label="Хлебные крошки"
            className="flex min-w-0 flex-wrap items-center gap-1.5 text-[13px] text-ink-muted"
          >
            <Link href="/overview" className="rounded-control hover:text-brand focus-visible:focus-ring">
              {project.data?.name ?? "Проект"}
            </Link>
            <ChevronRight className="size-3.5 shrink-0 text-ink-faint" />
            <span className="font-mono font-semibold text-brand">{detail.code}</span>
          </nav>

          <div className="flex items-center gap-2">
            {detail.externalLink?.synced ? (
              <Badge tone="success" dot size="sm">
                Синхронизировано ({detail.externalLink.provider} #
                {detail.externalLink.referenceId})
              </Badge>
            ) : null}
            <button
              type="button"
              aria-label="Развернуть на весь экран"
              className="rounded-control p-1 text-ink-faint hover:text-ink focus-visible:focus-ring"
            >
              <Maximize2 className="size-4" />
            </button>
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
              <Button variant="secondary">
                <Share2 />
                Поделиться
              </Button>
              <Button>
                <Check />
                Сохранить изменения
              </Button>
            </div>
          </div>

          <div className="flex flex-wrap items-center gap-2 border-t border-line pt-3">
            <Badge tone="brand" className="px-3 py-2 font-medium">
              <Waypoints className="size-3.5" />
              Связи и влияние (Что, если…)
            </Badge>
            <span className="flex items-center gap-2 rounded-control px-3 py-2 text-[13px] text-ink-muted">
              <History className="size-3.5" />
              История и журнал аудита
            </span>
            <span className="flex items-center gap-2 rounded-control px-3 py-2 text-[13px] text-ink-muted">
              <MessageSquare className="size-3.5" />
              Комментарии команды
              <span className="font-mono text-xs text-ink-faint">
                {detail.commentsCount}
              </span>
            </span>
          </div>
        </CardBody>
      </Card>

      <div className="grid gap-4 lg:grid-cols-[minmax(0,360px)_minmax(0,1fr)]">
        {project.data ? (
          <TaskParams
            task={detail}
            project={project.data}
            plannedProgressPercent={Math.max(detail.progressPercent - 5, 0)}
          />
        ) : (
          <PageLoading label="Загружаем проект…" />
        )}

        <div className="space-y-4">
          <DependencyGraph
            taskNumber={detail.wbsNumber ?? ""}
            predecessors={detail.predecessors}
            successors={detail.successors}
          />

          <Card>
            <CardBody className="flex items-start gap-3 py-4">
              <PlugZap className="mt-0.5 size-4 shrink-0 text-warning" />
              <p className="text-[13px] text-ink-muted">
                <span className="font-semibold text-ink">
                  Симулятор каскадного сдвига сроков появится здесь
                </span>{" "}
                после реализации{" "}
                <code className="font-mono">/tasks/{"{id}"}/simulate-shift</code> и{" "}
                <code className="font-mono">/apply-shift</code> — сейчас эти эндпоинты
                возвращают 501. Вёрстка каскадной диаграммы готова
                (<code className="font-mono">components/task/cascade-simulation.tsx</code>)
                и подключится без изменений экрана.
              </p>
            </CardBody>
          </Card>
        </div>
      </div>
    </div>
  );
}
