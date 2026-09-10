"use client";

import { PlugZap } from "lucide-react";

import {
  EmptyProjects,
  PageError,
  PageLoading,
} from "@/components/app/page-state";
import { AttentionTasks } from "@/components/dashboard/attention-tasks";
import { KpiCards } from "@/components/dashboard/kpi-cards";
import { MilestonesStrip } from "@/components/dashboard/milestones-strip";
import { Badge } from "@/components/ui/badge";
import { Card, CardBody } from "@/components/ui/card";
import { useCurrentProject } from "@/data/current-project";
import {
  useMilestones,
  useProject,
  useTasks,
} from "@/data/queries";
import { buildDashboardMetrics } from "@/lib/dashboard-metrics";

/**
 * Обзор и аналитика.
 *
 * Эндпоинт `/projects/{id}/dashboard` пока возвращает 501, поэтому показатели
 * считаются на клиенте из задач, вех и карточки проекта — все три источника
 * реализованы. Блоки, у которых источника нет вовсе (риски, скорость команды,
 * загрузка в часах), не заполняются выдуманными числами, а честно помечены
 * как ожидающие бэкенд.
 */
export default function OverviewPage() {
  const { projectId, isEmpty, error } = useCurrentProject();
  const project = useProject(projectId);
  const tasks = useTasks(projectId);
  const milestones = useMilestones(projectId);

  if (error) return <PageError error={error} />;
  if (isEmpty) return <EmptyProjects />;
  if (project.error) return <PageError error={project.error} />;
  if (tasks.error) return <PageError error={tasks.error} />;
  if (!project.data || !tasks.data) return <PageLoading />;

  const metrics = buildDashboardMetrics({
    project: project.data,
    tasks: tasks.data,
    milestones: milestones.data ?? [],
    today: new Date(),
  });

  return (
    <div className="mx-auto max-w-[1200px] space-y-4">
      <div className="min-w-0">
        <p className="flex flex-wrap items-center gap-2">
          <span className="font-mono text-xs tracking-wide text-ink-faint uppercase">
            Проект #{project.data.code}
          </span>
          {project.data.phase ? (
            <Badge tone="brand" size="sm" className="font-semibold tracking-wide uppercase">
              {project.data.phase}
            </Badge>
          ) : null}
        </p>
        <h1 className="mt-1.5 max-w-lg text-2xl leading-tight font-bold tracking-tight text-ink">
          Мониторинг здоровья и контрольных сроков
        </h1>
      </div>

      <KpiCards metrics={metrics} project={project.data} />

      {milestones.data && milestones.data.length > 0 ? (
        <MilestonesStrip milestones={milestones.data} />
      ) : null}

      <AttentionTasks tasks={metrics.laggingTasks} />

      <Card>
        <CardBody className="flex items-start gap-3 py-4">
          <PlugZap className="mt-0.5 size-4 shrink-0 text-warning" />
          <p className="text-[13px] text-ink-muted">
            <span className="font-semibold text-ink">
              Риски, скорость команды и загрузка по часам появятся здесь
            </span>{" "}
            после реализации <code className="font-mono">/projects/{"{id}"}/dashboard</code>{" "}
            и <code className="font-mono">/workload</code> — сейчас эти эндпоинты
            возвращают 501. Вёрстка блоков готова и подключится без изменений экрана.
          </p>
        </CardBody>
      </Card>
    </div>
  );
}
