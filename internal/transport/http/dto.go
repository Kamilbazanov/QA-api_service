package http

type CreateQuestionRequest struct {
	Text string `json:"text" binding:"required"`
}

type CreateAnswerRequest struct {
	UserID string `json:"user_id" binding:"required"`
	Text   string `json:"text" binding:"required"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
