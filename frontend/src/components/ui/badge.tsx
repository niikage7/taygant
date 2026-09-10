import { cva, type VariantProps } from "class-variance-authority";
import type { ComponentProps } from "react";

import { cn } from "@/lib/utils";

const badgeVariants = cva(
  "inline-flex items-center gap-1.5 rounded-control whitespace-nowrap font-medium",
  {
    variants: {
      tone: {
        neutral: "bg-surface-muted text-ink-muted",
        brand: "bg-brand-soft text-brand",
        success: "bg-success-tint text-success",
        warning: "bg-warning-tint text-warning",
        danger: "bg-danger-tint text-danger",
        accent: "bg-accent-tint text-accent",
        outline: "border border-line bg-surface text-ink-muted",
      },
      size: {
        sm: "px-1.5 py-0.5 text-[11px] leading-4",
        md: "px-2 py-1 text-xs leading-4",
      },
    },
    defaultVariants: { tone: "neutral", size: "md" },
  },
);

type BadgeProps = ComponentProps<"span"> &
  VariantProps<typeof badgeVariants> & {
    /** Показать точку-индикатор слева (как у статусов задач в макете). */
    dot?: boolean;
  };

export function Badge({ className, tone, size, dot, children, ...props }: BadgeProps) {
  return (
    <span className={cn(badgeVariants({ tone, size }), className)} {...props}>
      {dot ? <span className="size-1.5 shrink-0 rounded-full bg-current" /> : null}
      {children}
    </span>
  );
}

export { badgeVariants };
