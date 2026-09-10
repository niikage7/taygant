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
}: {
  taskId: string;
  /** Нужны, чтобы подписать затронутые задачи их номерами WBS, а не id. */
  projectTasks: Task[];
}) {
  const [shiftDays, setShiftDays] = useState(5);
  const simulate = useSimulateShift(taskId);
  const applyShift = useApplyShift(taskId);

  const simulation = applyShift.data ?? simulate.data;
  const wbsOf = new Map(projectTasks.map((task) => [task.id, task.wbsNumber]));

  const affectedNumbers = (simulation?.affectedTasks ?? [])
    .filter((affected) => affected.taskId !== taskId)
    .map((affected) => wbsOf.get(affected.taskId) ?? affected.taskId);

  return (
    <div className="space-y-3">
      <Card>
        <CardBody className="flex flex-wrap items-end justify-between gap-4">
          <div>
            <CardLabel>Моделирование сдвига</CardLabel>
            <p className="mt-1.5 text-[13px] text-ink-muted">
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
                onChange={(event) => setShiftDays(Number(event.target.value))}
                className="w-28"
              />
            </div>
            <Button
              onClick={() => simulate.mutate({ shiftDays })}
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
          {toUserMessage(simulate.error, {}, "Не удалось рассчитать сценарий")}
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
          isApplying={applyShift.isPending}
          onCancel={() => {
            simulate.reset();
            applyShift.reset();
          }}
          onApply={(compensateFromBuffer) =>
            applyShift.mutate({
              shiftDays,
              compensateFromBuffer,
              notifyAssignees: true,
            })
          }
        />
      ) : null}
    </div>
  );
}
