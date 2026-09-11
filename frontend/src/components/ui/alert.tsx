import { AlertCircle, CheckCircle2, Info, TriangleAlert } from "lucide-react";
import type { ComponentType, ReactNode } from "react";

import { cn } from "@/lib/utils";

type AlertTone = "info" | "success" | "warning" | "danger";

// success/warning/danger — текст *-ink, а не сам статусный цвет: на *-tint
// подложке насыщенный цвет не дотягивает до контраста 4.5:1 (WCAG AA).
const toneStyles: Record<AlertTone, string> = {
  info: "bg-brand-tint text-brand-ink",
  success: "bg-success-tint text-success-ink",
  warning: "bg-warning-tint text-warning-ink",
  danger: "bg-danger-tint text-danger-ink",
};

const toneIcons: Record<AlertTone, ComponentType<{ className?: string }>> = {
  info: Info,
  success: CheckCircle2,
  warning: TriangleAlert,
  danger: AlertCircle,
};

/** Компактный блок статуса/ошибки в тональностях дизайн-системы. */
export function Alert({
  tone = "info",
  className,
  children,
}: {
  tone?: AlertTone;
  className?: string;
  children: ReactNode;
}) {
  const Icon = toneIcons[tone];
  return (
    <div
      role={tone === "danger" ? "alert" : "status"}
      className={cn(
        "flex items-start gap-2 rounded-control px-3 py-2 text-13",
        toneStyles[tone],
        className,
      )}
    >
      <Icon className="mt-px size-4 shrink-0" />
      <span className="min-w-0">{children}</span>
    </div>
  );
}

export type { AlertTone };
