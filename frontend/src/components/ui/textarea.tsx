import type { ComponentProps } from "react";

import { cn } from "@/lib/utils";

export function Textarea({
  className,
  invalid = false,
  ...props
}: ComponentProps<"textarea"> & {
  /** Поле в состоянии ошибки — подсвечивается красной рамкой, как Input. */
  invalid?: boolean;
}) {
  return (
    <textarea
      data-slot="textarea"
      aria-invalid={invalid || undefined}
      className={cn(
        "w-full rounded-control bg-surface-muted px-3 py-2 text-sm leading-relaxed text-ink",
        "ring-1 ring-transparent outline-none transition-[box-shadow,background-color]",
        "placeholder:text-ink-faint focus:bg-surface focus:ring-brand",
        invalid && "ring-danger focus:ring-danger",
        className,
      )}
      {...props}
    />
  );
}
