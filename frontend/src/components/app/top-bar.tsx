"use client";

import { Menu, Plus, Search, X } from "lucide-react";
import { useRouter } from "next/navigation";
import { useEffect, useRef, useState, useSyncExternalStore, type KeyboardEvent } from "react";

import { ProfileMenu } from "@/components/app/profile-menu";
import { CreateTaskDialog } from "@/components/gantt/create-task-dialog";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { useCurrentProject } from "@/data/current-project";
import { useProjectAccess } from "@/data/project-access";
import { useTasks } from "@/data/queries";
import { TASK_STATUS_META } from "@/lib/task-status";
import { cn } from "@/lib/utils";
import type { Task } from "@/types";

const MAX_RESULTS = 8;

function matchesQuery(task: Task, query: string): boolean {
  return task.code.toLowerCase().includes(query) || task.title.toLowerCase().includes(query);
}

/** Задачи проекта, чей код или название содержат запрос (без учёта регистра), не больше MAX_RESULTS. */
function searchTasks(tasks: Task[] | undefined, query: string): Task[] {
  const trimmed = query.trim().toLowerCase();
  if (!trimmed || !tasks) return [];
  const results: Task[] = [];
  for (const task of tasks) {
    if (results.length >= MAX_RESULTS) break;
    if (matchesQuery(task, trimmed)) results.push(task);
  }
  return results;
}

const noSubscription = () => () => {};

/**
 * Подпись горячей клавиши под платформу macOS/остальные. Платформа известна
 * только в браузере, поэтому читаем её через useSyncExternalStore, а не через
 * эффект с setState: на сервере и при первом клиентском рендере отдаём
 * одинаковое «Ctrl+K», а настоящее значение подставляется сразу после
 * гидратации — без расхождения разметки и без лишнего состояния.
 */
function useShortcutHint(): string {
  return useSyncExternalStore(
    noSubscription,
    () => (/Mac|iPhone|iPad|iPod/.test(navigator.platform || navigator.userAgent) ? "⌘K" : "Ctrl+K"),
    () => "Ctrl+K",
  );
}

