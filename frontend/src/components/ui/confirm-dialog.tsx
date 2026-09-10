"use client";

import * as Dialog from "@radix-ui/react-dialog";
import { TriangleAlert } from "lucide-react";
import { useState, type ReactNode } from "react";

import { Alert } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { toUserMessage } from "@/lib/api-error-message";

/**
 * Подтверждение необратимого действия.
 *
 * Диалог не закрывается сам при ошибке: пользователь должен увидеть, почему
 * удаление не прошло, а не гадать, случилось оно или нет.
 */
export function ConfirmDialog({
  title,
  description,
  confirmLabel,
  pendingLabel,
  onConfirm,
  isPending,
  error,
  trigger,
}: {
  title: string;
  description: ReactNode;
  confirmLabel: string;
  pendingLabel: string;
  onConfirm: (close: () => void) => void;
  isPending: boolean;
  error?: unknown;
  trigger: ReactNode;
}) {
  const [open, setOpen] = useState(false);

  return (
    <Dialog.Root open={open} onOpenChange={setOpen}>
      <Dialog.Trigger asChild>{trigger}</Dialog.Trigger>

      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-0 z-50 bg-ink/30" />
        <Dialog.Content className="fixed top-1/2 left-1/2 z-50 w-[min(28rem,calc(100vw-2rem))] -translate-x-1/2 -translate-y-1/2 rounded-card bg-surface p-6 shadow-popover">
          <div className="flex items-start gap-3">
            <span className="flex size-9 shrink-0 items-center justify-center rounded-control bg-danger-tint text-danger">
              <TriangleAlert className="size-5" />
            </span>
            <div className="min-w-0">
              <Dialog.Title className="text-[15px] font-semibold text-ink">
                {title}
              </Dialog.Title>
              <Dialog.Description asChild>
                <div className="mt-1 text-[13px] leading-relaxed text-ink-muted">
                  {description}
                </div>
              </Dialog.Description>
            </div>
          </div>

          {error ? (
            <Alert tone="danger" className="mt-4">
              {toUserMessage(
                error,
                { 403: "Недостаточно прав для этого действия" },
                "Не удалось выполнить действие",
              )}
            </Alert>
          ) : null}

          <div className="mt-5 flex justify-end gap-2">
            <Dialog.Close asChild>
              <Button type="button" variant="ghost" disabled={isPending}>
                Отмена
              </Button>
            </Dialog.Close>
            <Button
              variant="danger"
              onClick={() => onConfirm(() => setOpen(false))}
              disabled={isPending}
            >
              {isPending ? pendingLabel : confirmLabel}
            </Button>
          </div>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
