-- +goose Up
-- Создаем таблицу вопросов.
CREATE TABLE IF NOT EXISTS questions (
    id SERIAL PRIMARY KEY,
    text TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Создаем таблицу ответов и включаем каскадное удаление по question_id.
CREATE TABLE IF NOT EXISTS answers (
    id SERIAL PRIMARY KEY,
    question_id INT NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    user_id VARCHAR(64) NOT NULL,
    text TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_answers_question_id ON answers(question_id);

-- +goose Down
DROP TABLE IF EXISTS answers;
DROP TABLE IF EXISTS questions;


