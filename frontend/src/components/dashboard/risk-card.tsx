/**
 * Активные риски проекта.
 *
 * ⚠️ Пока не подключён: риски приходят из `GET /projects/{id}/dashboard`,
 * который возвращает 501. Компонент готов и ждёт эндпоинт — см. заметку
 * на экране «Обзор и Аналитика».
 */
import { TriangleAlert } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Card, CardBody } from "@/components/ui/card";
import type { ProjectRisk } from "@/types";

/** Карточка активного риска с оценкой влияния на критический путь. */
export function RiskCard({ risk }: { risk: ProjectRisk }) {
  return (
    <Card>
      <CardBody className="flex items-start gap-3">
        <span className="flex size-9 shrink-0 items-center justify-center rounded-control bg-warning-tint text-warning">
          <TriangleAlert className="size-5" />
        </span>
        <div className="min-w-0 flex-1">
          <div className="flex items-start justify-between gap-3">
            <h3 className="text-15 leading-snug font-semibold text-ink">
              {risk.title}
            </h3>
            {risk.severity === "critical" ? (
              <Badge tone="danger" size="sm" className="shrink-0 font-bold tracking-wide uppercase">
                Critical path
              </Badge>
            ) : null}
          </div>
          {risk.description ? (
            <p className="mt-2 text-13 leading-relaxed text-ink-muted">
              {risk.description}
            </p>
          ) : null}
        </div>
      </CardBody>
    </Card>
  );
}
