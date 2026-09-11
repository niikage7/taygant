"use client";

import * as Popover from "@radix-ui/react-popover";
import { useQueryClient } from "@tanstack/react-query";
import { LogOut } from "lucide-react";
import { useRouter } from "next/navigation";

import { UserCard } from "@/components/user/user-card";
import { Avatar } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import { useMe } from "@/data/queries";
import { endSession } from "@/lib/session";

/**
 * Профиль того, под кем выполнен вход: аватар в верхней панели открывает
 * карточку со сведениями и кнопкой выхода.
 *
 * Редактирования и удаления аккаунта здесь нет: бэкенд отдаёт профиль только
 * на чтение (`GET /users/me`), других маршрутов для пользователя у него нет.
 */
export function ProfileMenu() {
  const router = useRouter();
  const queryClient = useQueryClient();
  const me = useMe();
  const user = me.data;

  if (!user) return null;

  const logout = () => {
    endSession();
    // Кеш держит проекты и задачи прежнего пользователя — следующий вошедший
    // на этой вкладке не должен увидеть их до первого перезапроса.
    queryClient.clear();
    router.replace("/login");
  };

  return (
    <Popover.Root>
      <Popover.Trigger
        aria-label={`Профиль: ${user.fullName}`}
        className="flex shrink-0 cursor-pointer items-center gap-2 rounded-control p-1 transition-colors hover:bg-surface-muted focus-visible:focus-ring data-[state=open]:bg-surface-muted"
      >
        <Avatar fullName={user.fullName} className="size-7 rounded-full text-2xs" />
        <span className="hidden max-w-40 truncate text-13 font-medium text-ink md:inline">
          {user.fullName}
        </span>
      </Popover.Trigger>

      <Popover.Portal>
        <Popover.Content
          align="end"
          sideOffset={6}
          collisionPadding={8}
          className="z-50 w-72 rounded-card border border-line bg-surface p-4 shadow-popover"
        >
          <p className="mb-3 text-2xs font-semibold tracking-wider text-ink-faint uppercase">
            Вы вошли как
          </p>
          <UserCard user={user} />
          <Button variant="secondary" size="sm" className="mt-4 w-full" onClick={logout}>
            <LogOut />
            Выйти
          </Button>
        </Popover.Content>
      </Popover.Portal>
    </Popover.Root>
  );
}
