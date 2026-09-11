"use client";

import { differenceInCalendarDays, format, parseISO } from "date-fns";
import { useEffect, useRef, useState } from "react";

import {
  buildRange,
  monthColumns,
  offsetPx,
  spanPx,
  subColumns,
  weekendBands,
  type TimeScale,
} from "@/lib/gantt";
import { cn } from "@/lib/utils";
import type { GanttRow } from "@/lib/gantt-rows";
import type { Milestone, Task, TaskDependency } from "@/types";

export const ROW_HEIGHT = 52;
const HEADER_HEIGHT = 48;
const BAR_HEIGHT = 26;
/** За сколько календарных дней до дедлайна задача считается рисковой. */
const DEADLINE_RISK_DAYS = 3;
/** Минимальная ширина колонки, при которой влезает полная подпись месяца. */
const MONTH_FULL_LABEL_MIN_WIDTH = 104;
/** Минимальная ширина колонки для короткой подписи месяца («авг 26»). */
const MONTH_SHORT_LABEL_MIN_WIDTH = 64;

/**
 * Правая часть диаграммы: шапка календаря, отрезки задач и SVG-стрелки связей.
 *
 * Геометрия считается в пикселях от начала диапазона (см. `lib/gantt.ts`), а не
 * в процентах: связи рисуются одним SVG поверх строк, и общая система координат
 * избавляет от рассинхрона стрелок с отрезками при прокрутке и смене масштаба.
 */
