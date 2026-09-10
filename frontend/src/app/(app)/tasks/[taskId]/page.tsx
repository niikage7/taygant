import {
  ChevronRight,
  History,
  Maximize2,
  MessageSquare,
  Share2,
  Waypoints,
  X,
  Check,
} from "lucide-react";
import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";

import { CascadeSimulation } from "@/components/task/cascade-simulation";
import { DependencyGraph } from "@/components/task/dependency-graph";
import { TaskParams } from "@/components/task/task-params";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardBody } from "@/components/ui/card";
import { demoProject, demoSimulation, demoTaskDetail, demoTasks } from "@/data/demo";
import { TASK_STATUS_META } from "@/lib/task-status";

export const metadata: Metadata = {
  title: "Детали и зависимости",
  description: "Редактор задачи и симулятор каскадного сдвига сроков",
};

/** Демо-данные содержат подробности только для одной задачи — см. src/data/demo.ts. */
export function generateStaticParams() {
  return demoTasks.map((task) => ({ taskId: task.id }));
}

export default async function TaskDetailPage({
  params,
}: PageProps<"/tasks/[taskId]">) {
  const { taskId } = await params;
  if (!demoTasks.some((task) => task.id === taskId)) notFound();

  const task = demoTaskDetail;
  const status = TASK_STATUS_META[task.status];

  return (
    <div className="mx-auto max-w-[1400px] space-y-4">
      <Card>
        <CardBody className="flex flex-wrap items-center justify-between gap-3 py-3">
          <nav
            aria-label="Хлебные крошки"
            className="flex min-w-0 flex-wrap items-center gap-1.5 text-[13px] text-ink-muted"
          >
            <Link href="/overview" className="rounded-control hover:text-brand focus-visible:focus-ring">
              ИС Мониторинга ТПУ
            </Link>
            <ChevronRight className="size-3.5 shrink-0 text-ink-faint" />
            <span>Спринт 4 · Архитектура и Визуализация</span>
            <ChevronRight className="size-3.5 shrink-0 text-ink-faint" />
            <span className="font-mono font-semibold text-brand">{task.code}</span>
          </nav>

          <div className="flex items-center gap-2">
            {task.externalLink?.synced ? (
              <Badge tone="success" dot size="sm">
                Синхронизировано ({task.externalLink.provider} #{task.externalLink.referenceId})
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
                <Badge tone="neutral" size="sm" className="font-mono">
                  #{task.wbsNumber}
                </Badge>
                <Badge tone={status.tone} size="sm" dot>
                  {status.label}
                </Badge>
                {task.isCriticalPath ? (
                  <Badge tone="danger" size="sm" className="font-semibold tracking-wide uppercase">
                    <Waypoints className="size-3" />
                    Критический путь (CPM)
                  </Badge>
                ) : null}
                <Badge tone="neutral" size="sm">
                  Вес задачи: {task.weightPercent}%
                </Badge>
              </div>
              <h1 className="mt-3 max-w-2xl text-2xl leading-tight font-bold tracking-tight text-ink">
                {task.title}
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
              <span className="font-mono text-xs text-ink-faint">{task.commentsCount}</span>
            </span>
          </div>
        </CardBody>
      </Card>

      <div className="grid gap-4 lg:grid-cols-[minmax(0,360px)_minmax(0,1fr)]">
        <TaskParams task={task} project={demoProject} plannedProgressPercent={40} />

        <div className="space-y-4">
          <DependencyGraph
            taskNumber={task.wbsNumber ?? ""}
            predecessors={task.predecessors}
            successors={task.successors}
          />
          <CascadeSimulation simulation={demoSimulation} affectedNumbers={["5", "6"]} />
        </div>
      </div>
    </div>
  );
}
