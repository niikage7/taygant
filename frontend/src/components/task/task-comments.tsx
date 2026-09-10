"use client";

import { formatDistanceToNow, parseISO } from "date-fns";
import { ru } from "date-fns/locale";
import { SendHorizontal } from "lucide-react";
import { useState } from "react";

import { PageError } from "@/components/app/page-state";
import { Alert } from "@/components/ui/alert";
import { Avatar } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import { Card, CardBody } from "@/components/ui/card";
import { Textarea } from "@/components/ui/textarea";
import { useAddComment, useTaskComments } from "@/data/queries";
import { toUserMessage } from "@/lib/api-error-message";

const MAX_LENGTH = 2000;

export function TaskComments({ taskId }: { taskId: string }) {
  const comments = useTaskComments(taskId);
  const addComment = useAddComment(taskId);
  const [text, setText] = useState("");

  const trimmed = text.trim();
  const canSend = trimmed.length > 0 && !addComment.isPending;

  function send() {
    if (!canSend) return;
    addComment.mutate(trimmed, { onSuccess: () => setText("") });
  }

  if (comments.error) return <PageError error={comments.error} />;

  return (
    <Card>
      <CardBody className="space-y-4">
        <form
          onSubmit={(event) => {
            event.preventDefault();
            send();
          }}
          className="space-y-2"
        >
          <Textarea
            rows={3}
            value={text}
            maxLength={MAX_LENGTH}
            onChange={(event) => setText(event.target.value)}
            placeholder="Написать комментарий команде…"
            aria-label="Текст комментария"
            onKeyDown={(event) => {
              // Ctrl/Cmd+Enter — привычная отправка для многострочного поля;
              // просто Enter оставляем за переносом строки.
              if ((event.metaKey || event.ctrlKey) && event.key === "Enter") {
                event.preventDefault();
                send();
              }
            }}
          />
          <div className="flex items-center justify-between gap-3">
            <span className="text-xs text-ink-faint">Ctrl + Enter — отправить</span>
            <Button type="submit" size="sm" disabled={!canSend}>
              <SendHorizontal />
              {addComment.isPending ? "Отправляем…" : "Отправить"}
            </Button>
          </div>
        </form>

        {addComment.isError ? (
          <Alert tone="danger">
            {toUserMessage(
              addComment.error,
              { 403: "Недостаточно прав для комментирования" },
              "Не удалось отправить комментарий",
            )}
          </Alert>
        ) : null}

        {comments.isPending ? (
          <p className="py-6 text-center text-[13px] text-ink-muted">Загружаем комментарии…</p>
        ) : comments.data.length === 0 ? (
          <p className="py-6 text-center text-[13px] text-ink-muted">
            Комментариев пока нет — напишите первый.
          </p>
        ) : (
          <ul className="space-y-3 border-t border-line pt-4">
            {comments.data.map((comment) => (
              <li key={comment.id} className="flex gap-3">
                <Avatar
                  fullName={comment.author.fullName}
                  className="size-8 rounded-full text-[11px]"
                />
                <div className="min-w-0 flex-1">
                  <p className="flex flex-wrap items-baseline gap-2">
                    <span className="text-[13px] font-semibold text-ink">
                      {comment.author.fullName}
                    </span>
                    <span className="text-xs text-ink-faint">
                      {formatDistanceToNow(parseISO(comment.createdAt), {
                        addSuffix: true,
                        locale: ru,
                      })}
                    </span>
                  </p>
                  <p className="mt-1 text-[13px] leading-relaxed whitespace-pre-wrap text-ink-muted">
                    {comment.text}
                  </p>
                </div>
              </li>
            ))}
          </ul>
        )}
      </CardBody>
    </Card>
  );
}
