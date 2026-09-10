"use client";

import * as Dialog from "@radix-ui/react-dialog";
import { Link2, X } from "lucide-react";
import { useState, type ReactNode } from "react";

import { Alert } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useCreateDependency } from "@/data/queries";
import { toUserMessage } from "@/lib/api-error-message";
import { plural } from "@/lib/format";
import type { Task } from "@/types";

/**
 * Массовая привязка задач к вехе.
 *
 * Привязки «задача → веха» в модели нет, принадлежность выражается связями
 * (см. docs/backend-request-milestones.md). Раньше каждую связь приходилось
 * создавать отдельной модалкой — здесь отмечаются сразу все нужные задачи.
 *
 * Запросы уходят по одному: эндпоинт принимает одну связь за раз. Ошибки
 * собираются и показываются списком, успешные связи при этом остаются —
 * откатывать уже созданное было бы хуже, чем сообщить о частичном результате.
 */
export function LinkTasksDialog({
  taskId,
  candidates,
  linkedTaskIds,
  trigger,
}: {
  taskId: string;
  candidates: Task[];
  linkedTaskIds: string[];
  trigger: ReactNode;
}) {
  const [open, setOpen] = useState(false);
  const [selected, setSelected] = useState<string[]>([]);
  const [lagDays, setLagDays] = useState(0);
  const [failed, setFailed] = useState<string[]>([]);
  const [busy, setBusy] = useState(false);

  const createDependency = useCreateDependency(taskId);
  const linked = new Set([...linkedTaskIds, taskId]);
  const available = candidates.filter((task) => !linked.has(task.id));

  async function submit() {
    setBusy(true);
    setFailed([]);
    const errors: string[] = [];

    for (const id of selected) {
      try {
        await createDependency.mutateAsync({
          relatedTaskId: id,
          direction: "predecessor",
          type: "FS",
          lagDays,
        });
      } catch (error) {
        const title = candidates.find((task) => task.id === id)?.title ?? id;
        errors.push(`${title}: ${toUserMessage(error, {}, "не удалось связать")}`);
      }
    }

    setBusy(false);
    if (errors.length > 0) {
      setFailed(errors);
      return;
    }
    setSelected([]);
    setOpen(false);
  }

  return (
    <Dialog.Root
      open={open}
      onOpenChange={(next) => {
        setOpen(next);
        if (!next) {
          setSelected([]);
          setFailed([]);
        }
      }}
    >
      <Dialog.Trigger asChild>{trigger}</Dialog.Trigger>

      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-0 z-50 bg-ink/30" />
        <Dialog.Content className="fixed top-1/2 left-1/2 z-50 max-h-[calc(100vh-2rem)] w-[min(32rem,calc(100vw-2rem))] -translate-x-1/2 -translate-y-1/2 overflow-y-auto rounded-card bg-surface p-6 shadow-popover">
          <div className="flex items-start justify-between gap-4">
            <div>
              <Dialog.Title className="text-[15px] font-semibold text-ink">
                Привязать задачи к вехе
              </Dialog.Title>
              <Dialog.Description className="mt-1 text-[13px] text-ink-muted">
                Отметьте работы, завершение которых означает достижение вехи.
                Связи создаются типом FS — веха наступает после их окончания.
              </Dialog.Description>
            </div>
            <Dialog.Close
              aria-label="Закрыть"
              className="rounded-control p-1 text-ink-faint hover:text-ink focus-visible:focus-ring"
            >
              <X className="size-4" />
            </Dialog.Close>
          </div>

          {available.length === 0 ? (
            <p className="py-8 text-center text-[13px] text-ink-muted">
              Все задачи проекта уже связаны с этой вехой.
            </p>
          ) : (
            <div className="mt-5 space-y-4">
              <ul className="max-h-64 space-y-1 overflow-y-auto rounded-control border border-line p-2">
                {available.map((task) => (
                  <li key={task.id}>
                    <label className="flex cursor-pointer items-center gap-2.5 rounded-control px-2 py-1.5 hover:bg-surface-muted">
                      <Checkbox
                        checked={selected.includes(task.id)}
                        onCheckedChange={(next) =>
                          setSelected((current) =>
                            next === true
                              ? [...current, task.id]
                              : current.filter((id) => id !== task.id),
                          )
                        }
                      />
                      <span className="min-w-0 flex-1 truncate text-[13px] text-ink">
                        <span className="font-mono text-ink-faint">#{task.wbsNumber}</span>{" "}
                        {task.title}
                      </span>
                    </label>
                  </li>
                ))}
              </ul>

              <div className="flex items-end justify-between gap-3">
                <div className="space-y-1.5">
                  <Label htmlFor="link-lag">Лаг, дней</Label>
                  <Input
                    id="link-lag"
                    type="number"
                    value={lagDays}
                    onChange={(event) => setLagDays(Number(event.target.value))}
                    className="w-28"
                  />
                </div>
                <span className="text-xs text-ink-faint">
                  {selected.length > 0
                    ? `Будет создано ${selected.length} ${plural(selected.length, ["связь", "связи", "связей"])}`
                    : "Ничего не выбрано"}
                </span>
              </div>

              {failed.length > 0 ? (
                <Alert tone="danger">
                  Часть связей создать не удалось:
                  <ul className="mt-1 list-disc pl-4">
                    {failed.map((message) => (
                      <li key={message}>{message}</li>
                    ))}
                  </ul>
                </Alert>
              ) : null}

              <div className="flex justify-end gap-2">
                <Dialog.Close asChild>
                  <Button type="button" variant="ghost">
                    Отмена
                  </Button>
                </Dialog.Close>
                <Button onClick={submit} disabled={selected.length === 0 || busy}>
                  <Link2 />
                  {busy ? "Связываем…" : "Привязать"}
                </Button>
              </div>
            </div>
          )}
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
