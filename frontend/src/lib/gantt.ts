import {
  addDays,
  differenceInCalendarDays,
  eachDayOfInterval,
  endOfWeek,
  format,
  isSameMonth,
  parseISO,
  startOfWeek,
} from "date-fns";
import { ru } from "date-fns/locale";

/** Масштаб таймлайна. */
export type TimeScale = "days" | "weeks" | "months";

/** Ширина одного календарного дня в пикселях для каждого масштаба. */
export const PX_PER_DAY: Record<TimeScale, number> = {
  days: 34,
  weeks: 11,
  months: 4,
};

export type TimelineRange = {
  start: Date;
  end: Date;
  totalDays: number;
  pxPerDay: number;
  width: number;
};

/**
 * Диапазон таймлайна с полями по краям: график начинается с начала недели,
 * чтобы колонки недель не обрывались посередине.
 */
export function buildRange(
  startIso: string,
  endIso: string,
  scale: TimeScale,
  padDays = 7,
): TimelineRange {
  const start = startOfWeek(addDays(parseISO(startIso), -padDays), { weekStartsOn: 1 });
  const end = endOfWeek(addDays(parseISO(endIso), padDays), { weekStartsOn: 1 });
  const totalDays = differenceInCalendarDays(end, start) + 1;
  const pxPerDay = PX_PER_DAY[scale];
  return { start, end, totalDays, pxPerDay, width: totalDays * pxPerDay };
}

/** Смещение даты от начала диапазона в пикселях. */
export function offsetPx(range: TimelineRange, iso: string): number {
  return differenceInCalendarDays(parseISO(iso), range.start) * range.pxPerDay;
}

/** Ширина отрезка задачи в пикселях (минимум один день). */
export function spanPx(range: TimelineRange, startIso: string, endIso: string): number {
  const days = differenceInCalendarDays(parseISO(endIso), parseISO(startIso)) + 1;
  return Math.max(days, 1) * range.pxPerDay;
}

export type TimelineColumn = {
  key: string;
  label: string;
  /** Короткая подпись для узких колонок (например, «авг 26»). */
  shortLabel?: string;
  left: number;
  width: number;
};

/** Верхний ряд шапки — месяцы. */
export function monthColumns(range: TimelineRange): TimelineColumn[] {
  const columns: TimelineColumn[] = [];
  for (const day of eachDayOfInterval({ start: range.start, end: range.end })) {
    const last = columns.at(-1);
    if (last && isSameMonth(parseISO(last.key), day)) {
      last.width += range.pxPerDay;
      continue;
    }
    columns.push({
      key: format(day, "yyyy-MM-dd"),
      label: format(day, "LLLL yyyy", { locale: ru }),
      shortLabel: format(day, "LLL yy", { locale: ru }),
      left: differenceInCalendarDays(day, range.start) * range.pxPerDay,
      width: range.pxPerDay,
    });
  }
  return columns;
}

/** Нижний ряд шапки: недели или дни в зависимости от масштаба. */
export function subColumns(range: TimelineRange, scale: TimeScale): TimelineColumn[] {
  const days = eachDayOfInterval({ start: range.start, end: range.end });

  if (scale === "days") {
    return days.map((day) => ({
      key: format(day, "yyyy-MM-dd"),
      label: format(day, "d"),
      left: differenceInCalendarDays(day, range.start) * range.pxPerDay,
      width: range.pxPerDay,
    }));
  }

  const weeks: TimelineColumn[] = [];
  for (const day of days) {
    const monday = startOfWeek(day, { weekStartsOn: 1 });
    const key = format(monday, "yyyy-MM-dd");
    const last = weeks.at(-1);
    if (last && last.key === key) {
      last.width += range.pxPerDay;
      continue;
    }
    weeks.push({
      key,
      label: `Н${format(monday, "w", { locale: ru })}`,
      left: differenceInCalendarDays(monday, range.start) * range.pxPerDay,
      width: range.pxPerDay,
    });
  }
  return weeks;
}

/** Выходные дни — для затенения фона таймлайна. */
export function weekendBands(range: TimelineRange): TimelineColumn[] {
  return eachDayOfInterval({ start: range.start, end: range.end })
    .filter((day) => [0, 6].includes(day.getDay()))
    .map((day) => ({
      key: format(day, "yyyy-MM-dd"),
      label: "",
      left: differenceInCalendarDays(day, range.start) * range.pxPerDay,
      width: range.pxPerDay,
    }));
}

/**
 * Ширина реестра задач слева от таймлайна.
 *
 * Живёт здесь, а не в компоненте, потому что нужна обоим: реестр задаёт по ней
 * свою ширину, а таймлайн — вычитает её из видимой области, когда подкручивает
 * график к сегодняшнему дню. Импорт друг из друга дал бы цикл.
 */
export const TASK_TABLE_WIDTH = 600;

/**
 * Реестр на телефоне: только номер и название. Полные 600px шире экрана
 * в 375px — приморожённый реестр закрывал бы таймлайн целиком.
 */
export const TASK_TABLE_COMPACT_WIDTH = 200;
