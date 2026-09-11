package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// TokenKind различает назначения токенов.
type TokenKind string

const (
	// KindAccess — короткоживущий токен, которым авторизуются запросы к API.
	KindAccess TokenKind = "access"
	// KindRefresh — долгоживущий токен, принимаемый только эндпоинтом /auth/refresh.
	KindRefresh TokenKind = "refresh"
)

// ErrInvalidToken — токен просрочен, подписан чужим ключом, повреждён
// или предъявлен не по назначению.
var ErrInvalidToken = errors.New("недействительный токен")

// Claims — полезная нагрузка токена. Subject содержит идентификатор пользователя.
type Claims struct {
	// Kind не даёт использовать refresh-токен как access.
	//
	// Токены хранятся только у клиента, отозвать их сервер не может, поэтому
	// единственная защита от подмены — само содержимое: без этого поля
	// долгоживущий refresh открывал бы доступ ко всему API на месяц вперёд.
	Kind TokenKind `json:"typ"`
	jwt.RegisteredClaims
}

// TokenIssuer выпускает и проверяет токены одним общим секретом.
//
// Refresh-токены сознательно не хранятся в БД: таблица сессий понадобилась бы
// ради одной возможности — отзывать выданные токены при логауте. Плата за
// отказ от неё в том, что logout очищает токен только на клиенте, а сам токен
// остаётся действительным до истечения срока. Для MVP это осознанный размен.
type TokenIssuer struct {
	secret      []byte
	accessTTL   time.Duration
	refreshTTL  time.Duration
	signMethod  jwt.SigningMethod
	parseOption []jwt.ParserOption
}

// NewTokenIssuer собирает выпускающий токены объект.
func NewTokenIssuer(secret string, accessTTL, refreshTTL time.Duration) *TokenIssuer {
	return &TokenIssuer{
		secret:     []byte(secret),
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
		signMethod: jwt.SigningMethodHS256,
		// Ограничение алгоритма обязательно: без него токен с alg=none или
		// подменённым на асимметричный алгоритм прошёл бы проверку подписи.
		parseOption: []jwt.ParserOption{
			jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		},
	}
}

// AccessTTL возвращает время жизни access-токена.
func (t *TokenIssuer) AccessTTL() time.Duration { return t.accessTTL }

// Issue выпускает токен указанного вида для пользователя.
func (t *TokenIssuer) Issue(userID uuid.UUID, kind TokenKind) (string, error) {
	ttl := t.accessTTL
	if kind == KindRefresh {
		ttl = t.refreshTTL
	}

	now := time.Now().UTC()
	claims := Claims{
		Kind: kind,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}

	signed, err := jwt.NewWithClaims(t.signMethod, claims).SignedString(t.secret)
	if err != nil {
		return "", fmt.Errorf("подписать токен: %w", err)
	}
	return signed, nil
}

// IssuePair выпускает access- и refresh-токены для успешного входа.
func (t *TokenIssuer) IssuePair(userID uuid.UUID) (access, refresh string, err error) {
	if access, err = t.Issue(userID, KindAccess); err != nil {
		return "", "", err
	}
	if refresh, err = t.Issue(userID, KindRefresh); err != nil {
		return "", "", err
	}
	return access, refresh, nil
}

// Parse проверяет подпись и срок токена и убеждается, что вид токена — ожидаемый.
// Возвращает идентификатор пользователя из Subject.
func (t *TokenIssuer) Parse(token string, expected TokenKind) (uuid.UUID, error) {
	claims := &Claims{}
	parsed, err := jwt.ParseWithClaims(token, claims, func(*jwt.Token) (any, error) {
		return t.secret, nil
	}, t.parseOption...)
	if err != nil || !parsed.Valid {
		return uuid.Nil, ErrInvalidToken
	}
	if claims.Kind != expected {
		return uuid.Nil, ErrInvalidToken
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, ErrInvalidToken
	}
	return userID, nil
}
