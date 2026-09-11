"use client";

import * as DropdownMenu from "@radix-ui/react-dropdown-menu";
import { CalendarDays, Check, Flag, MoveRight, Search } from "lucide-react";
import Link from "next/link";
import { useState, type DragEvent } from "react";

import { UserHoverCard } from "@/components/user/user-card";
import { Alert } from "@/components/ui/alert";
import { Avatar } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Progress } from "@/components/ui/progress";
import { Segmented } from "@/components/ui/segmented";
import { useProjectAccess } from "@/data/project-access";
import { useUpdateTaskStatus } from "@/data/queries";
import { toUserMessage } from "@/lib/api-error-message";
import {
  BOARD_COLUMNS,
  groupByColumn,
  rememberPlacement,
  type BoardColumnId,
} from "@/lib/board";
import { formatDayMonth, pluralizeCount } from "@/lib/format";
import { TASK_STATUS_META, isAssignableStatus } from "@/lib/task-status";
import { cn } from "@/lib/utils";
import type { Task } from "@/types";

type Scope = "all" | "mine";

const SCOPE_OPTIONS = [
  { value: "all", label: "Все задачи" },
  { value: "mine", label: "Мои" },
] as const;

const COLUMN_DOT: Record<BoardColumnId, string> = {
  planned: "bg-brand",
  in_progress: "bg-warning",
  done: "bg-success",
};

/** MIME-тип перетаскиваемой карточки: чужие перетаскивания (файлы, текст) доска не принимает. */
const DRAG_TYPE = "application/x-taygant-task";

function matches(task: Task, query: string): boolean {
  const trimmed = query.trim().toLowerCase();
  if (!trimmed) return true;
  return (
    task.code.toLowerCase().includes(trimmed) ||
    task.title.toLowerCase().includes(trimmed) ||
    (task.assignee?.fullName.toLowerCase().includes(trimmed) ?? false)
  );
}

/**
 * Реестр задач проекта в виде канбан-доски.
 *
 * Колонки фиксированы и совпадают со статусами, которые пользователь вправе
 * выставить сам: План, В работе, Завершены. Перенос карточки — это PATCH
 * статуса; права те же, что у карточки задачи (full — любые, edit — свои).
 * Для клавиатуры у карточки есть меню «Переместить»: перетаскивание мышью
 * без него было бы единственным способом сменить колонку.
 */
