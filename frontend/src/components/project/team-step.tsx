"use client";

import { UserPlus, X } from "lucide-react";
import { useState } from "react";

import { Avatar } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import type { ProjectMember } from "@/types";

const ROLE_LABEL: Record<string, string> = {
  project_manager: "Руководитель проекта",
  analyst: "Бизнес-аналитик",
  developer: "Разработчик",
  curator: "Куратор",
  other: "Участник",
};

const ACCESS_LABEL: Record<string, string> = {
  full: "Полный доступ",
  edit: "Редактирование",
  view: "Просмотр",
};

/** Шаг 2 мастера: владелец проекта и приглашённые участники. */
export function TeamStep({
  owner,
  initialMembers,
}: {
  owner: ProjectMember;
  initialMembers: ProjectMember[];
}) {
  const [members, setMembers] = useState(initialMembers);

  return (
    <div className="space-y-3">
      <div className="flex flex-wrap items-center justify-between gap-3 rounded-control bg-brand-tint p-3">
        <span className="flex min-w-0 items-center gap-3">
          <Avatar fullName={owner.user.fullName} className="size-11 rounded-full text-sm" />
          <span className="min-w-0">
            <span className="flex flex-wrap items-center gap-2">
              <span className="text-sm font-semibold text-ink">{owner.user.fullName}</span>
              <Badge tone="brand" size="sm" className="bg-brand text-white">
                Вы (PM)
              </Badge>
            </span>
            <span className="mt-0.5 block truncate text-xs text-ink-muted">
              {owner.user.email} · {owner.user.department}
            </span>
          </span>
        </span>
        <span className="flex items-center gap-3 text-right">
          <span>
            <span className="block text-[11px] font-semibold tracking-wider text-ink-faint uppercase">
              Назначенная роль
            </span>
            <span className="text-[13px] font-medium text-ink">
              {ROLE_LABEL[owner.projectRole]}
            </span>
          </span>
          <Badge tone="outline">{ACCESS_LABEL[owner.accessLevel]}</Badge>
        </span>
      </div>

      <ul className="grid gap-3 md:grid-cols-2 xl:grid-cols-3">
        {members.map((member) => (
          <li key={member.id} className="rounded-control bg-surface-subtle p-3">
            <div className="flex items-start justify-between gap-2">
              <span className="flex min-w-0 items-center gap-2.5">
                <Avatar fullName={member.user.fullName} className="size-9 rounded-full text-xs" />
                <span className="min-w-0">
                  <span className="block truncate text-[13px] font-semibold text-ink">
                    {member.user.fullName}
                  </span>
                  <span className="block truncate text-xs text-ink-faint">
                    {member.user.position}
                  </span>
                </span>
              </span>
              <button
                type="button"
                onClick={() =>
                  setMembers((current) => current.filter((item) => item.id !== member.id))
                }
                aria-label={`Убрать ${member.user.fullName} из команды`}
                className="shrink-0 rounded-control p-0.5 text-ink-faint hover:text-danger focus-visible:focus-ring"
              >
                <X className="size-4" />
              </button>
            </div>
            <div className="mt-3 flex items-center justify-between gap-2 border-t border-line pt-2.5">
              <span className="text-[11px] font-semibold tracking-wider text-ink-faint uppercase">
                Роль в спринтах
              </span>
              <Badge tone={member.projectRole === "curator" ? "warning" : "brand"} size="sm">
                {member.sprintRole}
              </Badge>
            </div>
          </li>
        ))}
      </ul>

      <Button variant="secondary" size="sm">
        <UserPlus />
        Добавить участника
      </Button>
    </div>
  );
}
