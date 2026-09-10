"use client";

import type { ComponentProps, ReactNode } from "react";

import { cn } from "@/lib/utils";

type InputProps = ComponentProps<"input"> & {
  /** Иконка слева внутри поля (как «@» и «замок» в макете авторизации). */
  icon?: ReactNode;
  /** Слот справа: кнопка показа пароля, счётчик символов и т.п. */
  trailing?: ReactNode;
  /** Поле в состоянии ошибки — подсвечивается красной рамкой. */
  invalid?: boolean;
};

export function Input({ className, icon, trailing, invalid = false, ...props }: InputProps) {
  return (
    <div
      className={cn(
        "flex h-9 w-full items-center gap-2 rounded-control bg-surface-muted px-3",
        "ring-1 ring-transparent transition-[box-shadow,background-color]",
        "focus-within:bg-surface focus-within:ring-brand",
        invalid && "ring-danger focus-within:ring-danger",
        className,
      )}
    >
      {icon ? (
        <span className="flex shrink-0 items-center text-ink-faint [&_svg]:size-4">{icon}</span>
      ) : null}
      <input
        data-slot="input"
        aria-invalid={invalid || undefined}
        className="min-w-0 flex-1 bg-transparent text-sm text-ink outline-none placeholder:text-ink-faint disabled:cursor-not-allowed disabled:opacity-60"
        {...props}
      />
      {trailing ? <span className="flex shrink-0 items-center">{trailing}</span> : null}
    </div>
  );
}

export type { InputProps };
