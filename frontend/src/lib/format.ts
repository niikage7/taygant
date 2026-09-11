import { format, parseISO } from "date-fns";
import { ru } from "date-fns/locale";

/** «2025-11-07» → «07.11». */
export function formatDayMonth(iso: string): string {
  return format(parseISO(iso), "dd.MM");
}

/** «2025-11-07» → «07.11.2025». */
export function formatDate(iso: string): string {
  return format(parseISO(iso), "dd.MM.yyyy");
}

/** «2025-11-07» → «7 ноя 2025». */
export function formatDateLong(iso: string): string {
  return format(parseISO(iso), "d MMM yyyy", { locale: ru });
}

/** Момент времени: «2025-11-07T09:05:00Z» → «7 ноя 2025, 16:05» (в поясе браузера). */
export function formatDateTime(iso: string): string {
  return format(parseISO(iso), "d MMM yyyy, HH:mm", { locale: ru });
}

/** Короткий день недели: «2025-11-15» → «Сб». */
export function formatWeekday(iso: string): string {
  const day = format(parseISO(iso), "EEEEEE", { locale: ru });
  return day.charAt(0).toUpperCase() + day.slice(1);
}

/** Число со знаком: 4.2 → «+4.2», -3 → «−3». */
export function formatSigned(value: number, fractionDigits = 0): string {
  const text = Math.abs(value).toFixed(fractionDigits);
  if (value > 0) return `+${text}`;
  if (value < 0) return `−${text}`;
  return text;
}

/** Склонение существительного по числу: дней / день / дня. */
export function plural(count: number, forms: [string, string, string]): string {
  const abs = Math.abs(count) % 100;
  const tail = abs % 10;
  if (abs > 10 && abs < 20) return forms[2];
  if (tail > 1 && tail < 5) return forms[1];
  if (tail === 1) return forms[0];
  return forms[2];
}
