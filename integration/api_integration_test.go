package integration

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	//"github.com/docker/docker/api/server/router"
	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	dockertest "github.com/ory/dockertest/v3"

	"subscribe_aggregation-main/internal/config"
	"subscribe_aggregation-main/internal/migrations"
)

var db *sqlx.DB
var router http.Handler

func TestMain(m *testing.M) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	resource, err := pool.Run("postgres", "15", []string{
		"POSTGRES_USER=postgres",
		"POSTGRES_PASSWORD=secret",
		"POSTGRES_DB=testdb",
	})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	port := resource.GetPort("5432/tcp")
	dsn := fmt.Sprintf("postgres://postgres:secret@localhost:%s/testdb?sslmode=disable", port)
	err = os.Setenv("DATABASE_URL", dsn)
	if err != nil {
		log.Fatalf("Failed to set DATABASE_URL env: %s", err)
	}
	pool.MaxWait = 120 * time.Second

	if err := pool.Retry(func() error {
		sqlDB, err := sql.Open("postgres", dsn)
		if err != nil {
			return err
		}
		return sqlDB.Ping()
	}); err != nil {
		log.Fatalf("Could not connect to docker Postgres: %s", err)
	}

	if err := config.InitDB(); err != nil {
		log.Fatalf("Failed to init DB: %s", err)
	}
	db = config.DB

	if err := migrations.RunMigrations(); err != nil {
		log.Fatalf("Could not run migrations: %s", err)
	}

	code := m.Run()

	// Очистка ресурса
	if err := pool.Purge(resource); err != nil {
		log.Printf("Could not purge resource: %s", err)
	}
	r := chi.NewRouter()

	router = r
	os.Exit(code)
}

func TestListSubscriptions(t *testing.T) {
	// Гарантируем, что таблица subscriptions пустая для чистого теста
	_, err := db.Exec("DELETE FROM subscriptions")
	if err != nil {
		t.Fatalf("Failed to clear subscriptions table: %v", err)
	}

	// Вставляем тестовые данные
	_, err = db.Exec(`INSERT INTO subscriptions (user_id, service_name, price, start_date, created_at) 
	VALUES ('123e4567-e89b-12d3-a456-426614174000', 'Test Service 1', '125', '2025-05-12', NOW())`)
	if err != nil {
		t.Fatalf("Failed to seed subscriptions: %v", err)
	}

	req := httptest.NewRequest("GET", "/subscriptions", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	// Опционально: проверьте тело ответа (например, распарсить JSON и проверить значения)
}
