package service

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

// Коды ошибок PostgreSQL, которые сервисы превращают в понятные доменные ошибки.
const (
	pgUniqueViolation     = "23505"
	pgForeignKeyViolation = "23503"
)

// isUniqueViolation распознаёт нарушение уникального индекса.
//
// Проверяем код ошибки драйвера, а не текст сообщения: текст зависит от локали
// сервера и от имени индекса и сломался бы при первом же переименовании.
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation
}

// isForeignKeyViolation распознаёт ссылку на несуществующую строку
// (например, userId, которого нет в таблице users).
func isForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == pgForeignKeyViolation
}
