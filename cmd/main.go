package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	_ "subscribe_aggregation-main/docs"
	"subscribe_aggregation-main/internal/api"
	"subscribe_aggregation-main/internal/config"
	"subscribe_aggregation-main/internal/storage"
	"subscribe_aggregation-main/pkg/logging"

	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/pressly/goose/v3"
)

func main() {
	config.InitDB()

	_, err := config.DB.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`)
	if err != nil {
		log.Fatalf("Failed to create extension: %v", err)
	}

	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("Failed to set goose dialect: %v", err)
	}

	if err := goose.Up(config.DB.DB, "internal/storage/migrations"); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	store := storage.NewStorage(config.DB)
	handler := api.NewHandler(store)

	r := chi.NewRouter()
	r.Use(logging.Middleware)

	r.Route("/subscriptions", func(r chi.Router) {
		r.Get("/", handler.ListSubscriptions)
		r.Post("/", handler.CreateSubscription)
		r.Get("/{id}", handler.GetSubscription)
		r.Put("/{id}", handler.UpdateSubscription)
		r.Delete("/{id}", handler.DeleteSubscription)
		r.Get("/sum", handler.SumSubscriptionsCostHandler)
	})

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/swagger/doc.json"),
	))

	srv := &http.Server{
		Addr:    ":" + config.ConfigInstance.ServerPort,
		Handler: r,
	}

	go func() {
		log.Printf("Start server :%s\n", config.ConfigInstance.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown Failed:%+v", err)
	}

	log.Println("Server exited properly")
}
