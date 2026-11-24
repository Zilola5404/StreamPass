package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

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

const paymentSelectCols = `
	id, user_id, provider_id, amount_rub, period_days, status, created_at,
	COALESCE(provider, 'yookassa'), COALESCE(currency, 'RUB'),
	telegram_user_id, COALESCE(tariff, ''), COALESCE(tx_hash, ''), paid_at,
	COALESCE(order_id, '')`

func scanPayment(scanner interface {
	Scan(dest ...any) error
}) (*subscription.Payment, error) {
	var p subscription.Payment
	var tgID sql.NullInt64
	var paidAt sql.NullTime
	err := scanner.Scan(
		&p.ID, &p.UserID, &p.ProviderID, &p.AmountRUB, &p.PeriodDays, &p.Status, &p.CreatedAt,
		&p.Provider, &p.Currency, &tgID, &p.Tariff, &p.TxHash, &paidAt, &p.OrderID,
	)
	if err != nil {
		return nil, err
	}
	if tgID.Valid {
		v := tgID.Int64
		p.TelegramUserID = &v
	}
	if paidAt.Valid {
		t := paidAt.Time
		p.PaidAt = &t
	}
	return &p, nil
}

// Create inserts a new pending payment row.
func (r *PaymentRepository) Create(ctx context.Context, p *subscription.Payment) error {
	if p.Provider == "" {
		p.Provider = "yookassa"
	}
	if p.Currency == "" {
		p.Currency = "RUB"
	}
	const q = `
		INSERT INTO payments (
			id, user_id, provider_id, amount_rub, period_days, status, created_at,
			provider, currency, telegram_user_id, tariff, tx_hash, paid_at, order_id
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`

	var tg any
	if p.TelegramUserID != nil {
		tg = *p.TelegramUserID
	}
	var orderID any
	if p.OrderID != "" {
		orderID = p.OrderID
	}
	_, err := r.db.ExecContext(ctx, q,
		p.ID, p.UserID, p.ProviderID, p.AmountRUB, p.PeriodDays, p.Status, p.CreatedAt,
		p.Provider, p.Currency, tg, p.Tariff, nullIfEmpty(p.TxHash), p.PaidAt, orderID,
	)
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to insert payment", err)
	}
	return nil
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}

// FindByProviderID looks up a payment by the payment provider's own ID —
// used to make webhook processing idempotent against redelivery.
func (r *PaymentRepository) FindByProviderID(ctx context.Context, providerID string) (*subscription.Payment, error) {
	q := `SELECT ` + paymentSelectCols + ` FROM payments WHERE provider_id = $1`
	p, err := scanPayment(r.db.QueryRowContext(ctx, q, providerID))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.New(apperrors.CodeNotFound, "payment not found").
			WithDetails(map[string]any{"provider_id": providerID})
	}
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to scan payment row", err)
	}
	return p, nil
}

// FindByID looks up a payment by our primary key (Telegram invoice payload).
func (r *PaymentRepository) FindByID(ctx context.Context, id string) (*subscription.Payment, error) {
	q := `SELECT ` + paymentSelectCols + ` FROM payments WHERE id = $1`
	p, err := scanPayment(r.db.QueryRowContext(ctx, q, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.New(apperrors.CodeNotFound, "payment not found").
			WithDetails(map[string]any{"id": id})
	}
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to scan payment row", err)
	}
	return p, nil
}

// FindByTxHash looks up a payment by blockchain tx hash (USDT idempotency).
func (r *PaymentRepository) FindByTxHash(ctx context.Context, txHash string) (*subscription.Payment, error) {
	q := `SELECT ` + paymentSelectCols + ` FROM payments WHERE tx_hash = $1`
	p, err := scanPayment(r.db.QueryRowContext(ctx, q, txHash))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.New(apperrors.CodeNotFound, "payment not found").
			WithDetails(map[string]any{"tx_hash": txHash})
	}
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to scan payment row", err)
	}
	return p, nil
}

