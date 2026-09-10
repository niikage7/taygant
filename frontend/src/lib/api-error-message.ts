import { ApiError } from "@/services/http-client";

/**
 * Превращает ошибку запроса в текст для пользователя.
 *
 * Сообщение бэкенда показываем только для 4xx-ошибок валидации — там оно осмысленное.
 * Технические строки вроде «Internal Server Error» заменяем на человеческие.
 *
 * @param error   пойманная ошибка (обычно `ApiError` или сетевой `TypeError`)
 * @param byStatus точечные тексты для конкретных статусов (401, 409 и т.п.)
 * @param fallback текст по умолчанию
 */
export function toUserMessage(
  error: unknown,
  byStatus: Record<number, string>,
  fallback: string,
): string {
  if (error instanceof ApiError) {
    const specific = byStatus[error.status];
    if (specific) return specific;

    if (error.status === 404 || error.status === 501) {
      return "Раздел ещё не готов на сервере. Попробуйте позже";
    }
    if (error.status >= 500) {
      return "Сервер временно недоступен. Попробуйте позже";
    }
    if (error.status === 429) {
      return "Слишком много попыток. Подождите немного и повторите";
    }
    if (error.status >= 400 && error.message) {
      return error.message;
    }
    return fallback;
  }

  // fetch бросает TypeError, когда сети нет вовсе.
  if (error instanceof TypeError) {
    return "Нет связи с сервером. Проверьте подключение";
  }

  return fallback;
}
