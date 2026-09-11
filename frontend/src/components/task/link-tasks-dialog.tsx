"use client";

import * as Dialog from "@radix-ui/react-dialog";
import { Link2, X } from "lucide-react";
import { useState, type ReactNode } from "react";

import { Alert } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { useAssignTasksToMilestone, useGantt } from "@/data/queries";
import { toUserMessage } from "@/lib/api-error-message";
import { plural } from "@/lib/format";
import { predecessorsLeadingTo } from "@/lib/milestone-links";
import type { Task } from "@/types";

/**
 * Массовая привязка задач к вехе через `milestoneId`.
 *
 * Отдельного эндпоинта нет, поэтому каждой задаче поле проставляется своим
 * PATCH. Ошибки собираются и показываются списком, а успешные привязки
 * остаются: откатывать их хуже, чем сообщить о частичном результате.
 *
 * Отметка задачи-вехи отмечает и работы, которые к ней ведут (прямых
 * предшественников, см. `predecessorsLeadingTo`), — их видно в списке и можно
 * снять. Снятие вехи снимает то, что было отмечено вместе с ней.
 */
export function LinkTasksDialog({
  projectId,
  milestoneId,
  milestoneName,
  candidates,
  trigger,
}: {
  projectId: string;
  /** Контрольная точка проекта, к которой привязываются работы. */
  milestoneId: string;
  milestoneName: string;
  candidates: Task[];
  trigger: ReactNode;
}) {
  const [open, setOpen] = useState(false);
  const [selected, setSelected] = useState<string[]>([]);
  const [failed, setFailed] = useState<string[]>([]);
  // Что отмечено автоматически вместе с каждой задачей-вехой: id вехи → id работ.
  const [autoAdded, setAutoAdded] = useState<Record<string, string[]>>({});

  const assignTasks = useAssignTasksToMilestone();
  // Связи берём из того же кэша, что и экран Ганта: отдельного запроса не будет.
  const dependencies = useGantt(projectId, "weeks").data?.dependencies ?? [];
  // Уже привязанные к этой вехе показываем отмеченными, чужие вехи не трогаем.
  // Задачи, уже привязанные к другой вехе, не показываем: перетаскивание между
  // вехами — отдельный сценарий, и молча переназначать чужую задачу неверно.
  const available = candidates.filter(
    (task) => !task.milestoneId || task.milestoneId === milestoneId,
  );

  function toggle(task: Task, checked: boolean) {
    if (checked) {
      const extra = task.isMilestone
        ? predecessorsLeadingTo(task, dependencies, candidates, milestoneId)
            .toLink.map((item) => item.id)
            .filter((id) => id !== task.id && !selected.includes(id))
        : [];
      setSelected([...selected, task.id, ...extra]);
      if (task.isMilestone) setAutoAdded({ ...autoAdded, [task.id]: extra });
      return;
    }

    // Снимаем веху — снимаем и то, что она отметила, если это же не отметила
    // другая ещё выбранная веха.
    const { [task.id]: ownExtra = [], ...rest } = autoAdded;
    const keptByOthers = new Set(
      Object.entries(rest)
        .filter(([ownerId]) => selected.includes(ownerId))
        .flatMap(([, ids]) => ids),
    );
    const removed = new Set([task.id, ...ownExtra.filter((id) => !keptByOthers.has(id))]);
    setSelected(selected.filter((id) => !removed.has(id)));
    setAutoAdded(rest);
  }

  // Для подписи «ведёт к вехе …» у автоматически отмеченной работы.
  const ownerTitleOf = new Map<string, string>();
  for (const [ownerId, ids] of Object.entries(autoAdded)) {
    if (!selected.includes(ownerId)) continue;
    const ownerTitle = candidates.find((task) => task.id === ownerId)?.title;
    if (!ownerTitle) continue;
    for (const id of ids) if (selected.includes(id)) ownerTitleOf.set(id, ownerTitle);
  }

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
          setAutoAdded({});
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
          setAutoAdded({});
        }
      }}
    >
      <Dialog.Trigger asChild>{trigger}</Dialog.Trigger>

      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-0 z-50 bg-ink/30" />
        <Dialog.Content className="fixed top-1/2 left-1/2 z-50 max-h-[calc(100vh-2rem)] w-[min(32rem,calc(100vw-2rem))] -translate-x-1/2 -translate-y-1/2 overflow-y-auto rounded-card bg-surface p-6 shadow-popover">
          <div className="flex items-start justify-between gap-4">
            <div>
              <Dialog.Title className="text-15 font-semibold text-ink">
                Привязать задачи к вехе
              </Dialog.Title>
              <Dialog.Description className="mt-1 text-13 text-ink-muted">
                Отметьте работы, которые ведут к вехе «{milestoneName}». Прогресс
                считается по ним. Вместе с задачей-вехой отметятся и работы,
                которые к ней ведут по связям.
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
            <p className="py-8 text-center text-13 text-ink-muted">
              Все задачи проекта уже связаны с этой вехой.
            </p>
          ) : (
            <div className="mt-5 space-y-4">
              <ul className="max-h-64 space-y-1 overflow-y-auto rounded-control border border-line p-2">
                {available.map((task) => {
                  const checked = selected.includes(task.id);
                  const ownerTitle = ownerTitleOf.get(task.id);
                  const leading = task.isMilestone
                    ? predecessorsLeadingTo(task, dependencies, candidates, milestoneId)
                    : null;
                  const addedCount = autoAdded[task.id]?.length ?? 0;
                  // Номера, а не только число: отмеченные работы могут быть за
                  // пределами видимой части прокручиваемого списка.
                  const numbersOf = (ids: string[]) =>
                    ids
                      .map((id) => `#${candidates.find((item) => item.id === id)?.wbsNumber ?? "?"}`)
                      .join(", ");
                  const otherCount = leading?.inOtherMilestones.length ?? 0;

                  return (
                    <li key={task.id}>
                      <label className="flex cursor-pointer items-start gap-2.5 rounded-control px-2 py-1.5 hover:bg-surface-muted">
                        <Checkbox
                          className="mt-0.5"
                          checked={checked}
                          onCheckedChange={(next) => toggle(task, next === true)}
                        />
                        <span className="min-w-0 flex-1">
                          <span
                            className={
                              task.isMilestone
                                ? "flex items-center gap-1.5 text-13 font-semibold text-brand"
                                : "block truncate text-13 text-ink"
                            }
                          >
                            {task.isMilestone ? (
                              <span
                                aria-hidden
                                className="size-2 shrink-0 rotate-45 rounded-[1px] bg-brand"
                              />
                            ) : null}
                            <span className="truncate">
                              <span className="font-mono font-normal text-ink-faint">
                                #{task.wbsNumber}
                              </span>{" "}
                              {task.isMilestone ? <span className="sr-only">Веха: </span> : null}
                              {task.title}
                            </span>
                          </span>
                          {ownerTitle ? (
                            <span className="block text-2xs text-ink-faint">
                              ведёт к вехе «{ownerTitle}»
                            </span>
                          ) : null}
                          {checked && leading ? (
                            <span className="block text-2xs text-ink-faint">
                              {addedCount === 1
                                ? `Вместе с вехой отмечена ведущая к ней работа ${numbersOf(autoAdded[task.id])}`
                                : addedCount > 1
                                  ? `Вместе с вехой отмечены ведущие к ней работы ${numbersOf(autoAdded[task.id])}`
                                  : "Новых работ, ведущих к вехе, нет"}
                              {otherCount === 1
                                ? ` · работа ${numbersOf(leading.inOtherMilestones.map((item) => item.id))} уже в другой КТ — не тронута`
                                : otherCount > 1
                                  ? ` · работы ${numbersOf(leading.inOtherMilestones.map((item) => item.id))} уже в других КТ — не тронуты`
                                  : ""}
                            </span>
                          ) : null}
                        </span>
                      </label>
                    </li>
                  );
                })}
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
