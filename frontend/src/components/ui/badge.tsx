import { cva, type VariantProps } from "class-variance-authority";
import type { ComponentProps } from "react";

import { cn } from "@/lib/utils";

const badgeVariants = cva(
  "inline-flex items-center gap-1.5 rounded-control whitespace-nowrap font-medium",
  {
    variants: {
      // Тона success/warning/danger используют *-ink текст, а не сам статусный
      // цвет: на *-tint подложке, например, text-success даёt контраст ≈2.6:1 —
      // ниже нормы WCAG AA (4.5:1). Насыщенный цвет остаётся у точки-индикатора.
      tone: {
        neutral: "bg-surface-muted text-ink-muted",
        brand: "bg-brand-soft text-brand",
        success: "bg-success-tint text-success-ink",
        warning: "bg-warning-tint text-warning-ink",
        danger: "bg-danger-tint text-danger-ink",
        accent: "bg-accent-tint text-accent",
        outline: "border border-line bg-surface text-ink-muted",
      },
      size: {
        sm: "px-1.5 py-0.5 text-2xs leading-4",
        md: "px-2 py-1 text-xs leading-4",
      },
    },
    defaultVariants: { tone: "neutral", size: "md" },
  },
);

/** Насыщенный цвет точки-индикатора — независим от *-ink текста тона. */
const dotToneStyles: Record<NonNullable<VariantProps<typeof badgeVariants>["tone"]>, string> = {
  neutral: "bg-ink-muted",
  brand: "bg-brand",
  success: "bg-success",
  warning: "bg-warning",
  danger: "bg-danger",
  accent: "bg-accent",
  outline: "bg-ink-muted",
};

type BadgeProps = ComponentProps<"span"> &
  VariantProps<typeof badgeVariants> & {
    /** Показать точку-индикатор слева (как у статусов задач в макете). */
    dot?: boolean;
  };

export function Badge({ className, tone, size, dot, children, ...props }: BadgeProps) {
  return (
    <span className={cn(badgeVariants({ tone, size }), className)} {...props}>
      {dot ? (
        <span
          className={cn("size-1.5 shrink-0 rounded-full", dotToneStyles[tone ?? "neutral"])}
        />
      ) : null}
      {children}
    </span>
  );
}

export { badgeVariants };
