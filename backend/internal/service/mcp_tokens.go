package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"taygant_backend/internal/models"
)

const (
	// MCPTokenPrefix начинает каждый личный ключ. По нему ключ узнают глазами
	// и сканеры секретов в репозиториях, а сервер отбрасывает всё остальное
	// (например, JWT браузера) ещё до запроса к БД.
	MCPTokenPrefix = "tgn_"

	// mcpTokensPerUser — сколько ключей может держать один человек. Ключ
	// заводят на каждый редактор или ноутбук, десятка хватает с запасом, а
	// бесконечный список забытых ключей — это лишние двери в аккаунт.
	// Проверка мягкая: два одновременных запроса могут превысить лимит на
	// единицу, ради этого не стоит блокировать таблицу.
	mcpTokensPerUser = 10

	// mcpTokenTouchInterval — как часто обновлять LastUsedAt. Ассистент шлёт
	// по несколько запросов на каждый вопрос, писать в БД на каждый ради
	// точности до секунды незачем.
	mcpTokenTouchInterval = time.Minute
)

// MCPTokens — личные ключи для подключения нейронок по MCP.
type MCPTokens struct {
	db *gorm.DB
}

// NewMCPTokens собирает сервис ключей.
func NewMCPTokens(db *gorm.DB) *MCPTokens {
	return &MCPTokens{db: db}
}

// Issue выпускает новый ключ. Секрет возвращается только здесь: в БД лежит
// его хеш, и показать ключ повторно невозможно — потерянный ключ отзывают
// и выпускают заново.
func (s *MCPTokens) Issue(ctx context.Context, userID uuid.UUID, name string) (models.MCPToken, string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return models.MCPToken{}, "", Invalid("укажите название ключа — например, «Claude Code на ноутбуке»")
	}
	if err := checkMaxLen("название ключа", name, 120); err != nil {
		return models.MCPToken{}, "", err
	}

	var count int64
	if err := s.db.WithContext(ctx).Model(&models.MCPToken{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
		return models.MCPToken{}, "", fmt.Errorf("посчитать ключи пользователя: %w", err)
	}
	if count >= mcpTokensPerUser {
		return models.MCPToken{}, "", Invalid("у вас уже %d ключей — отзовите ненужный, прежде чем создавать новый", count)
	}

	secret := newMCPSecret()
	token := models.MCPToken{
		UserID: userID,
		Name:   name,
		Hash:   hashMCPSecret(secret),
		Hint:   secret[len(secret)-4:],
	}
	if err := s.db.WithContext(ctx).Create(&token).Error; err != nil {
		return models.MCPToken{}, "", fmt.Errorf("сохранить ключ: %w", err)
	}
	return token, secret, nil
}

// List возвращает ключи пользователя, новые сверху.
func (s *MCPTokens) List(ctx context.Context, userID uuid.UUID) ([]models.MCPToken, error) {
	var tokens []models.MCPToken
	err := s.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Find(&tokens).Error
	if err != nil {
		return nil, fmt.Errorf("выбрать ключи: %w", err)
	}
	return tokens, nil
}

// Revoke удаляет ключ пользователя. Чужой ключ неотличим от несуществующего:
// ErrNotFound в обоих случаях, чтобы нельзя было перебирать чужие id.
func (s *MCPTokens) Revoke(ctx context.Context, userID, tokenID uuid.UUID) error {
	res := s.db.WithContext(ctx).Where("id = ? AND user_id = ?", tokenID, userID).Delete(&models.MCPToken{})
	if res.Error != nil {
		return fmt.Errorf("отозвать ключ: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// Authenticate находит ключ по секрету вместе с владельцем. Удалённый
// пользователь (мягкое удаление) не проходит: Preload его не подгружает.
//
// Сравнение идёт по хешу через индекс, а не перебором: время ответа зависит
// от хеша, а не от совпадающих символов секрета, поэтому подбирать ключ по
// таймингу бесполезно.
func (s *MCPTokens) Authenticate(ctx context.Context, secret string) (models.MCPToken, error) {
	if !strings.HasPrefix(secret, MCPTokenPrefix) {
		return models.MCPToken{}, ErrUnauthorized
	}

	var token models.MCPToken
	err := s.db.WithContext(ctx).Preload("User").Where("hash = ?", hashMCPSecret(secret)).First(&token).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return models.MCPToken{}, ErrUnauthorized
	case err != nil:
		return models.MCPToken{}, fmt.Errorf("найти ключ: %w", err)
	}
	if token.User == nil {
		return models.MCPToken{}, ErrUnauthorized
	}

	now := time.Now().UTC()
	if token.LastUsedAt == nil || now.Sub(*token.LastUsedAt) >= mcpTokenTouchInterval {
		err := s.db.WithContext(ctx).Model(&models.MCPToken{}).Where("id = ?", token.ID).Update("last_used_at", now).Error
		if err != nil {
			return models.MCPToken{}, fmt.Errorf("отметить использование ключа: %w", err)
		}
		token.LastUsedAt = &now
	}
	return token, nil
}

// newMCPSecret собирает ключ: префикс и 128 случайных бит в base32
// (crypto/rand.Text) — столько же энтропии, сколько у UUID, но без дефисов,
// и ключ целиком выделяется двойным щелчком.
func newMCPSecret() string {
	return MCPTokenPrefix + rand.Text()
}

// hashMCPSecret — SHA-256 в hex. Соль не нужна: ключ случайный и длинный,
// словарём его не подобрать, а без соли хеш можно искать индексом.
func hashMCPSecret(secret string) string {
	sum := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(sum[:])
}
