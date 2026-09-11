package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"taygant_backend/internal/models"
)

// Comments — комментарии команды к задачам.
type Comments struct {
	db *gorm.DB
}

// NewComments собирает сервис комментариев.
func NewComments(db *gorm.DB) *Comments {
	return &Comments{db: db}
}

// List возвращает комментарии задачи в хронологическом порядке.
func (s *Comments) List(ctx context.Context, taskID uuid.UUID) ([]models.TaskComment, error) {
	var comments []models.TaskComment
	err := s.db.WithContext(ctx).
		Preload("Author").
		Where("task_id = ?", taskID).
		Order("created_at").
		Find(&comments).Error
	if err != nil {
		return nil, fmt.Errorf("выбрать комментарии: %w", err)
	}
	return comments, nil
}

// Add публикует комментарий от имени текущего пользователя и отмечает это
// в журнале задачи — так история задачи остаётся полной картиной активности,
// а не только правок полей.
func (s *Comments) Add(ctx context.Context, taskID, authorID uuid.UUID, text string) (models.TaskComment, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return models.TaskComment{}, Invalid("комментарий не может быть пустым")
	}

	comment := models.TaskComment{TaskID: taskID, AuthorID: authorID, Text: text}
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&comment).Error; err != nil {
			return fmt.Errorf("создать комментарий: %w", err)
		}
		return recordHistory(tx, taskID, authorID, models.HistoryActionCommented, nil, nil, nil)
	})
	if err != nil {
		return models.TaskComment{}, err
	}

	comment.Author = &models.User{}
	if err := s.db.WithContext(ctx).First(comment.Author, "id = ?", authorID).Error; err != nil {
		return models.TaskComment{}, fmt.Errorf("найти автора комментария: %w", err)
	}
	return comment, nil
}
