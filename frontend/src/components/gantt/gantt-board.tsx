"use client";

import { addDays, differenceInCalendarDays, format, parseISO } from "date-fns";
import {
  CalendarClock,
  CircleAlert,
  FilterX,
  Plus,
  TriangleAlert,
  Undo2,
  User,
  Waypoints,
  X,
} from "lucide-react";
import { useRouter } from "next/navigation";
import { useRef, useState } from "react";

import { EmptyState } from "@/components/app/page-state";
import { CreateTaskDialog } from "@/components/gantt/create-task-dialog";
import { buildGanttRows } from "@/lib/gantt-rows";
import { TaskTable } from "@/components/gantt/task-table";
import { Timeline } from "@/components/gantt/timeline";
import { Button } from "@/components/ui/button";
import { Segmented } from "@/components/ui/segmented";
import { useProjectAccess } from "@/data/project-access";
import { useMoveTask, useRestoreTaskDates } from "@/data/queries";
import { toUserMessage } from "@/lib/api-error-message";
import { describeConflict, findDependencyConflicts } from "@/lib/dependency-conflicts";
import { cn } from "@/lib/utils";
import { TASK_TABLE_COMPACT_WIDTH, TASK_TABLE_WIDTH, type TimeScale } from "@/lib/gantt";
import { useMediaQuery } from "@/lib/use-media-query";
import { addWorkdays, nearestWorkday, workdaysAfter } from "@/lib/working-calendar";
import type { Milestone, Project, ShiftSimulation, Task, TaskDependency } from "@/types";

const SCALES = [
  { value: "days", label: "Дни" },
  { value: "weeks", label: "Недели" },
  { value: "months", label: "Месяцы" },
] as const satisfies readonly { value: TimeScale; label: string }[];

type Filter = "mine" | "critical" | "risks";

/** Последнее перетаскивание задачи: данных хватает, чтобы его отменить. */
type PendingMove = {
  task: Task;
  /** Сдвиг в днях, применённый при переносе. */
  deltaDays: number;
  /**
   * Прежние даты всех задач, которые сдвинул перенос: сама задача и, при
   * каскаде, её последователи. Отмена возвращает каждой её даты.
   */
  restore: { taskId: string; startDate: string; endDate: string }[];
  /** Каскадная сводка; null для одиночного переноса без пересчёта цепочки. */
  result: ShiftSimulation | null;
};

