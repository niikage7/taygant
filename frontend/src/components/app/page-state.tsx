"use client";

import { CircleAlert, Inbox, ListTodo, LoaderCircle, PlugZap } from "lucide-react";
import Link from "next/link";
import type { ComponentType, ReactNode } from "react";

import { Button } from "@/components/ui/button";
import { Card, CardBody } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { toUserMessage } from "@/lib/api-error-message";
import { ApiError } from "@/services/http-client";

/**
 * Общее пустое состояние: иконка, заголовок, короткое пояснение и опциональное
 * действие. `EmptyProjects` и другие пустые состояния экранов строятся на нём,
 * чтобы не изобретать разметку заново на каждой странице.
 */
export function EmptyState({
  icon: Icon,
  title,
  description,
  action,
}: {
  icon: ComponentType<{ className?: string }>;
  title: ReactNode;
  description?: ReactNode;
  action?: ReactNode;
}) {
  return (
    <Card>
      <CardBody className="flex flex-col items-center gap-3 py-16 text-center">
        <Icon className="size-6 text-ink-subtle" />
        <p className="text-sm font-semibold text-ink">{title}</p>
        {description ? <p className="max-w-md text-13 text-ink-muted">{description}</p> : null}
        {action ? <div className="mt-1">{action}</div> : null}
      </CardBody>
    </Card>
  );
}

/**
 * Скелет загрузки страницы: резервирует место под заголовок и пару карточек,
 * чтобы контент не «прыгал», когда данные приходят. Видимого текста нет — он
 * не нужен зрячему пользователю и путался бы с реальным контентом; `label`
 * озвучивается только скринридером через `aria-label`.
 */
export function PageLoading({ label = "Загружаем данные" }: { label?: string }) {
  return (
    <div role="status" aria-label={label} className="space-y-4">
      <Skeleton className="h-7 w-48" />
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
        {[0, 1].map((key) => (
          <Card key={key}>
            <CardBody className="space-y-3 py-6">
              <Skeleton className="h-3.5 w-24" />
              <Skeleton className="h-7 w-32" />
              <Skeleton className="h-3 w-full" />
            </CardBody>
          </Card>
        ))}
      </div>
    </div>
  );
}

/**
 * Текстовый вариант загрузки для проверки сессии в `(app)/layout.tsx`: до её
 * завершения ещё нет ни заголовка, ни карточек, под которые резервировать
 * место — скелет из `PageLoading` смотрелся бы как случайные серые блоки.
 */
export function SessionCheckLoading({ label = "Проверяем сессию…" }: { label?: string }) {
  return (
    <div role="status" className="flex items-center justify-center gap-3 py-16 text-13 text-ink-muted">
      <LoaderCircle className="size-4 animate-spin text-brand motion-reduce:animate-none" aria-hidden />
      {label}
    </div>
  );
}

/**
 * Ошибка загрузки. 501 показываем отдельным текстом: это не поломка, а ещё не
 * реализованный на бэкенде эндпоинт, и пользователю полезно видеть разницу.
 * `onRetry` необязателен и скрыт для 501 — повторный запрос там не поможет.
 */
export function PageError({ error, onRetry }: { error: unknown; onRetry?: () => void }) {
  const notImplemented =
    error instanceof ApiError && (error.status === 501 || error.status === 404);

  return (
    <Card>
      <CardBody role="alert" className="flex flex-col items-center gap-3 py-16 text-center">
        {notImplemented ? (
          <PlugZap className="size-6 text-warning" />
        ) : (
          <CircleAlert className="size-6 text-danger" />
        )}
        <p className="text-sm font-semibold text-ink">
          {notImplemented
            ? "Раздел ещё не реализован на сервере"
            : "Не удалось загрузить данные"}
        </p>
        <p className="max-w-md text-13 text-ink-muted">
          {toUserMessage(error, {}, "Неизвестная ошибка")}
        </p>
        {onRetry && !notImplemented ? (
          <Button variant="secondary" size="sm" onClick={onRetry} className="mt-1">
            Повторить
          </Button>
        ) : null}
      </CardBody>
    </Card>
  );
}

/** Пустое состояние: в проекте ещё нет задач. */
export function EmptyProjectTasks() {
  return (
    <EmptyState
      icon={ListTodo}
      title="Задач пока нет"
      description="Добавьте первую задачу — календарно-сетевой график появится автоматически."
    />
  );
}

/** Пустое состояние: у пользователя ещё нет проектов. */
export function EmptyProjects() {
  return (
    <EmptyState
      icon={Inbox}
      title="Проектов пока нет"
      description="Создайте первый проект — мастер инициации сформирует календарно-сетевой график и структуру задач."
      action={
        <Button asChild>
          <Link href="/projects/new">Создать проект</Link>
        </Button>
      }
    />
  );
}
