"use client";

import * as Dialog from "@radix-ui/react-dialog";
import { addDays, format } from "date-fns";
import { Diamond, ListTodo, Plus, X } from "lucide-react";
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
import { cn } from "@/lib/utils";
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
  const [predecessorIds, setPredecessorIds] = useState<string[]>([]);
  const [dependencyType, setDependencyType] = useState<DependencyType>("FS");

  const members = useProjectMembers(projectId);
  const createTask = useCreateTask(projectId);

  const validRange = endDate >= startDate;
  const canSubmit = title.trim().length > 0 && validRange && !createTask.isPending;

  function reset() {
    setTitle("");
    setDescription("");
    setAssigneeId("");
    setPredecessorIds([]);
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
        // Бэкенд принимает связи массивом, поэтому веху можно сразу привязать
        // ко всем задачам, которые к ней ведут, а не создавать связи по одной.
        predecessors:
          predecessorIds.length > 0
            ? predecessorIds.map((taskId) => ({
                taskId,
                type: dependencyType,
                lagDays: 0,
              }))
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
            <div
              role="radiogroup"
              aria-label="Тип элемента графика"
              className="grid grid-cols-2 gap-2"
            >
              <TypeOption
                active={!isMilestone}
                onClick={() => setIsMilestone(false)}
                icon={<ListTodo className="size-4" />}
                title="Задача"
                hint="Работа с длительностью и прогрессом"
              />
              <TypeOption
                active={isMilestone}
                onClick={() => setIsMilestone(true)}
                icon={<Diamond className="size-4" />}
                title="Веха"
                hint="Событие-контроль нулевой длительности"
              />
            </div>

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
              <div className="space-y-2">
                <div className="flex items-baseline justify-between gap-3">
                  <Label>
                    {isMilestone
                      ? "Задачи, ведущие к вехе"
                      : "Предшественники"}
                  </Label>
                  <span className="text-xs text-ink-faint">
                    {predecessorIds.length > 0
                      ? `выбрано: ${predecessorIds.length}`
                      : "можно несколько"}
                  </span>
                </div>

                <ul className="max-h-44 space-y-1 overflow-y-auto rounded-control border border-line p-2">
                  {projectTasks.map((task) => {
                    const checked = predecessorIds.includes(task.id);
                    return (
                      <li key={task.id}>
                        <label className="flex cursor-pointer items-center gap-2.5 rounded-control px-2 py-1.5 hover:bg-surface-muted">
                          <Checkbox
                            checked={checked}
                            onCheckedChange={(next) =>
                              setPredecessorIds((current) =>
                                next === true
                                  ? [...current, task.id]
                                  : current.filter((id) => id !== task.id),
                              )
                            }
                          />
                          <span className="min-w-0 flex-1 truncate text-[13px] text-ink">
                            <span className="font-mono text-ink-faint">
                              #{task.wbsNumber}
                            </span>{" "}
                            {task.title}
                          </span>
                        </label>
                      </li>
                    );
                  })}
                </ul>

                {predecessorIds.length > 0 ? (
                  <div className="space-y-1.5">
                    <Label htmlFor="task-dep-type">Тип связи</Label>
                    <Select
                      id="task-dep-type"
                      value={dependencyType}
                      onChange={(event) =>
                        setDependencyType(event.target.value as DependencyType)
                      }
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

function TypeOption({
  active,
  onClick,
  icon,
  title,
  hint,
}: {
  active: boolean;
  onClick: () => void;
  icon: React.ReactNode;
  title: string;
  hint: string;
}) {
  return (
    <button
      type="button"
      role="radio"
      aria-checked={active}
      onClick={onClick}
      className={cn(
        "flex items-start gap-2.5 rounded-control border p-3 text-left transition-colors focus-visible:focus-ring",
        active
          ? "border-brand bg-brand-tint"
          : "border-line bg-surface hover:bg-surface-subtle",
      )}
    >
      <span className={active ? "mt-0.5 text-brand" : "mt-0.5 text-ink-faint"}>{icon}</span>
      <span className="min-w-0">
        <span
          className={cn(
            "block text-[13px] font-semibold",
            active ? "text-brand" : "text-ink",
          )}
        >
          {title}
        </span>
        <span className="mt-0.5 block text-xs text-ink-muted">{hint}</span>
      </span>
    </button>
  );
}