/** Перенос, готовый к отправке: даты уже посчитаны по рабочему календарю. */
type PlannedMove = {
  task: Task;
  deltaDays: number;
  startDate: string;
  endDate: string;
  /** apply-shift с пересчётом цепочки (полный доступ) или PATCH одной задачи. */
  cascade: boolean;
  /** Какие связи нарушит перенос — пока список не пуст, ждём подтверждения. */
  warnings: string[];
};

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
  const router = useRouter();
  const [scale, setScale] = useState<TimeScale>("weeks");
  const [filters, setFilters] = useState<Filter[]>([]);
  const [alertVisible, setAlertVisible] = useState(true);
  const [collapsed, setCollapsed] = useState<ReadonlySet<string>>(new Set());
  // Прокрутка — общая для реестра и таймлайна: это один контейнер, скроллящийся
  // в обе стороны. Ссылка нужна таймлайну, чтобы подкрутить график к сегодня.
  const scrollRef = useRef<HTMLDivElement>(null);
  // Уже md реестр сжимается до номера и названия — иначе на телефоне он
  // закрывает таймлайн. Ширина нужна и таблице, и расчёту прокрутки, поэтому
  // решается в JS, а не классами.
  const compact = useMediaQuery("(max-width: 767px)");

  const toggleCollapse = (milestoneId: string) =>
    setCollapsed((current) => {
      const next = new Set(current);
      if (next.has(milestoneId)) next.delete(milestoneId);
      else next.add(milestoneId);
      return next;
    });

  const access = useProjectAccess();
  const moveTask = useMoveTask();
  const restoreDates = useRestoreTaskDates();
  // Итог последнего переноса: показываем, что изменилось, и даём отменить
  // случайное перетаскивание. Только последнее — новое перетаскивание затирает
  // предыдущее, чтобы «Отмена» не откатывала давно ушедший план.
  const [pendingMove, setPendingMove] = useState<PendingMove | null>(null);
  // Перенос, который нарушит связи задачи: без явного согласия не отправляем.
  const [pendingConfirm, setPendingConfirm] = useState<PlannedMove | null>(null);

  const toggleFilter = (filter: Filter) =>
    setFilters((current) =>
      current.includes(filter) ? current.filter((item) => item !== filter) : [...current, filter],
    );

  // При полном доступе перенос идёт через apply-shift: сервер пересчитывает по
  // CPM всю цепочку последователей. Участнику уровня edit остаётся PATCH дат
  // одной задачи — последователей он не двигает, поэтому связи проверяются
  // заранее (см. handleTaskMove).
  const commitMove = ({ task, deltaDays, startDate, endDate, cascade }: PlannedMove) => {
    setPendingConfirm(null);
    moveTask.mutate(
      {
        taskId: task.id,
        shiftDays: deltaDays,
        startDate,
        endDate,
        cascade,
      },
      {
        onSuccess: (response) => {
          const result = response && "affectedTasks" in response ? response : null;
          setPendingMove({
            task,
            deltaDays,
            // Каскад сообщает прежние даты каждой сдвинутой задачи, включая саму
            // перенесённую; одиночный PATCH трогает только её.
            restore: result
              ? result.affectedTasks.map((affected) => ({
                  taskId: affected.taskId,
                  startDate: affected.originalStartDate,
                  endDate: affected.originalEndDate,
                }))
              : [{ taskId: task.id, startDate: task.startDate, endDate: task.endDate }],
            result,
          });
        },
      },
    );
  };

  const calendar = project.workingCalendarType;
  const wbsOf = new Map(tasks.map((task) => [task.id, task.wbsNumber]));

  // Перенос задачи целиком. Начало примагничиваем к ближайшему рабочему дню, а
  // длительность держим в рабочих днях, как каскад на сервере: иначе задача
  // встаёт на выходной, и последователи ждут понедельника. Веха нулевой
  // длительности — точка, её начало и конец совпадают.
  const handleTaskMove = (task: Task, dragDays: number) => {
    const draggedStart = format(addDays(parseISO(task.startDate), dragDays), "yyyy-MM-dd");
    const startDate = nearestWorkday(calendar, draggedStart, dragDays > 0 ? 1 : -1);
    const deltaDays = differenceInCalendarDays(parseISO(startDate), parseISO(task.startDate));
    if (deltaDays === 0) return;
    const endDate = task.isMilestone
      ? startDate
      : addWorkdays(calendar, startDate, workdaysAfter(calendar, task.startDate, task.endDate));
    // Каскад требует и полного доступа, и включённого в настройках проекта
    // автопересчёта — галочка «Включить автоматический пересчёт зависимых
    // задач» из мастера проекта иначе ни на что не влияла бы.
    const cascade = access.isFull && project.autoRecalculateDependents;

    // Каскад сам отодвинет последователей, поэтому при нём сломать можно только
    // связь с предшественниками; PATCH не двигает никого — проверяем обе стороны.
    const warnings = findDependencyConflicts({
      taskId: task.id,
      startDate,
      endDate,
      dependencies,
      tasks,
      calendar,
      checkSuccessors: !cascade,
    }).map((conflict) => describeConflict(conflict, (id) => wbsOf.get(id)));

    const move = { task, deltaDays, startDate, endDate, cascade, warnings };
    setPendingMove(null);
    restoreDates.reset();
    if (warnings.length > 0) setPendingConfirm(move);
    else commitMove(move);
  };

  // Откат последнего переноса: каждой сдвинутой задаче возвращаем её прежние
  // даты. Обратный apply-shift не подходит — последователей назад он не тянет,
  // и после «Отмены» цепочка осталась бы сдвинутой.
  const handleUndoMove = () => {
    if (!pendingMove) return;
    restoreDates.mutate(pendingMove.restore, { onSuccess: () => setPendingMove(null) });
  };

  const hasActiveFilters = filters.length > 0;

  const visibleTasks = tasks.filter((task) => {
    if (filters.includes("mine") && task.assignee?.id !== currentUserId) return false;
    if (filters.includes("critical") && !task.isCriticalPath) return false;
    if (filters.includes("risks") && task.planVsActualDeviationDays >= 0) return false;
    return true;
  });

  // Порядок строк считаем один раз: реестр и таймлайн обязаны совпадать
  // построчно, иначе отрезки уедут относительно названий.
  const allRows = buildGanttRows(visibleTasks, milestones, collapsed);
  // Без фильтров пустая веха — честное отображение реального состояния
  // проекта (в неё правда ещё ничего не привязали). При активном фильтре
  // пустая группа — просто всё, что под ней было, отфильтровано, и держать
  // такой заголовок на экране незачем.
  const rows = hasActiveFilters
    ? allRows.filter((row) => row.kind !== "milestone" || row.childCount > 0)
    : allRows;
  const filteredToEmpty = hasActiveFilters && visibleTasks.length === 0;

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
            onClick={() => router.push(`/tasks/${criticalTask.id}`)}
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

      {moveTask.error || restoreDates.error ? (
        <div className="flex items-center gap-3 rounded-control bg-danger-tint px-3 py-2.5 text-[13px] text-danger">
          <TriangleAlert className="size-4 shrink-0" />
          <p className="min-w-0 flex-1">
            {moveTask.error
              ? toUserMessage(moveTask.error, {}, "Не удалось перенести задачу")
              : toUserMessage(restoreDates.error, {}, "Не удалось отменить перенос")}
          </p>
          <button
            type="button"
            onClick={() => {
              moveTask.reset();
              restoreDates.reset();
            }}
            aria-label="Скрыть сообщение"
            className="shrink-0 rounded-control text-ink-faint hover:text-ink focus-visible:focus-ring"
          >
            <X className="size-4" />
          </button>
        </div>
      ) : null}

      {pendingConfirm ? (
        <div
          role="alert"
          className="flex flex-wrap items-center gap-3 rounded-control bg-warning-tint px-3 py-2.5 text-13 text-warning-ink"
        >
          <TriangleAlert className="size-4 shrink-0 text-warning" />
          <p className="min-w-0 flex-1">
            <span className="font-semibold">«{pendingConfirm.task.title}»:</span>{" "}
            {pendingConfirm.warnings.join("; ")}. Всё равно перенести?
          </p>
          <span className="flex shrink-0 items-center gap-2">
            <Button size="sm" variant="secondary" onClick={() => setPendingConfirm(null)}>
              Не переносить
            </Button>
            <Button size="sm" onClick={() => commitMove(pendingConfirm)}>
              Перенести
            </Button>
          </span>
        </div>
      ) : null}

      {pendingMove ? (
        // Toast поверх диаграммы, вне потока документа: в потоке сводка сдвигала
        // бы график вниз при появлении, и только что отпущенная полоса уезжала
        // бы из-под курсора.
        <div className="fixed bottom-4 left-1/2 z-50 w-[min(640px,calc(100vw-2rem))] -translate-x-1/2">
          <ShiftSummary
            move={pendingMove}
            isUndoing={restoreDates.isPending}
            onUndo={handleUndoMove}
            onClose={() => setPendingMove(null)}
          />
        </div>
      ) : null}

      {filteredToEmpty ? (
        <EmptyState
          icon={FilterX}
          title="Нет задач по выбранным фильтрам"
          description="Попробуйте отключить один из фильтров или сбросить их все."
          action={
            <Button variant="secondary" onClick={() => setFilters([])}>
              Сбросить фильтры
            </Button>
          }
        />
      ) : (
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
              compact={compact}
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
              frozenWidth={compact ? TASK_TABLE_COMPACT_WIDTH : TASK_TABLE_WIDTH}
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
      )}

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

      <div className="flex w-full flex-wrap items-center gap-x-5 gap-y-2 px-1 text-xs text-ink-faint">
        <span className="font-semibold tracking-wider uppercase">Легенда полос:</span>
        <span className="flex items-center gap-1.5">
          <span className="size-2.5 rounded-[1px] bg-success" /> Готово
        </span>
        <span className="flex items-center gap-1.5">
          <span className="size-2.5 rounded-[1px] bg-warning" /> В плане или в работе
        </span>
        <span className="flex items-center gap-1.5">
          <span className="size-2.5 rounded-[1px] bg-danger" /> Критический путь, скоро дедлайн или
          заблокирована
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
 * Что именно сделал перенос и как его отменить. Для каскада показываем, сколько
 * задач сдвинулось следом и как это сказалось на дедлайне проекта: без этого
 * apply-shift меняет чужие сроки незаметно для того, кто подвинул полосу мышью.
 * Для одиночного PATCH-переноса хватает факта сдвига и кнопки отмены.
 */
