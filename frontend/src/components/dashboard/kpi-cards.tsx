import { CircleDot, Clock3 } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Card, CardBody, CardFooter, CardLabel } from "@/components/ui/card";
import { ProgressRing } from "@/components/ui/progress";
import { formatDateLong, plural } from "@/lib/format";
import type { DashboardMetrics } from "@/lib/dashboard-metrics";
import type { Project } from "@/types";

/**
 * Три верхние карточки обзора. Все значения — производные от задач, вех и
 * карточки проекта (см. `lib/dashboard-metrics.ts`); показателей, у которых нет
 * источника в API, здесь нет.
 */
export function KpiCards({
  metrics,
  project,
}: {
  metrics: DashboardMetrics;
  project: Project;
}) {
  const { breakdown } = metrics;
  const overdue = metrics.daysToDeadline < 0;

  const segments = [
    { key: "done", value: breakdown.done, className: "bg-success" },
    { key: "in_progress", value: breakdown.in_progress, className: "bg-warning" },
    { key: "planned", value: breakdown.planned, className: "bg-accent-soft" },
    { key: "blocked", value: breakdown.blocked, className: "bg-ink-faint" },
    { key: "overdue", value: breakdown.overdue, className: "bg-danger" },
  ].filter((segment) => segment.value > 0);

  return (
    <div className="grid gap-3 lg:grid-cols-3">
      <Card className="flex flex-col">
        <CardBody className="flex flex-1 items-start justify-between gap-4">
          <div>
            <CardLabel>Общий прогресс</CardLabel>
            <p className="mt-3 text-3xl font-bold tracking-tight text-ink">
              {Math.round(metrics.progressPercent)}%
            </p>
            <p className="mt-2 text-[13px] text-ink-muted">
              {metrics.milestonesTotal > 0
                ? `${metrics.milestonesClosed} из ${metrics.milestonesTotal} вех закрыто`
                : "Контрольных точек пока нет"}
            </p>
          </div>
          <ProgressRing value={metrics.progressPercent} />
        </CardBody>
        <CardFooter>
          <span className="text-ink-muted">
            Срок релиза: {formatDateLong(project.deadline)}
          </span>
          <span className={overdue ? "font-medium text-danger" : "font-medium text-ink"}>
            {overdue
              ? `просрочен на ${Math.abs(metrics.daysToDeadline)} ${plural(metrics.daysToDeadline, ["день", "дня", "дней"])}`
              : `${metrics.daysToDeadline} ${plural(metrics.daysToDeadline, ["день", "дня", "дней"])} до сдачи`}
          </span>
        </CardFooter>
      </Card>

      <Card className="flex flex-col">
        <CardBody className="flex-1">
          <div className="flex items-start justify-between gap-4">
            <div>
              <CardLabel>Отставание от плана</CardLabel>
              <Badge
                tone={metrics.laggingTasks.length > 0 ? "warning" : "success"}
                dot
                className="mt-3"
              >
                {metrics.laggingTasks.length > 0 ? "В зоне риска" : "В графике"}
              </Badge>
            </div>
            <span className="flex size-9 items-center justify-center rounded-control bg-warning-tint text-warning">
              <Clock3 className="size-5" />
            </span>
          </div>
          <p className="mt-4 flex items-baseline gap-3">
            <span
              className={
                metrics.worstLagDays > 0
                  ? "text-2xl font-bold text-danger"
                  : "text-2xl font-bold text-success"
              }
            >
              {metrics.worstLagDays > 0 ? `−${metrics.worstLagDays}` : "0"}{" "}
              {plural(metrics.worstLagDays, ["день", "дня", "дней"])}
            </span>
            <span className="text-[13px] text-ink-muted">максимальное отставание</span>
          </p>
        </CardBody>
        <CardFooter>
          <span className="text-ink-muted">Задач с отставанием:</span>
          <span className="font-mono text-ink">
            {metrics.laggingTasks.length} из {breakdown.total}
          </span>
        </CardFooter>
      </Card>

      <Card className="flex flex-col">
        <CardBody className="flex-1">
          <div className="flex items-center justify-between gap-4">
            <CardLabel>Статус задач</CardLabel>
            <span className="text-sm font-bold text-ink">{breakdown.total} всего</span>
          </div>
          {breakdown.total > 0 ? (
            <div
              className="mt-3 flex h-2 overflow-hidden rounded-control"
              role="img"
              aria-label={`Выполнено ${breakdown.done}, в работе ${breakdown.in_progress}, запланировано ${breakdown.planned}, просрочено ${breakdown.overdue}`}
            >
              {segments.map((segment) => (
                <span
                  key={segment.key}
                  className={segment.className}
                  style={{ width: `${(segment.value / breakdown.total) * 100}%` }}
                />
              ))}
            </div>
          ) : null}
          <dl className="mt-4 grid grid-cols-2 gap-x-4 gap-y-2 text-[13px]">
            <StatusLine color="text-success" value={breakdown.done} label="Выполнено" />
            <StatusLine color="text-warning" value={breakdown.in_progress} label="В работе" />
            <StatusLine color="text-accent-soft" value={breakdown.planned} label="Предстоит" />
            <StatusLine color="text-danger" value={breakdown.overdue} label="Просрочено" />
          </dl>
        </CardBody>
        <CardFooter>
          <span className="text-ink-muted">Заблокировано зависимостями:</span>
          <span className="font-mono text-ink">{breakdown.blocked}</span>
        </CardFooter>
      </Card>
    </div>
  );
}

function StatusLine({
  color,
  value,
  label,
}: {
  color: string;
  value: number;
  label: string;
}) {
  return (
    <div className="flex items-baseline gap-1.5">
      <CircleDot className={`size-2.5 shrink-0 translate-y-0.5 ${color}`} />
      <dt className="font-semibold text-ink">{value}</dt>
      <dd className="text-ink-muted">{label}</dd>
    </div>
  );
}
