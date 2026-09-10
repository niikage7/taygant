package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"taygant_backend/internal/models"
)

// Members — состав команды проекта.
type Members struct {
	db *gorm.DB
}

// NewMembers собирает сервис участников.
func NewMembers(db *gorm.DB) *Members {
	return &Members{db: db}
}

// List возвращает команду проекта.
func (s *Members) List(ctx context.Context, projectID uuid.UUID) ([]models.ProjectMember, error) {
	var members []models.ProjectMember
	err := s.db.WithContext(ctx).
		Preload("User").
		Where("project_id = ?", projectID).
		Order("created_at").
		Find(&members).Error
	if err != nil {
		return nil, fmt.Errorf("выбрать участников проекта: %w", err)
	}
	return members, nil
}

// Add добавляет существующего пользователя платформы в команду проекта.
func (s *Members) Add(ctx context.Context, projectID uuid.UUID, in MemberInput) (models.ProjectMember, error) {
	if !in.ProjectRole.Valid() {
		return models.ProjectMember{}, Invalid("недопустимая роль участника %q", in.ProjectRole)
	}
	if err := checkMaxLen("роль в спринте", in.SprintRole, 120); err != nil {
		return models.ProjectMember{}, err
	}
	accessLevel := in.AccessLevel
	if accessLevel == "" {
		accessLevel = models.AccessLevelEdit
	}
	if !accessLevel.Valid() {
		return models.ProjectMember{}, Invalid("недопустимый уровень доступа %q", accessLevel)
	}

	member := models.ProjectMember{
		ProjectID:        projectID,
		UserID:           in.UserID,
		ProjectRole:      in.ProjectRole,
		SprintRole:       in.SprintRole,
		AccessLevel:      accessLevel,
		WeeklyHoursLimit: 40,
	}
	if err := s.db.WithContext(ctx).Create(&member).Error; err != nil {
		switch {
		case isUniqueViolation(err):
			return models.ProjectMember{}, fmt.Errorf("%w: пользователь уже состоит в проекте", ErrConflict)
		case isForeignKeyViolation(err):
			return models.ProjectMember{}, Invalid("пользователь не найден")
		default:
			return models.ProjectMember{}, fmt.Errorf("добавить участника: %w", err)
		}
	}

	return s.get(ctx, member.ID)
}

// MemberUpdateInput — поля участника, доступные для частичного обновления.
type MemberUpdateInput struct {
	ProjectRole *models.ProjectRole
	SprintRole  *string
	AccessLevel *models.AccessLevel
}

// Update меняет роль и/или уровень доступа участника. memberID ищется строго
// внутри projectID — иначе участник чужого проекта, случайно совпавший по
// memberID, можно было бы отредактировать через свой projectId в URL.
func (s *Members) Update(ctx context.Context, projectID, memberID uuid.UUID, in MemberUpdateInput) (models.ProjectMember, error) {
	member, err := s.getInProject(ctx, projectID, memberID)
	if err != nil {
		return models.ProjectMember{}, err
	}

	if in.ProjectRole != nil {
		if !in.ProjectRole.Valid() {
			return models.ProjectMember{}, Invalid("недопустимая роль участника %q", *in.ProjectRole)
		}
		member.ProjectRole = *in.ProjectRole
	}
	if in.SprintRole != nil {
		if err := checkMaxLen("роль в спринте", *in.SprintRole, 120); err != nil {
			return models.ProjectMember{}, err
		}
		member.SprintRole = *in.SprintRole
	}
	if in.AccessLevel != nil {
		if !in.AccessLevel.Valid() {
			return models.ProjectMember{}, Invalid("недопустимый уровень доступа %q", *in.AccessLevel)
		}
		if member.AccessLevel == models.AccessLevelFull && *in.AccessLevel != models.AccessLevelFull {
			if err := s.ensureNotLastManager(ctx, member.ProjectID, member.ID); err != nil {
				return models.ProjectMember{}, err
			}
		}
		member.AccessLevel = *in.AccessLevel
	}

	if err := s.db.WithContext(ctx).Save(&member).Error; err != nil {
		return models.ProjectMember{}, fmt.Errorf("сохранить участника: %w", err)
	}
	return s.get(ctx, memberID)
}

// Remove исключает участника из команды проекта. memberID ищется строго
// внутри projectID — та же защита от подмены проекта в URL, что и в Update.
func (s *Members) Remove(ctx context.Context, projectID, memberID uuid.UUID) error {
	member, err := s.getInProject(ctx, projectID, memberID)
	if err != nil {
		return err
	}
	if member.AccessLevel == models.AccessLevelFull {
		if err := s.ensureNotLastManager(ctx, member.ProjectID, member.ID); err != nil {
			return err
		}
	}

	res := s.db.WithContext(ctx).Delete(&models.ProjectMember{}, "id = ? AND project_id = ?", memberID, projectID)
	if res.Error != nil {
		return fmt.Errorf("удалить участника: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// ensureNotLastManager не даёт понизить или удалить последнего участника
// с полным доступом: иначе управлять проектом стало бы физически некому.
func (s *Members) ensureNotLastManager(ctx context.Context, projectID, excludeMemberID uuid.UUID) error {
	var count int64
	err := s.db.WithContext(ctx).Model(&models.ProjectMember{}).
		Where("project_id = ? AND access_level = ? AND id <> ?", projectID, models.AccessLevelFull, excludeMemberID).
		Count(&count).Error
	if err != nil {
		return fmt.Errorf("проверить владельцев проекта: %w", err)
	}
	if count == 0 {
		return Invalid("в проекте должен остаться хотя бы один участник с полным доступом")
	}
	return nil
}

func (s *Members) get(ctx context.Context, memberID uuid.UUID) (models.ProjectMember, error) {
	var member models.ProjectMember
	err := s.db.WithContext(ctx).Preload("User").First(&member, "id = ?", memberID).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return models.ProjectMember{}, ErrNotFound
	case err != nil:
		return models.ProjectMember{}, fmt.Errorf("найти участника: %w", err)
	}
	return member, nil
}

// getInProject ищет участника по id И project_id одновременно: без этого
// пара (свой projectId в URL, чужой memberID) успешно находила бы участника
// другого проекта.
func (s *Members) getInProject(ctx context.Context, projectID, memberID uuid.UUID) (models.ProjectMember, error) {
	var member models.ProjectMember
	err := s.db.WithContext(ctx).Preload("User").First(&member, "id = ? AND project_id = ?", memberID, projectID).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return models.ProjectMember{}, ErrNotFound
	case err != nil:
		return models.ProjectMember{}, fmt.Errorf("найти участника: %w", err)
	}
	return member, nil
}