export function KanbanBoard({ projectId, tasks }: { projectId: string; tasks: Task[] }) {
  const access = useProjectAccess();
  const updateStatus = useUpdateTaskStatus(projectId);

  const [query, setQuery] = useState("");
  const [scope, setScope] = useState<Scope>("all");
  const [draggedId, setDraggedId] = useState<string | null>(null);
  const [overColumn, setOverColumn] = useState<BoardColumnId | null>(null);

  const visible = tasks.filter(
    (task) =>
      matches(task, query) &&
      (scope === "all" || task.assignee?.id === access.currentUserId),
  );
  const columns = groupByColumn(visible);
  const dragged = draggedId ? tasks.find((task) => task.id === draggedId) : undefined;

  const move = (task: Task, column: BoardColumnId, from: BoardColumnId) => {
    if (column === from || !access.canEditTask(task)) return;
    rememberPlacement(task.id, column);
    updateStatus.mutate({ taskId: task.id, status: column });
  };

  const onDrop = (event: DragEvent, column: BoardColumnId) => {
    event.preventDefault();
    setOverColumn(null);
    setDraggedId(null);
    const taskId = event.dataTransfer.getData(DRAG_TYPE);
    const task = tasks.find((item) => item.id === taskId);
    if (!task) return;
    const from = BOARD_COLUMNS.find(({ id }) => columns[id].some((item) => item.id === taskId))?.id;
    if (from) move(task, column, from);
  };

  const hint = access.isFull
    ? "Перетащите карточку в другую колонку, чтобы сменить статус."
    : access.isEdit
      ? "Переносить между колонками можно только задачи, где вы исполнитель."
      : "У вас доступ только на просмотр — статусы менять нельзя.";

  return (
    <div className="space-y-4">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <Input
          icon={<Search />}
          value={query}
          onChange={(event) => setQuery(event.target.value)}
          placeholder="Код, название или исполнитель"
          aria-label="Фильтр задач на доске"
          className="bg-surface sm:max-w-80"
        />
        <div className="flex flex-wrap items-center gap-3">
          <p className="text-13 text-ink-muted">{hint}</p>
          <Segmented
            options={SCOPE_OPTIONS}
            value={scope}
            onChange={setScope}
            ariaLabel="Какие задачи показывать"
          />
        </div>
      </div>

      {updateStatus.isError ? (
        <Alert tone="danger">
          {toUserMessage(
            updateStatus.error,
            { 403: "Недостаточно прав, чтобы сменить статус этой задачи" },
            "Не удалось сменить статус — карточка вернулась на место",
          )}
        </Alert>
      ) : null}

      <div className="grid items-start gap-4 md:grid-cols-3" data-tour="board">
        {BOARD_COLUMNS.map((column) => {
          const items = columns[column.id];
          const canDropHere =
            dragged !== undefined &&
            access.canEditTask(dragged) &&
            !items.some((item) => item.id === dragged.id);

          return (
            <section
              key={column.id}
              aria-labelledby={`board-column-${column.id}`}
              onDragOver={(event) => {
                if (!canDropHere) return;
                event.preventDefault();
                event.dataTransfer.dropEffect = "move";
                setOverColumn(column.id);
              }}
              onDragLeave={(event) => {
                // dragleave срабатывает и при переходе на дочернюю карточку —
                // подсветку снимаем, только когда курсор вышел из колонки целиком.
                if (!event.currentTarget.contains(event.relatedTarget as Node | null)) {
                  setOverColumn((current) => (current === column.id ? null : current));
                }
              }}
              onDrop={(event) => onDrop(event, column.id)}
              className={cn(
                "flex min-h-40 flex-col rounded-card border border-line bg-surface-subtle transition-colors",
                canDropHere && "border-dashed border-brand/50",
                overColumn === column.id && canDropHere && "border-solid border-brand bg-brand-tint",
              )}
            >
              <header className="flex items-center justify-between gap-2 px-3 pt-3 pb-2">
                <h2
                  id={`board-column-${column.id}`}
                  className="flex items-center gap-2 text-13 font-semibold text-ink"
                >
                  <span className={cn("size-2 rounded-full", COLUMN_DOT[column.id])} aria-hidden />
                  {column.title}
                </h2>
                <span
                  className="rounded-control bg-surface px-1.5 py-0.5 font-mono text-2xs text-ink-muted"
                  aria-label={pluralizeCount(items.length, ["задача", "задачи", "задач"])}
                >
                  {items.length}
                </span>
              </header>

              {items.length === 0 ? (
                <p className="mx-3 mb-3 flex flex-1 items-center justify-center rounded-control border border-dashed border-line py-8 text-13 text-ink-faint">
                  {query || scope === "mine" ? "Ничего не найдено" : "Задач нет"}
                </p>
              ) : (
                <ul className="space-y-2 px-3 pb-3">
                  {items.map((task) => (
                    <li key={task.id}>
                      <BoardCard
                        task={task}
                        column={column.id}
                        canMove={access.canEditTask(task)}
                        isDragging={draggedId === task.id}
                        onDragStart={(event) => {
                          event.dataTransfer.setData(DRAG_TYPE, task.id);
                          event.dataTransfer.effectAllowed = "move";
                          setDraggedId(task.id);
                        }}
                        onDragEnd={() => {
                          setDraggedId(null);
                          setOverColumn(null);
                        }}
                        onMove={(target) => move(task, target, column.id)}
                      />
                    </li>
                  ))}
                </ul>
              )}
            </section>
          );
        })}
      </div>
    </div>
  );
}

