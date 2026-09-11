"use client";

import * as Popover from "@radix-ui/react-popover";
import { format, isValid, parse } from "date-fns";
import { CalendarDays } from "lucide-react";
import { useState } from "react";

import { Calendar } from "@/components/ui/calendar";
import { Input } from "@/components/ui/input";
import { cn } from "@/lib/utils";

const MASK = "дд.мм.гггг";
const PATTERN = "dd.MM.yyyy";

/** ISO (2025-11-15) → «15.11.2025». Пустая строка, если даты нет. */
function toDisplay(iso: string): string {
  if (!iso) return "";
  const parsed = parse(iso, "yyyy-MM-dd", new Date());
  return isValid(parsed) ? format(parsed, PATTERN) : "";
}

/** «15.11.2025» → ISO. null, если дата неполная или несуществующая. */
function toIso(display: string): string | null {
  if (display.length !== MASK.length) return null;
  const parsed = parse(display, PATTERN, new Date());
  if (!isValid(parsed)) return null;
  // parse принимает 31.02.2025, подставляя март, — сверяем обратным форматированием.
  if (format(parsed, PATTERN) !== display) return null;
  return format(parsed, "yyyy-MM-dd");
}

/** Оставляет только цифры и расставляет точки: 15112025 → 15.11.2025. */
function applyMask(raw: string): string {
  const digits = raw.replace(/\D/g, "").slice(0, 8);
  const parts = [digits.slice(0, 2), digits.slice(2, 4), digits.slice(4, 8)];
  return parts.filter(Boolean).join(".");
}

/**
 * Поле даты: ввод с клавиатуры плюс календарь по кнопке.
 *
 * Нативный `<input type="date">` не подходит: браузер рисует его в своей
 * локали, и у пользователя с английской системой дата выглядела бы как
 * 11/15/2025 независимо от `lang` документа. Поэтому текстовый ввод с маской
 * (быстрее, когда дата известна) и всплывающий календарь (удобнее, когда её
 * надо выбрать глазами) — вместе, а не вместо друг друга.
 */
export function DateInput({
  id,
  value,
  onChange,
  className,
  weekdayHint,
  invalid,
  /** Ограничение снизу: например, дедлайн не раньше даты старта. */
  minDate,
  disabled,
}: {
  id: string;
  /** Значение в ISO — наружу компонент всегда отдаёт этот формат. */
  value: string;
  onChange: (iso: string) => void;
  className?: string;
  /** Короткий день недели справа (Сб, Вс) — подсказка из макета. */
  weekdayHint?: string;
  invalid?: boolean;
  minDate?: string;
  /** Только для чтения: поле и календарь недоступны. */
  disabled?: boolean;
}) {
  const [text, setText] = useState(() => toDisplay(value));
  const [open, setOpen] = useState(false);
  // Значение, на которое рассчитан текущий `text` — сравниваем с ним, а не
  // синхронизируем через эффект: обновление состояния прямо при рендере,
  // без лишнего кадра, и не задевает правило react-hooks/set-state-in-effect.
  const [syncedValue, setSyncedValue] = useState(value);

  // Значение может измениться снаружи (сброс формы, выбор в календаре) — тогда
  // синхронизируем видимый текст, но не мешаем набору внутри поля: если то,
  // что уже введено, само разбирается в новое значение, оставляем текст как есть.
  if (value !== syncedValue) {
    setSyncedValue(value);
    if (toIso(text) !== value) setText(toDisplay(value));
  }

  const incomplete = text.length > 0 && toIso(text) === null;
  const selected = value ? parse(value, "yyyy-MM-dd", new Date()) : undefined;
  const min = minDate ? parse(minDate, "yyyy-MM-dd", new Date()) : undefined;

  return (
    <Popover.Root open={open} onOpenChange={setOpen}>
      <Input
        id={id}
        type="text"
        inputMode="numeric"
        autoComplete="off"
        placeholder={MASK}
        value={text}
        disabled={disabled}
        invalid={invalid || incomplete}
        className={className}
        icon={
          <Popover.Trigger
            type="button"
            disabled={disabled}
            aria-label="Выбрать дату в календаре"
            className="flex cursor-pointer items-center rounded-control text-ink-faint transition-colors hover:text-brand focus-visible:focus-ring data-[state=open]:text-brand disabled:pointer-events-none disabled:cursor-not-allowed disabled:opacity-50"
          >
            <CalendarDays className="size-4" />
          </Popover.Trigger>
        }
        trailing={
          weekdayHint ? (
            <span className="font-mono text-xs text-ink-faint">{weekdayHint}</span>
          ) : undefined
        }
        onChange={(event) => {
          const masked = applyMask(event.target.value);
          setText(masked);
          const iso = toIso(masked);
          if (iso) onChange(iso);
        }}
      />

      <Popover.Portal>
        <Popover.Content
          align="start"
          sideOffset={6}
          className={cn(
            "z-50 rounded-card border border-line bg-surface shadow-popover",
            "outline-none",
          )}
        >
          <Calendar
            selected={selected && isValid(selected) ? selected : undefined}
            defaultMonth={selected && isValid(selected) ? selected : undefined}
            disabled={min ? (date) => date < min : undefined}
            onSelect={(date) => {
              if (!date) return;
              onChange(format(date, "yyyy-MM-dd"));
              setOpen(false);
            }}
          />
        </Popover.Content>
      </Popover.Portal>
    </Popover.Root>
  );
}

export { toDisplay as formatIsoAsRu, toIso as parseRuDate };