export function Timeline({
  rows,
  dependencies,
  milestones,
  scale,
  startDate,
  endDate,
  today,
  highlightCriticalPath,
  canMoveTask,
  onTaskMove,
  scrollRef,
  frozenWidth,
}: {
  /** Те же строки, что и в реестре: заголовки вех тоже занимают строку. */
  rows: GanttRow[];
  dependencies: TaskDependency[];
  /** Контрольные точки проекта — отдельная сущность от задач с isMilestone. */
  milestones: Milestone[];
  scale: TimeScale;
  startDate: string;
  endDate: string;
  /** «Сегодня» приходит снаружи, чтобы сервер и клиент отрисовали одну дату. */
  today: string;
  highlightCriticalPath: boolean;
  /** Какие задачи участник вправе перетаскивать по таймлайну. */
  canMoveTask: (task: Task) => boolean;
  /** Перенос задачи на delta календарных дней (сохранением длительности). */
  onTaskMove: (task: Task, deltaDays: number) => void;
  /** Общий на пару «реестр + таймлайн» контейнер прокрутки (см. GanttBoard). */
  scrollRef: React.RefObject<HTMLDivElement | null>;
  /** Ширина примороженного слева реестра — её не видно под графиком. */
  frozenWidth: number;
}) {
  const range = buildRange(startDate, endDate, scale);
  const months = monthColumns(range);
  const subs = subColumns(range, scale);
  const weekends = weekendBands(range);
  // Индекс строки нужен и отрезкам, и стрелкам связей: считаем один раз по
  // тем же строкам, что рисует реестр, иначе стрелки уедут от отрезков.
  const taskRows = rows.flatMap((row, index) =>
    row.kind === "task" ? [{ task: row.task, index }] : [],
  );
  const rowIndex = new Map(taskRows.map(({ task, index }) => [task.id, index]));
  const todayLeft = offsetPx(range, today) + range.pxPerDay / 2;
  const bodyHeight = rows.length * ROW_HEIGHT;

  // При открытии показываем окрестность сегодняшнего дня, а не начало графика:
  // проект длится месяцы, и без этого пользователь каждый раз мотал бы вручную.
  // Реестр приморожен слева и перекрывает начало графика, поэтому целимся в
  // треть именно видимой части таймлайна, а не всего контейнера.
  useEffect(() => {
    const container = scrollRef.current;
    if (!container) return;
    const visibleWidth = Math.max(0, container.clientWidth - frozenWidth);
    container.scrollLeft = Math.max(0, todayLeft - visibleWidth / 3);
  }, [scrollRef, frozenWidth, todayLeft, scale]);

  return (
    // Своей прокрутки у таймлайна нет: он часть общего скролл-контейнера, иначе
    // шапка календаря не примораживается, а горизонтальный ползунок уезжает
    // под последнюю строку — до него приходилось домотать весь список задач.
    // flex-1 добирает пустое место, когда график уже контейнера.
    <div className="relative" style={{ flex: `1 0 ${range.width}px` }}>
      <div
        className="sticky top-0 z-10 border-b border-line bg-surface"
        style={{ height: HEADER_HEIGHT }}
      >
        <div className="relative h-6 border-b border-line">
          {months.map((month) => (
            <span
              key={month.key}
              className="absolute top-0 flex h-6 items-center overflow-hidden px-2 text-2xs font-bold tracking-wide whitespace-nowrap text-brand uppercase"
              style={{ left: month.left, width: month.width }}
            >
              {month.width >= MONTH_FULL_LABEL_MIN_WIDTH
                ? month.label
                : month.width >= MONTH_SHORT_LABEL_MIN_WIDTH
                  ? month.shortLabel
                  : ""}
            </span>
          ))}
        </div>
        <div className="relative h-6">
          {subs.map((sub) => (
            <span
              key={sub.key}
              className="absolute top-0 flex h-6 items-center justify-center border-l border-line text-3xs text-ink-faint"
              style={{ left: sub.left, width: sub.width }}
            >
              {sub.label}
            </span>
          ))}
        </div>
      </div>

      <div className="relative" style={{ height: bodyHeight }}>
        {weekends.map((band) => (
          <span
            key={band.key}
            className="absolute top-0 bg-surface-subtle"
            style={{ left: band.left, width: band.width, height: bodyHeight }}
          />
        ))}

        {subs.map((sub) => (
          <span
            key={sub.key}
            className="absolute top-0 border-l border-line/70"
            style={{ left: sub.left, height: bodyHeight }}
          />
        ))}

        {rows.map((row, index) => (
          <span
            key={index}
            className={cn(
              "absolute left-0 w-full border-b border-line",
              row.kind === "milestone" && "bg-accent-tint/40",
            )}
            style={{
              top: row.kind === "milestone" ? index * ROW_HEIGHT : (index + 1) * ROW_HEIGHT - 1,
              height: row.kind === "milestone" ? ROW_HEIGHT : undefined,
            }}
          />
        ))}

        <DependencyArrows
          tasks={taskRows.map(({ task }) => task)}
          dependencies={dependencies}
          rowIndex={rowIndex}
          range={range}
          bodyHeight={bodyHeight}
          highlightCriticalPath={highlightCriticalPath}
        />

        {taskRows.map(({ task, index }) => (
          <TaskBar
            key={task.id}
            task={task}
            top={index * ROW_HEIGHT + (ROW_HEIGHT - BAR_HEIGHT) / 2}
            left={offsetPx(range, task.startDate)}
            width={spanPx(range, task.startDate, task.endDate)}
            pxPerDay={range.pxPerDay}
            today={today}
            highlightCriticalPath={highlightCriticalPath}
            movable={canMoveTask(task)}
            onMove={(deltaDays) => onTaskMove(task, deltaDays)}
          />
        ))}

        {milestones.map((milestone) => {
          const left = offsetPx(range, milestone.plannedDate) + range.pxPerDay / 2;
          if (left < 0 || left > range.width) return null;
          const done = milestone.status === "done";
          return (
            <span
              key={milestone.id}
              className="absolute top-0 flex flex-col items-center"
              style={{ left, height: bodyHeight }}
              title={`${milestone.code ? `${milestone.code} · ` : ""}${milestone.name} · ${format(parseISO(milestone.plannedDate), "dd.MM.yyyy")}`}
            >
              <span
                className={cn(
                  "h-full w-px border-l border-dashed",
                  done ? "border-success" : "border-accent",
                )}
              />
              <span
                className={cn(
                  "absolute -top-1 size-2.5 rotate-45 rounded-[1px]",
                  done ? "bg-success" : "bg-accent",
                )}
              />
            </span>
          );
        })}

        {differenceInCalendarDays(parseISO(today), range.start) >= 0 ? (
          <span
            className="absolute top-0 w-px bg-brand"
            style={{ left: todayLeft, height: bodyHeight }}
            aria-hidden
          />
        ) : null}
      </div>
    </div>
  );
}

