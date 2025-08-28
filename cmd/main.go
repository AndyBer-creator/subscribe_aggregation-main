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

// @title           Subscription API
// @version         1.0
// @description     API для управления подписками с поддержкой фильтрации и подсчёта стоимости.
// @host      localhost:8080
// @BasePath  /
func main() {
	config.InitDB()

	// Создаем расширение UUID, если оно отсутствует
	_, err := config.DB.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`)
	if err != nil {
		log.Fatalf("Failed to create extension: %v", err)
	}

	// Устанавливаем dialect для goose
	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("Failed to set goose dialect: %v", err)
	}

	// Запускаем миграции из папки storage/migrations
	if err := goose.Up(config.DB.DB, "storage/migrations"); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	store := storage.NewStorage(config.DB)
	handler := api.NewHandler(store)

	r := chi.NewRouter()
	r.Use(logging.Middleware)
	r.Post("/subscriptions", handler.CreateSubscription)
	r.Get("/subscriptions", handler.ListSubscriptions)
	r.Get("/subscriptions/{id}", handler.GetSubscription)
	r.Put("/subscriptions/{id}", handler.UpdateSubscription)
	r.Delete("/subscriptions/{id}", handler.DeleteSubscription)
	r.Get("/subscriptions/sum", handler.SumSubscriptionsCostHandler)
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/swagger/doc.json"),
	))

	log.Println("Start server :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
