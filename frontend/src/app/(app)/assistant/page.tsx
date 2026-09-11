"use client";

import { AssistantSetup } from "@/components/assistant/assistant-setup";
import { McpTokenList } from "@/components/assistant/mcp-token-list";

/**
 * Подключение приложения к нейронке пользователя по MCP (docs/mcp.md): личный
 * ключ, готовые настройки для Claude Code и Cursor и список выпущенных ключей.
 *
 * Страница не зависит от выбранного проекта: ключ личный и открывает
 * ассистенту все проекты, в которых человек состоит.
 */
export default function AssistantPage() {
  return (
    <div className="mx-auto max-w-[1000px] space-y-4">
      <div>
        <p className="font-mono text-xs tracking-wide text-ink-faint uppercase">
          Подключение по MCP
        </p>
        <h1 className="mt-1.5 text-2xl leading-tight font-bold tracking-tight text-ink">
          Ассистент в редакторе кода
        </h1>
        <p className="mt-2 max-w-2xl text-13 text-ink-muted">
          Подключите taygant к нейронке, которой вы уже пользуетесь, — Claude Code, Cursor или
          другой, что поддерживает MCP. Спросите «что у меня сегодня?», скажите «беру TASK-005 в
          работу» — и статус обновится, не выходя из редактора. Без вашего подтверждения ассистент
          ничего не меняет.
        </p>
      </div>

      <AssistantSetup />
      <McpTokenList />
    </div>
  );
}