function TaskBar({
  task,
  top,
  left,
  width,
  pxPerDay,
  today,
  highlightCriticalPath,
  movable,
  onMove,
}: {
  task: Task;
  top: number;
  left: number;
  width: number;
  pxPerDay: number;
  today: string;
  highlightCriticalPath: boolean;
  movable: boolean;
  onMove: (deltaDays: number) => void;
}) {
  const [dragDelta, setDragDelta] = useState<number | null>(null);
  const startXRef = useRef(0);

  // Перетаскивание сдвигает весь отрезок на целое число дней. Превью
  // (dragDelta) живёт до pointerup, а запрос уходит один раз — иначе PATCH
  // отправлялся бы на каждое движение мыши.
  const onPointerDown = (event: React.PointerEvent<HTMLSpanElement>) => {
    if (!movable) return;
    event.currentTarget.setPointerCapture(event.pointerId);
    startXRef.current = event.clientX;
    setDragDelta(0);
  };
  const onPointerMove = (event: React.PointerEvent<HTMLSpanElement>) => {
    if (dragDelta === null) return;
    setDragDelta(Math.round((event.clientX - startXRef.current) / pxPerDay));
  };
  const onPointerUp = () => {
    if (dragDelta === null) return;
    const delta = dragDelta;
    setDragDelta(null);
    if (delta !== 0) onMove(delta);
  };

  const offset = (dragDelta ?? 0) * pxPerDay;
  const dragHandlers = movable
    ? {
        onPointerDown,
        onPointerMove,
        onPointerUp,
        onPointerCancel: onPointerUp,
      }
    : {};
  const dragClass = movable ? "cursor-grab touch-none select-none active:cursor-grabbing" : "";

  const critical = highlightCriticalPath && task.isCriticalPath;
  // Мало времени до дедлайна (или уже просрочена) и задача не закрыта — тоже риск.
  const daysLeft = differenceInCalendarDays(parseISO(task.endDate), parseISO(today));
  const nearDeadline = task.status !== "done" && daysLeft <= DEADLINE_RISK_DAYS;

  if (task.isMilestone) {
    return (
      <span
        {...dragHandlers}
        className={`absolute flex items-center gap-2 ${dragClass}`}
        style={{ top, left: left + offset, height: BAR_HEIGHT }}
        title={`${task.title} · ${format(parseISO(task.startDate), "dd.MM.yyyy")}`}
      >
        <span className="size-3.5 rotate-45 rounded-[2px] bg-brand" />
        <span className="text-2xs font-semibold whitespace-nowrap text-brand">{task.title}</span>
      </span>
    );
  }

  return (
    <span
      {...dragHandlers}
      className={cn(
        "absolute flex items-center overflow-hidden rounded-control",
        critical || nearDeadline
          ? "bg-danger"
          : task.status === "done"
            ? "bg-success"
            : "bg-warning",
        dragClass,
      )}
      style={{ top, left: left + offset, width, height: BAR_HEIGHT }}
      title={`${task.title} · ${task.progressPercent}%`}
    >
      <span
        className="absolute inset-y-0 left-0 bg-black/15"
        style={{ width: `${task.progressPercent}%` }}
      />
      {width > 60 ? (
        <span className="relative z-1 truncate px-2 text-2xs font-semibold text-white">
          {task.progressPercent}% · {task.title}
        </span>
      ) : null}
      {critical && width > pxPerDay * 3 ? (
        <span className="absolute inset-0 rounded-control ring-1 ring-danger ring-offset-1 ring-offset-surface" />
      ) : null}
    </span>
  );
}

/** Стрелки Finish-to-Start и прочих связей между отрезками. */
function DependencyArrows({
  tasks,
  dependencies,
  rowIndex,
  range,
  bodyHeight,
  highlightCriticalPath,
}: {
  tasks: Task[];
  dependencies: TaskDependency[];
  rowIndex: Map<string, number>;
  range: ReturnType<typeof buildRange>;
  bodyHeight: number;
  highlightCriticalPath: boolean;
}) {
  const byId = new Map(tasks.map((task) => [task.id, task]));

  return (
    <svg
      className="pointer-events-none absolute top-0 left-0 overflow-visible"
      width={range.width}
      height={bodyHeight}
      aria-hidden
    >
      {dependencies.map((dependency) => {
        const from = byId.get(dependency.predecessorTaskId);
        const to = byId.get(dependency.successorTaskId);
        const fromRow = rowIndex.get(dependency.predecessorTaskId);
        const toRow = rowIndex.get(dependency.successorTaskId);
        if (!from || !to || fromRow === undefined || toRow === undefined) return null;

        const x1 = offsetPx(range, from.startDate) + spanPx(range, from.startDate, from.endDate);
        const y1 = fromRow * ROW_HEIGHT + ROW_HEIGHT / 2;
        const x2 = offsetPx(range, to.startDate);
        const y2 = toRow * ROW_HEIGHT + ROW_HEIGHT / 2;

        // Обходим отрезок-последователь слева, если он начинается раньше конца
        // предшественника — иначе стрелка прошла бы сквозь него.
        const elbow = Math.max(x1 + 10, x2 - 12);
        const path = `M ${x1} ${y1} H ${elbow} V ${y2} H ${x2}`;
        const critical = highlightCriticalPath && dependency.isCritical;

        return (
          <g key={dependency.id}>
            <path
              d={path}
              fill="none"
              stroke={critical ? "var(--color-danger)" : "var(--color-line-strong)"}
              strokeWidth={critical ? 1.5 : 1}
              strokeDasharray={critical ? "4 3" : undefined}
            />
            <path
              d={`M ${x2} ${y2} l -5 -3.5 v 7 z`}
              fill={critical ? "var(--color-danger)" : "var(--color-line-strong)"}
            />
          </g>
        );
      })}
    </svg>
  );
}
