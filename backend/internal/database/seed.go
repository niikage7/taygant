package database

import (
	"fmt"
	"log"

	"gorm.io/gorm"

	"taygant_backend/internal/auth"
	"taygant_backend/internal/models"
)

// DemoPassword — пароль всех демонстрационных учётных записей.
// Он же печатается в лог при первом запуске, чтобы не искать его в коде перед показом.
const DemoPassword = "demo1234"

// demoUsers — команда для демонстрационного проекта.
// Роли в проекте здесь не задаются: они относятся к проекту, а не к человеку,
// и назначаются при создании проекта в project_members.
var demoUsers = []models.User{
	{Email: "volkov@taygant.dev", FullName: "Александр Волков", Department: "Институт кибернетики", Position: "Lead Fullstack"},
	{Email: "orlova@taygant.dev", FullName: "Мария Орлова", Department: "Институт кибернетики", Position: "Бизнес-аналитик"},
	{Email: "petrov@taygant.dev", FullName: "Дмитрий Петров", Department: "Отдел разработки", Position: "Backend-разработчик"},
	{Email: "smirnova@taygant.dev", FullName: "Елена Смирнова", Department: "Отдел разработки", Position: "Frontend-разработчик"},
	{Email: "kuznetsov@taygant.dev", FullName: "Игорь Кузнецов", Department: "Дирекция", Position: "Куратор проекта"},
	{Email: "ivanova@taygant.dev", FullName: "Ольга Иванова", Department: "Отдел качества", Position: "QA-инженер"},
}

// Seed наполняет пустую БД демонстрационными пользователями.
//
// Идемпотентна: если в таблице уже есть хоть один пользователь, ничего не делает.
// Иначе повторный запуск приложения затирал бы данные, введённые на демо.
func Seed(db *gorm.DB) error {
	var existing int64
	if err := db.Model(&models.User{}).Count(&existing).Error; err != nil {
		return fmt.Errorf("проверить наличие пользователей: %w", err)
	}
	if existing > 0 {
		return nil
	}

	hash, err := auth.HashPassword(DemoPassword)
	if err != nil {
		return fmt.Errorf("подготовить пароль демо-пользователей: %w", err)
	}

	// Хеш считается один раз и переиспользуется: bcrypt намеренно медленный,
	// и шесть отдельных вычислений заметно задержали бы старт приложения.
	users := make([]models.User, len(demoUsers))
	for i, u := range demoUsers {
		u.PasswordHash = hash
		users[i] = u
	}

	if err := db.Create(&users).Error; err != nil {
		return fmt.Errorf("создать демо-пользователей: %w", err)
	}

	log.Printf("создано демо-пользователей: %d, пароль у всех — %q", len(users), DemoPassword)
	return nil
}
