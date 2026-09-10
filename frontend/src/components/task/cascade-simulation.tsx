/**
 * Каскадная диаграмма what-if: как сдвиг исходной задачи пробегает по цепочке
 * связей и во что превращается для дедлайна проекта.
 *
 * Расчёт делает бэкенд (`POST /tasks/{id}/simulate-shift`), применение —
 * `/apply-shift`. Сценарий не пересчитывается сам: это запрос по кнопке, иначе
 * пользователь видел бы, как цифры меняются без его действий.
 */
"use client";

import { ArrowDown, ArrowRight, CalendarX2, Info, Shield, Zap } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Card, CardBody, CardLabel } from "@/components/ui/card";
import { formatDate } from "@/lib/format";
import { cn } from "@/lib/utils";
import type { ShiftSimulation } from "@/types";

/**
 * Каскадная диаграмма what-if: как сдвиг исходной задачи пробегает по цепочке
 * связей и во что превращается для дедлайна проекта.
 */
export function CascadeSimulation({
  simulation,
  affectedNumbers,
  onCancel,
  onApply,
  isApplying,
}: {
  simulation: ShiftSimulation;
  /** Номера задач для текста уведомления — берутся из WBS, а не из id. */
  affectedNumbers: string[];
  onCancel: () => void;
  onApply: (compensateFromBuffer: boolean) => void;
  isApplying: boolean;
}) {
  const impact = simulation.projectDeadlineImpact;

  return (
    <Card>
      <CardBody className="space-y-3">
        <div className="flex flex-wrap items-baseline justify-between gap-2">
          <CardLabel>Каскадная диаграмма смещения связанных задач:</CardLabel>
          <span className="text-xs text-ink-faint">Алгоритм {simulation.algorithm}</span>
        </div>

        <div className="space-y-1 rounded-control border border-line p-3">
          {simulation.affectedTasks.map((task, index) => (
            <div key={task.taskId}>
              <div className="flex items-center gap-3 rounded-control border border-line bg-surface px-3 py-2.5">
                <span
                  className={cn(
                    "flex size-6 shrink-0 items-center justify-center rounded-control text-[11px] font-bold",
                    task.isCriticalPath
                      ? index === 0
                        ? "bg-brand text-white"
                        : "bg-danger text-white"
                      : "bg-surface-muted text-ink-muted",
                  )}
                >
                  {task.taskId.replace("t-", "")}
                </span>

                <div className="min-w-0 flex-1">
                  <p className="truncate text-[13px] font-semibold text-ink">{task.title}</p>
                  <p className="truncate text-xs text-ink-faint">
                    {task.isCriticalPath ? "Критический путь" : "Финальный этап"} · Исходно:{" "}
                    {formatDate(task.originalStartDate)} — {formatDate(task.originalEndDate)}
                  </p>
                </div>

                <span className="flex shrink-0 items-center gap-2">
                  <span className="font-mono text-[11px] text-ink-faint line-through">
                    {formatDate(task.originalEndDate)}
                  </span>
                  <ArrowRight className="size-3.5 text-ink-faint" />
                  <span className="rounded-control bg-danger-tint px-2 py-1 font-mono text-[11px] font-semibold text-danger">
                    {formatDate(task.newEndDate)}
                  </span>
                </span>
              </div>

              {index < simulation.affectedTasks.length - 1 ? (
                <p className="flex items-center gap-1.5 py-1.5 pl-3 text-xs text-danger">
                  <ArrowDown className="size-3.5 shrink-0" />
                  {index === 0
                    ? `Жёсткая связь ${task.dependencyType} (Finish-to-Start) без буфера`
                    : "Каскадный сдвиг финализации"}
                </p>
              ) : null}
            </div>
          ))}

          <div className="mt-2 flex flex-wrap items-center gap-3 rounded-control border border-line bg-surface px-3 py-2.5">
            <CalendarX2 className="size-5 shrink-0 text-danger" />
            <p className="min-w-0 flex-1 text-[13px] font-semibold text-ink">
              Итоговое влияние на дату сдачи проекта:
            </p>
            <span className="text-xs text-ink-faint">
              Было:
              <br />
              <span className="font-mono">{formatDate(impact.originalDeadline)}</span>
            </span>
            <ArrowRight className="size-4 shrink-0 text-ink-faint" />
            <span className="rounded-control bg-danger px-3 py-2 text-xs font-semibold text-white">
              Стало:
              <br />
              <span className="font-mono">{formatDate(impact.newDeadline)}</span>
            </span>
          </div>
        </div>

        <p className="flex items-start gap-1.5 text-xs text-ink-faint">
          <Info className="mt-px size-3.5 shrink-0" />
          Применение автоматического сдвига отправит уведомления ответственным за задачи{" "}
          {affectedNumbers.map((number) => `#${number}`).join(" и ")}.
        </p>

        <div className="flex flex-wrap items-center gap-2">
          <Button variant="ghost" onClick={onCancel} disabled={isApplying}>
            Отмена
          </Button>
          <Button
            variant="soft"
            onClick={() => onApply(true)}
            disabled={isApplying || simulation.bufferAvailableDays <= 0}
            title={
              simulation.bufferAvailableDays <= 0
                ? "Резерва нет — компенсировать сдвиг нечем"
                : undefined
            }
          >
            <Shield />
            Компенсировать из резерва
          </Button>
          <Button variant="danger" onClick={() => onApply(false)} disabled={isApplying}>
            <Zap />
            {isApplying ? "Применяем…" : `Применить сдвиг цепочки (+${simulation.shiftDays}д)`}
          </Button>
        </div>
      </CardBody>
    </Card>
  );
}
