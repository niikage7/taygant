/**
 * Личный ключ, которым нейронка пользователя (Claude Code, Cursor…) подключается
 * к MCP-серверу приложения. Сам ключ сервер не возвращает — только подписи.
 */
export interface McpToken {
  /** Идентификатор ключа. */
  id: string;
  /** Подпись ключа, по которой его узнают в списке и в истории задач. */
  name: string;
  /** Последние 4 символа ключа. */
  hint: string;
  /** Когда ключ выпущен. */
  createdAt: string;
  /** Когда ключом пользовались в последний раз; null — ещё ни разу. */
  lastUsedAt: string | null;
}

/** Тело POST /users/me/mcp-tokens. */
export interface McpTokenCreateRequest {
  /** Подпись ключа, например «Claude Code, ноутбук». */
  name: string;
}

/** Ответ на выпуск ключа — единственный, в котором есть сам ключ. */
export interface McpTokenCreated {
  token: McpToken;
  /** Ключ для заголовка `Authorization: Bearer` в настройках ассистента. */
  secret: string;
}
