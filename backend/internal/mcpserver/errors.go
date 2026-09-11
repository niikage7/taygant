package mcpserver

import (
	"errors"
	"fmt"
	"log"

	"taygant_backend/internal/service"
)

// toolFailure — ошибка, текст которой адресован ассистенту и человеку: что не
// так и как поправить запрос («уточните проект», «такого статуса нет»).
// SDK отдаёт её клиенту как результат инструмента с isError, и нейронка может
// сама исправить вызов.
type toolFailure struct {
	msg string
}

func (e *toolFailure) Error() string { return e.msg }

// failf собирает toolFailure с форматированным текстом.
func failf(format string, args ...any) error {
	return &toolFailure{msg: fmt.Sprintf(format, args...)}
}

// toolError переводит ошибку сервисов в текст для ассистента — так же, как
// fail в internal/api переводит её в HTTP-код. Неизвестная ошибка (обычно
// сбой БД) наружу не уходит: в её тексте может оказаться фрагмент SQL.
func toolError(err error) error {
	var failure *toolFailure
	switch {
	case err == nil:
		return nil
	case errors.As(err, &failure):
		return failure
	case errors.Is(err, service.ErrNotFound):
		return failf("не найдено: объект удалён или недоступен вам")
	case errors.Is(err, service.ErrForbidden):
		return failf("недостаточно прав для этого действия")
	}
	if validation, ok := service.AsValidation(err); ok {
		return failf("%s", validation.Message)
	}
	log.Printf("mcp: %v", err)
	return failf("внутренняя ошибка сервера — попробуйте ещё раз чуть позже")
}
