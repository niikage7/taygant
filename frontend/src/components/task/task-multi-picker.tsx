"use client";

import { Search } from "lucide-react";
import { useState, type Dispatch, type SetStateAction } from "react";

import { Checkbox } from "@/components/ui/checkbox";
import { Input } from "@/components/ui/input";
import { formatDayMonth, plural } from "@/lib/format";
import { cn } from "@/lib/utils";
import type { Task } from "@/types";

/** С какого размера списка появляется фильтр по названию и номеру. */
const FILTER_THRESHOLD = 8;

/**
 * Список задач с галочками — выбор нескольких работ сразу (например, тех, что
 * ведут к задаче-вехе).
 *
 * `disabledTasks` показываются выключенными с причиной, а не прячутся: иначе
 * пользователь искал бы нужную задачу и не понимал, куда она делась.
 * `lateAfter` помечает задачи, которые заканчиваются позже этой даты: связь
 * «окончание к началу» с ними сдвинет веху — это стоит увидеть до создания связи.
 */
export function TaskMultiPicker({
  label,
  tasks,
  selected,
  onChange,
  disabledTasks = [],
  disabledReason,
  lateAfter,
  emptyText,
}: {
  /** Подпись группы для читалок экрана. */
  label: string;
  tasks: Task[];
  selected: string[];
  /**
   * Сеттер состояния родителя. Именно сеттер, а не `(ids) => void`: переключение
   * считается от текущего значения, иначе два быстрых клика подряд строили бы
   * список от одного и того же устаревшего `selected`, и первый терялся бы.
   */
  onChange: Dispatch<SetStateAction<string[]>>;
  disabledTasks?: Task[];
  disabledReason?: string;
  /** Дата вехи (ISO): задачи, которые кончаются позже, получают предупреждение. */
  lateAfter?: string;
  emptyText: string;
}) {
  const [query, setQuery] = useState("");

  const total = tasks.length + disabledTasks.length;
  const normalized = query.trim().toLowerCase();
  const matches = (task: Task) =>
    !normalized ||
    task.title.toLowerCase().includes(normalized) ||
    task.wbsNumber.includes(normalized) ||
    task.code.toLowerCase().includes(normalized);
  const rows = [
    ...tasks.map((task) => ({ task, disabled: false })),
    ...disabledTasks.map((task) => ({ task, disabled: true })),
  ].filter(({ task }) => matches(task));

  const toggle = (taskId: string, checked: boolean) =>
    onChange((current) =>
      checked
        ? current.includes(taskId)
          ? current
          : [...current, taskId]
        : current.filter((id) => id !== taskId),
    );

  if (total === 0) {
    return (
      <p className="rounded-control border border-line px-3 py-6 text-center text-13 text-ink-muted">
        {emptyText}
      </p>
    );
  }

  return (
    <div className="space-y-2">
      {total > FILTER_THRESHOLD ? (
        <Input
          icon={<Search />}
          value={query}
          onChange={(event) => setQuery(event.target.value)}
          placeholder="Найти по названию или номеру"
          aria-label={`Фильтр списка «${label}»`}
        />
      ) : null}

      <ul
        role="group"
        aria-label={label}
        className="max-h-64 space-y-0.5 overflow-y-auto rounded-control border border-line p-1.5"
      >
        {rows.length === 0 ? (
          <li className="px-2 py-4 text-center text-13 text-ink-muted">Ничего не найдено</li>
        ) : (
          rows.map(({ task, disabled }) => {
            const late = !disabled && lateAfter !== undefined && task.endDate > lateAfter;
            return (
              <li key={task.id}>
                <label
                  className={cn(
                    "flex items-start gap-2.5 rounded-control px-2 py-1.5",
                    disabled ? "cursor-not-allowed opacity-60" : "cursor-pointer hover:bg-surface-muted",
                  )}
                >
                  <Checkbox
                    className="mt-0.5"
                    disabled={disabled}
                    checked={selected.includes(task.id)}
                    onCheckedChange={(next) => toggle(task.id, next === true)}
                  />
                  <span className="min-w-0 flex-1">
                    <span className="flex items-center gap-1.5 text-13">
                      {task.isMilestone ? (
                        <span aria-hidden className="size-2 shrink-0 rotate-45 rounded-[1px] bg-brand" />
                      ) : null}
                      <span
                        className={cn("truncate", task.isMilestone ? "font-semibold text-brand" : "text-ink")}
                      >
                        <span className="font-mono font-normal text-ink-faint">#{task.wbsNumber}</span>{" "}
                        {task.isMilestone ? <span className="sr-only">Веха: </span> : null}
                        {task.title}
                      </span>
                    </span>
                    {disabled && disabledReason ? (
                      <span className="block text-2xs text-ink-faint">{disabledReason}</span>
                    ) : null}
                    {late ? (
                      <span className="block text-2xs text-warning-ink">
                        заканчивается позже вехи — веха сдвинется
                      </span>
                    ) : null}
                  </span>
                  <span className="shrink-0 pt-px font-mono text-2xs text-ink-faint">
                    {task.isMilestone
                      ? formatDayMonth(task.startDate)
                      : `${formatDayMonth(task.startDate)}–${formatDayMonth(task.endDate)}`}
                  </span>
                </label>
              </li>
            );
          })
        )}
      </ul>

      <p className="text-xs text-ink-faint">
        {selected.length > 0
          ? `Выбрано ${selected.length} ${plural(selected.length, ["задача", "задачи", "задач"])}`
          : "Ничего не выбрано"}
      </p>
    </div>
  );
}
