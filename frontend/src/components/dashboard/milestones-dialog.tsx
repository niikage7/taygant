"use client";

import * as Dialog from "@radix-ui/react-dialog";
import { CircleCheck, Flag, Info, Pencil, Plus, Trash2, X } from "lucide-react";
import { useState, type ReactNode } from "react";

import { Alert } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { DateInput } from "@/components/ui/date-input";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  useCreateMilestone,
  useDeleteMilestone,
  useMilestones,
  useUpdateMilestone,
} from "@/data/queries";
import { toUserMessage } from "@/lib/api-error-message";
import { formatDate } from "@/lib/format";
import { MILESTONE_STATUS_META } from "@/lib/task-status";
import type { Milestone } from "@/types";

const today = () => new Date().toISOString().slice(0, 10);

/**
 * Управление контрольными точками проекта.
 *
 * Работает по полному списку `GET /projects/{id}/milestones`, а не по
 * `nearestMilestones` из дашборда: тот отфильтрован от завершённых и обрезан до
 * трёх, поэтому созданная веха могла бы в него не попасть и это выглядело бы
 * так, будто она не сохранилась.
 */
export function MilestonesDialog({
  projectId,
  trigger,
}: {
  projectId: string;
  trigger: ReactNode;
}) {
  const [open, setOpen] = useState(false);
  const milestones = useMilestones(projectId);

  return (
    <Dialog.Root open={open} onOpenChange={setOpen}>
      <Dialog.Trigger asChild>{trigger}</Dialog.Trigger>

      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-0 z-50 bg-ink/30" />
        <Dialog.Content className="fixed top-1/2 left-1/2 z-50 max-h-[calc(100vh-2rem)] w-[min(36rem,calc(100vw-2rem))] -translate-x-1/2 -translate-y-1/2 overflow-y-auto rounded-card bg-surface p-6 shadow-popover">
          <div className="flex items-start justify-between gap-4">
            <div>
              <Dialog.Title className="text-[15px] font-semibold text-ink">
                Контрольные точки проекта
              </Dialog.Title>
              <Dialog.Description className="mt-1 text-[13px] text-ink-muted">
                Обязательства проекта по датам. Задачи к ним не привязываются — для
                этого используйте спринты или задачу-веху со связями.
              </Dialog.Description>
            </div>
            <Dialog.Close
              aria-label="Закрыть"
              className="rounded-control p-1 text-ink-faint hover:text-ink focus-visible:focus-ring"
            >
              <X className="size-4" />
            </Dialog.Close>
          </div>

          <div className="mt-5 space-y-4">
            <MilestoneForm projectId={projectId} />

            {/*
              Статус вычисляется на сервере из actualDate и порядка плановых дат
              (deriveMilestoneStatuses). Поставить actualDate через API сейчас
              нельзя, поэтому «Готово» недостижимо — предупреждаем, иначе счётчик
              закрытых вех в обзоре выглядит сломанным.
            */}
            <p className="flex items-start gap-2 rounded-control bg-warning-tint px-3 py-2 text-xs text-warning-ink">
              <Info className="mt-px size-3.5 shrink-0" />
              Отметить веху достигнутой пока нельзя: статус выводится из фактической
              даты, а её приём не реализован на бэкенде. Поэтому в обзоре число
              закрытых вех остаётся нулевым.
            </p>

            {milestones.isPending ? (
              <p className="py-6 text-center text-[13px] text-ink-muted">
                Загружаем вехи…
              </p>
            ) : (milestones.data?.length ?? 0) === 0 ? (
              <p className="py-6 text-center text-[13px] text-ink-muted">
                Контрольных точек пока нет — добавьте первую.
              </p>
            ) : (
              <ul className="space-y-2 border-t border-line pt-4">
                {milestones.data?.map((milestone) => (
                  <MilestoneRow key={milestone.id} milestone={milestone} />
                ))}
              </ul>
            )}
          </div>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}

function MilestoneForm({ projectId }: { projectId: string }) {
  const [code, setCode] = useState("");
  const [name, setName] = useState("");
  const [plannedDate, setPlannedDate] = useState(today);
  const createMilestone = useCreateMilestone(projectId);

  const canSubmit = name.trim().length > 0 && plannedDate && !createMilestone.isPending;

  return (
    <form
      className="space-y-3 rounded-control bg-surface-subtle p-3"
      onSubmit={(event) => {
        event.preventDefault();
        if (!canSubmit) return;
        createMilestone.mutate(
          { code: code.trim() || undefined, name: name.trim(), plannedDate },
          {
            onSuccess: () => {
              setCode("");
              setName("");
              setPlannedDate(today());
            },
          },
        );
      }}
    >
      <div className="grid grid-cols-[6rem_1fr] gap-2">
        <div className="space-y-1.5">
          <Label htmlFor="milestone-code">Код</Label>
          <Input
            id="milestone-code"
            value={code}
            onChange={(event) => setCode(event.target.value)}
            placeholder="КТ-1"
          />
        </div>
        <div className="space-y-1.5">
          <Label htmlFor="milestone-name">
            Название <span className="text-danger">*</span>
          </Label>
          <Input
            id="milestone-name"
            value={name}
            onChange={(event) => setName(event.target.value)}
            placeholder="Например, Готовность MVP"
            required
          />
        </div>
      </div>

      <div className="flex items-end gap-2">
        <div className="flex-1 space-y-1.5">
          <Label htmlFor="milestone-date">Плановая дата</Label>
          <DateInput
            id="milestone-date"
            value={plannedDate}
            onChange={setPlannedDate}
          />
        </div>
        <Button type="submit" disabled={!canSubmit}>
          <Plus />
          {createMilestone.isPending ? "Добавляем…" : "Добавить"}
        </Button>
      </div>

      {createMilestone.isError ? (
        <Alert tone="danger">
          {toUserMessage(
            createMilestone.error,
            { 403: "Недостаточно прав для изменения вех проекта" },
            "Не удалось создать контрольную точку",
          )}
        </Alert>
      ) : null}
    </form>
  );
}

function MilestoneRow({ milestone }: { milestone: Milestone }) {
  const [editing, setEditing] = useState(false);
  const [name, setName] = useState(milestone.name);
  const [plannedDate, setPlannedDate] = useState(milestone.plannedDate);

  const updateMilestone = useUpdateMilestone();
  const deleteMilestone = useDeleteMilestone();
  const status = MILESTONE_STATUS_META[milestone.status];

  if (editing) {
    return (
      <li className="rounded-control border border-brand p-3">
        <form
          className="space-y-3"
          onSubmit={(event) => {
            event.preventDefault();
            updateMilestone.mutate(
              {
                milestoneId: milestone.id,
                payload: { code: milestone.code, name: name.trim(), plannedDate },
              },
              { onSuccess: () => setEditing(false) },
            );
          }}
        >
          <Input
            value={name}
            onChange={(event) => setName(event.target.value)}
            aria-label="Название контрольной точки"
            required
          />
          <div className="flex items-end gap-2">
            <DateInput
              id={`milestone-date-${milestone.id}`}
              value={plannedDate}
              onChange={setPlannedDate}
              className="flex-1"
            />
            <Button type="submit" size="sm" disabled={updateMilestone.isPending}>
              Сохранить
            </Button>
            <Button
              type="button"
              size="sm"
              variant="ghost"
              onClick={() => {
                setName(milestone.name);
                setPlannedDate(milestone.plannedDate);
                setEditing(false);
              }}
            >
              Отмена
            </Button>
          </div>
        </form>
      </li>
    );
  }

  return (
    <li className="flex items-center gap-3 rounded-control border border-line p-3">
      {milestone.status === "done" ? (
        <CircleCheck className="size-4 shrink-0 text-success" />
      ) : (
        <Flag className="size-4 shrink-0 text-ink-faint" />
      )}

      <div className="min-w-0 flex-1">
        <p className="truncate text-[13px] font-medium text-ink">
          {milestone.code ? (
            <span className="font-mono text-ink-faint">{milestone.code} </span>
          ) : null}
          {milestone.name}
        </p>
        <p className="mt-0.5 font-mono text-[11px] text-ink-faint">
          {formatDate(milestone.plannedDate)}
          {milestone.riskDays ? ` · риск +${milestone.riskDays} дн.` : ""}
        </p>
      </div>

      <Badge tone={status.tone} size="sm">
        {status.label}
      </Badge>

      <button
        type="button"
        onClick={() => setEditing(true)}
        aria-label={`Изменить «${milestone.name}»`}
        className="rounded-control p-1 text-ink-faint transition-colors hover:text-brand focus-visible:focus-ring"
      >
        <Pencil className="size-3.5" />
      </button>
      <button
        type="button"
        onClick={() => deleteMilestone.mutate(milestone.id)}
        disabled={deleteMilestone.isPending}
        aria-label={`Удалить «${milestone.name}»`}
        className="rounded-control p-1 text-ink-faint transition-colors hover:text-danger focus-visible:focus-ring disabled:opacity-50"
      >
        <Trash2 className="size-3.5" />
      </button>
    </li>
  );
}
