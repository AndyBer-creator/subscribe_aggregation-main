package storage

import (
	"context"
	"database/sql"
	"time"

	"subscribe_aggregation-main/internal/models"

	sq "github.com/Masterminds/squirrel"
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
	if sub.ID == uuid.Nil {
		sub.ID = uuid.New()
	}
<<<<<<< HEAD

	query := sq.Insert("subscriptions").
		Columns("id", "user_id", "service_name", "price", "start_date", "created_at", "updated_at").
		Values(sub.ID, sub.UserID, sub.ServiceName, sub.Price, sub.StartDate, sq.Expr("NOW()"), sq.Expr("NOW()")).
=======
	if sub.StartDate.IsZero() {
		sub.StartDate = time.Now()
	}

	query := sq.Insert("subscriptions").
		Columns("id", "user_id", "service_name", "price", "start_date", "end_date", "created_at", "updated_at").
		Values(sub.ID, sub.UserID, sub.ServiceName, sub.Price, sub.StartDate, sub.EndDate, time.Now(), time.Now()).
>>>>>>> 78bf63b (updated)
		PlaceholderFormat(sq.Dollar)

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, sqlStr, args...)
	return err
}

func (s *Storage) GetSubscriptionByID(ctx context.Context, id uuid.UUID) (*models.Subscription, error) {
	var sub models.Subscription

	query := sq.Select("*").
		From("subscriptions").
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar)

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	err = s.db.GetContext(ctx, &sub, sqlStr, args...)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &sub, err
}

func (s *Storage) ListSubscriptions(ctx context.Context) ([]models.Subscription, error) {
	var subs []models.Subscription
<<<<<<< HEAD
	query := sq.Select("*").From("subscriptions").
=======

	query := sq.Select("*").
		From("subscriptions").
>>>>>>> 78bf63b (updated)
		PlaceholderFormat(sq.Dollar)

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}
	err = s.db.SelectContext(ctx, &subs, sqlStr, args...)
	return subs, err
}

func (s *Storage) UpdateSubscription(ctx context.Context, sub *models.Subscription) error {
	query := sq.Update("subscriptions").
		Set("user_id", sub.UserID).
		Set("service_name", sub.ServiceName).
		Set("price", sub.Price).
		Set("start_date", sub.StartDate).
<<<<<<< HEAD
		Set("updated_at", sq.Expr("NOW()")).
=======
		Set("end_date", sub.EndDate).
		Set("updated_at", time.Now()).
>>>>>>> 78bf63b (updated)
		Where(sq.Eq{"id": sub.ID}).
		PlaceholderFormat(sq.Dollar)

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, sqlStr, args...)
	return err
}

func (s *Storage) DeleteSubscription(ctx context.Context, id uuid.UUID) error {
	query := sq.Delete("subscriptions").
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar)

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, sqlStr, args...)
	return err
}

<<<<<<< HEAD
func (s *Storage) SumSubscriptionsCost(ctx context.Context, userID, serviceName string, start, end time.Time) (int64, error) {
	query := sq.Select("COALESCE(SUM(price),0)").From("subscriptions").
		Where(sq.And{
			sq.GtOrEq{"start_date": start},
			sq.LtOrEq{"start_date": end},
=======
func (s *Storage) SumSubscriptionsCost(ctx context.Context, userID, serviceName string, start time.Time) (int64, error) {
	query := sq.Select("COALESCE(SUM(price), 0)").
		From("subscriptions").
		Where(sq.And{
			sq.GtOrEq{"start_date": start},
>>>>>>> 78bf63b (updated)
		}).
		PlaceholderFormat(sq.Dollar)

	if userID != "" {
		query = query.Where(sq.Eq{"user_id": userID})
	}
	if serviceName != "" {
		query = query.Where(sq.Eq{"service_name": serviceName})
	}

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return 0, err
	}

	var total int64
	err = s.db.QueryRowContext(ctx, sqlStr, args...).Scan(&total)
	if err != nil {
		return 0, err
	}

	return total, nil
}
