package models

import (
	"time"

	"github.com/google/uuid"
)

// MCPToken — личный ключ, которым нейронка пользователя (Claude Code, Cursor…)
// подключается к MCP-серверу приложения и действует от его имени.
//
// Отдельная сущность, а не JWT: access-токен живёт сутки и нужен браузеру, а
// ключ ассистента вписывают в настройки один раз — он должен жить долго, но
// отзываться по первому требованию. Отозвать JWT нельзя, строку в таблице — можно.
type MCPToken struct {
	Model

	UserID uuid.UUID `gorm:"type:uuid;not null;index"`
	User   *User     `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`

	// Name — подпись, по которой человек отличает ключи («Claude Code, ноутбук»).
	// Попадает в журнал задачи рядом с пометкой «через ассистента».
	Name string `gorm:"type:varchar(120);not null"`

	// Hash — SHA-256 от ключа в hex. Сам ключ не хранится: его показывают
	// один раз при выпуске, и утечка таблицы не даёт доступа к аккаунтам.
	Hash string `gorm:"type:char(64);not null;uniqueIndex"`
	// Hint — последние символы ключа, чтобы узнать его в списке, не храня целиком.
	Hint string `gorm:"type:varchar(8);not null"`

	// LastUsedAt — когда ключом пользовались в последний раз; помогает понять,
	// какой ключ забыт и его пора отозвать.
	LastUsedAt *time.Time
}

// TableName фиксирует имя таблицы.
func (MCPToken) TableName() string { return "mcp_tokens" }
