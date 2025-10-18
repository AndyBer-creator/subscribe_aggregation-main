package storage

import (
	"context"
	"database/sql"
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
	t.Cleanup(func() {
		db.Close()
	})

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

	// Создаем подписку с конкретной датой для точного сравнения
	now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	testSub := &models.Subscription{
		ID:          uuid.New(),
		UserID:      uuid.New(),
		ServiceName: "svc1",
		Price:       100,
		StartDate:   models.DataOnly(now),
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

			// Проверка для случая "not found"
			if tt.want == nil {
				if got != nil {
					t.Errorf("expected nil, got subscription with ID %s", got.ID)
				}
				return
			}

			// Проверка для случая "found"
			if got == nil {
				t.Errorf("expected subscription, got nil")
				return
			}

			// Сравнение полей
			if got.ID != tt.want.ID {
				t.Errorf("ID mismatch: got %s, want %s", got.ID, tt.want.ID)
			}
			if got.UserID != tt.want.UserID {
				t.Errorf("UserID mismatch: got %s, want %s", got.UserID, tt.want.UserID)
			}
			if got.ServiceName != tt.want.ServiceName {
				t.Errorf("ServiceName mismatch: got %s, want %s", got.ServiceName, tt.want.ServiceName)
			}
			if got.Price != tt.want.Price {
				t.Errorf("Price mismatch: got %d, want %d", got.Price, tt.want.Price)
			}

			// Сравнение дат через ToTime() и проверка года/месяца/дня
			if !got.StartDate.ToTime().Truncate(24 * time.Hour).Equal(tt.want.StartDate.ToTime().Truncate(24 * time.Hour)) {
				t.Errorf("StartDate mismatch: got %v, want %v", got.StartDate, tt.want.StartDate)
			}
		})
	}
}

func TestStorage_ListSubscriptions(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	store := storage.NewStorage(db)

	// Создаем фиктивные подписки
	subs := []models.Subscription{
		{
			ID:          uuid.New(),
			UserID:      uuid.New(),
			ServiceName: "svc1",
			Price:       100,
			StartDate:   models.DataOnly(time.Now()),
		},
		{
			ID:          uuid.New(),
			UserID:      uuid.New(),
			ServiceName: "svc2",
			Price:       200,
			StartDate:   models.DataOnly(time.Now()),
		},
		{
			ID:          uuid.New(),
			UserID:      uuid.New(),
			ServiceName: "svc3",
			Price:       300,
			StartDate:   models.DataOnly(time.Now()),
		},
	}

	for i := range subs {
		if err := store.CreateSubscription(context.Background(), &subs[i]); err != nil {
			t.Fatalf("failed to create subscription: %v", err)
		}
	}

	tests := []struct {
		name    string
		page    int
		limit   int
		wantLen int
		wantErr bool
	}{
		{"page 1, limit 2", 1, 2, 2, false},
		{"page 2, limit 2", 2, 2, 1, false},
		{"page 1, limit 5", 1, 5, 3, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := store.ListSubscriptions(context.Background(), tt.page, tt.limit)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListSubscriptions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.wantLen {
				t.Errorf("ListSubscriptions() length = %v, want %v", len(got), tt.wantLen)
			}
			// По желанию проверить ключевые поля элементов, например:
			for i, sub := range got {
				if sub.ServiceName != subs[(tt.page-1)*tt.limit+i].ServiceName {
					t.Errorf("Subscription #%d ServiceName = %v, want %v", i, sub.ServiceName, subs[(tt.page-1)*tt.limit+i].ServiceName)
				}
			}
		})
	}
}
func TestStorage_DeleteSubscription(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	store := storage.NewStorage(db)

	sub := &models.Subscription{
		ID:          uuid.New(),
		UserID:      uuid.New(),
		ServiceName: "svc1",
		Price:       100,
		StartDate:   models.DataOnly(time.Now()),
	}
	if err := store.CreateSubscription(context.Background(), sub); err != nil {
		t.Fatalf("failed to create subscription: %v", err)
	}

	t.Run("Delete existing subscription", func(t *testing.T) {
		err := store.DeleteSubscription(context.Background(), sub.ID)
		if err != nil {
			t.Fatalf("DeleteSubscription() error = %v", err)
		}
		// Проверить что(subscription реально удалена)
		got, err := store.GetSubscriptionByID(context.Background(), sub.ID)
		if err != nil && err != sql.ErrNoRows {
			t.Fatalf("GetSubscriptionByID() error = %v", err)
		}
		if got != nil {
			t.Error("expected subscription to be deleted, but it still exists")
		}
	})

	t.Run("Delete non-existing subscription", func(t *testing.T) {
		err := store.DeleteSubscription(context.Background(), uuid.New())
		if err != sql.ErrNoRows {
			t.Errorf("expected sql.ErrNoRows, got %v", err)
		}
	})
}

