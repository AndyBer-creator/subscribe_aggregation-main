package migrations

import (
	"fmt"
	"path/filepath"
	"subscribe_aggregation-main/internal/config"

	"github.com/pressly/goose/v3"
)

func RunMigrations() error {
	if err := config.InitDB(); err != nil {
		return fmt.Errorf("failed to init DB: %w", err)
	}

	if config.DB == nil || config.DB.DB == nil {
		return fmt.Errorf("DB not initialized")
	}

	// Обработка пути migrationsPath, например filepath.Join(os.Getwd(), "..", "migrations")
	migrationsPath := filepath.Join("..", "internal", "migrations")

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	if err := goose.Up(config.DB.DB, migrationsPath); err != nil {
		return err
	}

	return nil
}
