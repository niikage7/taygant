import { setAccessToken } from "@/services/http-client";
import type { LoginResponse, User } from "@/types";

const REFRESH_TOKEN_KEY = "taygant.refreshToken";
const USER_KEY = "taygant.user";

/**
 * Клиентская сессия поверх http-client: accessToken живёт в памяти/localStorage
 * (см. http-client), а refreshToken и профиль кладём в localStorage при «Запомнить меня»
 * и в sessionStorage — если галочка снята (тогда сессия умрёт вместе со вкладкой).
 */
function stores(): Storage[] {
  if (typeof window === "undefined") return [];
  return [window.localStorage, window.sessionStorage];
}

/** Сохраняет результат входа. `remember` выбирает долговременное хранилище. */
export function startSession(session: LoginResponse, remember: boolean): void {
  if (typeof window === "undefined") return;
  setAccessToken(session.accessToken);
  clearStorages();
  const store = remember ? window.localStorage : window.sessionStorage;
  store.setItem(REFRESH_TOKEN_KEY, session.refreshToken);
  store.setItem(USER_KEY, JSON.stringify(session.user));
}

/** Возвращает refresh-токен из любого из хранилищ. */
export function getRefreshToken(): string | null {
  for (const store of stores()) {
    const token = store.getItem(REFRESH_TOKEN_KEY);
    if (token) return token;
  }
  return null;
}

/** Возвращает профиль текущего пользователя, если он был сохранён при входе. */
export function getStoredUser(): User | null {
  for (const store of stores()) {
    const raw = store.getItem(USER_KEY);
    if (!raw) continue;
    try {
      return JSON.parse(raw) as User;
    } catch {
      store.removeItem(USER_KEY);
    }
  }
  return null;
}

function clearStorages(): void {
  for (const store of stores()) {
    store.removeItem(REFRESH_TOKEN_KEY);
    store.removeItem(USER_KEY);
  }
}

/** Полностью завершает сессию на клиенте. */
export function endSession(): void {
  setAccessToken(null);
  clearStorages();
}
