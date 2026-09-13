"use client";

import { ShieldCheck, Trash2, UserPlus } from "lucide-react";
import { useState } from "react";

import { Alert } from "@/components/ui/alert";
import { UserHoverCard } from "@/components/user/user-card";
import { Avatar } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardBody, CardHeader, CardTitle } from "@/components/ui/card";
import { Select } from "@/components/ui/select";
import {
  useAddMember,
  useProjectMembers,
  useRemoveMember,
  useUpdateMember,
  useUsers,
} from "@/data/queries";
import { toUserMessage } from "@/lib/api-error-message";
import { getStoredUser } from "@/lib/session";
import type { AccessLevel, ProjectMember, ProjectRole } from "@/types";

const ROLE_OPTIONS: { value: ProjectRole; label: string }[] = [
  { value: "project_manager", label: "Руководитель" },
  { value: "analyst", label: "Бизнес-аналитик" },
  { value: "developer", label: "Разработчик" },
  { value: "curator", label: "Куратор" },
  { value: "other", label: "Участник" },
];

const ACCESS_OPTIONS: { value: AccessLevel; label: string; hint: string }[] = [
  {
    value: "full",
    label: "Полный доступ",
    hint: "План, параметры проекта и команда",
  },
  {
    value: "edit",
    label: "Редактирование",
    hint: "Статус и сроки своих задач",
  },
  { value: "view", label: "Просмотр", hint: "Только чтение" },
];

const ACCESS_LABEL: Record<AccessLevel, string> = {
  full: "Полный доступ",
  edit: "Редактирование",
  view: "Просмотр",
};

const ACCESS_TONE: Record<AccessLevel, "brand" | "neutral" | "success"> = {
  full: "brand",
  edit: "success",
  view: "neutral",
};

/**
 * Управление составом команды проекта после его создания.
 *
 * Список пользователей берётся из `GET /users`: в проект можно добавить только
 * зарегистрированного пользователя. Добавлять, менять роли и исключать может
 * лишь участник с полным доступом — остальные видят состав read-only.
 */
export function TeamManager({
  projectId,
  projectName,
  canManage,
}: {
  projectId: string;
  projectName: string;
  canManage: boolean;
}) {
  const members = useProjectMembers(projectId);
  const users = useUsers();
  const addMember = useAddMember(projectId);
  const currentUserId = getStoredUser()?.id;

  const memberUserIds = new Set((members.data ?? []).map((member) => member.user.id));
  const available = (users.data ?? []).filter((user) => !memberUserIds.has(user.id));

  const [userId, setUserId] = useState("");
  const [projectRole, setProjectRole] = useState<ProjectRole>("developer");
  const [accessLevel, setAccessLevel] = useState<AccessLevel>("edit");

  const submit = () => {
    if (!userId) return;
    addMember.mutate(
      { userId, projectRole, accessLevel },
      {
        onSuccess: () => {
          setUserId("");
          setProjectRole("developer");
          setAccessLevel("edit");
        },
      },
    );
  };

  return (
    <div className="space-y-4">
      <Card>
        <CardHeader className="items-start">
          <div>
            <CardTitle>Команда проекта</CardTitle>
            <p className="mt-0.5 text-13 text-ink-muted">
              {projectName}: роли и уровни доступа участников.
            </p>
          </div>
          <Badge tone="brand" size="sm" className="gap-1">
            <ShieldCheck className="size-3.5" />
            {members.data?.length ?? 0} участников
          </Badge>
        </CardHeader>

        <CardBody className="space-y-4">
          {!canManage ? (
            <p className="rounded-control bg-surface-subtle px-3 py-2.5 text-13 text-ink-muted">
              Состав команды доступен только для просмотра. Управлять участниками может
              владелец проекта с полным доступом.
            </p>
          ) : null}

          {members.isPending ? (
            <p className="py-6 text-center text-13 text-ink-muted">Загружаем команду…</p>
          ) : members.isError ? (
            <Alert tone="danger">
              {toUserMessage(members.error, {}, "Не удалось загрузить состав команды")}
            </Alert>
          ) : (
            <ul className="space-y-2">
              {(members.data ?? []).map((member) => (
                <MemberRow
                  key={member.id}
                  projectId={projectId}
                  member={member}
                  canManage={canManage}
                  isSelf={member.user.id === currentUserId}
                />
              ))}
            </ul>
          )}
        </CardBody>
      </Card>

      {canManage ? (
        <Card>
          <CardHeader>
            <CardTitle>Добавить участника</CardTitle>
          </CardHeader>
          <CardBody className="space-y-3">
            <div className="grid grid-cols-1 gap-3 md:grid-cols-[minmax(0,1.4fr)_minmax(0,1fr)_minmax(0,1fr)_auto]">
              <div>
                <label htmlFor="team-user" className="text-xs text-ink-faint">
                  Пользователь
                </label>
                <Select
                  id="team-user"
                  value={userId}
                  onChange={(event) => setUserId(event.target.value)}
                  className="mt-1"
                >
                  <option value="">Выберите пользователя…</option>
                  {available.map((user) => (
                    <option key={user.id} value={user.id}>
                      {user.fullName}
                      {user.position ? ` — ${user.position}` : ""}
                    </option>
                  ))}
                </Select>
              </div>

              <div>
                <label htmlFor="team-role" className="text-xs text-ink-faint">
                  Роль в проекте
                </label>
                <Select
                  id="team-role"
                  value={projectRole}
                  onChange={(event) => setProjectRole(event.target.value as ProjectRole)}
                  className="mt-1"
                >
                  {ROLE_OPTIONS.map((option) => (
                    <option key={option.value} value={option.value}>
                      {option.label}
                    </option>
                  ))}
                </Select>
              </div>

              <div>
                <label htmlFor="team-access" className="text-xs text-ink-faint">
                  Уровень доступа
                </label>
                <Select
                  id="team-access"
                  value={accessLevel}
                  onChange={(event) => setAccessLevel(event.target.value as AccessLevel)}
                  className="mt-1"
                >
                  {ACCESS_OPTIONS.map((option) => (
                    <option key={option.value} value={option.value}>
                      {option.label}
                    </option>
                  ))}
                </Select>
                <p className="mt-1 text-xs text-ink-faint">
                  {ACCESS_OPTIONS.find((option) => option.value === accessLevel)?.hint}
                </p>
              </div>

              <div className="flex items-end">
                <Button
                  onClick={submit}
                  disabled={!userId || addMember.isPending || available.length === 0}
                  className="w-full md:w-auto"
                >
                  <UserPlus />
                  {addMember.isPending ? "Добавляем…" : "Добавить"}
                </Button>
              </div>
            </div>

            {available.length === 0 && !users.isPending ? (
              <p className="text-xs text-ink-faint">
                Все зарегистрированные пользователи уже состоят в проекте.
              </p>
            ) : null}

            {users.isError ? (
              <Alert tone="danger">
                {toUserMessage(users.error, {}, "Не удалось загрузить пользователей")}
              </Alert>
            ) : null}

            {addMember.isError ? (
              <Alert tone="danger">
                {toUserMessage(
                  addMember.error,
                  { 409: "Этот пользователь уже в команде", 400: "Проверьте данные участника" },
                  "Не удалось добавить участника",
                )}
              </Alert>
            ) : null}
          </CardBody>
        </Card>
      ) : null}
    </div>
  );
}

