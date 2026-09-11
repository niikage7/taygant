import { ChevronDown } from "lucide-react";
import type { ComponentProps } from "react";

import { cn } from "@/lib/utils";

/**
 * Нативный `<select>` в оформлении дизайн-системы. Радиксовый Select здесь не
 * нужен: в макетах это обычные списки без кастомных строк, а нативный элемент
 * бесплатно даёт мобильный пикер и клавиатуру.
 */
export function Select({ className, children, ...props }: ComponentProps<"select">) {
  return (
    <div className="relative">
      <select
        data-slot="select"
        className={cn(
          "h-9 w-full appearance-none rounded-control bg-surface-muted px-3 pr-8 text-sm text-ink",
          "ring-1 ring-transparent outline-none transition-[box-shadow,background-color]",
          "focus:bg-surface focus:ring-brand",
          "disabled:cursor-not-allowed disabled:opacity-60",
          className,
        )}
        {...props}
      >
        {children}
      </select>
      <ChevronDown className="pointer-events-none absolute top-1/2 right-2.5 size-4 -translate-y-1/2 text-ink-faint" />
    </div>
  );
}