/** Верхняя панель: гамбургер-меню (до `lg`), поиск по задачам и быстрое создание задачи. */
export function TopBar({
  onOpenMenu,
  menuOpen,
}: {
  /** Открыть выезжающую панель навигации — кнопка видна только до `lg`. */
  onOpenMenu: () => void;
  menuOpen: boolean;
}) {
  const router = useRouter();
  const { projectId } = useCurrentProject();
  const access = useProjectAccess();
  // Тот же ключ, что и в (app)/layout.tsx, — React Query отдаёт закешированный
  // ответ, лишнего запроса на каждое нажатие клавиши не будет.
  const tasks = useTasks(projectId);

  const shortcutHint = useShortcutHint();
  const inputRef = useRef<HTMLInputElement>(null);

  const [query, setQuery] = useState("");
  const [activeIndex, setActiveIndex] = useState(-1);
  const [focused, setFocused] = useState(false);
  const [mobileSearchOpen, setMobileSearchOpen] = useState(false);

  const results = searchTasks(tasks.data, query);
  const open = focused && query.trim().length > 0;
  const listboxId = "task-search-listbox";

  const focusSearch = () => {
    setMobileSearchOpen(true);
    // Поле на узком экране появляется в разметке только что — фокус ставим
    // следующим кадром, чтобы элемент успел смонтироваться.
    requestAnimationFrame(() => inputRef.current?.focus());
  };

  const closeSearch = () => {
    setQuery("");
    setActiveIndex(-1);
    setMobileSearchOpen(false);
    inputRef.current?.blur();
  };

  const selectTask = (task: Task) => {
    router.push(`/tasks/${task.id}`);
    closeSearch();
  };

  const onInputKeyDown = (event: KeyboardEvent<HTMLInputElement>) => {
    if (event.key === "ArrowDown") {
      event.preventDefault();
      if (results.length > 0) setActiveIndex((current) => Math.min(current + 1, results.length - 1));
    } else if (event.key === "ArrowUp") {
      event.preventDefault();
      if (results.length > 0) setActiveIndex((current) => Math.max(current - 1, 0));
    } else if (event.key === "Enter") {
      const task = results[activeIndex] ?? results[0];
      if (task) {
        event.preventDefault();
        selectTask(task);
      }
    } else if (event.key === "Escape") {
      event.preventDefault();
      closeSearch();
    }
  };

  // ⌘K/Ctrl+K фокусирует поиск с любого места страницы. Эффект только
  // подписывается на событие документа — setState вызывается позже, из
  // обработчика нажатия клавиши, а не синхронно при монтировании.
  useEffect(() => {
    const onKeyDown = (event: globalThis.KeyboardEvent) => {
      const isShortcut = (event.metaKey || event.ctrlKey) && event.key.toLowerCase() === "k";
      if (!isShortcut) return;
      event.preventDefault();
      setMobileSearchOpen(true);
      requestAnimationFrame(() => inputRef.current?.focus());
    };
    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, []);

  const activeOptionId =
    activeIndex >= 0 && results[activeIndex] ? `task-search-option-${results[activeIndex].id}` : undefined;

  return (
    <header className="flex h-14 shrink-0 items-center gap-2 border-b border-line bg-surface px-3 sm:gap-4 sm:px-6">
      <Button
        variant="ghost"
        size="icon"
        className="lg:hidden"
        aria-label="Открыть меню"
        aria-expanded={menuOpen}
        onClick={onOpenMenu}
      >
        <Menu />
      </Button>

      <div className="min-w-0 flex-1 sm:mx-auto sm:max-w-xl">
        <div className={cn("relative", mobileSearchOpen ? "block" : "hidden", "sm:block")}>
          <label className="flex h-8 w-full items-center gap-2 rounded-control border border-line bg-surface-subtle px-3 focus-within:border-brand">
            <Search className="size-4 shrink-0 text-ink-faint" aria-hidden />
            <input
              ref={inputRef}
              role="combobox"
              aria-expanded={open}
              aria-controls={listboxId}
              aria-activedescendant={activeOptionId}
              aria-autocomplete="list"
              aria-label="Поиск задач по коду или названию"
              type="text"
              placeholder="Поиск по задачам..."
              value={query}
              onChange={(event) => {
                setQuery(event.target.value);
                setActiveIndex(-1);
              }}
              onFocus={() => setFocused(true)}
              onBlur={() => setFocused(false)}
              onKeyDown={onInputKeyDown}
              className="min-w-0 flex-1 bg-transparent text-sm text-ink outline-none placeholder:text-ink-faint"
            />
            <button
              type="button"
              onClick={closeSearch}
              aria-label="Свернуть поиск"
              className="flex size-5 shrink-0 cursor-pointer items-center justify-center rounded-control text-ink-faint hover:text-ink sm:hidden"
            >
              <X className="size-4" />
            </button>
            <kbd className="hidden shrink-0 rounded-control border border-line bg-surface px-1.5 py-0.5 font-mono text-3xs text-ink-faint sm:inline-block">
              {shortcutHint}
            </kbd>
          </label>

          {open ? (
            <div
              id={listboxId}
              role="listbox"
              aria-label="Найденные задачи"
              className="absolute top-full left-0 z-40 mt-1 max-h-80 w-full min-w-72 overflow-y-auto rounded-card border border-line bg-surface py-1 shadow-popover"
            >
              {results.length === 0 ? (
                <p className="px-3 py-2 text-13 text-ink-muted">Ничего не найдено</p>
              ) : (
                results.map((task, index) => {
                  const meta = TASK_STATUS_META[task.status];
                  return (
                    <div
                      key={task.id}
                      id={`task-search-option-${task.id}`}
                      role="option"
                      aria-selected={index === activeIndex}
                      onMouseDown={(event) => event.preventDefault()}
                      onMouseEnter={() => setActiveIndex(index)}
                      onClick={() => selectTask(task)}
                      className={cn(
                        "flex cursor-pointer items-center gap-2 px-3 py-2 text-13",
                        index === activeIndex && "bg-surface-muted",
                      )}
                    >
                      <span className="shrink-0 font-mono text-2xs text-ink-muted">{task.code}</span>
                      <span className="min-w-0 flex-1 truncate text-ink">{task.title}</span>
                      <Badge tone={meta.tone} size="sm">
                        {meta.label}
                      </Badge>
                    </div>
                  );
                })
              )}
            </div>
          ) : null}
        </div>

        {!mobileSearchOpen ? (
          <Button
            variant="ghost"
            size="icon"
            className="sm:hidden"
            aria-label="Искать задачи"
            onClick={focusSearch}
          >
            <Search />
          </Button>
        ) : null}
      </div>

      {projectId && access.isFull ? (
        <CreateTaskDialog
          projectId={projectId}
          trigger={
            <Button size="sm" className="shrink-0" data-tour="create-task" aria-label="Создать задачу">
              <Plus />
              <span className="hidden sm:inline">Задача</span>
            </Button>
          }
        />
      ) : null}

      <ProfileMenu />
    </header>
  );
}
