"use client";

import { UserPlus, X } from "lucide-react";

import { Avatar } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Select } from "@/components/ui/select";
import type { ProjectRole, User } from "@/types";

const ROLE_OPTIONS: { value: ProjectRole; label: string }[] = [
  { value: "project_manager", label: "Руководитель проекта" },
  { value: "analyst", label: "Бизнес-аналитик" },
  { value: "developer", label: "Разработчик" },
  { value: "curator", label: "Куратор" },
  { value: "other", label: "Участник" },
];

export type DraftMember = { userId: string; projectRole: ProjectRole };

/**
 * Шаг 2 мастера: владелец проекта и приглашённые участники.
 *
 * Список берётся из `GET /users`, а не из выдуманных данных: мастер создаёт
 * настоящий проект, и назначить в него можно только реально существующих
 * пользователей — иначе бэкенд отклонит состав команды.
 */
export function TeamStep({
  users,
  currentUser,
  members,
  onChange,
}: {
  users: User[];
  currentUser: User | null;
  members: DraftMember[];
  onChange: (members: DraftMember[]) => void;
}) {
  const byId = new Map(users.map((user) => [user.id, user]));
  const selectedIds = new Set(members.map((member) => member.userId));
  const available = users.filter(
    (user) => user.id !== currentUser?.id && !selectedIds.has(user.id),
  );

  return (
    <div className="space-y-3">
      {currentUser ? (
        <div className="flex flex-wrap items-center justify-between gap-3 rounded-control bg-brand-tint p-3">
          <span className="flex min-w-0 items-center gap-3">
            <Avatar fullName={currentUser.fullName} className="size-11 rounded-full text-sm" />
            <span className="min-w-0">
              <span className="flex flex-wrap items-center gap-2">
                <span className="text-sm font-semibold text-ink">
                  {currentUser.fullName}
                </span>
                <Badge tone="brand" size="sm" className="bg-brand text-white">
                  Вы (PM)
                </Badge>
              </span>
              <span className="mt-0.5 block truncate text-xs text-ink-muted">
                {currentUser.email}
                {currentUser.department ? ` · ${currentUser.department}` : ""}
              </span>
            </span>
          </span>
          <span className="flex items-center gap-3 text-right">
            <span>
              <span className="block text-[11px] font-semibold tracking-wider text-ink-faint uppercase">
                Назначенная роль
              </span>
              <span className="text-[13px] font-medium text-ink">
                Руководитель проекта
              </span>
            </span>
            <Badge tone="outline">Полный доступ</Badge>
          </span>
        </div>
      ) : null}

      {members.length > 0 ? (
        <ul className="grid gap-3 md:grid-cols-2 xl:grid-cols-3">
          {members.map((member) => {
            const user = byId.get(member.userId);
            if (!user) return null;
            return (
              <li key={member.userId} className="rounded-control bg-surface-subtle p-3">
                <div className="flex items-start justify-between gap-2">
                  <span className="flex min-w-0 items-center gap-2.5">
                    <Avatar fullName={user.fullName} className="size-9 rounded-full text-xs" />
                    <span className="min-w-0">
                      <span className="block truncate text-[13px] font-semibold text-ink">
                        {user.fullName}
                      </span>
                      <span className="block truncate text-xs text-ink-faint">
                        {user.position || user.email}
                      </span>
                    </span>
                  </span>
                  <button
                    type="button"
                    onClick={() =>
                      onChange(members.filter((item) => item.userId !== member.userId))
                    }
                    aria-label={`Убрать ${user.fullName} из команды`}
                    className="shrink-0 rounded-control p-0.5 text-ink-faint hover:text-danger focus-visible:focus-ring"
                  >
                    <X className="size-4" />
                  </button>
                </div>
                <div className="mt-3 border-t border-line pt-2.5">
                  <label
                    htmlFor={`role-${member.userId}`}
                    className="block text-[11px] font-semibold tracking-wider text-ink-faint uppercase"
                  >
                    Роль в проекте
                  </label>
                  <Select
                    id={`role-${member.userId}`}
                    value={member.projectRole}
                    onChange={(event) =>
                      onChange(
                        members.map((item) =>
                          item.userId === member.userId
                            ? { ...item, projectRole: event.target.value as ProjectRole }
                            : item,
                        ),
                      )
                    }
                    className="mt-1.5 h-8"
                  >
                    {ROLE_OPTIONS.map((option) => (
                      <option key={option.value} value={option.value}>
                        {option.label}
                      </option>
                    ))}
                  </Select>
                </div>
              </li>
            );
          })}
        </ul>
      ) : (
        <p className="rounded-control bg-surface-subtle px-3 py-4 text-center text-[13px] text-ink-muted">
          Пока в проекте только вы. Добавьте участников из списка ниже.
        </p>
      )}

      {available.length > 0 ? (
        <label className="flex flex-wrap items-center gap-2">
          <span className="flex items-center gap-1.5 text-[13px] font-medium text-brand">
            <UserPlus className="size-4" />
            Добавить участника
          </span>
          <Select
            value=""
            aria-label="Добавить участника в проект"
            onChange={(event) => {
              if (!event.target.value) return;
              onChange([
                ...members,
                { userId: event.target.value, projectRole: "developer" },
              ]);
            }}
            className="h-8 w-64"
          >
            <option value="">Выберите пользователя…</option>
            {available.map((user) => (
              <option key={user.id} value={user.id}>
                {user.fullName}
                {user.position ? ` — ${user.position}` : ""}
              </option>
            ))}
          </Select>
        </label>
      ) : null}
    </div>
  );
}
