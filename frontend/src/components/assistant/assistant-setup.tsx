"use client";

import { KeyRound, Plug } from "lucide-react";
import { useState, useSyncExternalStore, type FormEvent } from "react";

import { CopyField } from "@/components/assistant/copy-field";
import { Alert } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Card, CardBody, CardHeader, CardTitle } from "@/components/ui/card";
import { FormField } from "@/components/ui/form-field";
import { Input } from "@/components/ui/input";
import { Segmented, type SegmentedOption } from "@/components/ui/segmented";
import { useCreateMcpToken } from "@/data/queries";
import { toUserMessage } from "@/lib/api-error-message";
import { mcpServerUrl } from "@/services";
import type { McpTokenCreated } from "@/types";

type ClientKind = "claude" | "cursor" | "other";

const CLIENT_OPTIONS: readonly SegmentedOption<ClientKind>[] = [
  { value: "claude", label: "Claude Code" },
  { value: "cursor", label: "Cursor" },
  { value: "other", label: "Другой клиент" },
];

/** Что показывать в инструкциях, пока ключ не выпущен на этой странице. */
const KEY_PLACEHOLDER = "tgn_ВАШ_КЛЮЧ";

/**
 * Настройка пишется в пользовательский конфиг (`--scope user`), а не в файл
 * проекта: ключ даёт доступ к аккаунту и не должен попасть в репозиторий.
 * `--header` стоит последним — он принимает несколько значений и иначе
 * забрал бы себе имя сервера и адрес.
 */
function claudeCommand(url: string, key: string): string {
  return `claude mcp add --transport http --scope user taygant ${url} --header "Authorization: Bearer ${key}"`;
}

/** Глобальный ~/.cursor/mcp.json, а не .cursor/mcp.json в проекте — по той же причине. */
function cursorConfig(url: string, key: string): string {
  return JSON.stringify(
    { mcpServers: { taygant: { url, headers: { Authorization: `Bearer ${key}` } } } },
    null,
    2,
  );
}

const noSubscription = () => () => {};

/**
 * Origin приложения. На сервере его нет — там отдаётся пустая строка, и адрес
 * дорисовывается в браузере без расхождения серверной и клиентской разметки.
 */
function useOrigin(): string {
  return useSyncExternalStore(
    noSubscription,
    () => window.location.origin,
    () => "",
  );
}

/** Шаги подключения: выпустить ключ и вписать его в настройки ассистента. */
export function AssistantSetup() {
  const createToken = useCreateMcpToken();
  const [name, setName] = useState("");
  const [created, setCreated] = useState<McpTokenCreated | null>(null);
  const [client, setClient] = useState<ClientKind>("claude");

  const url = mcpServerUrl(useOrigin());
  const key = created?.secret ?? KEY_PLACEHOLDER;

  const submit = (event: FormEvent) => {
    event.preventDefault();
    const trimmed = name.trim();
    if (!trimmed) return;
    createToken.mutate(
      { name: trimmed },
      {
        onSuccess: (result) => {
          setCreated(result);
          setName("");
        },
      },
    );
  };

  return (
    <>
      <Card>
        <CardHeader className="items-start">
          <div>
            <CardTitle>
              <KeyRound className="size-4 text-brand" />
              Шаг 1. Создайте личный ключ
            </CardTitle>
            <p className="mt-0.5 text-[13px] text-ink-muted">
              Ключ — как пароль для ассистента: он работает от вашего имени и видит только ваши
              проекты. Заведите отдельный ключ на каждый редактор — ненужный легко отозвать.
            </p>
          </div>
        </CardHeader>
        <CardBody className="space-y-3">
          <form onSubmit={submit} className="flex flex-col gap-3 md:flex-row md:items-end">
            <FormField
              id="mcp-token-name"
              label="Название ключа"
              hint="Видно в истории задач"
              className="flex-1"
            >
              <Input
                id="mcp-token-name"
                value={name}
                onChange={(event) => setName(event.target.value)}
                placeholder="Claude Code, ноутбук"
                maxLength={120}
                autoComplete="off"
              />
            </FormField>
            <Button type="submit" disabled={!name.trim() || createToken.isPending}>
              {createToken.isPending ? "Создаём…" : "Создать ключ"}
            </Button>
          </form>

          {createToken.isError ? (
            <Alert tone="danger">
              {toUserMessage(createToken.error, {}, "Не удалось создать ключ")}
            </Alert>
          ) : null}

          {created ? (
            <div className="space-y-2">
              <Alert tone="warning">
                Ключ «{created.token.name}» создан. Скопируйте его сейчас: показать его ещё раз
                нельзя — потерянный ключ отзывают и выпускают новый.
              </Alert>
              <CopyField value={created.secret} copyLabel="Скопировать ключ" />
            </div>
          ) : null}
        </CardBody>
      </Card>

      <Card>
        <CardHeader className="flex-wrap items-start">
          <CardTitle>
            <Plug className="size-4 text-brand" />
            Шаг 2. Подключите ассистента
          </CardTitle>
          <Segmented
            options={CLIENT_OPTIONS}
            value={client}
            onChange={setClient}
            ariaLabel="Какой ассистент подключаете"
          />
        </CardHeader>
        <CardBody className="space-y-3 text-[13px] text-ink-muted">
          {client === "claude" ? (
            <>
              <p>Выполните в терминале:</p>
              <CopyField value={claudeCommand(url, key)} copyLabel="Скопировать команду" />
            </>
          ) : client === "cursor" ? (
            <>
              <p>
                Добавьте в файл <code className="font-mono text-ink">~/.cursor/mcp.json</code> и
                перезапустите Cursor:
              </p>
              <CopyField value={cursorConfig(url, key)} copyLabel="Скопировать настройки" />
            </>
          ) : (
            <>
              <p>Подойдёт любой клиент, который умеет MCP по Streamable HTTP. Адрес сервера:</p>
              <CopyField value={url} copyLabel="Скопировать адрес" />
              <p>Заголовок авторизации:</p>
              <CopyField value={`Authorization: Bearer ${key}`} copyLabel="Скопировать заголовок" />
            </>
          )}

          {created ? null : (
            <p className="text-xs text-ink-faint">
              Вместо <span className="font-mono">{KEY_PLACEHOLDER}</span> подставьте ключ из шага 1.
            </p>
          )}

          <p>
            Готово — спросите ассистента: «Что у меня сегодня?», «Как дела у проекта?», «Что
            горит?». Статус он поменяет только после вашего «да», а в истории задачи такая правка
            будет помечена «через ассистента».
          </p>
        </CardBody>
      </Card>
    </>
  );
}
