"use client";

import { ChevronLeft, ChevronRight } from "lucide-react";
import { DayPicker } from "react-day-picker";
import { ru } from "react-day-picker/locale";

import { cn } from "@/lib/utils";

/**
 * Календарь в оформлении дизайн-системы.
 *
 * Стили задаются через `classNames`, а не подключением `react-day-picker/style.css`:
 * готовая таблица стилей принесла бы собственную палитру и радиусы, которые
 * разошлись бы с макетом.
 */
export function Calendar({
  selected,
  onSelect,
  defaultMonth,
  disabled,
  className,
}: {
  selected?: Date;
  onSelect: (date: Date | undefined) => void;
  defaultMonth?: Date;
  disabled?: (date: Date) => boolean;
  className?: string;
}) {
  return (
    <DayPicker
      mode="single"
      locale={ru}
      weekStartsOn={1}
      showOutsideDays
      selected={selected}
      onSelect={onSelect}
      defaultMonth={defaultMonth ?? selected}
      disabled={disabled}
      components={{
        PreviousMonthButton: (props) => (
          <button {...props} type="button">
            <ChevronLeft className="size-4" />
          </button>
        ),
        NextMonthButton: (props) => (
          <button {...props} type="button">
            <ChevronRight className="size-4" />
          </button>
        ),
      }}
      className={cn("p-3", className)}
      classNames={{
        months: "relative",
        month: "space-y-3",
        month_caption: "flex h-8 items-center justify-center",
        caption_label: "text-13 font-semibold text-ink capitalize",
        nav: "absolute inset-x-0 top-0 z-1 flex h-8 items-center justify-between",
        button_previous:
          "flex size-7 cursor-pointer items-center justify-center rounded-control text-ink-muted transition-colors hover:bg-surface-muted hover:text-ink disabled:pointer-events-none disabled:cursor-not-allowed disabled:opacity-40 focus-visible:focus-ring",
        button_next:
          "flex size-7 cursor-pointer items-center justify-center rounded-control text-ink-muted transition-colors hover:bg-surface-muted hover:text-ink disabled:pointer-events-none disabled:cursor-not-allowed disabled:opacity-40 focus-visible:focus-ring",
        month_grid: "w-full border-collapse",
        weekdays: "flex",
        weekday:
          "w-8 text-2xs font-semibold uppercase text-ink-faint",
        week: "mt-1 flex",
        day: "p-0",
        day_button:
          "flex size-8 cursor-pointer items-center justify-center rounded-control text-13 text-ink transition-colors hover:bg-surface-muted focus-visible:focus-ring",
        selected: "[&>button]:bg-brand [&>button]:text-white [&>button]:hover:bg-brand-hover",
        today: "[&>button]:font-bold [&>button]:text-brand",
        outside: "[&>button]:text-ink-faint/60",
        disabled: "[&>button]:pointer-events-none [&>button]:opacity-40",
        hidden: "invisible",
      }}
    />
  );
}
