"use client";

import * as Dialog from "@radix-ui/react-dialog";
import { Link2, X } from "lucide-react";
import { useState, type ReactNode } from "react";

import { TaskMultiPicker } from "@/components/task/task-multi-picker";
import { Alert } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { useGantt, useLinkPredecessors } from "@/data/queries";
import { toUserMessage } from "@/lib/api-error-message";
import { formatDayMonth } from "@/lib/format";
import { predecessorCandidates } from "@/lib/milestone-links";
import type { Task } from "@/types";

/**
 * Массовая привязка работ, ведущих к задаче-вехе.
 *
 * Каждая отмеченная задача становится предшественником вехи со связью
 * «окончание к началу» без лага: у события нулевой длительности другой смысл
 * у связи редко нужен, а выбор типа и направления на каждую задачу только мешал.
 * Для особых случаев остаётся карточка задачи-предшественника.
 */
export function LinkPredecessorsDialog({
  milestoneTask,
  projectTasks,
  trigger,
}: {
  milestoneTask: Task;
  projectTasks: Task[];
  trigger: ReactNode;
}) {
  const [open, setOpen] = useState(false);
  const [selected, setSelected] = useState<string[]>([]);
  const [failed, setFailed] = useState<string[]>([]);

  // Для проверки на цикл нужен весь граф проекта — тот же кэш, что у экрана Ганта.
  const gantt = useGantt(milestoneTask.projectId, "weeks");
  const linkPredecessors = useLinkPredecessors();
  const { available, wouldCycle } = predecessorCandidates(
    milestoneTask,
    projectTasks,
    gantt.data?.dependencies ?? [],
  );

  function submit() {
    setFailed([]);
    linkPredecessors.mutate(
      { successorTaskId: milestoneTask.id, predecessorIds: selected },
      {
        onSuccess: (errors) => {
          if (errors.length > 0) {
            setFailed(
              errors.map(({ taskId, error }) => {
                const title = projectTasks.find((task) => task.id === taskId)?.title ?? taskId;
                return `${title}: ${toUserMessage(error, { 409: "связь уже есть" }, "не удалось связать")}`;
              }),
            );
            // Отмеченными оставляем только несвязанные — удавшиеся уже в списке вехи.
            setSelected(errors.map(({ taskId }) => taskId));
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
        <Dialog.Content className="fixed top-1/2 left-1/2 z-50 max-h-[calc(100vh-2rem)] w-[min(34rem,calc(100vw-2rem))] -translate-x-1/2 -translate-y-1/2 overflow-y-auto rounded-card bg-surface p-6 shadow-popover">
          <div className="flex items-start justify-between gap-4">
            <div>
              <Dialog.Title className="text-15 font-semibold text-ink">
                Работы, ведущие к вехе
              </Dialog.Title>
              <Dialog.Description className="mt-1 text-13 text-ink-muted">
                Отметьте, что должно завершиться до вехи «{milestoneTask.title}» (
                {formatDayMonth(milestoneTask.startDate)}). Веха наступит не раньше, чем
                закончится последняя из них.
              </Dialog.Description>
            </div>
            <Dialog.Close
              aria-label="Закрыть"
              className="rounded-control p-1 text-ink-faint hover:text-ink focus-visible:focus-ring"
            >
              <X className="size-4" />
            </Dialog.Close>
          </div>

          <div className="mt-5 space-y-4">
            {gantt.isPending ? (
              <p className="py-8 text-center text-13 text-ink-muted">Загружаем связи проекта…</p>
            ) : (
              <TaskMultiPicker
                label="Работы, ведущие к вехе"
                tasks={available}
                selected={selected}
                onChange={setSelected}
                disabledTasks={wouldCycle}
                disabledReason="идёт после вехи — связь замкнула бы цикл"
                lateAfter={milestoneTask.startDate}
                emptyText="Все задачи проекта уже ведут к этой вехе."
              />
            )}

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
                disabled={selected.length === 0 || linkPredecessors.isPending}
                loading={linkPredecessors.isPending}
              >
                {linkPredecessors.isPending ? null : <Link2 />}
                Привязать
              </Button>
            </div>
          </div>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
