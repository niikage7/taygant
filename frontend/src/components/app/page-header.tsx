import type { ReactNode } from "react";

import { cn } from "@/lib/utils";

/**
 * Единый заголовок страницы: моноширинная надпись сверху (обычно «ПРОЕКТ
 * #КОД»), заголовок и мета-строка под ним, кнопки действий справа.
 *
 * На узких экранах `actions` переносятся под заголовок, а не сжимают его —
 * иначе на 375px заголовок и кнопки наезжали бы друг на друга.
 */
export function PageHeader({
  eyebrow,
  title,
  description,
  meta,
  actions,
  className,
}: {
  /** Моноширинная подпись над заголовком (например, «Проект #ABC-1»). */
  eyebrow?: ReactNode;
  title: ReactNode;
  /** Пояснение под заголовком. Синоним `meta` — используйте то, что подходит по смыслу вызова. */
  description?: ReactNode;
  meta?: ReactNode;
  /** Кнопки действий; переносятся под заголовок на экранах уже `sm`. */
  actions?: ReactNode;
  className?: string;
}) {
  const sub = description ?? meta;

  return (
    <div className={cn("flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between", className)}>
      <div className="min-w-0">
        {eyebrow ? (
          <p className="font-mono text-2xs font-semibold tracking-wider text-ink-muted uppercase">
            {eyebrow}
          </p>
        ) : null}
        <h1 className="mt-1 text-2xl leading-tight font-bold tracking-tight text-ink">{title}</h1>
        {sub ? <div className="mt-1.5 text-13 text-ink-muted">{sub}</div> : null}
      </div>
      {actions ? (
        <div className="flex shrink-0 flex-wrap items-center gap-2">{actions}</div>
      ) : null}
    </div>
  );
}
