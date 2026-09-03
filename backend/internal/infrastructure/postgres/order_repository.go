package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"streampass/backend/internal/domain/subscription"
	apperrors "streampass/shared/errors"
)

// OrderRepository implements subscription.OrderRepository.
type OrderRepository struct {
	db *sql.DB
}

// NewOrderRepository builds a Postgres-backed OrderRepository.
func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

// Create inserts a pending order.
func (r *OrderRepository) Create(ctx context.Context, o *subscription.Order) error {
	if o.Currency == "" {
		o.Currency = "RUB"
	}
	if o.Status == "" {
		o.Status = subscription.OrderPending
	}
	const q = `
		INSERT INTO orders (id, user_id, plan_code, amount_rub, period_days, currency, status, created_at, paid_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`
	_, err := r.db.ExecContext(ctx, q,
		o.ID, o.UserID, o.PlanCode, o.AmountRUB, o.PeriodDays, o.Currency, o.Status, o.CreatedAt, o.PaidAt,
	)
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to insert order", err)
	}
	return nil
}

// FindByID loads an order by id.
func (r *OrderRepository) FindByID(ctx context.Context, id string) (*subscription.Order, error) {
	const q = `
		SELECT id, user_id, plan_code, amount_rub, period_days, currency, status, created_at, paid_at
		FROM orders WHERE id = $1`
	var o subscription.Order
	var paidAt sql.NullTime
	err := r.db.QueryRowContext(ctx, q, id).Scan(
		&o.ID, &o.UserID, &o.PlanCode, &o.AmountRUB, &o.PeriodDays, &o.Currency, &o.Status, &o.CreatedAt, &paidAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.New(apperrors.CodeNotFound, "order not found").
			WithDetails(map[string]any{"id": id})
	}
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to scan order", err)
	}
	if paidAt.Valid {
		t := paidAt.Time
		o.PaidAt = &t
	}
	return &o, nil
}

// MarkPaidIfPending transitions PENDING→PAID once; returns false if already paid/canceled.
func (r *OrderRepository) MarkPaidIfPending(ctx context.Context, id string) (bool, error) {
	const q = `
		UPDATE orders
		SET status = $2, paid_at = COALESCE(paid_at, $3)
		WHERE id = $1 AND status = $4`
	now := time.Now().UTC()
	res, err := r.db.ExecContext(ctx, q, id, subscription.OrderPaid, now, subscription.OrderPending)
	if err != nil {
		return false, apperrors.Wrap(apperrors.CodeInternal, "failed to mark order paid", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, apperrors.Wrap(apperrors.CodeInternal, "failed to confirm order update", err)
	}
	return n > 0, nil
}