function MemberRow({
  projectId,
  member,
  canManage,
  isSelf,
}: {
  projectId: string;
  member: ProjectMember;
  canManage: boolean;
  isSelf: boolean;
}) {
  const updateMember = useUpdateMember(projectId);
  const removeMember = useRemoveMember(projectId);
  const busy = updateMember.isPending || removeMember.isPending;
  const roleLabel = ROLE_OPTIONS.find((option) => option.value === member.projectRole)?.label;

  const changeRole = (projectRole: ProjectRole) =>
    updateMember.mutate({ memberId: member.id, payload: { projectRole } });
  const changeAccess = (accessLevel: AccessLevel) =>
    updateMember.mutate({ memberId: member.id, payload: { accessLevel } });

  return (
    <li className="flex flex-wrap items-center gap-3 rounded-control bg-surface-subtle p-3">
      <UserHoverCard user={member.user} note={roleLabel}>
        <span className="shrink-0">
          <Avatar fullName={member.user.fullName} className="size-9 rounded-full text-xs" />
        </span>
      </UserHoverCard>

      <div className="min-w-0 flex-1">
        <p className="flex flex-wrap items-center gap-2">
          <UserHoverCard user={member.user} note={roleLabel}>
            <span className="truncate text-13 font-semibold text-ink">
              {member.user.fullName}
            </span>
          </UserHoverCard>
          {isSelf ? (
            <Badge tone="brand" size="sm">
              Вы
            </Badge>
          ) : null}
        </p>
        <p className="mt-0.5 truncate text-xs text-ink-faint">
          {member.user.email}
          {member.user.position ? ` · ${member.user.position}` : ""}
        </p>
      </div>

      {canManage ? (
        // На телефоне селекты уходят на свою строку во всю ширину: рядом с
        // аватаром от имени оставалось бы «Елен…».
        <div className="grid w-full grid-cols-[minmax(0,1fr)_auto] items-center gap-2 sm:flex sm:w-auto">
          <div className="col-start-1 min-w-0 sm:w-44">
            <Select
              value={member.projectRole}
              onChange={(event) => changeRole(event.target.value as ProjectRole)}
              disabled={busy}
              aria-label={`Роль ${member.user.fullName}`}
              className="h-8"
            >
              {ROLE_OPTIONS.map((option) => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </Select>
          </div>

          <div className="col-start-1 min-w-0 sm:w-44">
            <Select
              value={member.accessLevel}
              onChange={(event) => changeAccess(event.target.value as AccessLevel)}
              disabled={busy}
              aria-label={`Уровень доступа ${member.user.fullName}`}
              className="h-8"
            >
              {ACCESS_OPTIONS.map((option) => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </Select>
          </div>

          <button
            type="button"
            onClick={() => removeMember.mutate(member.id)}
            disabled={busy}
            aria-label={`Исключить ${member.user.fullName}`}
            className="col-start-2 row-span-2 row-start-1 shrink-0 rounded-control p-1.5 text-ink-faint transition-colors hover:text-danger focus-visible:focus-ring disabled:opacity-50"
          >
            <Trash2 className="size-4" />
          </button>
        </div>
      ) : (
        <Badge tone={ACCESS_TONE[member.accessLevel]} size="sm">
          {ACCESS_LABEL[member.accessLevel]}
        </Badge>
      )}

      {updateMember.isError || removeMember.isError ? (
        <p className="w-full text-xs text-danger">
          {toUserMessage(
            updateMember.error ?? removeMember.error,
            {
              400: "В проекте должен остаться хотя бы один участник с полным доступом",
            },
            "Не удалось изменить участника",
          )}
        </p>
      ) : null}
    </li>
  );
}
