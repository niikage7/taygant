import { httpClient } from "./http-client";
import type { User } from "@/types";

export const usersService = {
  /** Возвращает профиль пользователя, соответствующего переданному access-токену. */
  getMe(): Promise<User> {
    return httpClient.get<User>("/users/me");
  },

  /** Список пользователей платформы (для назначения ответственных). */
  list(search?: string): Promise<User[]> {
    return httpClient.get<User[]>("/users", { query: { search } });
  },
};
