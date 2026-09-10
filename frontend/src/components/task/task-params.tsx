"use client";

import { Building2, Plus, Trash2 } from "lucide-react";
import { useEffect, useState } from "react";

import { Avatar } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import { DateInput } from "@/components/ui/date-input";
import { Select } from "@/components/ui/select";
import { Card, CardBody, CardHeader, CardTitle } from "@/components/ui/card";
import { Checkbox } from "@/components/ui/checkbox";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import {
  useAddChecklistItem,
  useDeleteChecklistItem,
  useProjectMembers,
  useToggleChecklistItem,
} from "@/data/queries";
import { useProjectAccess } from "@/data/project-access";
import { formatSigned } from "@/lib/format";
import { TASK_STATUS_META } from "@/lib/task-status";
import { cn } from "@/lib/utils";
import type { Project, TaskDetail, TaskStatus, TaskUpdateRequest } from "@/types";

const SECTION = "text-[11px] font-semibold tracking-wider text-ink-faint uppercase";

/** Левая колонка редактора: паспорт задачи, прогресс, описание и DoD. */
export function TaskParams({
  task,
  project,
  plannedProgressPercent,
  onDraftChange,
}: {
  task: TaskDetail;
  project: Project;
  /** Плановый прогресс на сегодня — база для расчёта опережения/отставания. */
  plannedProgressPercent: number;
  /** Несохранённые правки — страница по ним решает, активна ли кнопка сохранения. */
  onDraftChange: (draft: TaskUpdateRequest | null) => void;
}) {
  const [progress, setProgress] = useState(task.progressPercent);
  const [status, setStatus] = useState<TaskStatus>(task.status);
  const [startDate, setStartDate] = useState(task.startDate);
  const [endDate, setEndDate] = useState(task.endDate);
  const [description, setDescription] = useState(task.description ?? "");
  const [title, setTitle] = useState(task.title);
  const [assigneeId, setAssigneeId] = useState(task.assignee?.id ?? "");
  const [newCriterion, setNewCriterion] = useState("");

  const members = useProjectMembers(project.id);
  const access = useProjectAccess();
  const toggleChecklistItem = useToggleChecklistItem(task.id);
  const addChecklistItem = useAddChecklistItem(task.id);
  const deleteChecklistItem = useDeleteChecklistItem(task.id);

  // Уровень edit правит только статус и сроки своих задач; полный доступ — всё.
  const canEdit = access.canEditTask(task);
  const restricted = access.isRestrictedTask(task);
  const canManageChecklist = access.isFull;

  // Сервер — источник истины: после сохранения приходит обновлённая задача,
  // и локальные поля надо подтянуть, иначе они «залипнут» на старых значениях.
  useEffect(() => {
    setProgress(task.progressPercent);
    setStatus(task.status);
    setStartDate(task.startDate);
    setEndDate(task.endDate);
    setDescription(task.description ?? "");
    setTitle(task.title);
    setAssigneeId(task.assignee?.id ?? "");
  }, [task]);

  useEffect(() => {
    if (!canEdit) {
      onDraftChange(null);
      return;
    }
    const draft: TaskUpdateRequest = {};
    if (status !== task.status) draft.status = status;
    if (startDate !== task.startDate) draft.startDate = startDate;
    if (endDate !== task.endDate) draft.endDate = endDate;
    // Остальные поля доступны только полному доступу: у уровня edit их правка
    // отклоняется бэкендом, и черновик не должен их содержать.
    if (!restricted) {
      if (progress !== task.progressPercent) draft.progressPercent = progress;
      if (description !== (task.description ?? "")) draft.description = description;
      if (title.trim() && title !== task.title) draft.title = title.trim();
      if (assigneeId !== (task.assignee?.id ?? "")) draft.assigneeId = assigneeId || null;
    }
    onDraftChange(Object.keys(draft).length > 0 ? draft : null);
  }, [
    progress,
    status,
    startDate,
    endDate,
    description,
    title,
    assigneeId,
    task,
    canEdit,
    restricted,
    onDraftChange,
  ]);

  const deviation = progress - plannedProgressPercent;
  const doneCount = task.checklist.filter((item) => item.isDone).length;

  return (
    <Card className="flex flex-col">
      <CardHeader>
        <CardTitle>Параметры задачи</CardTitle>
        <span className="font-mono text-xs text-ink-faint">ID: {task.id}</span>
      </CardHeader>

      <CardBody className="space-y-5">
        <div>
          <label htmlFor="task-title" className={SECTION}>
            Название задачи
          </label>
          <Textarea
            id="task-title"
            rows={2}
            value={title}
            disabled={!canEdit || restricted}
            onChange={(event) => setTitle(event.target.value)}
            className="mt-2 font-semibold"
            invalid={title.trim().length === 0}
          />
        </div>

        <div>
          <p className={SECTION}>Проект</p>
          <div className="mt-2 flex items-center justify-between gap-3 rounded-control bg-surface-muted px-3 py-2.5">
            <span className="flex min-w-0 items-center gap-2">
              <Building2 className="size-4 shrink-0 text-brand" />
              <span className="truncate text-[13px] font-medium text-ink">{project.name}</span>
            </span>
            <span className="shrink-0 text-xs text-ink-faint">Релиз 2.1</span>
          </div>
        </div>

        <div>
          <label htmlFor="task-assignee" className={SECTION}>
            Ответственный
          </label>
          <div className="mt-2 flex items-center gap-2.5">
            <Avatar
              fullName={
                members.data?.find((member) => member.user.id === assigneeId)?.user.fullName ?? "—"
              }
              className="size-9 shrink-0 rounded-full text-xs"
            />
            <Select
              id="task-assignee"
              value={assigneeId}
              disabled={!canEdit || restricted}
              onChange={(event) => setAssigneeId(event.target.value)}
            >
              <option value="">Не назначен</option>
              {(members.data ?? []).map((member) => (
                <option key={member.user.id} value={member.user.id}>
                  {member.user.fullName}
                  {member.user.position ? ` — ${member.user.position}` : ""}
                </option>
              ))}
            </Select>
          </div>
        </div>

        <div>
          <p className={SECTION}>График выполнения</p>
          <div className="mt-2 grid grid-cols-2 gap-2">
            <div>
              <label htmlFor="task-start" className="text-xs text-ink-faint">
                Начало
              </label>
              <DateInput
                id="task-start"
                value={startDate}
                disabled={!canEdit}
                onChange={setStartDate}
                className="mt-1"
              />
            </div>
            <div>
              <label htmlFor="task-end" className="text-xs text-ink-faint">
                Окончание
              </label>
              <DateInput
                id="task-end"
                value={endDate}
                disabled={!canEdit}
                onChange={setEndDate}
                className="mt-1"
                minDate={startDate}
                invalid={endDate < startDate}
              />
            </div>
          </div>
          <p className="mt-3 flex items-baseline justify-between gap-3 text-[13px]">
            <span className="text-ink-muted">Длительность задачи:</span>
            <span className="text-right font-semibold text-ink">
              {task.durationCalendarDays} календарных дней ({task.durationWorkingDays} раб.)
            </span>
          </p>
        </div>

        <div>
          <label htmlFor="task-status" className={SECTION}>
            Статус
          </label>
          <Select
            id="task-status"
            value={status}
            disabled={!canEdit}
            onChange={(event) => setStatus(event.target.value as TaskStatus)}
            className="mt-2"
          >
            {(restricted
              ? (["planned", "in_progress", "done"] as TaskStatus[])
              : (["planned", "in_progress", "done", "overdue", "blocked"] as TaskStatus[])
            ).map((value) => (
              <option key={value} value={value}>
                {TASK_STATUS_META[value].label}
              </option>
            ))}
          </Select>
        </div>

        <div>
          <div className="flex items-center justify-between gap-3">
            <label htmlFor="task-progress" className={SECTION}>
              Текущий прогресс
            </label>
            <span className="flex items-center gap-1.5">
              <Input
                id="task-progress"
                type="number"
                min={0}
                max={100}
                value={progress}
                disabled={!canEdit || restricted}
                onChange={(event) => setProgress(Number(event.target.value))}
                className="h-8 w-20"
              />
              <span className="text-[13px] text-ink-muted">%</span>
            </span>
          </div>
          <p className="mt-3 flex items-baseline justify-between gap-3 text-[13px]">
            <span className="text-ink-muted">План на сегодня: {plannedProgressPercent}%</span>
            <span
              className={cn(
                "font-medium",
                deviation > 0 && "text-success",
                deviation < 0 && "text-danger",
                deviation === 0 && "text-ink-muted",
              )}
            >
              {formatSigned(deviation)}%{" "}
              {deviation > 0 ? "опережение" : deviation < 0 ? "отставание" : "в плане"}
            </span>
          </p>
        </div>

        <div>
          <label htmlFor="task-description" className={SECTION}>
            Описание задачи
          </label>
          <Textarea
            id="task-description"
            rows={7}
            value={description}
            disabled={!canEdit || restricted}
            onChange={(event) => setDescription(event.target.value)}
            className="mt-2"
          />
        </div>

        <div>
          <div className="flex items-center justify-between gap-3">
            <p className={SECTION}>Критерии приёмки (DoD)</p>
            <span className="font-mono text-xs text-brand">
              {doneCount} / {task.checklist.length} готово
            </span>
          </div>
          <ul className="mt-2 space-y-1.5">
            {task.checklist.map((item) => (
              <li key={item.id}>
                <label
                  className={cn(
                    "flex items-start gap-2.5 rounded-control px-3 py-2.5 transition-colors",
                    canManageChecklist && "cursor-pointer",
                    item.isDone ? "bg-surface-muted" : "bg-surface-subtle hover:bg-surface-muted",
                  )}
                >
                  <Checkbox
                    checked={item.isDone}
                    disabled={!canManageChecklist || toggleChecklistItem.isPending}
                    onCheckedChange={(checked) =>
                      toggleChecklistItem.mutate({
                        itemId: item.id,
                        isDone: checked === true,
                      })
                    }
                    className="mt-0.5"
                  />
                  <span
                    className={cn(
                      "min-w-0 flex-1 text-[13px] leading-snug",
                      item.isDone ? "text-ink-faint line-through" : "text-ink",
                    )}
                  >
                    {item.text}
                  </span>
                  {canManageChecklist ? (
                    <button
                      type="button"
                      onClick={(event) => {
                        // Кнопка лежит внутри <label>, иначе клик переключил бы галочку.
                        event.preventDefault();
                        deleteChecklistItem.mutate(item.id);
                      }}
                      disabled={deleteChecklistItem.isPending}
                      aria-label={`Удалить критерий «${item.text}»`}
                      className="shrink-0 rounded-control p-0.5 text-ink-faint transition-colors hover:text-danger focus-visible:focus-ring disabled:opacity-50"
                    >
                      <Trash2 className="size-3.5" />
                    </button>
                  ) : null}
                </label>
              </li>
            ))}
          </ul>

          {canManageChecklist ? (
            <form
              className="mt-2 flex items-center gap-2"
              onSubmit={(event) => {
                event.preventDefault();
                const text = newCriterion.trim();
                if (!text) return;
                addChecklistItem.mutate(text, { onSuccess: () => setNewCriterion("") });
              }}
            >
              <Input
                value={newCriterion}
                onChange={(event) => setNewCriterion(event.target.value)}
                placeholder="Новый критерий приёмки"
                aria-label="Новый критерий приёмки"
                className="h-8"
              />
              <Button
                type="submit"
                variant="secondary"
                size="sm"
                disabled={!newCriterion.trim() || addChecklistItem.isPending}
              >
                <Plus />
                Добавить
              </Button>
            </form>
          ) : null}
        </div>
      </CardBody>
    </Card>
  );
}
