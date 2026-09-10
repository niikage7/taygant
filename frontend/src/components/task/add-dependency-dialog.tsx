"use client";

import * as Dialog from "@radix-ui/react-dialog";
import { Waypoints, X } from "lucide-react";
import { useState } from "react";

import { Alert } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select } from "@/components/ui/select";
import { useCreateDependency } from "@/data/queries";
import { toUserMessage } from "@/lib/api-error-message";
import type { DependencyType, Task, TaskDependencyDirection } from "@/types";

const TYPE_OPTIONS: { value: DependencyType; label: string }[] = [
  { value: "FS", label: "FS — окончание к началу" },
  { value: "SS", label: "SS — начало к началу" },
  { value: "FF", label: "FF — окончание к окончанию" },
  { value: "SF", label: "SF — начало к окончанию" },
];

const DIRECTION_OPTIONS: { value: TaskDependencyDirection; label: string }[] = [
  { value: "predecessor", label: "Предшественник — идёт до текущей задачи" },
  { value: "successor", label: "Последователь — идёт после текущей задачи" },
];

/**
 * Создание связи между текущей задачей и другой задачей проекта.
 *
 * Из списка исключены сама задача и те, с кем связь уже есть: бэкенд отклонил
 * бы такие запросы, и лучше не показывать заведомо неверный выбор.
 */
export function AddDependencyDialog({
  taskId,
  candidates,
  linkedTaskIds,
}: {
  taskId: string;
  /** Все задачи проекта — из них выбирается вторая сторона связи. */
  candidates: Task[];
  linkedTaskIds: string[];
}) {
  const [open, setOpen] = useState(false);
  const [relatedTaskId, setRelatedTaskId] = useState("");
  const [direction, setDirection] = useState<TaskDependencyDirection>("predecessor");
  const [type, setType] = useState<DependencyType>("FS");
  const [lagDays, setLagDays] = useState(0);

  const createDependency = useCreateDependency(taskId);
  const linked = new Set([...linkedTaskIds, taskId]);
  const available = candidates.filter((task) => !linked.has(task.id));

  function submit() {
    if (!relatedTaskId) return;
    createDependency.mutate(
      { relatedTaskId, direction, type, lagDays },
      {
        onSuccess: () => {
          setOpen(false);
          setRelatedTaskId("");
          setLagDays(0);
        },
      },
    );
  }

  return (
    <Dialog.Root open={open} onOpenChange={setOpen}>
      <Dialog.Trigger asChild>
        <Button variant="secondary" size="sm" disabled={available.length === 0}>
          <Waypoints />
          Добавить связь
        </Button>
      </Dialog.Trigger>

      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-0 z-50 bg-ink/30" />
        <Dialog.Content className="fixed top-1/2 left-1/2 z-50 w-[min(30rem,calc(100vw-2rem))] -translate-x-1/2 -translate-y-1/2 rounded-card bg-surface p-6 shadow-popover">
          <div className="flex items-start justify-between gap-4">
            <div>
              <Dialog.Title className="text-[15px] font-semibold text-ink">
                Новая связь задачи
              </Dialog.Title>
              <Dialog.Description className="mt-1 text-[13px] text-ink-muted">
                Связь участвует в расчёте критического пути и каскадном сдвиге сроков.
              </Dialog.Description>
            </div>
            <Dialog.Close
              aria-label="Закрыть"
              className="rounded-control p-1 text-ink-faint hover:text-ink focus-visible:focus-ring"
            >
              <X className="size-4" />
            </Dialog.Close>
          </div>

          <form
            className="mt-5 space-y-4"
            onSubmit={(event) => {
              event.preventDefault();
              submit();
            }}
          >
            <div className="space-y-1.5">
              <Label htmlFor="dependency-task">Связать с задачей</Label>
              <Select
                id="dependency-task"
                value={relatedTaskId}
                onChange={(event) => setRelatedTaskId(event.target.value)}
                required
              >
                <option value="">Выберите задачу…</option>
                {available.map((task) => (
                  <option key={task.id} value={task.id}>
                    #{task.wbsNumber} {task.title}
                  </option>
                ))}
              </Select>
            </div>

            <div className="space-y-1.5">
              <Label htmlFor="dependency-direction">Направление</Label>
              <Select
                id="dependency-direction"
                value={direction}
                onChange={(event) => setDirection(event.target.value as TaskDependencyDirection)}
              >
                {DIRECTION_OPTIONS.map((option) => (
                  <option key={option.value} value={option.value}>
                    {option.label}
                  </option>
                ))}
              </Select>
            </div>

            <div className="grid grid-cols-[1fr_auto] gap-3">
              <div className="space-y-1.5">
                <Label htmlFor="dependency-type">Тип связи</Label>
                <Select
                  id="dependency-type"
                  value={type}
                  onChange={(event) => setType(event.target.value as DependencyType)}
                >
                  {TYPE_OPTIONS.map((option) => (
                    <option key={option.value} value={option.value}>
                      {option.label}
                    </option>
                  ))}
                </Select>
              </div>
              <div className="space-y-1.5">
                <Label htmlFor="dependency-lag">Лаг, дней</Label>
                <Input
                  id="dependency-lag"
                  type="number"
                  value={lagDays}
                  onChange={(event) => setLagDays(Number(event.target.value))}
                  className="w-28"
                />
              </div>
            </div>

            {createDependency.isError ? (
              <Alert tone="danger">
                {toUserMessage(
                  createDependency.error,
                  {
                    409: "Такая связь уже существует или создаёт цикл",
                    403: "Недостаточно прав для изменения связей",
                  },
                  "Не удалось создать связь",
                )}
              </Alert>
            ) : null}

            <div className="flex justify-end gap-2 pt-1">
              <Dialog.Close asChild>
                <Button type="button" variant="ghost">
                  Отмена
                </Button>
              </Dialog.Close>
              <Button type="submit" disabled={!relatedTaskId || createDependency.isPending}>
                {createDependency.isPending ? "Создаём…" : "Создать связь"}
              </Button>
            </div>
          </form>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
