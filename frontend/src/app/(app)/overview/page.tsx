"use client";

import { Flag } from "lucide-react";

import { EmptyProjects, PageError, PageLoading } from "@/components/app/page-state";
import { AttentionTasks } from "@/components/dashboard/attention-tasks";
import { KpiCards } from "@/components/dashboard/kpi-cards";
import { MilestonesDialog } from "@/components/dashboard/milestones-dialog";
import { MilestonesStrip } from "@/components/dashboard/milestones-strip";
import { RiskCard } from "@/components/dashboard/risk-card";
import { WorkloadCard } from "@/components/dashboard/workload-card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { useCurrentProject } from "@/data/current-project";
import { useProjectAccess } from "@/data/project-access";
import { useDashboard, useProject, useWorkload } from "@/data/queries";

/**
 * Обзор и аналитика — данные из `GET /projects/{id}/dashboard` и
 * `GET /projects/{id}/workload`.
 *
 * Поле `teamWorkload` в ответе дашборда бэкенд отдаёт пустым — загрузка
 * команды берётся отдельным эндпоинтом `/workload` (по текущему спринту).
 */
export default function OverviewPage() {
  const { projectId, isEmpty, error } = useCurrentProject();
  const project = useProject(projectId);
  const dashboard = useDashboard(projectId);
  const access = useProjectAccess();
  const workload = useWorkload(projectId);

  if (error) return <PageError error={error} />;
  if (isEmpty) return <EmptyProjects />;
  if (project.error) return <PageError error={project.error} />;
  if (dashboard.error) return <PageError error={dashboard.error} />;
  if (workload.error) return <PageError error={workload.error} />;
  if (!project.data || !dashboard.data || !workload.data) return <PageLoading />;

  const data = dashboard.data;
  const workloadItems = workload.data;

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

      <div data-tour="kpi">
        <KpiCards data={data} project={project.data} />
      </div>

      <MilestonesStrip
        milestones={data.nearestMilestones}
        action={
          projectId && access.isFull ? (
            <MilestonesDialog
              projectId={projectId}
              trigger={
                <Button variant="secondary" size="sm" data-tour="milestones">
                  <Flag />
                  Контрольные точки
                </Button>
              }
            />
          ) : null
        }
      />

      <div className={workloadItems.length > 0 ? "grid grid-cols-1 gap-4 lg:grid-cols-[minmax(0,1.4fr)_minmax(0,1fr)]" : "space-y-4"}>
        <div className="space-y-4">
          {data.risks.map((risk) => (
            <RiskCard key={risk.title} risk={risk} />
          ))}
          <AttentionTasks tasks={data.attentionTasks} />
        </div>
        {workloadItems.length > 0 ? <WorkloadCard workload={workloadItems} /> : null}
      </div>
    </div>
  );
}
