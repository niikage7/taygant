"use client";

import { Play } from "lucide-react";
import { useState } from "react";

import { CascadeSimulation } from "@/components/task/cascade-simulation";
import { Alert } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Card, CardBody, CardLabel } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useApplyShift, useSimulateShift } from "@/data/queries";
import { toUserMessage } from "@/lib/api-error-message";
import type { Task } from "@/types";

/**
 * What-if сдвиг сроков задачи: сначала расчёт без изменений, затем применение.
 *
 * Разделено намеренно — «посмотреть, что будет» и «сделать» разные по
 * последствиям действия, и применение доступно только после того, как
 * пользователь увидел каскад.
 */
export function ShiftSimulator({
  taskId,
  projectTasks,
  canApply,
}: {
  taskId: string;
  /** Нужны, чтобы подписать затронутые задачи их номерами WBS, а не id. */
  projectTasks: Task[];
  /** Применение сдвига цепочки — только полный доступ. */
  canApply: boolean;
}) {
  const [shiftDays, setShiftDays] = useState(5);
  const simulate = useSimulateShift(taskId);
  const applyShift = useApplyShift(taskId);

  const simulation = applyShift.data ?? simulate.data;
  const wbsOf = new Map(projectTasks.map((task) => [task.id, task.wbsNumber]));

  // Отфильтровываем задачи, которых нет в списке проекта: показать сырой UUID
  // хуже, чем не упомянуть задачу в тексте уведомления.
  const affectedNumbers = (simulation?.affectedTasks ?? [])
    .filter((affected) => affected.taskId !== taskId)
    .map((affected) => wbsOf.get(affected.taskId))
    .filter((number): number is string => Boolean(number));

  return (
    <div className="space-y-3">
      <Card>
        <CardBody className="flex flex-wrap items-end justify-between gap-4">
          <div>
            <CardLabel>Моделирование сдвига</CardLabel>
            <p className="mt-1.5 text-13 text-ink-muted">
              Рассчитайте, как перенос сроков задачи отразится на связанных задачах и дате сдачи
              проекта.
            </p>
          </div>

          <div className="flex items-end gap-2">
            <div className="space-y-1.5">
              <Label htmlFor="shift-days">Сдвиг, дней</Label>
              <Input
                id="shift-days"
                type="number"
                value={shiftDays}
                step={1}
                // Сдвиг — целое число дней: дробь сервер не разберёт и ответит 400.
                onChange={(event) => setShiftDays(Math.trunc(Number(event.target.value)) || 0)}
                className="w-28"
              />
            </div>
            <Button
              onClick={() => {
                // Новый расчёт заменяет итог прошлого применения: иначе карточка
                // показывала бы старый результат, а «Применить» сработало бы второй раз.
                applyShift.reset();
                simulate.mutate({ shiftDays });
              }}
              disabled={simulate.isPending || shiftDays === 0}
            >
              <Play />
              {simulate.isPending ? "Считаем…" : "Смоделировать"}
            </Button>
          </div>
        </CardBody>
      </Card>

      {simulate.isError ? (
        <Alert tone="danger">
          {toUserMessage(
            simulate.error,
            { 400: "Сдвиг — целое число дней" },
            "Не удалось рассчитать сценарий",
          )}
        </Alert>
      ) : null}

      {applyShift.isError ? (
        <Alert tone="danger">
          {toUserMessage(
            applyShift.error,
            { 403: "Недостаточно прав для переноса сроков" },
            "Не удалось применить сдвиг",
          )}
        </Alert>
      ) : null}

      {applyShift.isSuccess ? (
        <Alert tone="success">
          Сдвиг применён: сроки задач пересчитаны по алгоритму {applyShift.data.algorithm}.
        </Alert>
      ) : null}

      {simulation ? (
        <CascadeSimulation
          simulation={simulation}
          affectedNumbers={affectedNumbers}
          numberOf={(id) => wbsOf.get(id)}
          isApplying={applyShift.isPending}
          canApply={canApply}
          onCancel={() => {
            simulate.reset();
            applyShift.reset();
          }}
          onApply={(compensateFromBuffer) =>
            applyShift.mutate({
              // Применяем ровно показанный сценарий, а не число, которое успели
              // поменять в поле после расчёта.
              shiftDays: simulation.shiftDays,
              compensateFromBuffer,
              notifyAssignees: true,
            })
          }
        />
      ) : null}
    </div>
  );
}