function ShiftSummary({
  move,
  isUndoing,
  onUndo,
  onClose,
}: {
  move: PendingMove;
  isUndoing: boolean;
  onUndo: () => void;
  onClose: () => void;
}) {
  const { task, result, deltaDays } = move;
  // В affectedTasks сервер кладёт и саму перенесённую задачу — зависимые без неё.
  const affected = result?.affectedTasks.filter((item) => item.taskId !== task.id).length ?? 0;
  // bufferAvailableDays — запас задачи до переноса. Сдвиг вправо его расходует,
  // поэтому показываем остаток; для сдвига влево честнее сказать, каким он был.
  const reserve =
    result && deltaDays > 0
      ? `остаток резерва: ${Math.max(0, result.bufferAvailableDays - deltaDays)} дн.`
      : `резерв до переноса: ${result?.bufferAvailableDays ?? 0} дн.`;
  const deadlineDelta = result?.projectDeadlineImpact.deltaDays ?? 0;
  const tone = deadlineDelta > 0 ? "danger" : "brand";

  return (
    <div
      className={cn(
        "flex items-center gap-3 rounded-control px-3 py-2.5 text-[13px] shadow-popover",
        tone === "danger" ? "bg-danger-tint text-danger" : "bg-brand-tint text-brand",
      )}
    >
      {tone === "danger" ? (
        <TriangleAlert className="size-4 shrink-0" />
      ) : result ? (
        <Waypoints className="size-4 shrink-0" />
      ) : (
        <CalendarClock className="size-4 shrink-0" />
      )}
      <p className="min-w-0 flex-1">
        <span className="font-semibold">«{task.title}»</span> перенесена на{" "}
        {deltaDays > 0 ? `+${deltaDays}` : deltaDays} дн.
        {result ? (
          <>
            {affected > 0
              ? ` Каскадно сдвинуто зависимых задач: ${affected}.`
              : " Зависимые задачи не затронуты."}
            {deadlineDelta !== 0
              ? ` Дедлайн проекта сместился на ${deadlineDelta > 0 ? "+" : ""}${deadlineDelta} дн.`
              : ` Дедлайн проекта не изменился (${reserve}).`}
          </>
        ) : null}
      </p>
      <button
        type="button"
        onClick={onUndo}
        disabled={isUndoing}
        className={cn(
          "flex shrink-0 items-center gap-1.5 rounded-control px-2.5 py-1 text-13 font-semibold text-white transition-colors focus-visible:focus-ring disabled:cursor-not-allowed disabled:opacity-60",
          tone === "danger" ? "bg-danger hover:bg-danger-soft" : "bg-brand hover:bg-brand-hover",
        )}
      >
        <Undo2 className="size-3.5" />
        {isUndoing ? "Отменяем…" : "Отмена"}
      </button>
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
