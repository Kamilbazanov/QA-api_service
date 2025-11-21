package storage

import (
	"context"
	"errors"

	"QA-api_service/internal/models"

	"gorm.io/gorm"
)

type Storage struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Storage {
	return &Storage{db: db}
}

func (s *Storage) CreateQuestion(ctx context.Context, question *models.Question) error {
	return s.db.WithContext(ctx).Create(question).Error
}

func (s *Storage) ListQuestions(ctx context.Context) ([]models.Question, error) {
	var questions []models.Question
	if err := s.db.WithContext(ctx).Order("created_at DESC").Find(&questions).Error; err != nil {
		return nil, err
	}
	return questions, nil
}

func (s *Storage) GetQuestionWithAnswers(ctx context.Context, id uint) (*models.Question, error) {
	var question models.Question
	err := s.db.WithContext(ctx).
		Preload("Answers", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC")
		}).
		First(&question, id).Error
	if err != nil {
		return nil, err
	}
	return &question, nil
}

func (s *Storage) DeleteQuestion(ctx context.Context, id uint) error {
	return s.db.WithContext(ctx).Delete(&models.Question{}, id).Error
}

func (s *Storage) CreateAnswer(ctx context.Context, answer *models.Answer) error {
	return s.db.WithContext(ctx).Create(answer).Error
}

func (s *Storage) GetAnswer(ctx context.Context, id uint) (*models.Answer, error) {
	var answer models.Answer
	if err := s.db.WithContext(ctx).First(&answer, id).Error; err != nil {
		return nil, err
	}
	return &answer, nil
}

func (s *Storage) DeleteAnswer(ctx context.Context, id uint) error {
	return s.db.WithContext(ctx).Delete(&models.Answer{}, id).Error
}

func (s *Storage) QuestionExists(ctx context.Context, id uint) error {
	result := s.db.WithContext(ctx).First(&models.Question{}, id)
	return result.Error
}

func IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
