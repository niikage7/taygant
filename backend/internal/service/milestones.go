package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"taygant_backend/internal/models"
)

// Milestones — контрольные точки (вехи) проекта.
type Milestones struct {
	db *gorm.DB
}

// NewMilestones собирает сервис вех.
func NewMilestones(db *gorm.DB) *Milestones {
	return &Milestones{db: db}
}

// ProjectID возвращает проект, которому принадлежит веха.
func (s *Milestones) ProjectID(ctx context.Context, milestoneID uuid.UUID) (uuid.UUID, error) {
	return projectIDOfMilestone(ctx, s.db, milestoneID)
}

// List возвращает вехи проекта в хронологическом порядке с вычисленным статусом.
//
// Статус вехи (done/current/planned/final) не хранится в таблице: спецификация
// не принимает его при создании, он всегда выводится из плановых дат — так же,
// как overdue/blocked у задач.
func (s *Milestones) List(ctx context.Context, projectID uuid.UUID) ([]models.Milestone, error) {
	var milestones []models.Milestone
	err := s.db.WithContext(ctx).Where("project_id = ?", projectID).Order("planned_date").Find(&milestones).Error
	if err != nil {
		return nil, fmt.Errorf("выбрать вехи: %w", err)
	}
	deriveMilestoneStatuses(milestones)
	return milestones, nil
}

// deriveMilestoneStatuses проставляет статус каждой вехи по месту, на основе
// ActualDate и относительного порядка плановых дат среди ещё не достигнутых вех.
func deriveMilestoneStatuses(milestones []models.Milestone) {
	pending := make([]int, 0, len(milestones))
	for i := range milestones {
		if milestones[i].ActualDate != nil {
			milestones[i].Status = models.MilestoneStatusDone
			continue
		}
		pending = append(pending, i)
	}
	sort.Slice(pending, func(a, b int) bool {
		return milestones[pending[a]].PlannedDate.Before(milestones[pending[b]].PlannedDate)
	})
	today := models.Today()
	for rank, idx := range pending {
		switch {
		case rank == 0:
			milestones[idx].Status = models.MilestoneStatusCurrent
		case rank == len(pending)-1:
			milestones[idx].Status = models.MilestoneStatusFinal
		default:
			milestones[idx].Status = models.MilestoneStatusPlanned
		}
		if milestones[idx].PlannedDate.Before(today) {
			risk := milestones[idx].PlannedDate.DaysUntil(today)
			milestones[idx].RiskDays = &risk
		}
	}
}

// MilestoneInput — данные для создания или изменения вехи.
type MilestoneInput struct {
	Code        string
	Name        string
	PlannedDate models.Date
}

func (in MilestoneInput) validate() error {
	if strings.TrimSpace(in.Name) == "" {
		return Invalid("укажите название контрольной точки")
	}
	if in.PlannedDate.IsZero() {
		return Invalid("укажите плановую дату контрольной точки")
	}
	return nil
}

// Create добавляет веху проекта.
func (s *Milestones) Create(ctx context.Context, projectID uuid.UUID, in MilestoneInput) (models.Milestone, error) {
	if err := in.validate(); err != nil {
		return models.Milestone{}, err
	}
	milestone := models.Milestone{
		ProjectID:   projectID,
		Code:        strings.TrimSpace(in.Code),
		Name:        strings.TrimSpace(in.Name),
		PlannedDate: in.PlannedDate,
		// Status пересчитывается при каждом чтении (см. deriveMilestoneStatuses);
		// колонка NOT NULL, поэтому здесь нужно любое валидное значение-заглушка.
		Status: models.MilestoneStatusPlanned,
	}
	if err := s.db.WithContext(ctx).Create(&milestone).Error; err != nil {
		return models.Milestone{}, fmt.Errorf("создать веху: %w", err)
	}
	return s.Get(ctx, milestone.ID)
}

// Update изменяет название, код или плановую дату вехи.
func (s *Milestones) Update(ctx context.Context, milestoneID uuid.UUID, in MilestoneInput) (models.Milestone, error) {
	if err := in.validate(); err != nil {
		return models.Milestone{}, err
	}
	milestone, err := s.get(ctx, milestoneID)
	if err != nil {
		return models.Milestone{}, err
	}
	milestone.Code = strings.TrimSpace(in.Code)
	milestone.Name = strings.TrimSpace(in.Name)
	milestone.PlannedDate = in.PlannedDate

	if err := s.db.WithContext(ctx).Save(&milestone).Error; err != nil {
		return models.Milestone{}, fmt.Errorf("сохранить веху: %w", err)
	}
	return s.Get(ctx, milestoneID)
}

// Delete удаляет веху из проекта.
func (s *Milestones) Delete(ctx context.Context, milestoneID uuid.UUID) error {
	res := s.db.WithContext(ctx).Delete(&models.Milestone{}, "id = ?", milestoneID)
	if res.Error != nil {
		return fmt.Errorf("удалить веху: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// Get возвращает веху вместе с вычисленным статусом: для одной вехи он зависит
// от плановых дат остальных вех проекта, поэтому пересчитывается по всему списку.
func (s *Milestones) Get(ctx context.Context, milestoneID uuid.UUID) (models.Milestone, error) {
	milestone, err := s.get(ctx, milestoneID)
	if err != nil {
		return models.Milestone{}, err
	}
	all, err := s.List(ctx, milestone.ProjectID)
	if err != nil {
		return models.Milestone{}, err
	}
	for _, m := range all {
		if m.ID == milestoneID {
			return m, nil
		}
	}
	return milestone, nil
}

func (s *Milestones) get(ctx context.Context, milestoneID uuid.UUID) (models.Milestone, error) {
	var milestone models.Milestone
	err := s.db.WithContext(ctx).First(&milestone, "id = ?", milestoneID).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return models.Milestone{}, ErrNotFound
	case err != nil:
		return models.Milestone{}, fmt.Errorf("найти веху: %w", err)
	}
	return milestone, nil
}
