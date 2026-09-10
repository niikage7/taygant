"use client";

import { PlugZap } from "lucide-react";

import { EmptyProjects, PageError, PageLoading } from "@/components/app/page-state";
import { AttentionTasks } from "@/components/dashboard/attention-tasks";
import { KpiCards } from "@/components/dashboard/kpi-cards";
import { MilestonesStrip } from "@/components/dashboard/milestones-strip";
import { RiskCard } from "@/components/dashboard/risk-card";
import { WorkloadCard } from "@/components/dashboard/workload-card";
import { Badge } from "@/components/ui/badge";
import { Card, CardBody } from "@/components/ui/card";
import { useCurrentProject } from "@/data/current-project";
import { useDashboard, useProject } from "@/data/queries";

/**
 * Обзор и аналитика — данные из `GET /projects/{id}/dashboard`.
 *
 * Два поля ответа бэкенд заполнить не может и честно об этом пишет:
 * `progressDeltaPercent` всегда 0 (нет исторических снимков прогресса) и
 * `teamWorkload` всегда пуст (в задачах нет оценки часов, а `/workload`
 * возвращает 501). Поэтому дельту не показываем вовсе, а карточку загрузки —
 * только когда список непустой.
 */
export default function OverviewPage() {
  const { projectId, isEmpty, error } = useCurrentProject();
  const project = useProject(projectId);
  const dashboard = useDashboard(projectId);

  if (error) return <PageError error={error} />;
  if (isEmpty) return <EmptyProjects />;
  if (project.error) return <PageError error={project.error} />;
  if (dashboard.error) return <PageError error={dashboard.error} />;
  if (!project.data || !dashboard.data) return <PageLoading />;

  const data = dashboard.data;
  const hasWorkload = data.teamWorkload.length > 0;

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

      <KpiCards data={data} project={project.data} />

      {data.nearestMilestones.length > 0 ? (
        <MilestonesStrip milestones={data.nearestMilestones} />
      ) : null}

      <div className={hasWorkload ? "grid gap-4 lg:grid-cols-[1.4fr_1fr]" : "space-y-4"}>
        <div className="space-y-4">
          {data.risks.map((risk) => (
            <RiskCard key={risk.title} risk={risk} />
          ))}
          <AttentionTasks tasks={data.attentionTasks} />
        </div>
        {hasWorkload ? <WorkloadCard workload={data.teamWorkload} /> : null}
      </div>

      {!hasWorkload ? (
        <Card>
          <CardBody className="flex items-start gap-3 py-4">
            <PlugZap className="mt-0.5 size-4 shrink-0 text-warning" />
            <p className="text-[13px] text-ink-muted">
              <span className="font-semibold text-ink">
                Загрузка команды по часам появится здесь
              </span>{" "}
              после реализации <code className="font-mono">/projects/{"{id}"}/workload</code>:
              сейчас эндпоинт возвращает 501, а в задачах нет оценки трудозатрат, поэтому бэкенд
              отдаёт пустой список. Вёрстка карточки готова.
            </p>
          </CardBody>
        </Card>
      ) : null}
    </div>
  );
}
