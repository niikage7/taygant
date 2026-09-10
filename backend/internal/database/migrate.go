package database

import (
	"fmt"

	"gorm.io/gorm"

	"taygant_backend/internal/models"
)

// schema перечисляет модели в порядке зависимостей: AutoMigrate создаёт внешние
// ключи сразу, поэтому таблица-цель должна существовать раньше ссылающейся на неё.
var schema = []any{
	&models.User{},
	&models.Project{},
	&models.ProjectMember{},
	&models.Sprint{},
	&models.Milestone{},
	&models.Task{},
	&models.TaskDependency{},
	&models.ChecklistItem{},
	&models.TaskComment{},
	&models.TaskHistoryEntry{},
}

// Migrate приводит схему БД в соответствие с моделями.
//
// AutoMigrate вместо версионированных SQL-миграций — осознанный выбор для MVP:
// схема ещё меняется, версионирование потребовало бы описывать каждое изменение
// дважды. AutoMigrate добавляет колонки и индексы, но не удаляет и не сужает их —
// значит, откатить неудачное изменение вручную придётся через psql.
func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(schema...); err != nil {
		return fmt.Errorf("применить схему: %w", err)
	}
	return nil
}
