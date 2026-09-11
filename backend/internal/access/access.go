// Package access определяет права участника на действия внутри проекта.
//
// Модель прав: view — только чтение; edit — изменение статуса и сроков своих
// (где участник исполнитель) задач; full — полное управление планом проекта,
// его параметрами и составом команды. Общее правило вынесено сюда, чтобы не
// дублировать его в каждом сервисе.
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

// CanManage сообщает, что участник может создавать и редактировать план проекта
// (задачи, спринты, вехи, чек-лист, связи), менять параметры проекта и состав
// команды. Это уровень доступа full.
func (m Membership) CanManage() bool {
	return m.Member.AccessLevel == models.AccessLevelFull
}

// CanEditTask сообщает, что участник может изменить задачу: либо у него полный
// доступ, либо он исполнитель задачи с уровнем edit.
func (m Membership) CanEditTask(task models.Task) bool {
	if m.CanManage() {
		return true
	}
	if m.Member.AccessLevel != models.AccessLevelEdit {
		return false
	}
	return task.AssigneeID != nil && *task.AssigneeID == m.Member.UserID
}

// IsTaskRestricted сообщает, что участник редактирует задачу как исполнитель с
// уровнем edit, а не как full. Такому участнику разрешены только статус и сроки
// своей задачи; набор полей проверяет HTTP-слой, который видит тело запроса.
func (m Membership) IsTaskRestricted(task models.Task) bool {
	return !m.CanManage() && m.CanEditTask(task)
}

// RequireManage возвращает ErrForbidden, если участнику не выдан полный доступ.
func (m Membership) RequireManage() error {
	if !m.CanManage() {
		return service.ErrForbidden
	}
	return nil
}
