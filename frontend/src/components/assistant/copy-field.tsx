"use client";

import { Check, Copy } from "lucide-react";
import { useEffect, useState } from "react";

import { cn } from "@/lib/utils";

/**
 * Текст для копирования: ключ, команда, кусок конфига.
 *
 * `navigator.clipboard` есть не везде (нужен https или localhost), поэтому
 * текст всегда виден и выделяется целиком одним щелчком — при отказе буфера
 * его можно скопировать вручную.
 */
export function CopyField({
  value,
  copyLabel,
  className,
}: {
  value: string;
  /** Подпись кнопки для скринридера: «Скопировать ключ», «Скопировать команду». */
  copyLabel: string;
  className?: string;
}) {
  const [state, setState] = useState<"idle" | "copied" | "failed">("idle");

  useEffect(() => {
    if (state === "idle") return;
    const timer = setTimeout(() => setState("idle"), 2500);
    return () => clearTimeout(timer);
  }, [state]);

  async function copy() {
    try {
      await navigator.clipboard.writeText(value);
      setState("copied");
    } catch {
      setState("failed");
    }
  }

  return (
    // Кнопка — отдельная колонка, а не поверх текста: на узком экране
    // абсолютная кнопка закрывала середину команды.
    <div className={cn("flex items-start gap-1 rounded-control bg-surface-muted", className)}>
      <pre className="min-w-0 flex-1 overflow-x-auto py-2.5 pl-3 font-mono text-xs leading-5 whitespace-pre text-ink select-all">
        {value}
      </pre>
      <button
        type="button"
        onClick={copy}
        aria-label={copyLabel}
        className="mt-1.5 mr-1.5 inline-flex shrink-0 items-center gap-1 rounded-control bg-surface px-2 py-1 text-2xs font-medium text-ink-muted shadow-card transition-colors hover:text-ink focus-visible:focus-ring"
      >
        {state === "copied" ? (
          <Check className="size-3.5 text-success" />
        ) : (
          <Copy className="size-3.5" />
        )}
        {state === "copied"
          ? "Скопировано"
          : state === "failed"
            ? "Выделите вручную"
            : "Копировать"}
      </button>
    </div>
  );
}
