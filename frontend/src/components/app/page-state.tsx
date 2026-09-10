"use client";

import { CircleAlert, Inbox, LoaderCircle, PlugZap } from "lucide-react";
import Link from "next/link";
import type { ReactNode } from "react";

import { Button } from "@/components/ui/button";
import { Card, CardBody } from "@/components/ui/card";
import { toUserMessage } from "@/lib/api-error-message";
import { ApiError } from "@/services/http-client";

/** Скелет на время загрузки экрана. */
export function PageLoading({ label = "Загружаем данные проекта…" }: { label?: string }) {
  return (
    <Card>
      <CardBody className="flex items-center justify-center gap-3 py-16 text-[13px] text-ink-muted">
        <LoaderCircle className="size-4 animate-spin text-brand" />
        {label}
      </CardBody>
    </Card>
  );
}

/**
 * Ошибка загрузки. 501 показываем отдельным текстом: это не поломка, а ещё не
 * реализованный на бэкенде эндпоинт, и пользователю полезно видеть разницу.
 */
export function PageError({ error }: { error: unknown }) {
  const notImplemented =
    error instanceof ApiError && (error.status === 501 || error.status === 404);

  return (
    <Card>
      <CardBody className="flex flex-col items-center gap-3 py-16 text-center">
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
        <p className="max-w-md text-[13px] text-ink-muted">
          {toUserMessage(error, {}, "Неизвестная ошибка")}
        </p>
      </CardBody>
    </Card>
  );
}

/** Пустое состояние: у пользователя ещё нет проектов. */
export function EmptyProjects() {
  return (
    <Card>
      <CardBody className="flex flex-col items-center gap-3 py-16 text-center">
        <Inbox className="size-6 text-ink-faint" />
        <p className="text-sm font-semibold text-ink">Проектов пока нет</p>
        <p className="max-w-md text-[13px] text-ink-muted">
          Создайте первый проект — мастер инициации сформирует календарно-сетевой
          график и структуру задач.
        </p>
        <Button asChild className="mt-1">
          <Link href="/projects/new">Создать проект</Link>
        </Button>
      </CardBody>
    </Card>
  );
}

/**
 * Плашка «данные демонстрационные». Нужна на экранах, где эндпоинт возвращает
 * 501 и содержимое приходит из фикстур: без явной пометки такие цифры легко
 * принять за настоящие метрики проекта.
 */
export function DemoDataNotice({ children }: { children: ReactNode }) {
  return (
    <p className="flex items-start gap-2 rounded-control bg-warning-tint px-3 py-2 text-[13px] text-warning-ink">
      <PlugZap className="mt-px size-4 shrink-0" />
      <span>{children}</span>
    </p>
  );
}
