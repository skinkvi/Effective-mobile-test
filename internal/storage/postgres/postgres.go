package postgres

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type Storage struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

func New(ctx context.Context, connString string, logger *zap.Logger) (*Storage, error) {
	const fn = "storage.postgres.New"

	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("%s, %w", fn, err)
	}

	db, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("%s, %w", fn, err)
	}

	return &Storage{db: db, logger: logger}, nil
}

func (s *Storage) Close() {
	s.db.Close()
}

func (s *Storage) GetDB() *pgxpool.Pool {
	return s.db
}

type Subscription struct {
	ID          int        `json:"id"`
	ServiceName string     `json:"service_name"`
	Price       int        `json:"price"`
	UserID      uuid.UUID  `json:"user_id"`
	StartDate   *time.Time `json:"start_date"`
	EndDate     *time.Time `json:"end_date,omitempty"`
}

func (s *Storage) CreateSubscription(ctx context.Context, sub Subscription) (int, error) {
	const fn = "storage.postgres.CreateSubscription"

	query := `INSERT INTO subscriptions (service_name, price, user_id, start_date, end_date) VALUES ($1, $2, $3, $4, $5) RETURNING id`

	var id int

	err := s.db.QueryRow(ctx, query, sub.ServiceName, sub.Price, sub.UserID, sub.StartDate, sub.EndDate).Scan(&id)
	if err != nil {
		s.logger.Error("failed to create subscription", zap.Error(err))
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, fmt.Errorf("%s: unexpected error, no rows returned from insert: %w", fn, err)
		}
		return 0, fmt.Errorf("%s: failed to create subscription: %w", fn, err)
	}

	return id, nil
}

func (s *Storage) GetSubscription(ctx context.Context, id int) (Subscription, error) {
	const fn = "storage.postgres.GetSubscription"

	query := `SELECT id, service_name, price, user_id, start_date, end_date FROM subscriptions WHERE id = $1`

	var sub Subscription

	err := s.db.QueryRow(ctx, query, id).Scan(&sub.ID, &sub.ServiceName, &sub.Price, &sub.UserID, &sub.StartDate, &sub.EndDate)
	if err != nil {
		s.logger.Error("failed to get subscription", zap.Error(err))
		if errors.Is(err, pgx.ErrNoRows) {
			return Subscription{}, fmt.Errorf("%s: subscription with id %d not found: %w", fn, id, err)
		}

		return Subscription{}, fmt.Errorf("%s: %w", fn, err)
	}

	return sub, nil
}

func (s *Storage) UpdateSubscription(ctx context.Context, sub Subscription) error {
	const fn = "storage.postgres.UpdateSubscription"

	var exists bool
	existsQuery := `SELECT EXISTS(SELECT 1 FROM subscriptions WHERE id = $1)`
	err := s.db.QueryRow(ctx, existsQuery, sub.ID).Scan(&exists)
	if err != nil {
		s.logger.Error("failed to check subscription existence", zap.Error(err))
		return fmt.Errorf("%s: %w", fn, err)
	}

	if !exists {
		s.logger.Warn("subscription not found for update", zap.Int("id", sub.ID))
		return fmt.Errorf("%s: subscription with id %d not found", fn, sub.ID)
	}

	query := `UPDATE subscriptions SET service_name = $1, price = $2, user_id = $3, start_date = $4, end_date = $5 WHERE id = $6`
	_, err = s.db.Exec(ctx, query, sub.ServiceName, sub.Price, sub.UserID, sub.StartDate, sub.EndDate, sub.ID)
	if err != nil {
		s.logger.Error("failed to update subscription", zap.Error(err))
		return fmt.Errorf("%s: %w", fn, err)
	}

	return nil
}

func (s *Storage) DeleteSubscription(ctx context.Context, id int) error {
	const fn = "storage.postgres.DeleteSubscription"

	var exists bool
	existsQuery := `SELECT EXISTS(SELECT 1 FROM subscriptions WHERE id = $1)`
	err := s.db.QueryRow(ctx, existsQuery, id).Scan(&exists)
	if err != nil {
		s.logger.Error("failed to check subscription existence", zap.Error(err))
		return fmt.Errorf("%s: %w", fn, err)
	}

	if !exists {
		s.logger.Warn("subscription not found for deletion", zap.Int("id", id))
		return fmt.Errorf("%s: subscription with id %d not found", fn, id)
	}

	query := `DELETE FROM subscriptions WHERE id = $1`
	_, err = s.db.Exec(ctx, query, id)
	if err != nil {
		s.logger.Error("failed to delete subscription", zap.Error(err))
		return fmt.Errorf("%s: %w", fn, err)
	}

	return nil
}

func (s *Storage) ListSubscriptions(ctx context.Context, userID *uuid.UUID, serviceName *string) ([]Subscription, error) {
	const fn = "storage.postgres.ListSubscriptions"

	query := `SELECT id, service_name, price, user_id, start_date, end_date FROM subscriptions WHERE 1=1`
	var args []interface{}

	if userID != nil {
		query += ` AND user_id = $` + strconv.Itoa(len(args)+1)
		args = append(args, *userID)
	}

	if serviceName != nil {
		query += ` AND service_name = $` + strconv.Itoa(len(args)+1)
		args = append(args, *serviceName)
	}

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		s.logger.Error("failed to query subscriptions", zap.Error(err))
		return nil, fmt.Errorf("%s: failed to query subscriptions: %w", fn, err)
	}

	defer rows.Close()

	var subs []Subscription
	for rows.Next() {
		var sub Subscription

		err := rows.Scan(&sub.ID, &sub.ServiceName, &sub.Price, &sub.UserID, &sub.StartDate, &sub.EndDate)
		if err != nil {
			s.logger.Error("failed to scan subscription row", zap.Error(err))
			return nil, fmt.Errorf("%s: failed to scan subscription row: %w", fn, err)
		}

		subs = append(subs, sub)
	}

	if rows.Err() != nil {
		s.logger.Error("error iterating over rows", zap.Error(rows.Err()))
		return nil, fmt.Errorf("%s: error iterating over rows: %w", fn, rows.Err())
	}

	return subs, nil
}

func (s *Storage) CalculateTotalCost(ctx context.Context, startDate, endDate time.Time, userID uuid.UUID, serviceName string) (int, error) {
	const fn = "storage.postgres.CalculateTotalCost"

	query := `SELECT COALESCE(SUM(price), 0) FROM subscriptions WHERE user_id = $1 AND service_name = $2 AND ((start_date >= $3 AND start_date <= $4) OR (start_date <= $3 AND (end_date IS NULL OR end_date >= $3)))`

	var totalCost int

	err := s.db.QueryRow(ctx, query, userID, serviceName, startDate, endDate).Scan(&totalCost)
	if err != nil {
		s.logger.Error("failed to calculate total cost", zap.Error(err))
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, fmt.Errorf("%s: no subscriptions found for user %s and service %s: %w", fn, userID, serviceName, err)
		}

		return 0, fmt.Errorf("%s: %w", fn, err)
	}

	return totalCost, nil
}
