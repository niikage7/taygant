"use client";

import * as Tooltip from "@radix-ui/react-tooltip";
import { Building2, BriefcaseBusiness, Mail } from "lucide-react";
import type { ReactNode } from "react";

import { Avatar } from "@/components/ui/avatar";
import { cn } from "@/lib/utils";
import type { User } from "@/types";

/**
 * Сведения о пользователе: аватар, имя, должность, отдел и email.
 *
 * Один блок и для своего профиля в верхней панели, и для всплывающей карточки
 * над чужим именем — иначе одни и те же поля расходились бы по двум вёрсткам.
 * Поля, которых нет в профиле, не выводятся: пустая строка «Отдел: —» только шумит.
 */
export function UserCard({
  user,
  note,
  className,
}: {
  user: User;
  /** Дополнительная строка под именем — например, роль в проекте. */
  note?: ReactNode;
  className?: string;
}) {
  const details = [
    { icon: BriefcaseBusiness, label: "Должность", value: user.position },
    { icon: Building2, label: "Отдел", value: user.department },
    { icon: Mail, label: "Email", value: user.email },
  ].filter((item) => Boolean(item.value));

  return (
    <div className={cn("min-w-0", className)}>
      <div className="flex items-center gap-3">
        <Avatar fullName={user.fullName} className="size-10 rounded-full text-sm" />
        <div className="min-w-0">
          <p className="truncate text-15 font-semibold text-ink">{user.fullName}</p>
          {note ? <div className="mt-0.5 truncate text-xs text-ink-muted">{note}</div> : null}
        </div>
      </div>

      {details.length > 0 ? (
        <dl className="mt-3 space-y-1.5 border-t border-line pt-3">
          {details.map(({ icon: Icon, label, value }) => (
            <div key={label} className="flex items-center gap-2 text-13">
              <dt className="flex shrink-0 items-center">
                <Icon className="size-3.5 text-ink-faint" aria-hidden />
                <span className="sr-only">{label}</span>
              </dt>
              <dd className="min-w-0 truncate text-ink-muted" title={value}>
                {value}
              </dd>
            </div>
          ))}
        </dl>
      ) : null}
    </div>
  );
}

/**
 * Всплывающая карточка пользователя при наведении на его аватар или имя.
 *
 * Сделана на Tooltip, а не на Popover: карточка только показывает сведения и
 * не содержит кнопок, а открываться должна по наведению и фокусу без клика —
 * клик по имени внутри ссылки (строка задачи) должен остаться переходом.
 * Триггером становится сам переданный элемент (`asChild`), поэтому обёртка не
 * добавляет в разметку лишних узлов и не ломает раскладку строк и таблиц.
 */
export function UserHoverCard({
  user,
  note,
  children,
  side = "top",
}: {
  user: User;
  note?: ReactNode;
  /** Ровно один элемент — он станет триггером. */
  children: ReactNode;
  side?: "top" | "bottom" | "left" | "right";
}) {
  return (
    <Tooltip.Root delayDuration={350}>
      <Tooltip.Trigger asChild>{children}</Tooltip.Trigger>
      <Tooltip.Portal>
        <Tooltip.Content
          side={side}
          align="start"
          sideOffset={6}
          collisionPadding={8}
          className="z-50 w-72 rounded-card border border-line bg-surface p-4 shadow-popover"
        >
          <UserCard user={user} note={note} />
        </Tooltip.Content>
      </Tooltip.Portal>
    </Tooltip.Root>
  );
}
