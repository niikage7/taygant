import { CircleCheck, CircleDashed, Flag, Hourglass } from "lucide-react";
import type { ComponentType, ReactNode } from "react";

import { Card, CardBody, CardHeader, CardTitle } from "@/components/ui/card";
import { formatDayMonth } from "@/lib/format";
import { cn } from "@/lib/utils";
import type { Milestone, MilestoneStatus } from "@/types";

const STATUS_STYLE: Record<
  MilestoneStatus,
  { icon: ComponentType<{ className?: string }>; tile: string; code: string; icon_: string }
> = {
  done: { icon: CircleCheck, tile: "bg-surface-muted", code: "text-ink-faint", icon_: "text-success" },
  current: { icon: Hourglass, tile: "bg-brand-soft", code: "text-brand", icon_: "text-warning" },
  planned: { icon: CircleDashed, tile: "bg-surface-subtle", code: "text-ink-faint", icon_: "text-ink-faint" },
  final: { icon: Flag, tile: "bg-surface-subtle", code: "text-ink-faint", icon_: "text-brand" },
};

/**
 * Полоса ближайших контрольных точек.
 *
 * Показывает `nearestMilestones` из дашборда — он отфильтрован от завершённых и
 * обрезан до трёх. Полный список и управление живут в диалоге, который
 * открывается из `action`.
 */
export function MilestonesStrip({
  milestones,
  action,
}: {
  milestones: Milestone[];
  /** Кнопка управления вехами в заголовке карточки. */
  action?: ReactNode;
}) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>
          <Flag className="size-4 text-brand" />
          Ближайшие контрольные точки (Milestones)
        </CardTitle>
        {action}
      </CardHeader>
      <CardBody>
        {milestones.length === 0 ? (
          <p className="py-6 text-center text-[13px] text-ink-muted">
            Предстоящих контрольных точек нет.
          </p>
        ) : (
        <ul className="grid gap-2 sm:grid-cols-2 xl:grid-cols-4">
          {milestones.map((milestone) => {
            const style = STATUS_STYLE[milestone.status];
            const Icon = style.icon;
            return (
              <li
                key={milestone.id}
                className={cn("rounded-control p-3", style.tile)}
              >
                <div className="flex items-start justify-between gap-2">
                  <span className={cn("font-mono text-xs", style.code)}>
                    {milestone.code}
                    {milestone.status === "current" ? " (Текущая)" : ""}
                    {milestone.status === "final" ? " (Финал)" : ""}
                  </span>
                  <Icon className={cn("size-4 shrink-0", style.icon_)} />
                </div>
                <p className="mt-2 text-[13px] font-medium text-ink">{milestone.name}</p>
                <p className="mt-1 text-xs">
                  {milestone.actualDate ? (
                    <span className="text-success">
                      Сдано {formatDayMonth(milestone.actualDate)}
                    </span>
                  ) : milestone.riskDays ? (
                    <span className="text-warning">
                      Дедлайн: {formatDayMonth(milestone.plannedDate)} (Риск +
                      {milestone.riskDays}d)
                    </span>
                  ) : (
                    <span className="text-ink-faint">
                      {milestone.status === "final" ? "Финал" : "План"}:{" "}
                      {formatDayMonth(milestone.plannedDate)}
                    </span>
                  )}
                </p>
              </li>
            );
          })}
        </ul>
        )}
      </CardBody>
    </Card>
  );
}
