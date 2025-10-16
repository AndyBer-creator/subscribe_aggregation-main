package storage

import (
	"context"
	"database/sql"
	"log"
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

type SubscriptionPeriod struct {
	Price     int64      `db:"price"`
	StartDate time.Time  `db:"start_date"`
	EndDate   *time.Time `db:"end_date"`
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
	var subs []SubscriptionPeriod

	query := sq.Select("price", "start_date", "end_date").
		From("subscriptions").
		Where(sq.And{
			sq.Or{
				sq.GtOrEq{"end_date": filterStart},
				sq.Expr("end_date IS NULL"),
			},
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
	log.Printf("SumSubscripionsCost SQL: %s\nARGS:%v|n", sqlStr, args)

	err = s.db.SelectContext(ctx, &subs, sqlStr, args...)
	if err != nil {
		return 0, err
	}
	log.Printf("Fetched subscriptions: %+v\n", subs)

	total := int64(0)
	merged := MergeIntervals(subs, filterStart, filterEnd)
	for _, sub := range merged {
		// Обработка EndDate в случае nil - подставляем filterEnd
		end := filterEnd
		if sub.EndDate != nil {
			end = *sub.EndDate
		}
		months := MonthsBetween(sub.StartDate, end)
		total += sub.Price * int64(months)
	}
	return total, nil
}

func MonthsBetween(start, end time.Time) int {
	if end.Before(start) {
		return 0
	}

	yearDiff := end.Year() - start.Year()
	monthDiff := int(end.Month()) - int(start.Month())
	months := yearDiff*12 + monthDiff

	if end.Day() < start.Day() {
		months--
	}
	return months + 1
}

func MergeIntervals(subs []SubscriptionPeriod, filterStart, filterEnd time.Time) []SubscriptionPeriod {
	if len(subs) == 0 {
		return nil
	}

	sort.Slice(subs, func(i, j int) bool {
		return subs[i].StartDate.Before(subs[j].StartDate)
	})

	var merged []SubscriptionPeriod

	current := SubscriptionPeriod{
		Price:     subs[0].Price,
		StartDate: maxTime(subs[0].StartDate, filterStart),
		EndDate:   minTimePtr(subs[0].EndDate, &filterEnd),
	}

	for i := 1; i < len(subs); i++ {
		start := maxTime(subs[i].StartDate, filterStart)
		end := minTimePtr(subs[i].EndDate, &filterEnd)

		// Проверяем пересечение интервалов (с допуском в 1 день)
		if !start.After(addOneDay(current.EndDate)) {
			// Объединяем интервалы
			if end.After(*current.EndDate) {
				current.EndDate = end
			}
			if subs[i].Price > current.Price {
				current.Price = subs[i].Price
			}
		} else {
			merged = append(merged, current)
			current = SubscriptionPeriod{
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

func minTimePtr(a, b *time.Time) *time.Time {
	if a == nil {
		return b
	}
	if b == nil {
		return a
	}
	if a.Before(*b) {
		return a
	}
	return b
}

func addOneDay(t *time.Time) time.Time {
	if t == nil {
		return time.Time{} // или filterEnd, если лучше
	}
	return t.AddDate(0, 0, 1)
}
