package integrationtest

import (
	"context"
	"testing"
	"time"

	"streampass/backend/internal/domain/subscription"
	"streampass/backend/internal/domain/user"
	"streampass/backend/internal/infrastructure/postgres"
)

func TestPaymentCreditRollbackAndConcurrentRenewals(t *testing.T) {
	db := StartPostgres(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)
	users := postgres.NewUserRepository(db)
	orders := postgres.NewOrderRepository(db)
	payments := postgres.NewPaymentRepository(db)
	account := user.NewUser("billing-user", "credit@example.com", "test-hash", now)
	if err := users.Create(ctx, account); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"first", "second"} {
		order := &subscription.Order{ID: id, UserID: string(account.ID), PlanCode: subscription.PlanPersonalBasic,
			AmountRUB: 299, Currency: "RUB", PeriodDays: 30, Status: subscription.OrderPending, CreatedAt: now}
		if err := orders.Create(ctx, order); err != nil {
			t.Fatal(err)
		}
		payment := &subscription.Payment{ID: id, ProviderID: "provider-" + id, OrderID: order.ID, UserID: order.UserID,
			AmountRUB: 299, PeriodDays: 30, Status: subscription.PaymentPending, Tariff: order.PlanCode, CreatedAt: now}
		if err := payments.Create(ctx, payment); err != nil {
			t.Fatal(err)
		}
	}

	if _, err := db.ExecContext(ctx, "ALTER TABLE payments ADD CONSTRAINT test_payment_failure CHECK (status <> 'SUCCEEDED')"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := payments.Credit(ctx, "first", "", nil, now); err == nil {
		t.Fatal("expected settlement to fail")
	}
	storedUser, err := users.FindByID(ctx, account.ID)
	if err != nil || storedUser.SubscriptionActiveUntil != nil {
		t.Fatalf("subscription changed after rollback: %+v, %v", storedUser, err)
	}
	storedOrder, err := orders.FindByID(ctx, "first")
	if err != nil || storedOrder.Status != subscription.OrderPending {
		t.Fatalf("order changed after rollback: %+v, %v", storedOrder, err)
	}
	if _, err := db.ExecContext(ctx, "ALTER TABLE payments DROP CONSTRAINT test_payment_failure"); err != nil {
		t.Fatal(err)
	}

	type result struct {
		credited bool
		err      error
	}
	results := make(chan result, 3)
	for _, id := range []string{"first", "second", "first"} {
		go func(id string) {
			_, credited, err := payments.Credit(ctx, id, "", nil, now)
			results <- result{credited: credited, err: err}
		}(id)
	}
	count := 0
	for range 3 {
		out := <-results
		if out.err != nil {
			t.Fatal(out.err)
		}
		if out.credited {
			count++
		}
	}
	storedUser, err = users.FindByID(ctx, account.ID)
	want := now.Add(60 * 24 * time.Hour)
	if err != nil || storedUser.SubscriptionActiveUntil == nil || !storedUser.SubscriptionActiveUntil.Equal(want) || count != 2 {
		t.Fatalf("unexpected subscription after concurrent deliveries: %+v, credited=%d, err=%v", storedUser, count, err)
	}
}
