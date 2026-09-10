package service

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"gorm.io/gorm"

	"taygant_backend/internal/models"
)

// usersListLimit ограничивает выдачу справочника пользователей.
// Список используется в выпадающих меню выбора ответственного, где длинная
// выдача бесполезна: нужное имя ищут поиском, а не прокруткой.
const usersListLimit = 100

// Users — справочник пользователей платформы.
type Users struct {
	db *gorm.DB
}

// NewUsers собирает сервис пользователей.
func NewUsers(db *gorm.DB) *Users {
	return &Users{db: db}
}

// List возвращает пользователей, отфильтрованных по подстроке имени или email.
func (s *Users) List(ctx context.Context, search string) ([]models.User, error) {
	query := s.db.WithContext(ctx).Model(&models.User{}).Order("full_name")

	if search = strings.TrimSpace(search); search != "" {
		// PostgreSQL отвергает строки с битой кодировкой на уровне соединения,
		// и такой запрос упал бы как 500. Клиент, приславший не-UTF-8, получит
		// понятную ошибку ввода.
		if !utf8.ValidString(search) {
			return nil, Invalid("строка поиска должна быть в кодировке UTF-8")
		}

		// ILIKE вместо LIKE: поиск по имени должен находить «волков» и «Волков».
		// Спецсимволы шаблона экранируем, иначе введённый пользователем "%"
		// превратил бы запрос в выборку всех строк.
		pattern := "%" + escapeLikePattern(search) + "%"
		query = query.Where("full_name ILIKE ? OR email ILIKE ?", pattern, pattern)
	}

	var users []models.User
	if err := query.Limit(usersListLimit).Find(&users).Error; err != nil {
		return nil, fmt.Errorf("выбрать пользователей: %w", err)
	}
	return users, nil
}

// escapeLikePattern обезвреживает символы шаблона LIKE во вводе пользователя.
func escapeLikePattern(s string) string {
	replacer := strings.NewReplacer(`\`, `\`, `%`, `\%`, `_`, `\_`)
	return replacer.Replace(s)
}
