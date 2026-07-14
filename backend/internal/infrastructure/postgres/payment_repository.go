package postgres

import (
	"context"
	"database/sql"
	"errors"

	"streampass/backend/internal/domain/subscription"
	apperrors "streampass/shared/errors"
)

// PaymentRepository implements subscription.PaymentRepository against the
// "payments" table.
type PaymentRepository struct {
	db *sql.DB
}

// NewPaymentRepository builds a Postgres-backed subscription.PaymentRepository.
func NewPaymentRepository(db *sql.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

// Create inserts a new pending payment row.
func (r *PaymentRepository) Create(ctx context.Context, p *subscription.Payment) error {
	const q = `
		INSERT INTO payments (id, user_id, provider_id, amount_rub, period_days, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.db.ExecContext(ctx, q, p.ID, p.UserID, p.ProviderID, p.AmountRUB, p.PeriodDays, p.Status, p.CreatedAt)
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to insert payment", err)
	}
	return nil
}

// FindByProviderID looks up a payment by the payment provider's own ID —
// used to make webhook processing idempotent against redelivery.
func (r *PaymentRepository) FindByProviderID(ctx context.Context, providerID string) (*subscription.Payment, error) {
	const q = `
		SELECT id, user_id, provider_id, amount_rub, period_days, status, created_at
		FROM payments WHERE provider_id = $1`

	var p subscription.Payment
	err := r.db.QueryRowContext(ctx, q, providerID).Scan(
		&p.ID, &p.UserID, &p.ProviderID, &p.AmountRUB, &p.PeriodDays, &p.Status, &p.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.New(apperrors.CodeNotFound, "payment not found").
			WithDetails(map[string]any{"provider_id": providerID})
	}
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to scan payment row", err)
	}
	return &p, nil
}

// MarkSucceeded transitions a payment to SUCCEEDED.
func (r *PaymentRepository) MarkSucceeded(ctx context.Context, providerID string) error {
	const q = `UPDATE payments SET status = $2 WHERE provider_id = $1`

	res, err := r.db.ExecContext(ctx, q, providerID, subscription.PaymentSucceeded)
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to mark payment succeeded", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to confirm payment update", err)
	}
	if n == 0 {
		return apperrors.New(apperrors.CodeNotFound, "payment not found").
			WithDetails(map[string]any{"provider_id": providerID})
	}
	return nil
}
