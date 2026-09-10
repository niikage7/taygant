import type { ComponentProps } from "react";

import { cn } from "@/lib/utils";

export function Textarea({ className, ...props }: ComponentProps<"textarea">) {
  return (
    <textarea
      data-slot="textarea"
      className={cn(
        "w-full rounded-control bg-surface-muted px-3 py-2 text-sm leading-relaxed text-ink",
        "ring-1 ring-transparent outline-none transition-[box-shadow,background-color]",
        "placeholder:text-ink-faint focus:bg-surface focus:ring-brand",
        className,
      )}
      {...props}
    />
  );
}
