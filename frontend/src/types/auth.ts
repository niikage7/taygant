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
