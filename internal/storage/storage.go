package storage

import (
	"context"
	"database/sql"
	"subscribe_aggregation-main/internal/models"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Storage struct {
	db *sqlx.DB
}

func NewStorage(db *sqlx.DB) *Storage {
	return &Storage{db: db}
}

func (s *Storage) CreateSubscription(ctx context.Context, sub *models.Subscription) error {
	query := `INSERT INTO subscriptions (id, user_id, service_name, price, start_date, created_at, updated_at)
              VALUES (:id, :user_id, :service_name, :price, :start_date, NOW(), NOW())`
	if sub.ID == uuid.Nil {
		sub.ID = uuid.New()
	}
	_, err := s.db.NamedExecContext(ctx, query, sub)
	return err
}

func (s *Storage) GetSubscriptionByID(ctx context.Context, id uuid.UUID) (*models.Subscription, error) {
	var sub models.Subscription
	query := `SELECT * FROM subscriptions WHERE id = $1`
	err := s.db.GetContext(ctx, &sub, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &sub, err
}

func (s *Storage) ListSubscriptions(ctx context.Context) ([]models.Subscription, error) {
	var subs []models.Subscription
	query := `SELECT * FROM subscriptions`
	err := s.db.SelectContext(ctx, &subs, query)
	return subs, err
}

func (s *Storage) UpdateSubscription(ctx context.Context, sub *models.Subscription) error {
	query := `UPDATE subscriptions SET 
        user_id = :user_id, 
        service_name = :service_name, 
        price = :price, 
        start_date = :start_date, 
        updated_at = NOW()
        WHERE id = :id`
	_, err := s.db.NamedExecContext(ctx, query, sub)
	return err
}

func (s *Storage) DeleteSubscription(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM subscriptions WHERE id = $1`
	_, err := s.db.ExecContext(ctx, query, id)
	return err
}

func (s *Storage) SumSubscriptionsCost(ctx context.Context, userID, serviceName string, start, end time.Time) (int64, error) {
	var total int64
	query := `SELECT COALESCE(SUM(price), 0) FROM subscriptions WHERE start_date >= $1 AND start_date <= $2`
	args := []interface{}{start, end}

	if userID != "" {
		query += ` AND user_id = $3`
		args = append(args, userID)
	}
	if serviceName != "" {
		if userID == "" {
			query += ` AND service_name = $3`
		} else {
			query += ` AND service_name = $4`
		}
		args = append(args, serviceName)
	}

	err := s.db.GetContext(ctx, &total, query, args...)
	return total, err
}
