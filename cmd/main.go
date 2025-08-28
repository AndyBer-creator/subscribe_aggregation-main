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
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title           Subscription API
// @version         1.0
// @description     API для управления подписками с поддержкой фильтрации и подсчёта стоимости.
// @host      localhost:8080
// @BasePath  /
func main() {
	config.InitDB()

	// Create UUID extension if not exists
	_, err := config.DB.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`)
	if err != nil {
		log.Fatalf("Failed to create extension: %v", err)
	}

	// TODO: Run your migrations here before starting the server,
	// e.g. using golang-migrate or another migration tool

	storage := storage.NewStorage(config.DB)
	handler := api.NewHandler(storage)

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
	http.ListenAndServe(":8080", r)
}
