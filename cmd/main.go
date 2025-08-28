package main

import (
	"log"
	"net/http"

	_ "subscribe_aggregation-main/docs"
	"subscribe_aggregation-main/internal/api"
	"subscribe_aggregation-main/internal/config"
	"subscribe_aggregation-main/internal/storage"
	"subscribe_aggregation-main/pkg/logging"

	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"
	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/pressly/goose/v3"
)

func main() {
	config.InitDB()

	// Создаем расширение UUID, если не создано
	_, err := config.DB.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`)
	if err != nil {
		log.Fatalf("Failed to create extension: %v", err)
	}

	// Устанавливаем диалект для goose
	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("Failed to set goose dialect: %v", err)
	}

	// Запускаем миграции из папки internal/storage/migrations
	if err := goose.Up(config.DB.DB, "internal/storage/migrations"); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	store := storage.NewStorage(config.DB)
	handler := api.NewHandler(store)

	r := chi.NewRouter()
	r.Use(logging.Middleware) // Здесь ваше middleware для логирования (например, api.LoggingMiddleware)

	// Регистрация роутов
	r.Post("/subscriptions", handler.CreateSubscription)
	r.Get("/subscriptions", handler.ListSubscriptions)
	r.Get("/subscriptions/{id}", handler.GetSubscription)
	r.Put("/subscriptions/{id}", handler.UpdateSubscription)
	r.Delete("/subscriptions/{id}", handler.DeleteSubscription)
	r.Get("/subscriptions/sum", handler.SumSubscriptionsCostHandler)

	// Swagger UI
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/swagger/doc.json"),
	))

	log.Println("Start server :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
