package http

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"QA-api_service/internal/models"
	"QA-api_service/internal/storage"
)

// Handler агрегирует зависимости HTTP-слоя (хранилище и логгер).
type Handler struct {
	store  *storage.Storage
	logger *slog.Logger
}

// NewHandler инициализирует обработчик.
func NewHandler(store *storage.Storage, logger *slog.Logger) *Handler {
	return &Handler{store: store, logger: logger}
}

// RegisterRoutes вешает обработчики на стандартный ServeMux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/healthz", h.handleHealth)
	mux.HandleFunc("/questions", h.questionsRouter)
	mux.HandleFunc("/questions/", h.questionsRouter)
	mux.HandleFunc("/answers/", h.answersRouter)
	mux.HandleFunc("/answers", h.answersRouter)
}

// handleHealth нужен для быстрой проверки статуса контейнера.
func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// questionsRouter разбирает URL и перенаправляет на нужную операцию.
func (h *Handler) questionsRouter(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/questions")
	path = strings.Trim(path, "/")

	switch {
	case path == "":
		// Если дополнительных сегментов нет — работаем со списком вопросов.
		h.handleQuestionsCollection(w, r)
	default:
		// Делим путь на сегменты, чтобы узнать ID и вложенный ресурс.
		segments := filterEmpty(strings.Split(path, "/"))
		if len(segments) == 0 {
			h.handleQuestionsCollection(w, r)
			return
		}

		// Первый сегмент — это ID вопроса.
		id, err := strconv.Atoi(segments[0])
		if err != nil || id <= 0 {
			writeError(w, http.StatusBadRequest, "invalid question id")
			return
		}

		// /questions/{id}
		if len(segments) == 1 {
			h.handleQuestionResource(w, r, uint(id))
			return
		}

		// /questions/{id}/answers
		if len(segments) == 2 && segments[1] == "answers" {
			h.handleQuestionAnswers(w, r, uint(id))
			return
		}

		writeError(w, http.StatusNotFound, "route not found")
	}
}

// answersRouter обслуживает /answers/{id}.
func (h *Handler) answersRouter(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/answers")
	path = strings.Trim(path, "/")
	if path == "" {
		// Без ID не понятно, какой ответ нужно обработать.
		writeError(w, http.StatusMethodNotAllowed, "answer id is required")
		return
	}

	// Преобразуем сегмент в число.
	id, err := strconv.Atoi(path)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid answer id")
		return
	}

	h.handleAnswerResource(w, r, uint(id))
}

// handleQuestionsCollection обслуживает GET/POST /questions.
func (h *Handler) handleQuestionsCollection(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	switch r.Method {
	case http.MethodGet:
		// Получаем все вопросы для списка.
		questions, err := h.store.ListQuestions(ctx)
		if err != nil {
			h.logger.Error("failed to list questions", slog.String("error", err.Error()))
			writeError(w, http.StatusInternalServerError, "failed to list questions")
			return
		}
		writeJSON(w, http.StatusOK, questions)
	case http.MethodPost:
		// Читаем полезную нагрузку с текстом вопроса.
		var req struct {
			Text string `json:"text"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json body")
			return
		}
		// Минимальная валидация.
		if strings.TrimSpace(req.Text) == "" {
			writeError(w, http.StatusBadRequest, "text is required")
			return
		}
		// Создаем модель и сохраняем в БД.
		question := &models.Question{Text: req.Text}
		if err := h.store.CreateQuestion(ctx, question); err != nil {
			h.logger.Error("failed to create question", slog.String("error", err.Error()))
			writeError(w, http.StatusInternalServerError, "failed to create question")
			return
		}
		writeJSON(w, http.StatusCreated, question)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// handleQuestionResource обслуживает GET/DELETE /questions/{id}.
func (h *Handler) handleQuestionResource(w http.ResponseWriter, r *http.Request, id uint) {
	ctx := r.Context()
	switch r.Method {
	case http.MethodGet:
		// Загружаем вопрос и его ответы.
		question, err := h.store.GetQuestionWithAnswers(ctx, id)
		if err != nil {
			if storage.IsNotFound(err) {
				writeError(w, http.StatusNotFound, "question not found")
				return
			}
			h.logger.Error("failed to get question", slog.String("error", err.Error()))
			writeError(w, http.StatusInternalServerError, "failed to get question")
			return
		}
		writeJSON(w, http.StatusOK, question)
	case http.MethodDelete:
		// Удаляем вопрос; каскад удалит ответы.
		if err := h.store.DeleteQuestion(ctx, id); err != nil {
			h.logger.Error("failed to delete question", slog.String("error", err.Error()))
			writeError(w, http.StatusInternalServerError, "failed to delete question")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// handleQuestionAnswers обслуживает POST /questions/{id}/answers.
func (h *Handler) handleQuestionAnswers(w http.ResponseWriter, r *http.Request, id uint) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	ctx := r.Context()

	// Проверяем, что вопрос существует.
	if err := h.store.QuestionExists(ctx, id); err != nil {
		if storage.IsNotFound(err) {
			writeError(w, http.StatusNotFound, "question not found")
			return
		}
		h.logger.Error("failed to check question", slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, "failed to create answer")
		return
	}

	var req struct {
		UserID string `json:"user_id"`
		Text   string `json:"text"`
	}

	// Считываем тело запроса.
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	// Проверяем, что пришли оба обязательных поля.
	if strings.TrimSpace(req.UserID) == "" || strings.TrimSpace(req.Text) == "" {
		writeError(w, http.StatusBadRequest, "user_id and text are required")
		return
	}

	// Собираем и сохраняем ответ.
	answer := &models.Answer{
		QuestionID: id,
		UserID:     req.UserID,
		Text:       req.Text,
	}

	if err := h.store.CreateAnswer(ctx, answer); err != nil {
		h.logger.Error("failed to create answer", slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, "failed to create answer")
		return
	}

	writeJSON(w, http.StatusCreated, answer)
}

// handleAnswerResource обслуживает GET/DELETE /answers/{id}.
func (h *Handler) handleAnswerResource(w http.ResponseWriter, r *http.Request, id uint) {
	ctx := r.Context()
	switch r.Method {
	case http.MethodGet:
		// Читаем ответ.
		answer, err := h.store.GetAnswer(ctx, id)
		if err != nil {
			if storage.IsNotFound(err) {
				writeError(w, http.StatusNotFound, "answer not found")
				return
			}
			h.logger.Error("failed to get answer", slog.String("error", err.Error()))
			writeError(w, http.StatusInternalServerError, "failed to get answer")
			return
		}
		writeJSON(w, http.StatusOK, answer)
	case http.MethodDelete:
		// Удаляем конкретный ответ.
		if err := h.store.DeleteAnswer(ctx, id); err != nil {
			h.logger.Error("failed to delete answer", slog.String("error", err.Error()))
			writeError(w, http.StatusInternalServerError, "failed to delete answer")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// writeJSON сериализует ответ в JSON с корректными заголовками.
func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

// writeError возвращает структуру с сообщением об ошибке.
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

// filterEmpty убирает пустые сегменты пути (например, двойные слэши).
func filterEmpty(parts []string) []string {
	var cleaned []string
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			cleaned = append(cleaned, trimmed)
		}
	}
	return cleaned
}


