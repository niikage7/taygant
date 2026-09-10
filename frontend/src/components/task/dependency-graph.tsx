import { ArrowLeftToLine, ArrowRightFromLine, Waypoints } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardBody, CardHeader, CardTitle } from "@/components/ui/card";
import { formatDate, plural } from "@/lib/format";
import { cn } from "@/lib/utils";
import type { DependencyType, TaskDependency } from "@/types";

const DEPENDENCY_LABEL: Record<DependencyType, string> = {
  FS: "FS (Окончание-к-Началу)",
  SS: "SS (Начало-к-Началу)",
  FF: "FF (Окончание-к-Окончанию)",
  SF: "SF (Начало-к-Окончанию)",
};

/** Сетевой граф связей задачи: предшественники и последователи. */
export function DependencyGraph({
  taskNumber,
  predecessors,
  successors,
}: {
  taskNumber: string;
  predecessors: TaskDependency[];
  successors: TaskDependency[];
}) {
  const criticalSuccessors = successors.filter((item) => item.isCritical).length;

  return (
    <Card>
      <CardHeader className="items-start">
        <div>
          <CardTitle>Сетевой граф связей задачи #{taskNumber}</CardTitle>
          <p className="mt-0.5 text-[13px] text-ink-muted">
            Определяет динамический расчёт сроков по алгоритму Critical Path Method (CPM)
          </p>
        </div>
      </CardHeader>
      <CardBody className="space-y-3">
        <Button variant="secondary" size="sm">
          <Waypoints />
          Добавить связь
        </Button>

        <div className="grid gap-3 lg:grid-cols-2">
          <DependencyColumn
            icon={<ArrowLeftToLine className="size-3.5" />}
            title="Входящие связи (предшественники)"
            counter={`${predecessors.length} ${plural(predecessors.length, ["связь", "связи", "связей"])}`}
            counterTone="neutral"
            items={predecessors}
            relatedTitle={(item) => item.predecessorTitle ?? "Задача"}
            footer={(item) => `Финиш: ${formatDate("2025-11-01")}${item.lagDays ? "" : ""}`}
          />
          <DependencyColumn
            icon={<ArrowRightFromLine className="size-3.5" />}
            title="Исходящие связи (последователи)"
            counter={
              criticalSuccessors > 0
                ? `${successors.length} связи · Критичные`
                : `${successors.length} ${plural(successors.length, ["связь", "связи", "связей"])}`
            }
            counterTone={criticalSuccessors > 0 ? "danger" : "neutral"}
            items={successors}
            relatedTitle={(item) => item.successorTitle ?? "Задача"}
            footer={(item) =>
              item.isCritical ? "След. старт: 19.11.2025" : "Финиш: 28.11.2025"
            }
            footerTone={(item) => (item.isCritical ? "text-danger" : "text-ink-faint")}
          />
        </div>
      </CardBody>
    </Card>
  );
}

function DependencyColumn({
  icon,
  title,
  counter,
  counterTone,
  items,
  relatedTitle,
  footer,
  footerTone,
}: {
  icon: React.ReactNode;
  title: string;
  counter: string;
  counterTone: "neutral" | "danger";
  items: TaskDependency[];
  relatedTitle: (item: TaskDependency) => string;
  footer: (item: TaskDependency) => string;
  footerTone?: (item: TaskDependency) => string;
}) {
  return (
    <section className="rounded-control border border-line">
      <header className="flex items-start justify-between gap-3 border-b border-line px-3 py-2.5">
        <h4 className="flex items-start gap-1.5 text-[11px] font-semibold tracking-wider text-ink-muted uppercase">
          <span className="mt-px text-ink-faint">{icon}</span>
          {title}
        </h4>
        <span
          className={cn(
            "shrink-0 text-right text-[11px] font-semibold",
            counterTone === "danger" ? "text-danger" : "text-ink-muted",
          )}
        >
          {counter}
        </span>
      </header>

      <ul className="space-y-2 p-3">
        {items.map((item) => (
          <li key={item.id} className="rounded-control bg-surface-subtle p-3">
            <div className="flex items-start justify-between gap-2">
              <p className="min-w-0 text-[13px] leading-snug font-medium text-ink">
                <span className="font-mono text-brand">
                  #{item.predecessorTitle ? item.predecessorTaskId.replace("t-", "") : item.successorTaskId.replace("t-", "")}
                </span>{" "}
                {relatedTitle(item)}
              </p>
              <Badge
                tone={item.isCritical ? "brand" : "success"}
                size="sm"
                className="shrink-0"
              >
                {item.isCritical ? "Ожидает" : "Завершена"}
              </Badge>
            </div>
            <p className="mt-2 flex flex-wrap items-center gap-x-3 gap-y-1">
              <span className="rounded-control bg-surface-muted px-2 py-1 font-mono text-[11px] text-ink-muted">
                {DEPENDENCY_LABEL[item.type]}
              </span>
              <span className="font-mono text-[11px] text-ink-faint">
                Лаг: {item.lagDays} дн.
              </span>
            </p>
            <p
              className={cn(
                "mt-2 font-mono text-[11px]",
                footerTone?.(item) ?? "text-ink-faint",
              )}
            >
              {footer(item)}
            </p>
          </li>
        ))}
      </ul>
    </section>
  );
}
