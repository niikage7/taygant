import type { ComponentProps } from "react";

import { cn } from "@/lib/utils";

/**
 * Плейсхолдер загрузки: прямоугольник формы будущего контента вместо пустоты
 * или спиннера — так виднее, что именно появится, и меньше «прыжков» разметки.
 */
export function Skeleton({ className, ...props }: ComponentProps<"div">) {
  return (
    <div
      aria-hidden
      className={cn(
        "animate-pulse rounded-control bg-surface-muted motion-reduce:animate-none",
        className,
      )}
      {...props}
    />
  );
}
