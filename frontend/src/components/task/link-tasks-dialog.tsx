"use client";

import * as Dialog from "@radix-ui/react-dialog";
import { Link2, X } from "lucide-react";
import { useState, type ReactNode } from "react";

import { Alert } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { useAssignTasksToMilestone } from "@/data/queries";
import { toUserMessage } from "@/lib/api-error-message";
import { plural } from "@/lib/format";
import type { Task } from "@/types";

/**
 * Массовая привязка задач к вехе через `milestoneId`.
 *
 * Отдельного эндпоинта нет, поэтому каждой задаче поле проставляется своим
 * PATCH. Ошибки собираются и показываются списком, а успешные привязки
 * остаются: откатывать их хуже, чем сообщить о частичном результате.
 */
export function LinkTasksDialog({
  milestoneId,
  milestoneName,
  candidates,
  trigger,
}: {
  /** Контрольная точка проекта, к которой привязываются работы. */
  milestoneId: string;
  milestoneName: string;
  candidates: Task[];
  trigger: ReactNode;
}) {
  const [open, setOpen] = useState(false);
  const [selected, setSelected] = useState<string[]>([]);
  const [failed, setFailed] = useState<string[]>([]);

  const assignTasks = useAssignTasksToMilestone();
  // Уже привязанные к этой вехе показываем отмеченными, чужие вехи не трогаем.
  // Задачи, уже привязанные к другой вехе, не показываем: перетаскивание между
  // вехами — отдельный сценарий, и молча переназначать чужую задачу неверно.
  const available = candidates.filter(
    (task) => !task.milestoneId || task.milestoneId === milestoneId,
  );

  function submit() {
    setFailed([]);
    assignTasks.mutate(
      { taskIds: selected, milestoneId },
      {
        onSuccess: (errors) => {
          if (errors.length > 0) {
            setFailed(
              errors.map(({ taskId, error }) => {
                const title = candidates.find((task) => task.id === taskId)?.title ?? taskId;
                return `${title}: ${toUserMessage(error, {}, "не удалось привязать")}`;
              }),
            );
            return;
          }
          setSelected([]);
          setOpen(false);
        },
      },
    );
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
                Отметьте работы, которые ведут к вехе «{milestoneName}». Прогресс
                считается по ним.
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

              <div className="flex items-center justify-end gap-3">
                <span className="text-xs text-ink-faint">
                  {selected.length > 0
                    ? `Будет привязано ${selected.length} ${plural(selected.length, ["задача", "задачи", "задач"])}`
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
                <Button
                  onClick={submit}
                  disabled={selected.length === 0 || assignTasks.isPending}
                >
                  <Link2 />
                  {assignTasks.isPending ? "Привязываем…" : "Привязать"}
                </Button>
              </div>
            </div>
          )}
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
