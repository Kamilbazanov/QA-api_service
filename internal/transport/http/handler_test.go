package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"QA-api_service/internal/models"
	"QA-api_service/internal/storage"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestQuestionLifecycle(t *testing.T) {
	router, store := setupTestRouter(t)

	createQuestionPayload := `{"text":"Что такое Go?"}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/questions", bytes.NewBufferString(createQuestionPayload))
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	var question models.Question
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&question))
	require.NotZero(t, question.ID)

	createAnswerPayload := `{"user_id":"user-123","text":"Это язык программирования"}` // #nosec G101 тестовые данные
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/questions/%d/answers", question.ID), bytes.NewBufferString(createAnswerPayload))
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	var answer models.Answer
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&answer))
	require.NotZero(t, answer.ID)
	require.Equal(t, question.ID, answer.QuestionID)

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/questions/%d", question.ID), nil)
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var questionWithAnswers models.Question
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&questionWithAnswers))
	require.Len(t, questionWithAnswers.Answers, 1)
	require.Equal(t, answer.ID, questionWithAnswers.Answers[0].ID)

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/questions/%d", question.ID), nil)
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusNoContent, rec.Code)

	_, err := store.GetAnswer(context.Background(), answer.ID)
	require.Error(t, err)
	require.True(t, storage.IsNotFound(err))
}

func setupTestRouter(t *testing.T) (*gin.Engine, *storage.Storage) {
	t.Helper()

	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	require.NoError(t, db.Exec("PRAGMA foreign_keys = ON").Error)

	require.NoError(t, db.AutoMigrate(&models.Question{}, &models.Answer{}))

	store := storage.New(db)
	logger := slog.New(slogDiscardHandler{})
	handler := NewHandler(store, logger)

	router := gin.New()
	handler.RegisterRoutes(router)

	return router, store
}

type slogDiscardHandler struct{}

func (s slogDiscardHandler) Enabled(context.Context, slog.Level) bool { return false }
func (s slogDiscardHandler) Handle(context.Context, slog.Record) error {
	return nil
}
func (s slogDiscardHandler) WithAttrs([]slog.Attr) slog.Handler { return s }
func (s slogDiscardHandler) WithGroup(string) slog.Handler      { return s }
