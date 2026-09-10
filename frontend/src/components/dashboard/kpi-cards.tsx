import { differenceInCalendarDays, parseISO } from "date-fns";
import { CircleDot, Clock3 } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Card, CardBody, CardFooter, CardLabel } from "@/components/ui/card";
import { ProgressRing } from "@/components/ui/progress";
import { formatDateLong, plural } from "@/lib/format";
import type { Project, ProjectDashboard } from "@/types";

const ADHERENCE_META = {
  on_track: { label: "В графике", tone: "success" as const },
  at_risk: { label: "В зоне риска", tone: "warning" as const },
  delayed: { label: "Срок сорван", tone: "danger" as const },
};

/** Три верхние карточки обзора — показатели из `GET /projects/{id}/dashboard`. */
export function KpiCards({ data, project }: { data: ProjectDashboard; project: Project }) {
  const breakdown = data.taskStatusBreakdown;
  const schedule = data.scheduleAdherence;
  const adherence = ADHERENCE_META[schedule.status];
  const daysToDeadline = differenceInCalendarDays(parseISO(project.deadline), new Date());
  const overdue = daysToDeadline < 0;
  const bufferCritical = schedule.projectBufferDays <= 1;

  const segments = [
    { key: "done", value: breakdown.done, className: "bg-success" },
    { key: "inProgress", value: breakdown.inProgress, className: "bg-warning" },
    { key: "planned", value: breakdown.planned, className: "bg-accent-soft" },
    { key: "overdue", value: breakdown.overdue, className: "bg-danger" },
  ].filter((segment) => segment.value > 0);

  return (
    <div className="grid gap-3 lg:grid-cols-3">
      <Card className="flex flex-col">
        <CardBody className="flex flex-1 items-start justify-between gap-4">
          <div>
            <CardLabel>Общий прогресс</CardLabel>
            <p className="mt-3 text-3xl font-bold tracking-tight text-ink">
              {Math.round(data.overallProgressPercent)}%
            </p>
            <p className="mt-2 text-[13px] text-ink-muted">
              {data.milestonesTotal > 0
                ? `${data.milestonesClosed} из ${data.milestonesTotal} вех закрыто`
                : "Контрольных точек пока нет"}
            </p>
          </div>
          <ProgressRing value={data.overallProgressPercent} />
        </CardBody>
        <CardFooter>
          <span className="text-ink-muted">Срок релиза: {formatDateLong(project.deadline)}</span>
          <span className={overdue ? "font-medium text-danger" : "font-medium text-ink"}>
            {overdue
              ? `просрочен на ${Math.abs(daysToDeadline)} ${plural(daysToDeadline, ["день", "дня", "дней"])}`
              : `${daysToDeadline} ${plural(daysToDeadline, ["день", "дня", "дней"])} до сдачи`}
          </span>
        </CardFooter>
      </Card>

      <Card className="flex flex-col">
        <CardBody className="flex-1">
          <div className="flex items-start justify-between gap-4">
            <div>
              <CardLabel>Соблюдение сроков</CardLabel>
              <Badge tone={adherence.tone} dot className="mt-3">
                {adherence.label}
              </Badge>
            </div>
            <span className="flex size-9 items-center justify-center rounded-control bg-warning-tint text-warning">
              <Clock3 className="size-5" />
            </span>
          </div>
          <p className="mt-4 flex items-baseline gap-3">
            <span
              className={
                schedule.forecastDeviationDays < 0
                  ? "text-2xl font-bold text-danger"
                  : "text-2xl font-bold text-success"
              }
            >
              {schedule.forecastDeviationDays > 0 ? "+" : ""}
              {schedule.forecastDeviationDays}{" "}
              {plural(schedule.forecastDeviationDays, ["день", "дня", "дней"])}
            </span>
            <span className="text-[13px] text-ink-muted">прогноз отклонения</span>
          </p>
        </CardBody>
        <CardFooter>
          <span className="text-ink-muted">Буфер проекта:</span>
          <span className={bufferCritical ? "font-mono text-danger" : "font-mono text-ink"}>
            {schedule.projectBufferDays} дн.{bufferCritical ? " (критично)" : ""}
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
              aria-label={`Выполнено ${breakdown.done}, в работе ${breakdown.inProgress}, запланировано ${breakdown.planned}, просрочено ${breakdown.overdue}`}
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
            <StatusLine color="text-warning" value={breakdown.inProgress} label="В работе" />
            <StatusLine color="text-accent-soft" value={breakdown.planned} label="Предстоит" />
            <StatusLine color="text-danger" value={breakdown.overdue} label="Просрочено" />
          </dl>
        </CardBody>
        <CardFooter>
          <span className="text-ink-muted">Скорость команды:</span>
          <span className="font-mono text-brand">
            {data.teamVelocity.tasksPerWeek} задач / нед.
          </span>
        </CardFooter>
      </Card>
    </div>
  );
}

function StatusLine({ color, value, label }: { color: string; value: number; label: string }) {
  return (
    <div className="flex items-baseline gap-1.5">
      <CircleDot className={`size-2.5 shrink-0 translate-y-0.5 ${color}`} />
      <dt className="font-semibold text-ink">{value}</dt>
      <dd className="text-ink-muted">{label}</dd>
    </div>
  );
}
