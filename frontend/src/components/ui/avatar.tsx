import { cn } from "@/lib/utils";

/** Инициалы из ФИО: «Елена Васильева» → «ЕВ». */
export function initialsOf(fullName: string): string {
  return fullName
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0]?.toUpperCase() ?? "")
    .join("");
}

/**
 * Цвет плашки выводится из имени, а не хранится: одинаковый человек всегда
 * получает один и тот же цвет, а новый участник не требует правок в коде.
 */
const TONES = [
  "bg-brand-soft text-brand",
  "bg-accent-tint text-accent",
  "bg-success-tint text-success",
  "bg-warning-tint text-warning-ink",
  "bg-danger-chip text-danger",
];

function toneOf(seed: string): string {
  let hash = 0;
  for (const char of seed) hash = (hash * 31 + char.charCodeAt(0)) % 997;
  return TONES[hash % TONES.length];
}

export function Avatar({
  fullName,
  className,
  title,
}: {
  fullName: string;
  className?: string;
  title?: string;
}) {
  return (
    <span
      title={title ?? fullName}
      aria-hidden
      className={cn(
        "inline-flex size-6 shrink-0 items-center justify-center rounded-control text-3xs font-bold",
        toneOf(fullName),
        className,
      )}
    >
      {initialsOf(fullName)}
    </span>
  );
}
