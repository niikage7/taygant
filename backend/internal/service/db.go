package service

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

// pgUniqueViolation — код ошибки PostgreSQL «нарушение уникального ограничения».
const pgUniqueViolation = "23505"

// isUniqueViolation распознаёт нарушение уникального индекса.
//
// Проверяем код ошибки драйвера, а не текст сообщения: текст зависит от локали
// сервера и от имени индекса и сломался бы при первом же переименовании.
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation
}
