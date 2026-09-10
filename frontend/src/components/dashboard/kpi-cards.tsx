import { CircleDot, Clock3 } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Card, CardBody, CardFooter, CardLabel } from "@/components/ui/card";
import { ProgressRing } from "@/components/ui/progress";
import { formatDateLong, formatSigned, plural } from "@/lib/format";
import type { ProjectDashboard } from "@/types";

const ADHERENCE_META = {
  on_track: { label: "В графике", tone: "success" as const },
  at_risk: { label: "В зоне риска", tone: "warning" as const },
  delayed: { label: "Срок сорван", tone: "danger" as const },
};

/** Три верхние карточки дашборда: прогресс, сроки, разбивка задач. */
export function KpiCards({ data }: { data: ProjectDashboard }) {
  const { scheduleAdherence: schedule, taskStatusBreakdown: tasks } = data;
  const adherence = ADHERENCE_META[schedule.status];
  const bufferCritical = schedule.projectBufferDays <= 1;

  const segments = [
    { key: "done", value: tasks.done, className: "bg-success" },
    { key: "inProgress", value: tasks.inProgress, className: "bg-warning" },
    { key: "planned", value: tasks.planned, className: "bg-accent-soft" },
    { key: "overdue", value: tasks.overdue, className: "bg-danger" },
  ];

  return (
    <div className="grid gap-3 lg:grid-cols-3">
      <Card className="flex flex-col">
        <CardBody className="flex flex-1 items-start justify-between gap-4">
          <div>
            <CardLabel>Общий прогресс</CardLabel>
            <p className="mt-3 flex items-baseline gap-2">
              <span className="text-3xl font-bold tracking-tight text-ink">
                {data.overallProgressPercent}%
              </span>
              <span className="text-sm font-medium text-success">
                {formatSigned(data.progressDeltaPercent, 1)}%
              </span>
            </p>
            <p className="mt-2 text-[13px] text-ink-muted">
              {data.milestonesClosed} из {data.milestonesTotal} вех закрыто
            </p>
          </div>
          <ProgressRing value={data.overallProgressPercent} />
        </CardBody>
        <CardFooter>
          <span className="text-ink-muted">
            Срок релиза: {formatDateLong(data.project.deadline)}
          </span>
          <span className="font-medium text-ink">9 дн. до сдачи</span>
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
            <span className="text-2xl font-bold text-danger">
              {formatSigned(schedule.forecastDeviationDays)}{" "}
              {plural(schedule.forecastDeviationDays, ["день", "дня", "дней"])}
            </span>
            <span className="text-[13px] text-ink-muted">прогноз отставания</span>
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
            <span className="text-sm font-bold text-ink">{tasks.total} всего</span>
          </div>
          <div
            className="mt-3 flex h-2 overflow-hidden rounded-control"
            role="img"
            aria-label={`Выполнено ${tasks.done}, в процессе ${tasks.inProgress}, предстоит ${tasks.planned}, просрочено ${tasks.overdue}`}
          >
            {segments.map((segment) => (
              <span
                key={segment.key}
                className={segment.className}
                style={{ width: `${(segment.value / tasks.total) * 100}%` }}
              />
            ))}
          </div>
          <dl className="mt-4 grid grid-cols-2 gap-x-4 gap-y-2 text-[13px]">
            <StatusLine color="text-success" value={tasks.done} label="Выполнено" />
            <StatusLine color="text-warning" value={tasks.inProgress} label="В процессе" />
            <StatusLine color="text-accent-soft" value={tasks.planned} label="Предстоит" />
            <StatusLine color="text-danger" value={tasks.overdue} label="Просрочена" />
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
