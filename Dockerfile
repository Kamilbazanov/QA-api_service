FROM golang:1.21 AS builder
WORKDIR /app

# Ставим зависимости заранее, чтобы использовать кеш.
COPY go.mod go.sum ./
RUN go mod download

# Копируем остальной код и собираем бинарник.
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o qa-api ./cmd/server

# Устанавливаем goose, чтобы потом перенести бинарник в рантайм-слой.
RUN go install github.com/pressly/goose/v3/cmd/goose@v3.19.4

FROM debian:bookworm-slim AS runtime
WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && rm -rf /var/lib/apt/lists/*

COPY --from=builder /go/bin/goose /usr/local/bin/goose
COPY --from=builder /app/qa-api /usr/local/bin/qa-api
COPY migrations ./migrations
COPY scripts/entrypoint.sh /entrypoint.sh

RUN chmod +x /entrypoint.sh

ENV DATABASE_URL=postgres://qa_user:qa_password@db:5432/qa_db?sslmode=disable
ENV HTTP_PORT=8080

ENTRYPOINT ["/entrypoint.sh"]
CMD ["/usr/local/bin/qa-api"]


