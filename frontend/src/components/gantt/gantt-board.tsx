"use client";

import { addDays, format, parseISO } from "date-fns";
import { CircleAlert, Plus, TriangleAlert, User, Waypoints, X } from "lucide-react";
import { useRef, useState } from "react";

import { CreateTaskDialog } from "@/components/gantt/create-task-dialog";
import { buildGanttRows } from "@/lib/gantt-rows";
import { TaskTable } from "@/components/gantt/task-table";
import { Timeline } from "@/components/gantt/timeline";
import { Segmented } from "@/components/ui/segmented";
import { useProjectAccess } from "@/data/project-access";
import { useMoveTask } from "@/data/queries";
import { toUserMessage } from "@/lib/api-error-message";
import { cn } from "@/lib/utils";
import { TASK_TABLE_WIDTH, type TimeScale } from "@/lib/gantt";
import type { Milestone, Project, ShiftSimulation, Task, TaskDependency } from "@/types";

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
  milestones,
  currentUserId,
  today,
}: {
  project: Project;
  tasks: Task[];
  dependencies: TaskDependency[];
  milestones: Milestone[];
  currentUserId?: string;
  today: string;
}) {
  const [scale, setScale] = useState<TimeScale>("weeks");
  const [filters, setFilters] = useState<Filter[]>([]);
  const [alertVisible, setAlertVisible] = useState(true);
  const [collapsed, setCollapsed] = useState<ReadonlySet<string>>(new Set());
  // Прокрутка — общая для реестра и таймлайна: это один контейнер, скроллящийся
  // в обе стороны. Ссылка нужна таймлайну, чтобы подкрутить график к сегодня.
  const scrollRef = useRef<HTMLDivElement>(null);

  const toggleCollapse = (milestoneId: string) =>
    setCollapsed((current) => {
      const next = new Set(current);
      if (next.has(milestoneId)) next.delete(milestoneId);
      else next.add(milestoneId);
      return next;
    });

  const access = useProjectAccess();
  const moveTask = useMoveTask();
  // Итог последнего каскадного переноса: сдвиг чужих задач нельзя проводить
  // молча — пользователь должен видеть, что тронул не только свою полосу.
  const [shift, setShift] = useState<{ task: Task; result: ShiftSimulation } | null>(null);

  const toggleFilter = (filter: Filter) =>
    setFilters((current) =>
      current.includes(filter) ? current.filter((item) => item !== filter) : [...current, filter],
    );

  // Перенос задачи целиком сдвигает сроки на delta дней, сохраняя длительность.
  // Веха нулевой длительности — точка, поэтому её начало и конец совпадают.
  //
  // При полном доступе перенос идёт через apply-shift: сервер пересчитывает по
  // CPM всю цепочку последователей. Обычный PATCH дат сдвинул бы одну задачу и
  // оставил её последователя начинаться раньше её конца — молча сломанный план.
  const handleTaskMove = (task: Task, deltaDays: number) => {
    if (deltaDays === 0) return;
    const startDate = format(addDays(parseISO(task.startDate), deltaDays), "yyyy-MM-dd");
    const endDate = task.isMilestone
      ? startDate
      : format(addDays(parseISO(task.endDate), deltaDays), "yyyy-MM-dd");
    setShift(null);
    moveTask.mutate(
      {
        taskId: task.id,
        shiftDays: deltaDays,
        startDate,
        endDate,
        cascade: access.isFull,
      },
      {
        onSuccess: (result) => {
          if (result && "affectedTasks" in result) setShift({ task, result });
        },
      },
    );
  };

  const visibleTasks = tasks.filter((task) => {
    if (filters.includes("mine") && task.assignee?.id !== currentUserId) return false;
    if (filters.includes("critical") && !task.isCriticalPath) return false;
    if (filters.includes("risks") && task.planVsActualDeviationDays >= 0) return false;
    return true;
  });

  // Порядок строк считаем один раз: реестр и таймлайн обязаны совпадать
  // построчно, иначе отрезки уедут относительно названий.
  const rows = buildGanttRows(visibleTasks, milestones, collapsed);

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
      </div>

      {alertVisible && criticalTask ? (
        <div className="flex items-center gap-3 rounded-control bg-danger-tint px-3 py-2.5">
          <TriangleAlert className="size-4 shrink-0 text-danger" />
          <p className="min-w-0 flex-1 truncate text-13 text-danger">
            <span className="font-semibold">Критический путь под угрозой:</span> Задача #
            {criticalTask.wbsNumber} <span className="font-mono">«{criticalTask.title}»</span> имеет
            отставание. Задержка на {Math.abs(criticalTask.planVsActualDeviationDays)} дн.
          </p>
          <button
            type="button"
            className="shrink-0 rounded-control text-13 font-semibold text-brand hover:text-brand-hover focus-visible:focus-ring"
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

      {moveTask.error ? (
        <div className="flex items-center gap-3 rounded-control bg-danger-tint px-3 py-2.5 text-[13px] text-danger">
          <TriangleAlert className="size-4 shrink-0" />
          <p className="min-w-0 flex-1">
            {toUserMessage(moveTask.error, {}, "Не удалось перенести задачу")}
          </p>
          <button
            type="button"
            onClick={() => moveTask.reset()}
            aria-label="Скрыть сообщение"
            className="shrink-0 rounded-control text-ink-faint hover:text-ink focus-visible:focus-ring"
          >
            <X className="size-4" />
          </button>
        </div>
      ) : null}

      {shift ? (
        <ShiftSummary task={shift.task} result={shift.result} onClose={() => setShift(null)} />
      ) : null}

      <div className="overflow-hidden rounded-card bg-surface shadow-card" data-tour="gantt">
        <div
          ref={scrollRef}
          className="flex max-h-[calc(100dvh-20rem)] min-h-[18rem] overflow-auto"
        >
          <TaskTable
            rows={rows}
            allTasks={tasks}
            dependencies={dependencies}
            collapsed={collapsed}
            onToggleCollapse={toggleCollapse}
            highlightCriticalPath={project.highlightCriticalPath}
          />
          <Timeline
            rows={rows}
            dependencies={dependencies}
            milestones={milestones}
            scale={scale}
            startDate={project.startDate}
            endDate={project.deadline}
            today={today}
            highlightCriticalPath={project.highlightCriticalPath}
            canMoveTask={access.canEditTask}
            onTaskMove={handleTaskMove}
            scrollRef={scrollRef}
            frozenWidth={TASK_TABLE_WIDTH}
          />
        </div>
        {access.isFull ? (
          <CreateTaskDialog
            projectId={project.id}
            trigger={
              <button
                type="button"
                className="flex w-full items-center gap-2 border-t border-line px-4 py-3 text-left text-13 text-ink-faint transition-colors hover:bg-surface-subtle hover:text-ink-muted focus-visible:focus-ring"
              >
                <Plus className="size-4" />
                Добавить задачу или веху…
              </button>
            }
          />
        ) : null}{" "}
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
          <span className="size-2.5 rotate-45 rounded-[1px] bg-brand" /> Задача-веха
        </span>
        <span className="flex items-center gap-1.5 text-accent">
          <span className="size-2.5 rotate-45 rounded-[1px] bg-accent" /> Контрольная точка проекта
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
        "flex items-center gap-1.5 rounded-control border px-2.5 py-1.5 text-13 font-medium transition-colors focus-visible:focus-ring",
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

/**
 * Что именно сделал каскадный перенос: сколько задач сдвинулось следом и как
 * это сказалось на дедлайне проекта. Без этого apply-shift меняет чужие сроки
 * незаметно для того, кто просто подвинул полосу мышью.
 */
function ShiftSummary({
  task,
  result,
  onClose,
}: {
  task: Task;
  result: ShiftSimulation;
  onClose: () => void;
}) {
  const affected = result.affectedTasks.length;
  const deadlineDelta = result.projectDeadlineImpact.deltaDays;
  const tone = deadlineDelta > 0 ? "danger" : "brand";

  return (
    <div
      className={cn(
        "flex items-center gap-3 rounded-control px-3 py-2.5 text-[13px]",
        tone === "danger" ? "bg-danger-tint text-danger" : "bg-brand-tint text-brand",
      )}
    >
      {tone === "danger" ? (
        <TriangleAlert className="size-4 shrink-0" />
      ) : (
        <Waypoints className="size-4 shrink-0" />
      )}
      <p className="min-w-0 flex-1">
        <span className="font-semibold">«{task.title}»</span> перенесена на{" "}
        {result.shiftDays > 0 ? `+${result.shiftDays}` : result.shiftDays} дн.
        {affected > 0
          ? ` Каскадно сдвинуто зависимых задач: ${affected}.`
          : " Зависимые задачи не затронуты."}
        {deadlineDelta !== 0
          ? ` Дедлайн проекта сместился на ${deadlineDelta > 0 ? "+" : ""}${deadlineDelta} дн.`
          : ` Дедлайн проекта не изменился (резерв: ${result.bufferAvailableDays} дн.).`}
      </p>
      <button
        type="button"
        onClick={onClose}
        aria-label="Скрыть сводку переноса"
        className="shrink-0 rounded-control opacity-70 transition-opacity hover:opacity-100 focus-visible:focus-ring"
      >
        <X className="size-4" />
      </button>
    </div>
  );
}
