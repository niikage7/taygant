"use client";

import { format, isValid, parse } from "date-fns";
import { CalendarDays } from "lucide-react";
import { useEffect, useState } from "react";

import { Input } from "@/components/ui/input";

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
 * Поле даты в русском формате.
 *
 * Нативный `<input type="date">` не подходит: браузер рисует его в своей
 * локали, и у пользователя с английской системой дата выглядела бы как
 * 11/15/2025 независимо от `lang` документа. Поэтому обычный текстовый ввод с
 * маской — формат тогда одинаковый у всех.
 */
export function DateInput({
  id,
  value,
  onChange,
  className,
  weekdayHint,
  invalid,
}: {
  id: string;
  /** Значение в ISO — наружу компонент всегда отдаёт этот формат. */
  value: string;
  onChange: (iso: string) => void;
  className?: string;
  /** Короткий день недели справа (Сб, Вс) — подсказка из макета. */
  weekdayHint?: string;
  invalid?: boolean;
}) {
  const [text, setText] = useState(() => toDisplay(value));

  // Значение может измениться снаружи (сброс формы, автозаполнение) — тогда
  // синхронизируем видимый текст, но не мешаем набору внутри поля.
  useEffect(() => {
    const next = toDisplay(value);
    setText((current) => (toIso(current) === value ? current : next));
  }, [value]);

  const incomplete = text.length > 0 && toIso(text) === null;

  return (
    <Input
      id={id}
      type="text"
      inputMode="numeric"
      autoComplete="off"
      placeholder={MASK}
      value={text}
      aria-describedby={incomplete ? `${id}-format` : undefined}
      invalid={invalid || incomplete}
      icon={<CalendarDays />}
      className={className}
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
  );
}

export { toDisplay as formatIsoAsRu, toIso as parseRuDate };
