// Package models описывает таблицы PostgreSQL и структуру доменных данных.
//
// Модели намеренно не содержат json-тегов: они не являются представлением API.
// Ответы собираются отдельными DTO в internal/api — иначе PasswordHash и другие
// внутренние поля рано или поздно утекут в ответ вместе с забытым полем.
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Model — общие поля всех сущностей.
//
// Идентификаторы генерируются приложением, а не БД: сервисы собирают связанные
// объекты (проект + участники, задача + связи) в одной транзакции и должны знать
// id родителя до вставки.
type Model struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// BeforeCreate проставляет идентификатор, если вызывающий код его не задал.
func (m *Model) BeforeCreate(*gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}
