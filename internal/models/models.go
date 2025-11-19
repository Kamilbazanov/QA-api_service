package models

import "time"

// Question описывает вопрос; GORM добавляет ID автоматически, но мы явно укажем поля.
type Question struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Text      string    `json:"text" gorm:"type:text;not null"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	Answers   []Answer  `json:"answers,omitempty" gorm:"constraint:OnDelete:CASCADE"`
}

// Answer описывает ответ на конкретный вопрос.
type Answer struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	QuestionID uint      `json:"question_id" gorm:"not null;index"`
	UserID     string    `json:"user_id" gorm:"type:varchar(64);not null"`
	Text       string    `json:"text" gorm:"type:text;not null"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
}


