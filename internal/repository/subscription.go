package repository

import (
	"context"
	"fmt"

	"effective-go/internal/model"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type SubscriptionRepo struct {
	db *sqlx.DB
}

func NewSubscriptionRepo(db *sqlx.DB) *SubscriptionRepo {
	return &SubscriptionRepo{db: db}
}

func (r *SubscriptionRepo) Create(ctx context.Context, sub *model.Subscription) error {
	query := `
		INSERT INTO subscriptions (id, service_name, price, user_id, start_date, end_date)
		VALUES (:id, :service_name, :price, :user_id, :start_date, :end_date)
		RETURNING created_at, updated_at
	`

	rows, err := sqlx.NamedQueryContext(ctx, r.db, query, sub)
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.StructScan(sub)
	}
	return err
}

func (r *SubscriptionRepo) List(ctx context.Context, userID *uuid.UUID, serviceName string) ([]model.Subscription, error) {
	query := `SELECT * FROM subscriptions WHERE 1=1`
	args := []interface{}{}
	argID := 1

	if userID != nil {
		query += fmt.Sprintf(` AND user_id = $%d`, argID)
		args = append(args, *userID)
		argID++
	}

	if serviceName != "" {
		query += fmt.Sprintf(` AND service_name ILIKE $%d`, argID)
		args = append(args, "%"+serviceName+"%")
		argID++
	}

	var subs []model.Subscription
	err := r.db.SelectContext(ctx, &subs, query, args...)
	return subs, err
}

func (r *SubscriptionRepo) CalculateTotal(ctx context.Context, userID *uuid.UUID, serviceName string) (int, error) {
	query := `SELECT COALESCE(SUM(price), 0) FROM subscriptions WHERE 1=1`
	args := []interface{}{}
	argID := 1

	if userID != nil {
		query += fmt.Sprintf(` AND user_id = $%d`, argID)
		args = append(args, *userID)
		argID++
	}

	if serviceName != "" {
		query += fmt.Sprintf(` AND service_name ILIKE $%d`, argID)
		args = append(args, "%"+serviceName+"%")
		argID++
	}

	var total int
	err := r.db.GetContext(ctx, &total, query, args...)
	return total, err
}
