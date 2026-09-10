package api

import (
	"github.com/google/uuid"

	"taygant_backend/internal/models"
)

// DTO повторяют схемы из api-spec.yml: имена полей в camelCase, набор полей —
// ровно тот, что описан в спецификации. Отдельный слой нужен, чтобы модель БД
// можно было менять, не ломая контракт, и чтобы служебные поля вроде
// PasswordHash физически не могли попасть в ответ.

// userDTO — схема User.
type userDTO struct {
	ID         uuid.UUID `json:"id"`
	FullName   string    `json:"fullName"`
	Email      string    `json:"email"`
	Department string    `json:"department"`
	Position   string    `json:"position"`
	AvatarURL  *string   `json:"avatarUrl"`
}

// newUserDTO собирает представление пользователя для ответа.
func newUserDTO(u models.User) userDTO {
	return userDTO{
		ID:         u.ID,
		FullName:   u.FullName,
		Email:      u.Email,
		Department: u.Department,
		Position:   u.Position,
		AvatarURL:  u.AvatarURL,
	}
}

// newUserDTOPtr возвращает nil для незаполненной ссылки на пользователя:
// у задачи может не быть ответственного, и подставлять пустой объект нельзя —
// фронт отличает «не назначен» именно по null.
func newUserDTOPtr(u *models.User) *userDTO {
	if u == nil || u.ID == uuid.Nil {
		return nil
	}
	dto := newUserDTO(*u)
	return &dto
}

// authTokensDTO — ответ /auth/login и /auth/register.
type authTokensDTO struct {
	AccessToken  string  `json:"accessToken"`
	RefreshToken string  `json:"refreshToken"`
	User         userDTO `json:"user"`
}

// accessTokenDTO — ответ /auth/refresh.
type accessTokenDTO struct {
	AccessToken string `json:"accessToken"`
}
