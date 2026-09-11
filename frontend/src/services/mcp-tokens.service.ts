import { httpClient } from "./http-client";
import type { McpToken, McpTokenCreated, McpTokenCreateRequest } from "@/types";

/**
 * Личные ключи для подключения ассистента по MCP (тег McpTokens в api-spec.yml).
 * Сам MCP-сервер живёт вне /api/v1 — по адресу /backend/mcp, см. `mcpServerUrl`.
 */
export const mcpTokensService = {
  /** Ключи текущего пользователя, новые сверху. */
  list(): Promise<McpToken[]> {
    return httpClient.get<McpToken[]>("/users/me/mcp-tokens");
  },

  /** Выпускает ключ. Секрет приходит только в этом ответе. */
  create(payload: McpTokenCreateRequest): Promise<McpTokenCreated> {
    return httpClient.post<McpTokenCreated>("/users/me/mcp-tokens", payload);
  },

  /** Отзывает ключ: подключённый им ассистент теряет доступ со следующего запроса. */
  revoke(tokenId: string): Promise<void> {
    return httpClient.delete<void>(`/users/me/mcp-tokens/${tokenId}`);
  },
};

/**
 * Адрес MCP-сервера для настроек ассистента. Тот же прокси Next.js, что и у
 * API (`/backend/*` → бэкенд), поэтому адрес строится от origin приложения.
 */
export function mcpServerUrl(origin: string): string {
  return `${origin}/backend/mcp`;
}
