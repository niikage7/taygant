"use client";

import { format, parseISO } from "date-fns";
import { ru } from "date-fns/locale";
import { Bot } from "lucide-react";

import { PageError } from "@/components/app/page-state";
import { UserHoverCard } from "@/components/user/user-card";
import { Avatar } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Card, CardBody } from "@/components/ui/card";
import { useTaskHistory } from "@/data/queries";
import { TASK_STATUS_META } from "@/lib/task-status";
import type { HistoryEntry, TaskStatus } from "@/types";

/** Человекочитаемое название изменённого поля. */
const FIELD_LABELS: Record<string, string> = {
  title: "Название",
  description: "Описание",
  status: "Статус",
  assigneeId: "Ответственный",
  startDate: "Дата начала",
  endDate: "Дата окончания",
  progressPercent: "Прогресс",
  weightPercent: "Вес задачи",
  sprintId: "Спринт",
  milestoneId: "Контрольная точка",
  checklist: "Чек-лист",
  dependency: "Связь",
};

/**
 * Коды действий — из `backend/internal/models/task_history.go`. Правка поля
 * читается как «изменил · Поле», поэтому у них общая подпись.
 */
const ACTION_LABELS: Record<string, string> = {
  created: "создал задачу",
  updated: "изменил",
  status_changed: "изменил",
  assigned: "изменил",
  date_shifted: "изменил",
  progress_set: "изменил",
  dependency_added: "добавил связь",
  dependency_removed: "удалил связь",
  commented: "оставил комментарий",
};

/** Значение поля так, как его видно в приложении: статус — подписью, а не кодом. */
function displayValue(field: string | null, value: string): string {
  if (field === "status" && value in TASK_STATUS_META) {
    return TASK_STATUS_META[value as TaskStatus].label;
  }
  return value;
}

export function TaskHistory({ taskId }: { taskId: string }) {
  const history = useTaskHistory(taskId);

  if (history.error) return <PageError error={history.error} />;

  return (
    <Card>
      <CardBody>
        {history.isPending ? (
          <p className="py-8 text-center text-13 text-ink-muted">Загружаем журнал…</p>
        ) : history.data.length === 0 ? (
          <p className="py-8 text-center text-13 text-ink-muted">
            Изменений пока нет — журнал заполнится после первой правки задачи.
          </p>
        ) : (
          <ol className="space-y-3">
            {history.data.map((entry) => (
              <HistoryRow key={entry.id} entry={entry} />
            ))}
          </ol>
        )}
      </CardBody>
    </Card>
  );
}

function HistoryRow({ entry }: { entry: HistoryEntry }) {
  const fieldLabel = entry.field ? (FIELD_LABELS[entry.field] ?? entry.field) : null;
  const action = ACTION_LABELS[entry.action] ?? entry.action;

  return (
    <li className="flex gap-3">
      <span className="flex flex-col items-center">
        <Avatar fullName={entry.actor.fullName} className="size-7 rounded-full text-3xs" />
        <span className="mt-1 w-px flex-1 bg-line" />
      </span>

      <div className="min-w-0 flex-1 pb-1">
        <p className="flex flex-wrap items-center gap-x-1.5 gap-y-1 text-13 text-ink">
          <span>
            <UserHoverCard user={entry.actor}>
              <span className="font-semibold">{entry.actor.fullName}</span>
            </UserHoverCard>{" "}
            {action}
            {fieldLabel ? <span className="text-ink-muted"> · {fieldLabel}</span> : null}
          </span>
          {entry.source === "assistant" ? <AssistantBadge via={entry.via} /> : null}
        </p>

        {entry.oldValue || entry.newValue ? (
          <p className="mt-1 flex flex-wrap items-center gap-2 text-xs">
            {entry.oldValue ? (
              <span className="rounded-control bg-surface-muted px-2 py-0.5 font-mono text-ink-faint line-through">
                {displayValue(entry.field, entry.oldValue)}
              </span>
            ) : null}
            {entry.newValue ? (
              <span className="rounded-control bg-brand-tint px-2 py-0.5 font-mono text-brand">
                {displayValue(entry.field, entry.newValue)}
              </span>
            ) : null}
          </p>
        ) : null}

        <p className="mt-1 font-mono text-2xs text-ink-faint">
          {format(parseISO(entry.createdAt), "d MMMM yyyy, HH:mm", { locale: ru })}
        </p>
      </div>
    </li>
  );
}

/**
 * Пометка «через ассистента»: правку сделала нейронка автора записи по MCP,
 * с его подтверждения (docs/mcp.md). Подпись ключа показывает, какой именно
 * ассистент это был.
 */
function AssistantBadge({ via }: { via: string | null }) {
  return (
    <Badge tone="accent" size="sm" title="Изменено нейронкой по MCP — с подтверждения автора">
      <Bot className="size-3" />
      через ассистента{via ? ` · ${via}` : ""}
    </Badge>
  );
}
