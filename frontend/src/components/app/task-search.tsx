"use client";

import { Search } from "lucide-react";
import { useRouter } from "next/navigation";
import { useEffect, useMemo, useRef, useState } from "react";

import { Avatar } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { TASK_STATUS_META } from "@/lib/task-status";
import { cn } from "@/lib/utils";
import type { Task } from "@/types";

/** Сколько совпадений показываем в выпадающем списке. */
const MAX_RESULTS = 8;

/**
 * Поиск по задачам текущего проекта в верхней панели.
 *
 * Фильтрация клиентская: полный список задач проекта уже загружен оболочкой
 * (`useTasks` в `(app)/layout.tsx`) и приходит сюда пропом, поэтому каждый
 * ввод не порождает запрос. Совпадение — по названию или коду задачи.
 *
 * Клавиши: ⌘K/Ctrl+K — фокус на поле, ↑/↓ — навигация, Enter — открыть
 * карточку, Esc — закрыть список.
 */
export function TaskSearch({ tasks }: { tasks: Task[] }) {
  const router = useRouter();
  const [query, setQuery] = useState("");
  const [open, setOpen] = useState(false);
  const [activeIndex, setActiveIndex] = useState(0);
  const containerRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);

  const results = useMemo(() => {
    const needle = query.trim().toLowerCase();
    if (!needle) return [];
    return tasks
      .filter(
        (task) =>
          task.title.toLowerCase().includes(needle) ||
          task.code.toLowerCase().includes(needle),
      )
      .slice(0, MAX_RESULTS);
  }, [tasks, query]);

  const showResults = open && query.trim().length > 0;
  const activeTask = showResults ? results[activeIndex] : undefined;

  // ⌘K / Ctrl+K фокусирует поиск — бейдж в поле обещает именно это.
  useEffect(() => {
    const onShortcut = (event: KeyboardEvent) => {
      if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === "k") {
        event.preventDefault();
        inputRef.current?.focus();
        setOpen(true);
      }
    };
    document.addEventListener("keydown", onShortcut);
    return () => document.removeEventListener("keydown", onShortcut);
  }, []);

  // Закрываем список по клику вне поля.
  useEffect(() => {
    const onPointerDown = (event: MouseEvent) => {
      if (!containerRef.current?.contains(event.target as Node)) {
        setOpen(false);
      }
    };
    document.addEventListener("mousedown", onPointerDown);
    return () => document.removeEventListener("mousedown", onPointerDown);
  }, []);

  const openTask = (task: Task) => {
    setOpen(false);
    setQuery("");
    router.push(`/tasks/${task.id}`);
  };

  const onKeyDown = (event: React.KeyboardEvent<HTMLInputElement>) => {
    if (event.key === "Escape") {
      setOpen(false);
      return;
    }
    if (!showResults || results.length === 0) return;

    if (event.key === "ArrowDown") {
      event.preventDefault();
      setActiveIndex((index) => (index + 1) % results.length);
    } else if (event.key === "ArrowUp") {
      event.preventDefault();
      setActiveIndex((index) => (index - 1 + results.length) % results.length);
    } else if (event.key === "Enter") {
      const task = results[activeIndex];
      if (task) {
        event.preventDefault();
        openTask(task);
      }
    }
  };

  return (
    <div ref={containerRef} className="relative mx-auto w-full max-w-xl">
      <label className="flex h-8 w-full items-center gap-2 rounded-control border border-line bg-surface-subtle px-3">
        <Search className="size-4 shrink-0 text-ink-faint" />
        <input
          ref={inputRef}
          type="search"
          value={query}
          onChange={(event) => {
            setQuery(event.target.value);
            setActiveIndex(0);
            setOpen(true);
          }}
          onFocus={() => setOpen(true)}
          onKeyDown={onKeyDown}
          placeholder="Поиск по задачам..."
          role="combobox"
          aria-expanded={showResults}
          aria-controls="task-search-results"
          aria-activedescendant={activeTask ? `task-search-option-${activeTask.id}` : undefined}
          aria-autocomplete="list"
          className="min-w-0 flex-1 bg-transparent text-sm text-ink outline-none placeholder:text-ink-faint"
        />
        <kbd className="shrink-0 rounded-control border border-line bg-surface px-1.5 py-0.5 font-mono text-[10px] text-ink-faint">
          ⌘K
        </kbd>
      </label>

      {showResults ? (
        <div
          id="task-search-results"
          role="listbox"
          className="absolute top-full left-0 z-50 mt-1.5 max-h-80 w-full overflow-y-auto rounded-card border border-line bg-surface p-1 shadow-popover"
        >
          {results.length === 0 ? (
            <p className="px-3 py-4 text-center text-[13px] text-ink-faint">
              Ничего не найдено
            </p>
          ) : (
            results.map((task, index) => {
              const status = TASK_STATUS_META[task.status];
              return (
                <button
                  key={task.id}
                  id={`task-search-option-${task.id}`}
                  type="button"
                  role="option"
                  aria-selected={index === activeIndex}
                  onMouseEnter={() => setActiveIndex(index)}
                  onMouseDown={(event) => event.preventDefault()}
                  onClick={() => openTask(task)}
                  className={cn(
                    "flex w-full items-center gap-2 rounded-control px-2 py-2 text-left",
                    index === activeIndex ? "bg-surface-muted" : "hover:bg-surface-subtle",
                  )}
                >
                  <span className="w-16 shrink-0 truncate font-mono text-[11px] text-ink-faint">
                    {task.code}
                  </span>
                  <span className="min-w-0 flex-1 truncate text-[13px] text-ink">
                    {task.title}
                  </span>
                  {task.assignee ? (
                    <Avatar fullName={task.assignee.fullName} className="size-5 text-[9px]" />
                  ) : null}
                  <Badge tone={status.tone} size="sm" dot={task.status !== "planned"}>
                    {status.label}
                  </Badge>
                </button>
              );
            })
          )}
        </div>
      ) : null}
    </div>
  );
}
