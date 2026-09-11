"use client";

import { KeyRound, Trash2 } from "lucide-react";
import { useState } from "react";

import { Alert } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardBody, CardHeader, CardTitle } from "@/components/ui/card";
import { useMcpTokens, useRevokeMcpToken } from "@/data/queries";
import { toUserMessage } from "@/lib/api-error-message";
import { formatDateLong, formatDateTime } from "@/lib/format";
import type { McpToken } from "@/types";

/**
 * Выпущенные ключи. По «последнему запросу» видно, каким ключом давно не
 * пользовались — такой лучше отозвать: это лишняя дверь в аккаунт.
 */
export function McpTokenList() {
  const tokens = useMcpTokens();

  return (
    <Card>
      <CardHeader>
        <CardTitle>Ваши ключи</CardTitle>
        <Badge tone="brand" size="sm">
          {tokens.data?.length ?? 0} из 10
        </Badge>
      </CardHeader>
      <CardBody>
        {tokens.isPending ? (
          <p className="py-6 text-center text-13 text-ink-muted">Загружаем ключи…</p>
        ) : tokens.isError ? (
          <Alert tone="danger">
            {toUserMessage(tokens.error, {}, "Не удалось загрузить ключи")}
          </Alert>
        ) : tokens.data.length === 0 ? (
          <p className="py-6 text-center text-13 text-ink-muted">
            Ключей пока нет — создайте первый в шаге 1.
          </p>
        ) : (
          <ul className="space-y-2">
            {tokens.data.map((token) => (
              <TokenRow key={token.id} token={token} />
            ))}
          </ul>
        )}
      </CardBody>
    </Card>
  );
}

function TokenRow({ token }: { token: McpToken }) {
  const revoke = useRevokeMcpToken();
  const [confirming, setConfirming] = useState(false);

  return (
    <li className="flex flex-wrap items-center gap-3 rounded-control bg-surface-subtle p-3">
      <span className="flex size-9 shrink-0 items-center justify-center rounded-control bg-brand-tint text-brand">
        <KeyRound className="size-4" />
      </span>

      <div className="min-w-0 flex-1">
        <p className="truncate text-13 font-semibold text-ink">{token.name}</p>
        <p className="mt-0.5 text-xs text-ink-faint">
          <span className="font-mono">tgn_…{token.hint}</span>
          {" · "}создан {formatDateLong(token.createdAt)}
          {" · "}
          {token.lastUsedAt
            ? `последний запрос ${formatDateTime(token.lastUsedAt)}`
            : "ещё не использовался"}
        </p>
      </div>

      {confirming ? (
        <span className="flex flex-wrap items-center gap-2">
          <span className="text-xs text-ink-muted">Ассистент с этим ключом потеряет доступ.</span>
          <Button
            size="sm"
            variant="danger"
            onClick={() => revoke.mutate(token.id)}
            disabled={revoke.isPending}
          >
            {revoke.isPending ? "Отзываем…" : "Отозвать"}
          </Button>
          <Button size="sm" variant="ghost" onClick={() => setConfirming(false)}>
            Отмена
          </Button>
        </span>
      ) : (
        <Button
          size="sm"
          variant="secondary"
          onClick={() => setConfirming(true)}
          aria-label={`Отозвать ключ «${token.name}»`}
        >
          <Trash2 />
          Отозвать
        </Button>
      )}

      {revoke.isError ? (
        <p className="w-full text-xs text-danger">
          {toUserMessage(revoke.error, { 404: "Ключ уже отозван" }, "Не удалось отозвать ключ")}
        </p>
      ) : null}
    </li>
  );
}