// Credit serializes payment delivery and extends access using the latest expiry.
func (r *PaymentRepository) Credit(ctx context.Context, id, chargeID string, tgUserID *int64, now time.Time) (time.Time, bool, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return time.Time{}, false, apperrors.Wrap(apperrors.CodeInternal, "failed to start payment settlement", err)
	}
	defer tx.Rollback()

	until, credited, err := creditPayment(ctx, tx, id, chargeID, tgUserID, now)
	if err == nil {
		err = tx.Commit()
	}
	if err != nil {
		return time.Time{}, false, apperrors.Wrap(apperrors.CodeInternal, "failed to settle payment", err)
	}
	return until, credited, nil
}

func creditPayment(ctx context.Context, tx *sql.Tx, id, chargeID string, tgUserID *int64, now time.Time) (time.Time, bool, error) {
	payment, err := scanPayment(tx.QueryRowContext(ctx, `SELECT `+paymentSelectCols+` FROM payments WHERE id = $1 FOR UPDATE`, id))
	if err != nil {
		return time.Time{}, false, err
	}
	var expiry sql.NullTime
	if err := tx.QueryRowContext(ctx, `SELECT subscription_active_until FROM users WHERE id = $1 FOR UPDATE`, payment.UserID).Scan(&expiry); err != nil {
		return time.Time{}, false, err
	}
	if payment.Status != subscription.PaymentPending {
		return expiry.Time, false, nil
	}
	base := now
	if expiry.Valid && expiry.Time.After(base) {
		base = expiry.Time
	}
	until := base.Add(time.Duration(payment.PeriodDays) * 24 * time.Hour)
	plan := subscription.NormalizePlanCode(payment.Tariff)
	if plan == "" {
		plan = subscription.PlanPersonalBasic
	}
	if _, err := tx.ExecContext(ctx, `UPDATE users SET subscription_active_until = $2,
		entitlement_source = 'paid', plan_code = $3, subscription_canceled_at = NULL,
		updated_at = $4 WHERE id = $1`, payment.UserID, until, plan, now); err != nil {
		return time.Time{}, false, err
	}
	if payment.OrderID != "" {
		result, err := tx.ExecContext(ctx, `UPDATE orders SET status = $2, paid_at = COALESCE(paid_at, $5)
			WHERE id = $1 AND user_id = $3 AND status IN ($4, $2)`,
			payment.OrderID, subscription.OrderPaid, payment.UserID, subscription.OrderPending, now)
		if err != nil {
			return time.Time{}, false, err
		}
		count, err := result.RowsAffected()
		if err != nil {
			return time.Time{}, false, err
		}
		if count != 1 {
			return time.Time{}, false, errors.New("payment order unavailable")
		}
	}
	if _, err := tx.ExecContext(ctx, `UPDATE payments SET status = $2, paid_at = COALESCE(paid_at, $3),
		tx_hash = COALESCE(NULLIF($4, ''), tx_hash),
		telegram_user_id = COALESCE($5, telegram_user_id) WHERE id = $1`,
		id, subscription.PaymentSucceeded, now, chargeID, tgUserID); err != nil {
		return time.Time{}, false, err
	}
	return until, true, nil
}

// ListByUserID returns payments for a user, newest first.
func (r *PaymentRepository) ListByUserID(ctx context.Context, userID string) ([]*subscription.Payment, error) {
	q := `SELECT ` + paymentSelectCols + ` FROM payments WHERE user_id = $1 ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, q, userID)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to list payments", err)
	}
	defer rows.Close()

	var out []*subscription.Payment
	for rows.Next() {
		p, err := scanPayment(rows)
		if err != nil {
			return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to scan payment row", err)
		}
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed while iterating payments", err)
	}
	if out == nil {
		out = []*subscription.Payment{}
	}
	return out, nil
}
