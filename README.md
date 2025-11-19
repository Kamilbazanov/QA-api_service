## QA API Service

Минималистичный сервис вопросов и ответов на Go. Использует `net/http`, GORM, PostgreSQL и миграции Goose. Каждый обработчик покрыт комментариями, чтобы junior-разработчику было проще разобраться в коде.

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

### Полезные переменные окружения
| Переменная | Значение по умолчанию | Назначение |
|------------|----------------------|------------|
| `HTTP_PORT` | `8080` | Порт HTTP-сервера |
| `DATABASE_URL` | `postgres://qa_user:qa_password@db:5432/qa_db?sslmode=disable` | Подключение к PostgreSQL |

### Миграции
Используем [Goose](https://github.com/pressly/goose). Пример запуска локально:
```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
DATABASE_URL=postgres://qa_user:qa_password@localhost:5432/qa_db?sslmode=disable \
goose -dir migrations postgres "$DATABASE_URL" up
```
В `docker-compose` миграции применяются автоматически во время старта контейнера `app`.

### Тесты
```bash
go test ./...
```

### API (кратко)
| Метод | Путь | Описание |
|-------|------|----------|
| `GET` | `/questions` | список вопросов |
| `POST` | `/questions` | создать вопрос (`{"text":"..."}`) |
| `GET` | `/questions/{id}` | вопрос + ответы |
| `DELETE` | `/questions/{id}` | удалить вопрос и ответы |
| `POST` | `/questions/{id}/answers` | добавить ответ (`{"user_id":"...","text":"..."}`) |
| `GET` | `/answers/{id}` | получить ответ |
| `DELETE` | `/answers/{id}` | удалить ответ |

### Разработка без Docker
1. Поднимите PostgreSQL (например, через `docker compose up db`).
2. Примените миграции (см. блок выше).
3. Запустите приложение:
   ```bash
   go run ./cmd/server
   ```


