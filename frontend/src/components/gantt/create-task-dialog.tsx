"use client";

import * as Dialog from "@radix-ui/react-dialog";
import { addDays, format } from "date-fns";
import { Plus, X } from "lucide-react";
import { useState, type ReactNode } from "react";

import { Alert } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { DateInput } from "@/components/ui/date-input";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select } from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import { useCreateTask, useProjectMembers } from "@/data/queries";
import { toUserMessage } from "@/lib/api-error-message";
import type { DependencyType, Task } from "@/types";

const toIso = (date: Date) => format(date, "yyyy-MM-dd");

const TYPE_OPTIONS: { value: DependencyType; label: string }[] = [
  { value: "FS", label: "FS — окончание к началу" },
  { value: "SS", label: "SS — начало к началу" },
  { value: "FF", label: "FF — окончание к окончанию" },
  { value: "SF", label: "SF — начало к окончанию" },
];

/**
 * Создание задачи проекта.
 *
 * Предшественника можно указать сразу: бэкенд принимает связи в теле создания
 * (`predecessors`), и это избавляет от второго шага «создать, потом связать».
 */
export function CreateTaskDialog({
  projectId,
  projectTasks,
  trigger,
}: {
  projectId: string;
  projectTasks: Task[];
  /** Кнопка-триггер: в шапке это «+ Задача», внизу таблицы — строка-подсказка. */
  trigger: ReactNode;
}) {
  const [open, setOpen] = useState(false);
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [assigneeId, setAssigneeId] = useState("");
  const [startDate, setStartDate] = useState(() => toIso(new Date()));
  const [endDate, setEndDate] = useState(() => toIso(addDays(new Date(), 5)));
  const [isMilestone, setIsMilestone] = useState(false);
  const [predecessorId, setPredecessorId] = useState("");
  const [dependencyType, setDependencyType] = useState<DependencyType>("FS");

  const members = useProjectMembers(projectId);
  const createTask = useCreateTask(projectId);

  const validRange = endDate >= startDate;
  const canSubmit = title.trim().length > 0 && validRange && !createTask.isPending;

  function reset() {
    setTitle("");
    setDescription("");
    setAssigneeId("");
    setPredecessorId("");
    setIsMilestone(false);
  }

  function submit() {
    if (!canSubmit) return;
    createTask.mutate(
      {
        title: title.trim(),
        description: description.trim() || undefined,
        assigneeId: assigneeId || undefined,
        startDate,
        // У вехи нулевая длительность — конец совпадает с началом.
        endDate: isMilestone ? startDate : endDate,
        isMilestone,
        predecessors: predecessorId
          ? [{ taskId: predecessorId, type: dependencyType, lagDays: 0 }]
          : undefined,
      },
      {
        onSuccess: () => {
          setOpen(false);
          reset();
        },
      },
    );
  }

  return (
    <Dialog.Root
      open={open}
      onOpenChange={(next) => {
        setOpen(next);
        if (!next) createTask.reset();
      }}
    >
      <Dialog.Trigger asChild>{trigger}</Dialog.Trigger>

      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-0 z-50 bg-ink/30" />
        <Dialog.Content className="fixed top-1/2 left-1/2 z-50 max-h-[calc(100vh-2rem)] w-[min(34rem,calc(100vw-2rem))] -translate-x-1/2 -translate-y-1/2 overflow-y-auto rounded-card bg-surface p-6 shadow-popover">
          <div className="flex items-start justify-between gap-4">
            <div>
              <Dialog.Title className="text-[15px] font-semibold text-ink">
                Новая задача
              </Dialog.Title>
              <Dialog.Description className="mt-1 text-[13px] text-ink-muted">
                Задача появится в реестре и на диаграмме Ганта.
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
              <Label htmlFor="task-title">
                Название <span className="text-danger">*</span>
              </Label>
              <Input
                id="task-title"
                value={title}
                onChange={(event) => setTitle(event.target.value)}
                placeholder="Например, Нагрузочное тестирование"
                required
                autoFocus
              />
            </div>

            <div className="space-y-1.5">
              <Label htmlFor="task-desc">Описание</Label>
              <Textarea
                id="task-desc"
                rows={3}
                value={description}
                onChange={(event) => setDescription(event.target.value)}
              />
            </div>

            <div className="space-y-1.5">
              <Label htmlFor="task-assignee">Ответственный</Label>
              <Select
                id="task-assignee"
                value={assigneeId}
                onChange={(event) => setAssigneeId(event.target.value)}
              >
                <option value="">Не назначен</option>
                {(members.data ?? []).map((member) => (
                  <option key={member.user.id} value={member.user.id}>
                    {member.user.fullName}
                  </option>
                ))}
              </Select>
            </div>

            <label className="flex cursor-pointer items-start gap-2.5 rounded-control bg-surface-subtle p-3">
              <Checkbox
                checked={isMilestone}
                onCheckedChange={(checked) => setIsMilestone(checked === true)}
                className="mt-0.5"
              />
              <span>
                <span className="block text-[13px] font-medium text-ink">
                  Контрольная точка (веха)
                </span>
                <span className="mt-0.5 block text-xs text-ink-muted">
                  Событие с нулевой длительностью — дата окончания не нужна.
                </span>
              </span>
            </label>

            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-1.5">
                <Label htmlFor="task-start">Начало</Label>
                <DateInput id="task-start" value={startDate} onChange={setStartDate} />
              </div>
              {!isMilestone ? (
                <div className="space-y-1.5">
                  <Label htmlFor="task-end">Окончание</Label>
                  <DateInput
                    id="task-end"
                    value={endDate}
                    onChange={setEndDate}
                    minDate={startDate}
                    invalid={!validRange}
                  />
                </div>
              ) : null}
            </div>

            {projectTasks.length > 0 ? (
              <div className="grid grid-cols-[1fr_auto] gap-3">
                <div className="space-y-1.5">
                  <Label htmlFor="task-predecessor">Предшественник</Label>
                  <Select
                    id="task-predecessor"
                    value={predecessorId}
                    onChange={(event) => setPredecessorId(event.target.value)}
                  >
                    <option value="">Без связи</option>
                    {projectTasks.map((task) => (
                      <option key={task.id} value={task.id}>
                        #{task.wbsNumber} {task.title}
                      </option>
                    ))}
                  </Select>
                </div>
                {predecessorId ? (
                  <div className="space-y-1.5">
                    <Label htmlFor="task-dep-type">Тип</Label>
                    <Select
                      id="task-dep-type"
                      value={dependencyType}
                      onChange={(event) => setDependencyType(event.target.value as DependencyType)}
                      className="w-44"
                    >
                      {TYPE_OPTIONS.map((option) => (
                        <option key={option.value} value={option.value}>
                          {option.label}
                        </option>
                      ))}
                    </Select>
                  </div>
                ) : null}
              </div>
            ) : null}

            {createTask.isError ? (
              <Alert tone="danger">
                {toUserMessage(
                  createTask.error,
                  { 403: "Недостаточно прав для создания задач" },
                  "Не удалось создать задачу",
                )}
              </Alert>
            ) : null}

            <div className="flex justify-end gap-2 pt-1">
              <Dialog.Close asChild>
                <Button type="button" variant="ghost">
                  Отмена
                </Button>
              </Dialog.Close>
              <Button type="submit" disabled={!canSubmit}>
                <Plus />
                {createTask.isPending ? "Создаём…" : "Создать задачу"}
              </Button>
            </div>
          </form>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
