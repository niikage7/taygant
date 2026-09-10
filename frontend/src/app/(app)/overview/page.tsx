import { Bell, FileBarChart, GitBranch } from "lucide-react";
import type { Metadata } from "next";

import { AttentionTasks } from "@/components/dashboard/attention-tasks";
import { KpiCards } from "@/components/dashboard/kpi-cards";
import { MilestonesStrip } from "@/components/dashboard/milestones-strip";
import { RiskCard } from "@/components/dashboard/risk-card";
import { WorkloadCard } from "@/components/dashboard/workload-card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { demoDashboard } from "@/data/demo";

export const metadata: Metadata = {
  title: "Обзор и аналитика",
  description: "Мониторинг здоровья проекта, рисков и контрольных сроков",
};

export default function OverviewPage() {
  const data = demoDashboard;

  return (
    <div className="mx-auto max-w-[1200px] space-y-4">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div className="min-w-0">
          <p className="flex flex-wrap items-center gap-2">
            <span className="font-mono text-xs tracking-wide text-ink-faint uppercase">
              Проект #{data.project.code}
            </span>
            <Badge tone="brand" size="sm" className="font-semibold tracking-wide uppercase">
              {data.project.phase}
            </Badge>
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
