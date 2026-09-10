"use client";

import { CircleAlert, Plus, TriangleAlert, User, Waypoints, X } from "lucide-react";
import { useState } from "react";

import { CreateTaskDialog } from "@/components/gantt/create-task-dialog";
import { TaskTable } from "@/components/gantt/task-table";
import { Timeline } from "@/components/gantt/timeline";
import { Button } from "@/components/ui/button";
import { Segmented } from "@/components/ui/segmented";
import { cn } from "@/lib/utils";
import type { TimeScale } from "@/lib/gantt";
import type { Project, Task, TaskDependency } from "@/types";

const SCALES = [
  { value: "days", label: "Дни" },
  { value: "weeks", label: "Недели" },
  { value: "months", label: "Месяцы" },
] as const satisfies readonly { value: TimeScale; label: string }[];

type Filter = "mine" | "critical" | "risks";

/**
 * Интерактивная часть экрана Ганта: масштаб, фильтры и синхронная пара
 * «реестр задач + таймлайн». Обе половины прокручиваются вертикально вместе,
 * поэтому строки не разъезжаются.
 */
export function GanttBoard({
  project,
  tasks,
  dependencies,
  currentUserId,
  today,
}: {
  project: Project;
  tasks: Task[];
  dependencies: TaskDependency[];
  currentUserId?: string;
  today: string;
}) {
  const [scale, setScale] = useState<TimeScale>("weeks");
  const [filters, setFilters] = useState<Filter[]>([]);
  const [alertVisible, setAlertVisible] = useState(true);

  const toggleFilter = (filter: Filter) =>
    setFilters((current) =>
      current.includes(filter) ? current.filter((item) => item !== filter) : [...current, filter],
    );

  const visibleTasks = tasks.filter((task) => {
    if (filters.includes("mine") && task.assignee?.id !== currentUserId) return false;
    if (filters.includes("critical") && !task.isCriticalPath) return false;
    if (filters.includes("risks") && task.planVsActualDeviationDays >= 0) return false;
    return true;
  });

  const criticalTask = tasks.find(
    (task) => task.isCriticalPath && task.planVsActualDeviationDays < 0,
  );

  return (
    <div className="space-y-3">
      <div className="flex flex-wrap items-center gap-2">
        <Segmented
          options={SCALES}
          value={scale}
          onChange={setScale}
          ariaLabel="Масштаб таймлайна"
        />

        <span className="flex flex-wrap items-center gap-1">
          <FilterChip
            active={filters.includes("mine")}
            onClick={() => toggleFilter("mine")}
            icon={<User className="size-3.5" />}
            label="Мои задачи"
          />
          <FilterChip
            active={filters.includes("critical")}
            onClick={() => toggleFilter("critical")}
            icon={<Waypoints className="size-3.5" />}
            label="Критический путь"
            tone="danger"
          />
          <FilterChip
            active={filters.includes("risks")}
            onClick={() => toggleFilter("risks")}
            icon={<CircleAlert className="size-3.5" />}
            label="Только риски"
            tone="warning"
          />
        </span>

        <span className="ml-auto flex items-center gap-2">
          <CreateTaskDialog
            projectId={project.id}
            projectTasks={tasks}
            trigger={
              <Button>
                <Plus />
                Задача
              </Button>
            }
          />
        </span>
      </div>

      {alertVisible && criticalTask ? (
        <div className="flex items-center gap-3 rounded-control bg-danger-tint px-3 py-2.5">
          <TriangleAlert className="size-4 shrink-0 text-danger" />
          <p className="min-w-0 flex-1 truncate text-[13px] text-danger">
            <span className="font-semibold">Критический путь под угрозой:</span> Задача #
            {criticalTask.wbsNumber} <span className="font-mono">«{criticalTask.title}»</span> имеет
            отставание. Задержка на {Math.abs(criticalTask.planVsActualDeviationDays)} дн.
          </p>
          <button
            type="button"
            className="shrink-0 rounded-control text-[13px] font-semibold text-brand hover:text-brand-hover focus-visible:focus-ring"
          >
            Оптимизировать связи
          </button>
          <button
            type="button"
            onClick={() => setAlertVisible(false)}
            aria-label="Скрыть предупреждение"
            className="shrink-0 rounded-control text-ink-faint hover:text-ink focus-visible:focus-ring"
          >
            <X className="size-4" />
          </button>
        </div>
      ) : null}

      <div className="overflow-hidden rounded-card bg-surface shadow-card">
        <div className="flex max-h-[540px] overflow-y-auto">
          <TaskTable
            tasks={visibleTasks}
            allTasks={tasks}
            dependencies={dependencies}
            highlightCriticalPath={project.highlightCriticalPath}
          />
          <Timeline
            tasks={visibleTasks}
            dependencies={dependencies}
            scale={scale}
            startDate={project.startDate}
            endDate={project.deadline}
            today={today}
            highlightCriticalPath={project.highlightCriticalPath}
          />
        </div>

        <CreateTaskDialog
          projectId={project.id}
          projectTasks={tasks}
          trigger={
            <button
              type="button"
              className="flex w-full items-center gap-2 border-t border-line px-4 py-3 text-left text-[13px] text-ink-faint transition-colors hover:bg-surface-subtle hover:text-ink-muted focus-visible:focus-ring"
            >
              <Plus className="size-4" />
              Добавить задачу или веху…
            </button>
          }
        />
      </div>

      <div className="flex flex-wrap items-center gap-x-5 gap-y-2 px-1 text-xs text-ink-faint">
        <span className="font-semibold tracking-wider uppercase">Легенда связей:</span>
        <span className="flex items-center gap-1.5">
          <span className="h-px w-4 bg-line-strong" /> Обычная (FS)
        </span>
        <span className="flex items-center gap-1.5">
          <span className="h-px w-4 border-t border-dashed border-danger" /> Критическая
        </span>
        <span className="flex items-center gap-1.5 text-brand">
          <span className="size-2.5 rotate-45 rounded-[1px] bg-brand" /> Контрольная точка
        </span>
      </div>
    </div>
  );
}

function FilterChip({
  active,
  onClick,
  icon,
  label,
  tone = "brand",
}: {
  active: boolean;
  onClick: () => void;
  icon: React.ReactNode;
  label: string;
  tone?: "brand" | "danger" | "warning";
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      aria-pressed={active}
      className={cn(
        "flex items-center gap-1.5 rounded-control border px-2.5 py-1.5 text-[13px] font-medium transition-colors focus-visible:focus-ring",
        active
          ? "border-transparent bg-brand-tint text-brand"
          : "border-line bg-surface hover:bg-surface-muted",
        !active && tone === "brand" && "text-ink-muted",
        !active && tone === "danger" && "text-danger",
        !active && tone === "warning" && "text-warning",
      )}
    >
      {icon}
      {label}
    </button>
  );
}
