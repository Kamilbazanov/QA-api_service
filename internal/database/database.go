package database

import (
	"fmt"

	"QA-api_service/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Connect открывает подключение к постгрест
func Connect(cfg config.Config) (*gorm.DB, error) {
	dsn := cfg.DatabaseURL
	if dsn == "" {
		dsn = fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			cfg.DBHost,
			cfg.DBPort,
			cfg.DBUser,
			cfg.DBPassword,
			cfg.DBName,
		)
	}

	// Инициализируем GORM с драйвером postgres.
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}
