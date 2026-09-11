import type { ComponentType, ReactNode } from "react";

import { Card, CardBody } from "@/components/ui/card";

/** Секция мастера инициации: номерная иконка, заголовок и произвольное содержимое. */
export function WizardStep({
  icon: Icon,
  title,
  subtitle,
  aside,
  children,
}: {
  icon: ComponentType<{ className?: string }>;
  title: string;
  subtitle: string;
  /** Правый верхний угол секции: ID проекта, кнопка, версия движка. */
  aside?: ReactNode;
  children: ReactNode;
}) {
  return (
    <Card>
      <CardBody className="space-y-4">
        <div className="flex flex-wrap items-start justify-between gap-3">
          <div className="flex items-start gap-3">
            <span className="flex size-9 shrink-0 items-center justify-center rounded-control bg-brand-tint text-brand">
              <Icon className="size-4.5" />
            </span>
            <div>
              <h2 className="text-15 font-semibold text-ink">{title}</h2>
              <p className="mt-0.5 text-13 text-ink-muted">{subtitle}</p>
            </div>
          </div>
          {aside}
        </div>
        {children}
      </CardBody>
    </Card>
  );
}
