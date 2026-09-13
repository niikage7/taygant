import { addDays, format, parseISO } from "date-fns";

import type { WorkingCalendarType } from "@/types";

/**
 * Рабочий календарь проекта на клиенте — копия правил бэкенда
 * (`backend/internal/schedule/schedule.go`: `isWorkday`, `addWorkdays`).
 * Расхождение с сервером дало бы предупреждения там, где план цел, и наоборот.
 */

const toIso = (date: Date) => format(date, "yyyy-MM-dd");

/** Воскресенье выходной в обоих календарях, суббота — только в 5/2. */
export function isWorkday(calendar: WorkingCalendarType, iso: string): boolean {
  const weekday = parseISO(iso).getDay();
  if (weekday === 0) return false;
  return !(calendar === "5/2" && weekday === 6);
}

/** Дата через n рабочих дней (n < 0 — назад). Выходные не считаются. */
export function addWorkdays(calendar: WorkingCalendarType, iso: string, n: number): string {
  let date = parseISO(iso);
  const step = n < 0 ? -1 : 1;
  let left = Math.abs(n);
  while (left > 0) {
    date = addDays(date, step);
    if (isWorkday(calendar, toIso(date))) left--;
  }
  return toIso(date);
}

/**
 * Длительность задачи в рабочих днях: сколько их в (начало, конец]. Конец
 * задачи от нового начала — `addWorkdays(начало, workdaysAfter(...))`.
 */
export function workdaysAfter(calendar: WorkingCalendarType, startIso: string, endIso: string): number {
  const end = parseISO(endIso);
  let count = 0;
  for (let date = addDays(parseISO(startIso), 1); date <= end; date = addDays(date, 1)) {
    if (isWorkday(calendar, toIso(date))) count++;
  }
  return count;
}

/** Ближайший рабочий день; при равном расстоянии — в сторону `direction`. */
export function nearestWorkday(calendar: WorkingCalendarType, iso: string, direction: 1 | -1): string {
  if (isWorkday(calendar, iso)) return iso;
  const date = parseISO(iso);
  for (let distance = 1; ; distance++) {
    const toward = toIso(addDays(date, direction * distance));
    if (isWorkday(calendar, toward)) return toward;
    const away = toIso(addDays(date, -direction * distance));
    if (isWorkday(calendar, away)) return away;
  }
}
