import { httpClient, setAccessToken } from "./http-client";
import type {
  LoginRequest,
  LoginResponse,
  RefreshRequest,
  RefreshResponse,
  RegisterRequest,
  RegisterResponse,
} from "@/types";

export const authService = {
  /** Проверяет логин/пароль и выдаёт пару access/refresh токенов. */
  async login(payload: LoginRequest): Promise<LoginResponse> {
    const result = await httpClient.post<LoginResponse>("/auth/login", payload, {
      auth: false,
    });
    setAccessToken(result.accessToken);
    return result;
  },

  /**
   * Регистрирует нового пользователя и сразу выдаёт пару токенов.
   *
   * ⚠️ `/auth/register` ещё не описан в `backend/api-spec.yml` и не реализован в Go-хендлерах —
   * до появления эндпоинта запрос вернёт 404/501. Форма регистрации из макета работает,
   * ошибка показывается пользователю.
   */
  async register(payload: RegisterRequest): Promise<RegisterResponse> {
    const result = await httpClient.post<RegisterResponse>("/auth/register", payload, {
      auth: false,
    });
    setAccessToken(result.accessToken);
    return result;
  },

  /** Выдаёт новый accessToken по действующему refreshToken. */
  async refresh(payload: RefreshRequest): Promise<RefreshResponse> {
    const result = await httpClient.post<RefreshResponse>("/auth/refresh", payload, {
      auth: false,
    });
    setAccessToken(result.accessToken);
    return result;
  },

  /** Очищает сохранённый access-токен на клиенте. */
  logout(): void {
    setAccessToken(null);
  },
};
