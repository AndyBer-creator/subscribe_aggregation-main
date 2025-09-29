package storage

import (
	"context"
	"database/sql"
	"sort"
	"time"

	"subscribe_aggregation-main/internal/models"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Storage struct {
	db *sqlx.DB
}

type subscriptionPeriod struct {
	Price     int64     `db:"price"`
	StartDate time.Time `db:"start_date"`
	EndDate   time.Time `db:"end_date"`
}

type StorageInterface interface {
	CreateSubscription(ctx context.Context, sub *models.Subscription) error
	GetSubscriptionByID(ctx context.Context, id uuid.UUID) (*models.Subscription, error)
	ListSubscriptions(ctx context.Context, page, limit int) ([]models.Subscription, error)
	UpdateSubscription(ctx context.Context, sub *models.Subscription) error
	DeleteSubscription(ctx context.Context, id uuid.UUID) error
	SumSubscriptionsCost(ctx context.Context, userID, serviceName string, filterStart, filterEnd time.Time) (int64, error)
}

func NewStorage(db *sqlx.DB) *Storage {
	return &Storage{db: db}
}

func (s *Storage) CreateSubscription(ctx context.Context, sub *models.Subscription) error {
	if sub.ID == uuid.Nil {
		sub.ID = uuid.New()
	}

	startDate := time.Time(sub.StartDate)

	query := sq.Insert("subscriptions").
		Columns("id", "user_id", "service_name", "price", "start_date", "created_at", "updated_at").
		Values(sub.ID, sub.UserID, sub.ServiceName, sub.Price, startDate, sq.Expr("NOW()"), sq.Expr("NOW()")).
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

func (s *Storage) ListSubscriptions(ctx context.Context, page, limit int) ([]models.Subscription, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 1000
	}
	offset := (page - 1) * limit

	query := sq.Select("*").
		From("subscriptions").
		Limit(uint64(limit)).
		Offset(uint64(offset)).
		PlaceholderFormat(sq.Dollar)

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	var subs []models.Subscription
	err = s.db.SelectContext(ctx, &subs, sqlStr, args...)
	return subs, err
}

func (s *Storage) UpdateSubscription(ctx context.Context, sub *models.Subscription) error {
	startDate := time.Time(sub.StartDate)

	var endDate *time.Time
	if sub.EndDate != nil {
		ed := time.Time(*sub.EndDate)
		endDate = &ed
	}

	query := sq.Update("subscriptions").
		Set("service_name", sub.ServiceName).
		Set("price", sub.Price).
		Set("start_date", startDate).
		Set("end_date", endDate).
		Set("updated_at", sq.Expr("NOW()")).
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

	res, err := s.db.ExecContext(ctx, sqlStr, args...)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (s *Storage) SumSubscriptionsCost(ctx context.Context, userID, serviceName string, filterStart, filterEnd time.Time) (int64, error) {
	var subs []subscriptionPeriod

	query := sq.Select("price", "start_date", "end_date").
		From("subscriptions").
		Where(sq.And{
			sq.GtOrEq{"end_date": filterStart},
			sq.LtOrEq{"start_date": filterEnd},
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

	err = s.db.SelectContext(ctx, &subs, sqlStr, args...)
	if err != nil {
		return 0, err
	}

	total := int64(0)
	merged := mergeIntervals(subs, filterStart, filterEnd)
	for _, sub := range merged {
		months := monthsBetween(sub.StartDate, sub.EndDate)
		total += sub.Price * int64(months)
	}

	return total, nil
}

func monthsBetween(start, end time.Time) int {
	years := end.Year() - start.Year()
	months := int(end.Month()) - int(start.Month())
	return years*12 + months + 1
}

func mergeIntervals(subs []subscriptionPeriod, filterStart, filterEnd time.Time) []subscriptionPeriod {
	if len(subs) == 0 {
		return nil
	}

	sort.Slice(subs, func(i, j int) bool {
		return subs[i].StartDate.Before(subs[j].StartDate)
	})

	var merged []subscriptionPeriod
	current := subscriptionPeriod{
		Price:     subs[0].Price,
		StartDate: maxTime(subs[0].StartDate, filterStart),
		EndDate:   minTime(subs[0].EndDate, filterEnd),
	}

	for i := 1; i < len(subs); i++ {
		start := maxTime(subs[i].StartDate, filterStart)
		end := minTime(subs[i].EndDate, filterEnd)

		if !start.After(current.EndDate.AddDate(0, 0, 1)) {
			if end.After(current.EndDate) {
				current.EndDate = end
			}
			if subs[i].Price > current.Price {
				current.Price = subs[i].Price
			}
		} else {
			merged = append(merged, current)
			current = subscriptionPeriod{
				Price:     subs[i].Price,
				StartDate: start,
				EndDate:   end,
			}
		}
	}
	merged = append(merged, current)
	return merged
}

func maxTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

func minTime(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}
