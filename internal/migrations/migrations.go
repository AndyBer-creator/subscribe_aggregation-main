package migrations

import (
	"log"
	"subscribe_aggregation-main/internal/config"

	"github.com/pressly/goose/v3"
)

func RunMigrations() {
	config.InitDB() // Инициализация базы, если нужно

	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("Failed to set goose dialect: %v", err)
	}

	if err := goose.Up(config.DB.DB, "internal/storage/migrations"); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	log.Println("Migrations applied successfully")
}
