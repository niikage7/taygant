import type { User } from "./user";

/** Данные для входа пользователя. */
export interface LoginRequest {
  /** Email пользователя, используемый как логин. */
  email: string;
  /** Пароль пользователя в открытом виде (передаётся только по HTTPS). */
  password: string;
}

/** Ответ на успешную аутентификацию. */
export interface LoginResponse {
  /** Короткоживущий JWT для авторизации запросов (передаётся в заголовке Authorization). */
  accessToken: string;
  /** Долгоживущий токен для получения нового accessToken без повторного ввода пароля. */
  refreshToken: string;
  user: User;
}

/** Данные для обновления access-токена. */
export interface RefreshRequest {
  /** Refresh-токен, полученный при входе. */
  refreshToken: string;
}

/** Ответ с новым access-токеном. */
export interface RefreshResponse {
  /** Новый короткоживущий JWT. */
  accessToken: string;
}

/**
 * Данные для регистрации нового пользователя.
 *
 * `POST /auth/register` намеренно отсутствует в `backend/api-spec.yml`: спецификация
 * описывает только вход, предполагая заранее заведённые учётные записи. Эндпоинт
 * реализован в `backend/internal/api/auth.go` — сверяться нужно с ним.
 */
export interface RegisterRequest {
  /** Email пользователя, он же логин. */
  email: string;
  /** Полное имя пользователя (ФИО). */
  fullName: string;
  /**
   * Пароль в открытом виде (передаётся только по HTTPS).
   * Бэкенд требует не короче 8 символов, не длиннее 72 байт,
   * минимум одну букву и одну цифру.
   */
  password: string;
  /** Организационное подразделение. Форма регистрации его не собирает. */
  department?: string;
  /** Должность/специализация. Форма регистрации её не собирает. */
  position?: string;
}

/** Ответ на успешную регистрацию (201): сразу выдаётся пара токенов, как при входе. */
export type RegisterResponse = LoginResponse;
