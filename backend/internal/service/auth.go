package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"taygant_backend/internal/auth"
	"taygant_backend/internal/models"
)

// Auth — регистрация, вход и обновление токенов.
type Auth struct {
	db     *gorm.DB
	tokens *auth.TokenIssuer
}

// NewAuth собирает сервис аутентификации.
func NewAuth(db *gorm.DB, tokens *auth.TokenIssuer) *Auth {
	return &Auth{db: db, tokens: tokens}
}

// RegisterInput — данные для создания учётной записи.
type RegisterInput struct {
	Email      string
	Password   string
	FullName   string
	Department string
	Position   string
}

// TokenPair — выпущенные при входе токены.
type TokenPair struct {
	Access  string
	Refresh string
}

// Register создаёт пользователя и сразу выдаёт токены: после регистрации
// человек должен оказаться в приложении, а не на форме входа.
func (s *Auth) Register(ctx context.Context, in RegisterInput) (models.User, TokenPair, error) {
	email := normalizeEmail(in.Email)
	if email == "" || !strings.Contains(email, "@") {
		return models.User{}, TokenPair{}, Invalid("укажите корректный email")
	}
	if strings.TrimSpace(in.FullName) == "" {
		return models.User{}, TokenPair{}, Invalid("укажите имя пользователя")
	}
	if err := auth.ValidatePassword(in.Password); err != nil {
		return models.User{}, TokenPair{}, &ValidationError{Message: err.Error()}
	}

	hash, err := auth.HashPassword(in.Password)
	if err != nil {
		return models.User{}, TokenPair{}, err
	}

	user := models.User{
		Email:        email,
		PasswordHash: hash,
		FullName:     strings.TrimSpace(in.FullName),
		Department:   strings.TrimSpace(in.Department),
		Position:     strings.TrimSpace(in.Position),
	}

	if err := s.db.WithContext(ctx).Create(&user).Error; err != nil {
		// Занятый email отлавливаем по ошибке уникального индекса, а не проверкой
		// «есть ли такой» перед вставкой: две одновременные регистрации прошли бы
		// такую проверку обе.
		if isUniqueViolation(err) {
			return models.User{}, TokenPair{}, fmt.Errorf("%w: email уже зарегистрирован", ErrConflict)
		}
		return models.User{}, TokenPair{}, fmt.Errorf("создать пользователя: %w", err)
	}

	pair, err := s.issue(user.ID)
	if err != nil {
		return models.User{}, TokenPair{}, err
	}
	return user, pair, nil
}

// Login проверяет учётные данные и выдаёт пару токенов.
func (s *Auth) Login(ctx context.Context, email, password string) (models.User, TokenPair, error) {
	var user models.User
	err := s.db.WithContext(ctx).Where("email = ?", normalizeEmail(email)).First(&user).Error

	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		// Пароль всё равно «проверяем», чтобы ответ занимал столько же времени,
		// сколько при существующем пользователе: разница во времени ответа
		// позволяет перебором выяснить, какие адреса зарегистрированы.
		auth.VerifyPassword(dummyHash, password)
		return models.User{}, TokenPair{}, ErrInvalidCredentials
	case err != nil:
		return models.User{}, TokenPair{}, fmt.Errorf("найти пользователя: %w", err)
	}

	if !auth.VerifyPassword(user.PasswordHash, password) {
		return models.User{}, TokenPair{}, ErrInvalidCredentials
	}

	pair, err := s.issue(user.ID)
	if err != nil {
		return models.User{}, TokenPair{}, err
	}
	return user, pair, nil
}

// Refresh выдаёт новый access-токен по действующему refresh-токену.
func (s *Auth) Refresh(ctx context.Context, refreshToken string) (string, error) {
	userID, err := s.tokens.Parse(refreshToken, auth.KindRefresh)
	if err != nil {
		return "", ErrUnauthorized
	}

	// Пользователь мог быть удалён уже после выдачи токена — токен об этом
	// не знает, поэтому существование проверяем на каждом обновлении.
	if _, err := s.UserByID(ctx, userID); err != nil {
		return "", ErrUnauthorized
	}

	return s.tokens.Issue(userID, auth.KindAccess)
}

// UserByID возвращает пользователя по идентификатору из токена.
func (s *Auth) UserByID(ctx context.Context, id uuid.UUID) (models.User, error) {
	var user models.User
	err := s.db.WithContext(ctx).First(&user, "id = ?", id).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return models.User{}, ErrNotFound
	case err != nil:
		return models.User{}, fmt.Errorf("найти пользователя: %w", err)
	}
	return user, nil
}

func (s *Auth) issue(userID uuid.UUID) (TokenPair, error) {
	access, refresh, err := s.tokens.IssuePair(userID)
	if err != nil {
		return TokenPair{}, err
	}
	return TokenPair{Access: access, Refresh: refresh}, nil
}

// normalizeEmail приводит адрес к виду, в котором он лежит в БД:
// без пробелов по краям и в нижнем регистре, иначе "User@x.ru" и "user@x.ru"
// оказались бы разными учётными записями.
func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// dummyHash — заранее посчитанный bcrypt-хеш произвольной строки.
// Нужен только чтобы вход несуществующего пользователя стоил столько же
// времени, сколько вход существующего.
const dummyHash = "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"
