package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

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

func (r *SubscriptionRepo) CalculateTotal(ctx context.Context, userID *uuid.UUID, serviceName string, periodStart, periodEnd time.Time) (int, error) {
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
	query += fmt.Sprintf(` AND start_date <= $%d AND (end_date IS NULL OR end_date >= $%d)`, argID, argID+1)
	args = append(args, periodEnd, periodStart)

	var total int
	err := r.db.GetContext(ctx, &total, query, args...)
	return total, err
}

func (r *SubscriptionRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Subscription, error) {
	var sub model.Subscription
	query := `SELECT * FROM subscriptions WHERE id = $1`
	err := r.db.GetContext(ctx, &sub, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &sub, nil
}

func (r *SubscriptionRepo) Update(ctx context.Context, sub *model.Subscription) error {
	query := `
		UPDATE subscriptions
		SET service_name = :service_name, 
		    price = :price, 
		    start_date = :start_date, 
		    end_date = :end_date, 
		    updated_at = now()
		WHERE id = :id
		RETURNING updated_at
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
func (r *SubscriptionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM subscriptions WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
