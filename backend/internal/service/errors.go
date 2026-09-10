// Package service содержит бизнес-логику и работу с БД через GORM.
// HTTP-слой (internal/api) разбирает запрос, вызывает сервис и собирает DTO;
// SQL-запросов в хендлерах нет.
package service

import (
	"errors"
	"fmt"
	"unicode/utf8"
)

// Доменные ошибки. Сервисы не знают про HTTP-коды — сопоставление кодам
// живёт в одном месте, в internal/api. Иначе номер статуса пришлось бы
// дублировать в каждом сервисе и следить за его согласованностью.
var (
	// ErrNotFound — запрошенный объект не существует или недоступен пользователю.
	ErrNotFound = errors.New("объект не найден")
	// ErrConflict — действие нарушает уникальность или текущее состояние данных.
	ErrConflict = errors.New("конфликт данных")
	// ErrInvalidCredentials — неверная пара email/пароль.
	ErrInvalidCredentials = errors.New("неверный email или пароль")
	// ErrForbidden — прав пользователя недостаточно для действия.
	ErrForbidden = errors.New("недостаточно прав")
	// ErrUnauthorized — запрос без действительного токена.
	ErrUnauthorized = errors.New("требуется авторизация")
)

// ValidationError — ошибка входных данных с текстом, пригодным для показа пользователю.
type ValidationError struct {
	Message string
}

// Error реализует интерфейс error.
func (e *ValidationError) Error() string { return e.Message }

// Invalid собирает ValidationError с форматированным сообщением.
func Invalid(format string, args ...any) error {
	return &ValidationError{Message: fmt.Sprintf(format, args...)}
}

// AsValidation извлекает ValidationError из цепочки ошибок.
func AsValidation(err error) (*ValidationError, bool) {
	var target *ValidationError
	ok := errors.As(err, &target)
	return target, ok
}

// checkMaxLen проверяет длину строки в символах (не байтах — utf8.RuneCountInString),
// иначе кириллица и другой не-ASCII текст ограничивались бы втрое строже, чем
// заявлено в спецификации и колонках БД (varchar считает символы, не байты).
// Ограничение из БД (varchar) само по себе на 500 — эта проверка нужна, чтобы
// вернуть 400 до INSERT.
func checkMaxLen(field, value string, max int) error {
	if utf8.RuneCountInString(value) > max {
		return Invalid("%s не может быть длиннее %d символов", field, max)
	}
	return nil
}
