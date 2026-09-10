"use client";

import { Plus, Search } from "lucide-react";

import { CreateTaskDialog } from "@/components/gantt/create-task-dialog";
import { Button } from "@/components/ui/button";
import { useCurrentProject } from "@/data/current-project";
import { useProjectAccess } from "@/data/project-access";
import { useTasks } from "@/data/queries";

/** Верхняя панель: глобальный поиск и быстрое создание задачи. */
export function TopBar() {
  const { projectId } = useCurrentProject();
  const access = useProjectAccess();
  // Тот же ключ, что и на экране Ганта, — React Query переиспользует ответ.
  const tasks = useTasks(projectId);

  return (
    <header className="flex h-14 shrink-0 items-center gap-4 border-b border-line bg-surface px-6">
      <label className="mx-auto flex h-8 w-full max-w-xl items-center gap-2 rounded-control border border-line bg-surface-subtle px-3">
        <Search className="size-4 shrink-0 text-ink-faint" />
        <input
          type="search"
          placeholder="Поиск по задачам..."
          className="min-w-0 flex-1 bg-transparent text-sm text-ink outline-none placeholder:text-ink-faint"
        />
        <kbd className="shrink-0 rounded-control border border-line bg-surface px-1.5 py-0.5 font-mono text-[10px] text-ink-faint">
          ⌘K
        </kbd>
      </label>

      {projectId && access.isFull ? (
        <CreateTaskDialog
          projectId={projectId}
          projectTasks={tasks.data ?? []}
          trigger={
            <Button size="sm" className="shrink-0">
              <Plus />
              Задача
            </Button>
          }
        />
      ) : null}
    </header>
  );
}