function BoardCard({
  task,
  column,
  canMove,
  isDragging,
  onDragStart,
  onDragEnd,
  onMove,
}: {
  task: Task;
  column: BoardColumnId;
  canMove: boolean;
  isDragging: boolean;
  onDragStart: (event: DragEvent) => void;
  onDragEnd: () => void;
  onMove: (column: BoardColumnId) => void;
}) {
  // Вычисленный сервером статус (просрочена, заблокирована) колонкой не
  // выразить — он показывается бейджем, чтобы риск не терялся на доске.
  const derived = isAssignableStatus(task.status) ? null : TASK_STATUS_META[task.status];
  const atRisk = task.isCriticalPath && task.planVsActualDeviationDays < 0;

  return (
    <article
      draggable={canMove}
      onDragStart={onDragStart}
      onDragEnd={onDragEnd}
      className={cn(
        "rounded-control border border-line bg-surface p-3 shadow-card transition-opacity",
        canMove && "cursor-grab active:cursor-grabbing",
        isDragging && "opacity-40",
        atRisk && "border-danger/40",
      )}
    >
      <div className="flex items-center gap-2">
        <span className="font-mono text-2xs text-ink-faint">{task.code}</span>
        {task.isMilestone ? (
          <Badge tone="brand" size="sm">
            <span aria-hidden className="size-2 rotate-45 rounded-[1px] bg-brand" />
            Веха
          </Badge>
        ) : null}
        {derived ? (
          <Badge tone={derived.tone} size="sm" dot>
            {derived.label}
          </Badge>
        ) : null}
        {canMove ? <MoveMenu task={task} column={column} onMove={onMove} /> : null}
      </div>

      <Link
        href={`/tasks/${task.id}`}
        draggable={false}
        className="mt-1.5 block rounded-control text-13 font-medium text-ink hover:text-brand focus-visible:focus-ring"
      >
        {task.title}
      </Link>

      {!task.isMilestone && column !== "planned" ? (
        <Progress
          value={task.progressPercent}
          className="mt-2.5 h-1"
          label={`Выполнено: ${Math.round(task.progressPercent)}%`}
        />
      ) : null}

      <div className="mt-2.5 flex items-center justify-between gap-2">
        {task.assignee ? (
          <UserHoverCard user={task.assignee}>
            <span className="flex min-w-0 items-center gap-1.5">
              <Avatar fullName={task.assignee.fullName} className="size-5 rounded-full text-3xs" />
              <span className="truncate text-xs text-ink-muted">{task.assignee.fullName}</span>
            </span>
          </UserHoverCard>
        ) : (
          <span className="text-xs text-ink-faint">Не назначен</span>
        )}

        <span
          className={cn(
            "flex shrink-0 items-center gap-1 font-mono text-2xs",
            task.status === "overdue" ? "text-danger-ink" : "text-ink-muted",
          )}
          title={task.isMilestone ? "Дата вехи" : "Плановое окончание"}
        >
          {atRisk ? <Flag className="size-3 text-danger" aria-label="Критический путь" /> : null}
          <CalendarDays className="size-3" aria-hidden />
          {formatDayMonth(task.endDate)}
        </span>
      </div>
    </article>
  );
}

function MoveMenu({
  task,
  column,
  onMove,
}: {
  task: Task;
  column: BoardColumnId;
  onMove: (column: BoardColumnId) => void;
}) {
  return (
    <DropdownMenu.Root>
      <DropdownMenu.Trigger
        aria-label={`Переместить задачу ${task.code}`}
        className="ml-auto flex size-6 shrink-0 cursor-pointer items-center justify-center rounded-control text-ink-faint transition-colors hover:bg-surface-muted hover:text-ink focus-visible:focus-ring data-[state=open]:bg-surface-muted"
      >
        <MoveRight className="size-3.5" />
      </DropdownMenu.Trigger>
      <DropdownMenu.Portal>
        <DropdownMenu.Content
          align="end"
          sideOffset={4}
          className="z-50 w-44 rounded-card border border-line bg-surface p-1 shadow-popover"
        >
          <DropdownMenu.Label className="px-2 py-1.5 text-2xs font-semibold tracking-wider text-ink-faint uppercase">
            Переместить в
          </DropdownMenu.Label>
          {BOARD_COLUMNS.map((target) => {
            const current = target.id === column;
            return (
              <DropdownMenu.Item
                key={target.id}
                disabled={current}
                onSelect={() => onMove(target.id)}
                className="flex cursor-pointer items-center gap-2 rounded-control px-2 py-1.5 text-13 text-ink outline-none data-[disabled]:cursor-default data-[disabled]:text-ink-faint data-[highlighted]:bg-surface-muted"
              >
                <Check className={cn("size-3.5 shrink-0", !current && "opacity-0")} />
                {target.title}
              </DropdownMenu.Item>
            );
          })}
        </DropdownMenu.Content>
      </DropdownMenu.Portal>
    </DropdownMenu.Root>
  );
}
