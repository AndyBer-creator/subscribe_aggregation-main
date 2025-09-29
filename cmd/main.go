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
)

func main() {
	// Создаем корневой контекст с функцией отмены
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config.InitDB()

	// Передаем только DB в конструктор, ctx передаем методам явно
	store := storage.NewStorage(config.DB)
	handler := api.NewHandler(store)

	r := chi.NewRouter()

	// Добавляем middleware логирования и передачи контекста запроса
	r.Use(logging.Middleware)

	r.Route("/subscriptions", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			handler.ListSubscriptions(w, r.WithContext(ctx))
		})
		r.Post("/", func(w http.ResponseWriter, r *http.Request) {
			handler.CreateSubscription(w, r.WithContext(ctx))
		})
		r.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
			handler.GetSubscription(w, r.WithContext(ctx))
		})
		r.Put("/{id}", func(w http.ResponseWriter, r *http.Request) {
			handler.UpdateSubscription(w, r.WithContext(ctx))
		})
		r.Delete("/{id}", func(w http.ResponseWriter, r *http.Request) {
			handler.DeleteSubscription(w, r.WithContext(ctx))
		})
		r.Get("/sum", func(w http.ResponseWriter, r *http.Request) {
			handler.SumSubscriptionsCostHandler(w, r.WithContext(ctx))
		})
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

	cancel() // уведомляем зависимости контекста для безопасного завершения

	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()

	if err := srv.Shutdown(ctxShutdown); err != nil {
		log.Fatalf("Server Shutdown Failed:%+v", err)
	}

	log.Println("Server exited properly")
}
