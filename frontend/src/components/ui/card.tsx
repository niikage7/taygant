import type { ComponentProps } from "react";

import { cn } from "@/lib/utils";

/** Базовая карточка: белый фон, радиус 8px и мягкая тень — как во всех макетах. */
export function Card({ className, ...props }: ComponentProps<"div">) {
  return (
    <div
      className={cn("rounded-card bg-surface shadow-card", className)}
      {...props}
    />
  );
}

export function CardHeader({ className, ...props }: ComponentProps<"div">) {
  return (
    <div
      className={cn("flex flex-wrap items-center justify-between gap-3 px-4 pt-4", className)}
      {...props}
    />
  );
}

/** Мелкий прописной заголовок секции (ОБЩИЙ ПРОГРЕСС, СТАТУС ЗАДАЧ и т.п.). */
export function CardLabel({ className, ...props }: ComponentProps<"h3">) {
  return (
    <h3
      className={cn(
        "text-2xs font-semibold tracking-wider text-ink-faint uppercase",
        className,
      )}
      {...props}
    />
  );
}

export function CardTitle({ className, ...props }: ComponentProps<"h3">) {
  return (
    <h3
      className={cn("flex items-center gap-2 text-15 font-semibold text-ink", className)}
      {...props}
    />
  );
}

export function CardBody({ className, ...props }: ComponentProps<"div">) {
  return <div className={cn("px-4 py-4", className)} {...props} />;
}

/** Подвал карточки, отделённый линией — строка «Срок релиза», «Буфер проекта». */
export function CardFooter({ className, ...props }: ComponentProps<"div">) {
  return (
    <div
      className={cn(
        "flex flex-wrap items-center justify-between gap-x-3 gap-y-1 border-t border-line px-4 py-3 text-xs",
        className,
      )}
      {...props}
    />
  );
}
