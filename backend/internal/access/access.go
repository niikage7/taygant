// Package access определяет права участника на действия внутри проекта.
// Правило одно и то же для всех сущностей проекта (задач, спринтов, вех и т.д.),
// поэтому оно вынесено сюда, а не продублировано в каждом сервисе.
package access

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"taygant_backend/internal/models"
	"taygant_backend/internal/service"
)

// Membership — проект вместе с записью участника, снятые одним запросом.
// Почти каждый хендлер, работающий с конкретным проектом, нуждается в обоих сразу.
type Membership struct {
	Project models.Project
	Member  models.ProjectMember
}

// Load проверяет, что проект существует и что пользователь входит в его команду.
//
// Пользователю, который не состоит в проекте, отдаём 404, а не 403: иначе сам факт
// отличия кода ответа подтверждал бы существование projectId, до которого у него
// нет доступа.
func Load(ctx context.Context, db *gorm.DB, projectID, userID uuid.UUID) (Membership, error) {
	var project models.Project
	if err := db.WithContext(ctx).First(&project, "id = ?", projectID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Membership{}, service.ErrNotFound
		}
		return Membership{}, err
	}

	var member models.ProjectMember
	err := db.WithContext(ctx).
		Where("project_id = ? AND user_id = ?", projectID, userID).
		First(&member).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return Membership{}, service.ErrNotFound
	case err != nil:
		return Membership{}, err
	}

	return Membership{Project: project, Member: member}, nil
}

// CanEdit сообщает, что участник может создавать и редактировать задачи, спринты,
// вехи, чек-лист и связи — всё, что описывает план проекта.
func (m Membership) CanEdit() bool {
	return m.Member.AccessLevel == models.AccessLevelFull || m.Member.AccessLevel == models.AccessLevelEdit
}

// CanManage сообщает, что участник может менять параметры проекта и состав команды.
// Именно это спецификация называет уровнем full: "управление проектом и командой".
func (m Membership) CanManage() bool {
	return m.Member.AccessLevel == models.AccessLevelFull
}

// RequireEdit возвращает ErrForbidden, если у участника только просмотр.
func (m Membership) RequireEdit() error {
	if !m.CanEdit() {
		return service.ErrForbidden
	}
	return nil
}

// RequireManage возвращает ErrForbidden, если участнику не выдан полный доступ.
func (m Membership) RequireManage() error {
	if !m.CanManage() {
		return service.ErrForbidden
	}
	return nil
}
