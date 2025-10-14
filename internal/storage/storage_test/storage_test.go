package storage

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"subscribe_aggregation-main/internal/models"
	"subscribe_aggregation-main/internal/storage"

	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq" // PostgreSQL driver
)

func init() {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("unable to get current filepath")
	}
	envPath := filepath.Join(filepath.Dir(filename), "../../../.env") // поднимаемся на 3 уровня выше
	if err := godotenv.Load(envPath); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
}

func setupTestDB(t *testing.T) *sqlx.DB {
	dsn := os.Getenv("DATABASE_URL")
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatalf("Failed to connect to test db: %v", err)
	}

	// Выполните миграции таблиц подписок, например:
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS subscriptions (
            id UUID PRIMARY KEY,
            user_id TEXT NOT NULL,
            service_name TEXT NOT NULL,
            price BIGINT NOT NULL,
            start_date TIMESTAMPTZ NOT NULL,
            end_date TIMESTAMPTZ,
            created_at TIMESTAMPTZ NOT NULL,
            updated_at TIMESTAMPTZ NOT NULL
        )
    `)

	if err != nil {
		t.Fatalf("Failed creating subscriptions table: %v", err)
	}

	// Очистка таблицы перед каждым тестом
	_, err = db.Exec("TRUNCATE TABLE subscriptions")
	if err != nil {
		t.Fatalf("Failed to truncate subscriptions table: %v", err)
	}

	return db
}

func TestStorage_CreateSubscription(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	userID := uuid.New()
	store := storage.NewStorage(db)

	sub := &models.Subscription{
		UserID:      userID,
		ServiceName: "svc1",
		Price:       100,
		StartDate:   models.DataOnly(time.Now()),
	}

	err := store.CreateSubscription(context.Background(), sub)
	if err != nil {
		t.Fatalf("CreateSubscription failed: %v", err)
	}

	if sub.ID == uuid.Nil {
		t.Error("Expected non-nil UUID after CreateSubscription")
	}

	// Можно проверить, что данные реально в базе
	storedSub, err := store.GetSubscriptionByID(context.Background(), sub.ID)
	if err != nil {
		t.Fatalf("GetSubscriptionByID failed: %v", err)
	}
	if storedSub == nil {
		t.Fatal("Expected subscription in DB but got nil")
	}
	if storedSub.UserID != sub.UserID {
		t.Errorf("Expected UserID %s, got %s", sub.UserID, storedSub.UserID)
	}
}

func TestStorage_GetSubscriptionByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	store := storage.NewStorage(db)

	testSub := &models.Subscription{
		ID:          uuid.New(),
		UserID:      uuid.New(),
		ServiceName: "svc1",
		Price:       100,
		StartDate:   models.DataOnly(time.Now()),
	}
	if err := store.CreateSubscription(context.Background(), testSub); err != nil {
		t.Fatalf("failed to create subscription: %v", err)
	}

	tests := []struct {
		name    string
		id      uuid.UUID
		want    *models.Subscription
		wantErr bool
	}{
		{name: "found", id: testSub.ID, want: testSub, wantErr: false},
		{name: "not found", id: uuid.New(), want: nil, wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := store.GetSubscriptionByID(context.Background(), tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetSubscriptionByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got == nil && tt.want == nil {
				return
			}
			if got == nil || tt.want == nil {
				t.Errorf("got %v, want %v", got, tt.want)
				return
			}
			if got.ID != tt.want.ID {
				t.Errorf("ID mismatch: got %v, want %v", got.ID, tt.want.ID)
			}
			if got.UserID != tt.want.UserID {
				t.Errorf("UserID mismatch: got %v, want %v", got.UserID, tt.want.UserID)
			}
			if got.ServiceName != tt.want.ServiceName {
				t.Errorf("ServiceName mismatch: got %v, want %v", got.ServiceName, tt.want.ServiceName)
			}
			if got.Price != tt.want.Price {
				t.Errorf("Price mismatch: got %v, want %v", got.Price, tt.want.Price)
			}
			if time.Time(got.StartDate).Unix() != time.Time(tt.want.StartDate).Unix() {
				t.Errorf("StartDate mismatch: got %v, want %v", got.StartDate, tt.want.StartDate)
			}
		})

	}
}
