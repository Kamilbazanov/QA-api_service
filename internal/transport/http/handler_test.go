package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	stdhttp "net/http"
	"net/http/httptest"
	"testing"

	"QA-api_service/internal/models"
	"QA-api_service/internal/storage"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestQuestionLifecycle проверяет полный путь: создание вопроса, ответ, чтение и каскадное удаление.
func TestQuestionLifecycle(t *testing.T) {
	handler, store := setupTestHandler(t)

	// 1. Создаем вопрос.
	createQuestionPayload := `{"text":"Что такое Go?"}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(stdhttp.MethodPost, "/questions", bytes.NewBufferString(createQuestionPayload))
	handler.questionsRouter(rec, req)
	require.Equal(t, stdhttp.StatusCreated, rec.Code)

	var question models.Question
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&question))
	require.NotZero(t, question.ID)

	// 2. Добавляем ответ к вопросу.
	createAnswerPayload := `{"user_id":"user-123","text":"Это язык программирования"}` // #nosec G101 тестовые данные
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(stdhttp.MethodPost, fmt.Sprintf("/questions/%d/answers", question.ID), bytes.NewBufferString(createAnswerPayload))
	handler.questionsRouter(rec, req)
	require.Equal(t, stdhttp.StatusCreated, rec.Code)

	var answer models.Answer
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&answer))
	require.NotZero(t, answer.ID)
	require.Equal(t, question.ID, answer.QuestionID)

	// 3. Получаем вопрос с ответами и убеждаемся, что ответ присутствует.
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(stdhttp.MethodGet, fmt.Sprintf("/questions/%d", question.ID), nil)
	handler.questionsRouter(rec, req)
	require.Equal(t, stdhttp.StatusOK, rec.Code)

	var questionWithAnswers models.Question
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&questionWithAnswers))
	require.Len(t, questionWithAnswers.Answers, 1)
	require.Equal(t, answer.ID, questionWithAnswers.Answers[0].ID)

	// 4. Удаляем вопрос и проверяем, что ответ удалился каскадно.
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(stdhttp.MethodDelete, fmt.Sprintf("/questions/%d", question.ID), nil)
	handler.questionsRouter(rec, req)
	require.Equal(t, stdhttp.StatusNoContent, rec.Code)

	_, err := store.GetAnswer(context.Background(), answer.ID)
	require.Error(t, err)
	require.True(t, storage.IsNotFound(err))
}

// setupTestHandler разворачивает in-memory SQLite, чтобы тесты были быстрыми и повторяемыми.
func setupTestHandler(t *testing.T) (*Handler, *storage.Storage) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	// Включаем поддержку внешних ключей, чтобы каскадное удаление работало в SQLite так же, как в PostgreSQL.
	require.NoError(t, db.Exec("PRAGMA foreign_keys = ON").Error)

	require.NoError(t, db.AutoMigrate(&models.Question{}, &models.Answer{}))

	store := storage.New(db)
	logger := slog.New(slogDiscardHandler{})
	handler := NewHandler(store, logger)
	return handler, store
}

// slogDiscardHandler реализует минимальный Handler, который молча игнорирует сообщения.
type slogDiscardHandler struct{}

func (s slogDiscardHandler) Enabled(context.Context, slog.Level) bool { return false }
func (s slogDiscardHandler) Handle(context.Context, slog.Record) error {
	return nil
}
func (s slogDiscardHandler) WithAttrs([]slog.Attr) slog.Handler { return s }
func (s slogDiscardHandler) WithGroup(string) slog.Handler      { return s }


