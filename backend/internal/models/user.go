package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User — пользователь платформы.
//
// Соответствует схеме User из api-spec.yml плюс PasswordHash, которого в ответах
// нет и быть не должно. Роли здесь нет намеренно: один и тот же человек может быть
// руководителем в одном проекте и наблюдателем в другом — это хранит ProjectMember.
type User struct {
	Model

	// Email — логин пользователя. Уникальность обеспечивает ограничение PostgreSQL,
	// а не только проверка в сервисе: гонка двух одновременных регистраций иначе
	// создала бы двух пользователей с одним адресом.
	//
	// Индекс уникален по всей таблице, включая мягко удалённых: повторно
	// зарегистрироваться на освободившийся адрес нельзя. Для MVP это приемлемо.
	Email string `gorm:"type:varchar(320);not null;uniqueIndex"`

	// PasswordHash — bcrypt-хеш пароля. Открытый пароль не хранится и не логируется.
	PasswordHash string `gorm:"type:text;not null"`

	FullName   string  `gorm:"type:varchar(255);not null"`
	Department string  `gorm:"type:varchar(255)"`
	Position   string  `gorm:"type:varchar(255)"`
	AvatarURL  *string `gorm:"type:text"`

	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// TableName фиксирует имя таблицы, чтобы оно не зависело от правил
// множественного числа GORM.
func (User) TableName() string { return "users" }

// NewUserID выдаёт идентификатор пользователя заранее — нужно сервисам,
// которые собирают пользователя и связанные записи в одной транзакции.
func NewUserID() uuid.UUID { return uuid.New() }
