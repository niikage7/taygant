"use client";

import { Bell, FileBarChart, GitBranch } from "lucide-react";

import { DemoDataNotice, EmptyProjects, PageError } from "@/components/app/page-state";
import { AttentionTasks } from "@/components/dashboard/attention-tasks";
import { KpiCards } from "@/components/dashboard/kpi-cards";
import { MilestonesStrip } from "@/components/dashboard/milestones-strip";
import { RiskCard } from "@/components/dashboard/risk-card";
import { WorkloadCard } from "@/components/dashboard/workload-card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { demoDashboard } from "@/data/demo";
import { useCurrentProjectId, useProject } from "@/data/queries";

/**
 * Дашборд остаётся на демо-данных: оба его источника —
 * `GET /projects/{id}/dashboard` и `GET /projects/{id}/workload` — на бэкенде
 * пока возвращают 501. Шапка при этом уже берёт реальный проект, чтобы название
 * и код не расходились с остальными экранами.
 */
export default function OverviewPage() {
  const { projectId, isEmpty, error } = useCurrentProjectId();
  const project = useProject(projectId);

  if (error) return <PageError error={error} />;
  if (isEmpty) return <EmptyProjects />;

  const data = demoDashboard;
  const code = project.data?.code ?? data.project.code;
  const phase = project.data?.phase ?? data.project.phase;

  return (
    <div className="mx-auto max-w-[1200px] space-y-4">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div className="min-w-0">
          <p className="flex flex-wrap items-center gap-2">
            <span className="font-mono text-xs tracking-wide text-ink-faint uppercase">
              Проект #{code}
            </span>
            {phase ? (
              <Badge tone="brand" size="sm" className="font-semibold tracking-wide uppercase">
                {phase}
              </Badge>
            ) : null}
          </p>
          <h1 className="mt-1.5 max-w-lg text-2xl leading-tight font-bold tracking-tight text-ink">
            Мониторинг здоровья и контрольных сроков
          </h1>
        </div>

        <div className="flex flex-wrap items-center justify-end gap-2">
          <Button variant="secondary">
            <GitBranch />
            Смоделировать сценарий
          </Button>
          <Button variant="secondary">
            <Bell />
            Уведомить исполнителей
          </Button>
          <Button>
            <FileBarChart />
            Отчёт для жюри ТПУ
          </Button>
        </div>
      </div>

      <DemoDataNotice>
        Показатели демонстрационные: эндпоинты{" "}
        <code className="font-mono">/dashboard</code> и{" "}
        <code className="font-mono">/workload</code> на бэкенде ещё возвращают 501.
        Экран переключится на реальные данные без правок вёрстки.
      </DemoDataNotice>

      <KpiCards data={data} />
      <MilestonesStrip milestones={data.nearestMilestones} />

      <div className="grid gap-4 lg:grid-cols-[1.4fr_1fr]">
        <div className="space-y-4">
          {data.risks.map((risk) => (
            <RiskCard key={risk.title} risk={risk} />
          ))}
          <AttentionTasks tasks={data.attentionTasks} />
        </div>
        <WorkloadCard workload={data.teamWorkload} />
      </div>
    </div>
  );
}