func TestStorage_SumSubscriptionsCost(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	store := storage.NewStorage(db)

	userID := uuid.New()
	service := "svc1"

	// Создаем подписки с пересекающимися периодами
	start1 := models.DataOnly(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))
	end1 := models.DataOnly(time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC))
	start2 := models.DataOnly(time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC))
	end2 := models.DataOnly(time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC))

	subs := []models.Subscription{
		{
			ID:          uuid.New(),
			UserID:      userID,
			ServiceName: service,
			Price:       100,
			StartDate:   start1,
			EndDate:     &end1,
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			ServiceName: service,
			Price:       150,
			StartDate:   start2,
			EndDate:     &end2,
		},
	}

	// Создаем записи в БД
	for i := range subs {
		err := store.CreateSubscription(context.Background(), &subs[i])
		if err != nil {
			t.Fatalf("failed to create subscription: %v", err)
		}
	}

	tests := []struct {
		name        string
		filterStart time.Time
		filterEnd   time.Time
		want        int64
		wantErr     bool
	}{
		{
			name:        "overlapping subscriptions full period",
			filterStart: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			filterEnd:   time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC),
			want:        600, // 150 руб * 4 месяца (январь-апрель)
			wantErr:     false,
		},
		{
			name:        "partial period",
			filterStart: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			filterEnd:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			want:        100, // 100 руб * 1 месяц
			wantErr:     false,
		},
		{
			name:        "non overlapping period",
			filterStart: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			filterEnd:   time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC),
			want:        0,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := store.SumSubscriptionsCost(context.Background(), userID.String(), service, tt.filterStart, tt.filterEnd)
			if (err != nil) != tt.wantErr {
				t.Errorf("SumSubscriptionsCost() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("SumSubscriptionsCost() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMergeIntervals(t *testing.T) {
	filterStart := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	filterEnd := time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC)

	endDate1 := time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)
	endDate2 := time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC)

	subs := []storage.SubscriptionPeriod{
		{
			Price:     100,
			StartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:   &endDate1,
		},
		{
			Price:     150,
			StartDate: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
			EndDate:   &endDate2,
		},
	}

	merged := storage.MergeIntervals(subs, filterStart, filterEnd)

	if len(merged) != 1 {
		t.Errorf("Expected 1 merged interval, got %d", len(merged))
	}
	if merged[0].Price != 150 {
		t.Errorf("Expected price 150, got %d", merged[0].Price)
	}
	if !merged[0].StartDate.Equal(filterStart) || !merged[0].EndDate.Equal(filterEnd) {
		t.Errorf("Merged interval dates incorrect: %+v", merged[0])
	}
}

func TestMonthsBetween(t *testing.T) {
	tests := []struct {
		name     string
		start    time.Time
		end      time.Time
		expected int
	}{
		{
			name:     "exact months",
			start:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			end:      time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC),
			expected: 4,
		},
		{
			name:     "partial month",
			start:    time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			end:      time.Date(2024, 2, 14, 0, 0, 0, 0, time.UTC),
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := storage.MonthsBetween(tt.start, tt.end)
			if got != tt.expected {
				t.Errorf("MonthsBetween(%v, %v) = %d, want %d", tt.start, tt.end, got, tt.expected)
			}
		})
	}
}
