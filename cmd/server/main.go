package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"QA-api_service/internal/config"
	"QA-api_service/internal/database"
	"QA-api_service/internal/storage"
	transport "QA-api_service/internal/transport/http"
)

func main() {
	// создаем JSON-логгер, чтобы видеть структурированные сообщения в консоли контейнера
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// конфиг из переменных окружения
	cfg := config.Load()

	// устанавливаем подключение к постгрест через GORM
	db, err := database.Connect(cfg)
	if err != nil {
		logger.Error("failed to connect database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	store := storage.New(db)
	// создаем HTTP-обработчики и регистрируем маршруты
	handler := transport.NewHandler(store, logger)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	// инициализируем сервер
	server := &http.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: mux,
	}

	// запускаем сервер в отдельной горутине, чтобы можно было перехватывать сигналы остановки
	go func() {
		logger.Info("server is starting", slog.String("addr", server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", slog.String("error", err.Error()))
		}
	}()

	// ждем сисколы
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	logger.Info("server is shutting down")

	// даём серверу до 5 секунд, чтобы завершить активные запросы
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("graceful shutdown failed", slog.String("error", err.Error()))
	}
}
