## QA API Service

Сервис вопросов и ответов на Go. Использует `net/http`, GORM, PostgreSQL и миграции Goose.

### Функциональность
- CRUD для вопросов (`/questions`).
- Добавление и получение ответов (`/questions/{id}/answers`, `/answers/{id}`).
- Каскадное удаление ответов при удалении вопроса.
- Простые комментарии и логирование через `slog`.
- Один интеграционный тест на `httptest`.

### Быстрый старт
```bash
  docker-compose up --build
```
Сервис станет доступен на `http://localhost:8080`.

#### Чтобы остановить токен 

```bash
  docker-compose down
```
### Переменные окружения
Порт HTTP-сервера - `HTTP_PORT`  `8080`

Подключение к PostgreSQL - `DATABASE_URL`  `postgres://qa_user:qa_password@db:5432/qa_db?sslmode=disable`


### Тесты
```bash
  go test ./...
```

### API

### Questions
- `GET /questions` - список вопросов
- `POST /questions` - создать вопрос
- `GET /questions/{id}` - вопрос + ответы
- `DELETE /questions/{id}` - удалить вопрос и ответы

### Answers
- `POST /questions/{id}/answers` - добавить ответ
- `GET /answers/{id}` - получить ответ
- `DELETE /answers/{id}` - удалить ответ



