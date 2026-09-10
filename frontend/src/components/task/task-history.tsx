"use client";

import { format, parseISO } from "date-fns";
import { ru } from "date-fns/locale";

import { PageError } from "@/components/app/page-state";
import { Avatar } from "@/components/ui/avatar";
import { Card, CardBody } from "@/components/ui/card";
import { useTaskHistory } from "@/data/queries";
import type { HistoryEntry } from "@/types";

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
};

const ACTION_LABELS: Record<string, string> = {
  created: "создал задачу",
  updated: "изменил",
  deleted: "удалил задачу",
  dependency_added: "добавил связь",
  dependency_removed: "удалил связь",
  comment_added: "оставил комментарий",
};

export function TaskHistory({ taskId }: { taskId: string }) {
  const history = useTaskHistory(taskId);

  if (history.error) return <PageError error={history.error} />;

  return (
    <Card>
      <CardBody>
        {history.isPending ? (
          <p className="py-8 text-center text-[13px] text-ink-muted">Загружаем журнал…</p>
        ) : history.data.length === 0 ? (
          <p className="py-8 text-center text-[13px] text-ink-muted">
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
        <Avatar fullName={entry.actor.fullName} className="size-7 rounded-full text-[10px]" />
        <span className="mt-1 w-px flex-1 bg-line" />
      </span>

      <div className="min-w-0 flex-1 pb-1">
        <p className="text-[13px] text-ink">
          <span className="font-semibold">{entry.actor.fullName}</span> {action}
          {fieldLabel ? <span className="text-ink-muted"> · {fieldLabel}</span> : null}
        </p>

        {entry.oldValue || entry.newValue ? (
          <p className="mt-1 flex flex-wrap items-center gap-2 text-xs">
            {entry.oldValue ? (
              <span className="rounded-control bg-surface-muted px-2 py-0.5 font-mono text-ink-faint line-through">
                {entry.oldValue}
              </span>
            ) : null}
            {entry.newValue ? (
              <span className="rounded-control bg-brand-tint px-2 py-0.5 font-mono text-brand">
                {entry.newValue}
              </span>
            ) : null}
          </p>
        ) : null}

        <p className="mt-1 font-mono text-[11px] text-ink-faint">
          {format(parseISO(entry.createdAt), "d MMMM yyyy, HH:mm", { locale: ru })}
        </p>
      </div>
    </li>
  );
}
