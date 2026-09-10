package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"taygant_backend/internal/models"
)

// Sprints — спринты и релизы внутри проекта.
type Sprints struct {
	db *gorm.DB
}

// NewSprints собирает сервис спринтов.
func NewSprints(db *gorm.DB) *Sprints {
	return &Sprints{db: db}
}

// ProjectID возвращает проект, которому принадлежит спринт.
func (s *Sprints) ProjectID(ctx context.Context, sprintID uuid.UUID) (uuid.UUID, error) {
	return projectIDOfSprint(ctx, s.db, sprintID)
}

// List возвращает спринты проекта.
func (s *Sprints) List(ctx context.Context, projectID uuid.UUID) ([]models.Sprint, error) {
	var sprints []models.Sprint
	err := s.db.WithContext(ctx).Where("project_id = ?", projectID).Order("start_date").Find(&sprints).Error
	if err != nil {
		return nil, fmt.Errorf("выбрать спринты: %w", err)
	}
	return sprints, nil
}

// SprintInput — данные для создания или изменения спринта.
type SprintInput struct {
	Name           string
	ReleaseVersion string
	StartDate      models.Date
	EndDate        models.Date
}

func (in SprintInput) validate() error {
	if strings.TrimSpace(in.Name) == "" {
		return Invalid("укажите название спринта")
	}
	if err := checkMaxLen("название спринта", strings.TrimSpace(in.Name), 255); err != nil {
		return err
	}
	if err := checkMaxLen("версия релиза", strings.TrimSpace(in.ReleaseVersion), 50); err != nil {
		return err
	}
	if in.StartDate.IsZero() || in.EndDate.IsZero() {
		return Invalid("укажите сроки спринта")
	}
	if in.EndDate.Before(in.StartDate) {
		return Invalid("дата окончания спринта не может быть раньше даты начала")
	}
	return nil
}

// Create создаёт спринт в проекте.
func (s *Sprints) Create(ctx context.Context, projectID uuid.UUID, in SprintInput) (models.Sprint, error) {
	if err := in.validate(); err != nil {
		return models.Sprint{}, err
	}
	sprint := models.Sprint{
		ProjectID:      projectID,
		Name:           strings.TrimSpace(in.Name),
		ReleaseVersion: strings.TrimSpace(in.ReleaseVersion),
		StartDate:      in.StartDate,
		EndDate:        in.EndDate,
	}
	if err := s.db.WithContext(ctx).Create(&sprint).Error; err != nil {
		return models.Sprint{}, fmt.Errorf("создать спринт: %w", err)
	}
	return sprint, nil
}

// Update изменяет название, версию релиза или даты спринта.
func (s *Sprints) Update(ctx context.Context, sprintID uuid.UUID, in SprintInput) (models.Sprint, error) {
	if err := in.validate(); err != nil {
		return models.Sprint{}, err
	}
	sprint, err := s.get(ctx, sprintID)
	if err != nil {
		return models.Sprint{}, err
	}
	sprint.Name = strings.TrimSpace(in.Name)
	sprint.ReleaseVersion = strings.TrimSpace(in.ReleaseVersion)
	sprint.StartDate = in.StartDate
	sprint.EndDate = in.EndDate

	if err := s.db.WithContext(ctx).Save(&sprint).Error; err != nil {
		return models.Sprint{}, fmt.Errorf("сохранить спринт: %w", err)
	}
	return sprint, nil
}

// Delete удаляет спринт. Связанные задачи остаются в проекте: внешний ключ
// Task.SprintID настроен на SET NULL, а не CASCADE.
func (s *Sprints) Delete(ctx context.Context, sprintID uuid.UUID) error {
	res := s.db.WithContext(ctx).Delete(&models.Sprint{}, "id = ?", sprintID)
	if res.Error != nil {
		return fmt.Errorf("удалить спринт: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Sprints) get(ctx context.Context, sprintID uuid.UUID) (models.Sprint, error) {
	var sprint models.Sprint
	err := s.db.WithContext(ctx).First(&sprint, "id = ?", sprintID).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return models.Sprint{}, ErrNotFound
	case err != nil:
		return models.Sprint{}, fmt.Errorf("найти спринт: %w", err)
	}
	return sprint, nil
}
