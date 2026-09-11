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
import { formatDate, formatSigned } from "@/lib/format";
import { cn } from "@/lib/utils";
import type { ShiftSimulation } from "@/types";

export function CascadeSimulation({
  simulation,
  affectedNumbers,
  numberOf,
  onCancel,
  onApply,
  isApplying,
  canApply,
}: {
  simulation: ShiftSimulation;
  /** Номера задач для текста уведомления — берутся из WBS, а не из id. */
  affectedNumbers: string[];
  /** WBS-номер задачи по её id. Идентификаторы — UUID, показывать их нельзя. */
  numberOf: (taskId: string) => string | undefined;
  onCancel: () => void;
  onApply: (compensateFromBuffer: boolean) => void;
  isApplying: boolean;
  /** Применение сдвига цепочки доступно только полному доступу. */
  canApply: boolean;
}) {
  const impact = simulation.projectDeadlineImpact;
  // Сдвиг задач не обязан двигать дедлайн: если у цепочки есть запас, проект
  // сдаётся в тот же срок. Такой исход нужно показывать спокойным, а не красным.
  const deadlineMoved = impact.deltaDays > 0;

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
                    "flex size-6 shrink-0 items-center justify-center rounded-control text-2xs font-bold",
                    task.isCriticalPath
                      ? index === 0
                        ? "bg-brand text-white"
                        : "bg-danger text-white"
                      : "bg-surface-muted text-ink-muted",
                  )}
                >
                  {numberOf(task.taskId) ?? "•"}
                </span>

                <div className="min-w-0 flex-1">
                  <p className="truncate text-13 font-semibold text-ink">{task.title}</p>
                  <p className="truncate text-xs text-ink-faint">
                    {task.isCriticalPath ? "Критический путь" : `Связь ${task.dependencyType}`}{" "}
                    · Исходно: {formatDate(task.originalStartDate)} —{" "}
                    {formatDate(task.originalEndDate)}
                  </p>
                </div>

                <span className="flex shrink-0 items-center gap-2">
                  <span className="font-mono text-2xs text-ink-faint line-through">
                    {formatDate(task.originalEndDate)}
                  </span>
                  <ArrowRight className="size-3.5 text-ink-faint" />
                  <span className="rounded-control bg-danger-tint px-2 py-1 font-mono text-2xs font-semibold text-danger">
                    {formatDate(task.newEndDate)}
                  </span>
                </span>
              </div>

              {index < simulation.affectedTasks.length - 1 ? (
                <p className="flex items-center gap-1.5 py-1.5 pl-3 text-xs text-ink-muted">
                  <ArrowDown className="size-3.5 shrink-0" />
                  Связь {simulation.affectedTasks[index + 1].dependencyType} —
                  сдвиг передаётся дальше по цепочке
                </p>
              ) : null}
            </div>
          ))}

          <div className="mt-2 flex flex-wrap items-center gap-3 rounded-control border border-line bg-surface px-3 py-2.5">
            <CalendarX2
              className={cn(
                "size-5 shrink-0",
                deadlineMoved ? "text-danger" : "text-success",
              )}
            />
            <div className="min-w-0 flex-1">
              <p className="text-13 font-semibold text-ink">
                Итоговое влияние на дату сдачи проекта:
              </p>
              {!deadlineMoved ? (
                <p className="mt-0.5 text-xs text-ink-muted">
                  Дедлайн не двигается — у цепочки хватает запаса, чтобы поглотить сдвиг.
                </p>
              ) : null}
            </div>
            <span className="text-xs text-ink-faint">
              Было:
              <br />
              <span className="font-mono">{formatDate(impact.originalDeadline)}</span>
            </span>
            <ArrowRight className="size-4 shrink-0 text-ink-faint" />
            <span
              className={cn(
                "rounded-control px-3 py-2 text-xs font-semibold",
                deadlineMoved
                  ? "bg-danger text-white"
                  : "bg-success-tint text-success",
              )}
            >
              {deadlineMoved ? `Стало (+${impact.deltaDays} дн.):` : "Без изменений:"}
              <br />
              <span className="font-mono">{formatDate(impact.newDeadline)}</span>
            </span>
          </div>
        </div>

        <p className="flex items-start gap-1.5 text-xs text-ink-faint">
          <Info className="mt-px size-3.5 shrink-0" />
          {affectedNumbers.length > 0
            ? `Применение сдвига отправит уведомления ответственным за задачи ${affectedNumbers
                .map((number) => `#${number}`)
                .join(", ")}.`
            : "Кроме самой задачи сдвиг ничего не затрагивает."}
        </p>

        <div className="flex flex-wrap items-center gap-2">
          <Button variant="ghost" onClick={onCancel} disabled={isApplying}>
            {canApply ? "Отмена" : "Сбросить"}
          </Button>
          {canApply ? (
            <>
              <Button
                variant="soft"
                onClick={() => onApply(true)}
                // Показан уже применённый сдвиг: повторное нажатие сдвинуло бы
                // цепочку ещё раз. Для нового применения нужен новый расчёт.
                disabled={isApplying || simulation.applied || simulation.bufferAvailableDays <= 0}
                title={
                  simulation.bufferAvailableDays <= 0
                    ? "Резерва нет — компенсировать сдвиг нечем"
                    : undefined
                }
              >
                <Shield />
                Компенсировать из резерва
              </Button>
              <Button
                variant="danger"
                onClick={() => onApply(false)}
                disabled={isApplying || simulation.applied}
              >
                <Zap />
                {isApplying
                  ? "Применяем…"
                  : `Применить сдвиг цепочки (${formatSigned(simulation.shiftDays)}д)`}
              </Button>
            </>
          ) : null}
        </div>
      </CardBody>
    </Card>
  );
}
