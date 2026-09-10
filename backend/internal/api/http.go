package api

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"taygant_backend/internal/service"
)

// fail переводит доменную ошибку в HTTP-ответ.
//
// Единственное место, где сервисные ошибки превращаются в коды: сервисы про
// HTTP не знают, а хендлеры не должны каждый раз решать, что вернуть на
// «не найдено». Неопознанная ошибка становится 500 и её текст наружу не уходит —
// в сообщении может оказаться фрагмент SQL-запроса.
func fail(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, service.ErrNotFound):
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	case errors.Is(err, service.ErrConflict):
		return fiber.NewError(fiber.StatusConflict, err.Error())
	case errors.Is(err, service.ErrInvalidCredentials):
		return fiber.NewError(fiber.StatusUnauthorized, err.Error())
	case errors.Is(err, service.ErrUnauthorized):
		return fiber.NewError(fiber.StatusUnauthorized, err.Error())
	case errors.Is(err, service.ErrForbidden):
		return fiber.NewError(fiber.StatusForbidden, err.Error())
	}

	if validation, ok := service.AsValidation(err); ok {
		return fiber.NewError(fiber.StatusBadRequest, validation.Message)
	}

	return fiber.NewError(fiber.StatusInternalServerError, "внутренняя ошибка сервера")
}

// parseBody читает JSON-тело запроса в структуру.
func parseBody[T any](c *fiber.Ctx) (T, error) {
	var body T
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.NewError(fiber.StatusBadRequest, "не удалось разобрать тело запроса")
	}
	return body, nil
}

// pathUUID читает UUID из параметра пути и отвечает 400 на некорректный формат.
// Без этой проверки нераспознанный идентификатор ушёл бы в запрос к БД
// и вернулся бы как 500 вместо понятной ошибки ввода.
func pathUUID(c *fiber.Ctx, name string) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Params(name))
	if err != nil {
		return uuid.Nil, fiber.NewError(fiber.StatusBadRequest, "некорректный идентификатор "+name)
	}
	return id, nil
}

// queryUUID читает необязательный UUID из query-параметра. Пустое значение —
// не ошибка (параметр просто не передан), а нераспознанный формат — 400,
// иначе опечатка в фильтре молча вернула бы «ничего не найдено».
func queryUUID(c *fiber.Ctx, name string) (*uuid.UUID, error) {
	raw := c.Query(name)
	if raw == "" {
		return nil, nil
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusBadRequest, "некорректный идентификатор в параметре "+name)
	}
	return &id, nil
}

// queryBool читает query-параметр вида ?flag=true. Любое значение, кроме "true",
// трактуется как false — так withDefault-парсинг не нужен ни в одном хендлере.
func queryBool(c *fiber.Ctx, name string) bool {
	return c.Query(name) == "true"
}
