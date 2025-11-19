package config

import (
	"os"
)

// Config хранит все параметры приложения, чтобы не дергать переменные окружения по всему коду.
type Config struct {
	HTTPPort    string
	DBHost      string
	DBPort      string
	DBUser      string
	DBPassword  string
	DBName      string
	DatabaseURL string
}

// Load собирает конфигурацию с дефолтами для локального запуска (docker-compose).
func Load() Config {
	return Config{
		HTTPPort:    getEnv("HTTP_PORT", "8080"),
		DBHost:      getEnv("DB_HOST", "db"),
		DBPort:      getEnv("DB_PORT", "5432"),
		DBUser:      getEnv("DB_USER", "qa_user"),
		DBPassword:  getEnv("DB_PASSWORD", "qa_password"),
		DBName:      getEnv("DB_NAME", "qa_db"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}
}

// getEnv помогает аккуратно подставить значение по умолчанию, если переменной нет.
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}


