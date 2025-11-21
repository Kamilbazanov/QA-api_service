package http

import (
	"log/slog"
	"net/http"
	"strconv"

	"QA-api_service/internal/models"
	"QA-api_service/internal/storage"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	store  *storage.Storage
	logger *slog.Logger
}

func NewHandler(store *storage.Storage, logger *slog.Logger) *Handler {
	return &Handler{store: store, logger: logger}
}

func (h *Handler) RegisterRoutes(router *gin.Engine) {
	router.GET("/healthz", h.handleHealth)

	questions := router.Group("/questions")
	{
		questions.GET("", h.listQuestions)
		questions.POST("", h.createQuestion)
		questions.GET("/:id", h.getQuestion)
		questions.DELETE("/:id", h.deleteQuestion)
		questions.POST("/:id/answers", h.createAnswer)
	}

	answers := router.Group("/answers")
	{
		answers.GET("/:id", h.getAnswer)
		answers.DELETE("/:id", h.deleteAnswer)
	}
}

func (h *Handler) handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handler) listQuestions(c *gin.Context) {
	questions, err := h.store.ListQuestions(c.Request.Context())
	if err != nil {
		h.logger.Error("failed to list questions", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to list questions"})
		return
	}
	c.JSON(http.StatusOK, questions)
}

func (h *Handler) createQuestion(c *gin.Context) {
	var req CreateQuestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid json body"})
		return
	}

	question := &models.Question{Text: req.Text}
	if err := h.store.CreateQuestion(c.Request.Context(), question); err != nil {
		h.logger.Error("failed to create question", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to create question"})
		return
	}
	c.JSON(http.StatusCreated, question)
}

func (h *Handler) getQuestion(c *gin.Context) {
	id, err := h.parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid question id"})
		return
	}

	question, err := h.store.GetQuestionWithAnswers(c.Request.Context(), id)
	if err != nil {
		if storage.IsNotFound(err) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "question not found"})
			return
		}
		h.logger.Error("failed to get question", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to get question"})
		return
	}
	c.JSON(http.StatusOK, question)
}

func (h *Handler) deleteQuestion(c *gin.Context) {
	id, err := h.parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid question id"})
		return
	}

	if err := h.store.DeleteQuestion(c.Request.Context(), id); err != nil {
		h.logger.Error("failed to delete question", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to delete question"})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) createAnswer(c *gin.Context) {
	questionID, err := h.parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid question id"})
		return
	}

	if err := h.store.QuestionExists(c.Request.Context(), questionID); err != nil {
		if storage.IsNotFound(err) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "question not found"})
			return
		}
		h.logger.Error("failed to check question", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to create answer"})
		return
	}

	var req CreateAnswerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid json body"})
		return
	}

	answer := &models.Answer{
		QuestionID: questionID,
		UserID:     req.UserID,
		Text:       req.Text,
	}

	if err := h.store.CreateAnswer(c.Request.Context(), answer); err != nil {
		h.logger.Error("failed to create answer", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to create answer"})
		return
	}

	c.JSON(http.StatusCreated, answer)
}

func (h *Handler) getAnswer(c *gin.Context) {
	id, err := h.parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid answer id"})
		return
	}

	answer, err := h.store.GetAnswer(c.Request.Context(), id)
	if err != nil {
		if storage.IsNotFound(err) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "answer not found"})
			return
		}
		h.logger.Error("failed to get answer", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to get answer"})
		return
	}
	c.JSON(http.StatusOK, answer)
}

func (h *Handler) deleteAnswer(c *gin.Context) {
	id, err := h.parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid answer id"})
		return
	}

	if err := h.store.DeleteAnswer(c.Request.Context(), id); err != nil {
		h.logger.Error("failed to delete answer", slog.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to delete answer"})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) parseID(idStr string) (uint, error) {
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		return 0, err
	}
	return uint(id), nil
}
