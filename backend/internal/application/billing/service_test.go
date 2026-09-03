package billing_test

import (
	"context"
	"testing"
	"time"

	"streampass/backend/internal/application/billing"
	"streampass/backend/internal/domain/subscription"
	"streampass/backend/internal/domain/user"
	"streampass/shared/logger"
)

type fixedClock struct{ t time.Time }

func (c fixedClock) Now() time.Time { return c.t }

type memPayRepo struct {
	byProvider map[string]*subscription.Payment
	byID       map[string]*subscription.Payment
}

func newMemPay() *memPayRepo {
	return &memPayRepo{byProvider: map[string]*subscription.Payment{}, byID: map[string]*subscription.Payment{}}
}

func (m *memPayRepo) FindByProviderID(_ context.Context, providerID string) (*subscription.Payment, error) {
	p, ok := m.byProvider[providerID]
	if !ok {
		return nil, subscriptionErrNotFound("provider")
	}
	cp := *p
	return &cp, nil
}
func (m *memPayRepo) FindByID(_ context.Context, id string) (*subscription.Payment, error) {
	p, ok := m.byID[id]
	if !ok {
		return nil, subscriptionErrNotFound("id")
	}
	cp := *p
	return &cp, nil
}
func (m *memPayRepo) FindByTxHash(context.Context, string) (*subscription.Payment, error) {
	return nil, subscriptionErrNotFound("tx")
}
func (m *memPayRepo) Create(_ context.Context, p *subscription.Payment) error {
	cp := *p
	m.byID[p.ID] = &cp
	m.byProvider[p.ProviderID] = &cp
	return nil
}
func (m *memPayRepo) MarkSucceededIfPending(_ context.Context, providerID string) (bool, error) {
	p, ok := m.byProvider[providerID]
	if !ok {
		return false, subscriptionErrNotFound("provider")
	}
	if p.Status == subscription.PaymentSucceeded {
		return false, nil
	}
	p.Status = subscription.PaymentSucceeded
	return true, nil
}
func (m *memPayRepo) MarkSucceededByIDIfPending(_ context.Context, id, chargeID string, tgUserID *int64) (bool, error) {
	p, ok := m.byID[id]
	if !ok {
		return false, subscriptionErrNotFound("id")
	}
	if p.Status == subscription.PaymentSucceeded {
		return false, nil
	}
	p.Status = subscription.PaymentSucceeded
	p.TxHash = chargeID
	_ = tgUserID
	return true, nil
}
func (m *memPayRepo) ListByUserID(context.Context, string) ([]*subscription.Payment, error) {
	return nil, nil
}

type memOrders struct {
	byID map[string]*subscription.Order
}

func newMemOrders() *memOrders { return &memOrders{byID: map[string]*subscription.Order{}} }

func (m *memOrders) Create(_ context.Context, o *subscription.Order) error {
	cp := *o
	m.byID[o.ID] = &cp
	return nil
}
func (m *memOrders) FindByID(_ context.Context, id string) (*subscription.Order, error) {
	o, ok := m.byID[id]
	if !ok {
		return nil, subscriptionErrNotFound("order")
	}
	cp := *o
	return &cp, nil
}
func (m *memOrders) MarkPaidIfPending(_ context.Context, id string) (bool, error) {
	o, ok := m.byID[id]
	if !ok {
		return false, subscriptionErrNotFound("order")
	}
	if o.Status == subscription.OrderPaid {
		return false, nil
	}
	o.Status = subscription.OrderPaid
	return true, nil
}

type memUsersBilling struct {
	u            *user.User
	activateCalls int
}

func (m *memUsersBilling) Create(context.Context, *user.User) error { return nil }
func (m *memUsersBilling) FindByEmail(context.Context, string) (*user.User, error) {
	return nil, user.ErrNotFound("x")
}
func (m *memUsersBilling) FindByID(context.Context, user.ID) (*user.User, error) {
	cp := *m.u
	return &cp, nil
}
func (m *memUsersBilling) ExtendSubscription(_ context.Context, _ user.ID, until time.Time) error {
	m.u.SubscriptionActiveUntil = &until
	m.u.EntitlementSource = "paid"
	return nil
}
func (m *memUsersBilling) ActivatePaidPlan(_ context.Context, _ user.ID, until time.Time, planCode string) error {
	m.activateCalls++
	m.u.SubscriptionActiveUntil = &until
	m.u.EntitlementSource = "paid"
	m.u.PlanCode = planCode
	m.u.SubscriptionCanceledAt = nil
	return nil
}
func (m *memUsersBilling) CancelAutoRenew(_ context.Context, _ user.ID, now time.Time) error {
	m.u.SubscriptionCanceledAt = &now
	return nil
}
func (m *memUsersBilling) ClearSubscription(context.Context, user.ID, time.Time) error { return nil }
func (m *memUsersBilling) SetBanned(context.Context, user.ID, *time.Time, time.Time) error {
	return nil
}
func (m *memUsersBilling) SearchByEmail(context.Context, string) ([]*user.User, error) {
	return nil, nil
}
func (m *memUsersBilling) List(context.Context) ([]*user.User, error) { return nil, nil }
func (m *memUsersBilling) UpdatePasswordHash(context.Context, user.ID, string, time.Time) error {
	return nil
}
func (m *memUsersBilling) Delete(context.Context, user.ID) error { return nil }

type fakeProvider struct {
	status billing.PaymentStatus
}

func (f *fakeProvider) CreatePayment(context.Context, string, int64, string) (string, string, error) {
	return "prov-1", "https://pay.example/1", nil
}
func (f *fakeProvider) FetchPaymentStatus(context.Context, string) (billing.PaymentStatus, error) {
	return f.status, nil
}

func subscriptionErrNotFound(k string) error {
	return user.ErrNotFound(k)
}

func TestWebhookIdempotency_NoDuplicateActivate(t *testing.T) {
	now := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	u := &user.User{ID: "u1", Email: "a@b.c", EntitlementSource: "trial"}
	users := &memUsersBilling{u: u}
	pays := newMemPay()
	orders := newMemOrders()
	prov := &fakeProvider{status: billing.PaymentStatusSucceeded}

	svc := billing.NewService(users, pays, orders, prov, []billing.Plan{
		{Code: "personal_basic", Title: "Basic", AmountRUB: 299, PeriodDays: 30},
	}, fixedClock{t: now}, logger.New("test", "error"))

	p := &subscription.Payment{
		ID: "pay1", UserID: "u1", ProviderID: "prov-1",
		AmountRUB: 299, PeriodDays: 30, Status: subscription.PaymentPending,
		CreatedAt: now, Tariff: "personal_basic", OrderID: "ord1",
	}
	_ = pays.Create(context.Background(), p)
	_ = orders.Create(context.Background(), &subscription.Order{
		ID: "ord1", UserID: "u1", PlanCode: "personal_basic",
		AmountRUB: 299, PeriodDays: 30, Status: subscription.OrderPending, CreatedAt: now,
	})

	if err := svc.HandleWebhook(context.Background(), "prov-1"); err != nil {
		t.Fatal(err)
	}
	if users.activateCalls != 1 {
		t.Fatalf("activateCalls=%d want 1", users.activateCalls)
	}
	if err := svc.HandleWebhook(context.Background(), "prov-1"); err != nil {
		t.Fatal(err)
	}
	if users.activateCalls != 1 {
		t.Fatalf("duplicate webhook activateCalls=%d want 1", users.activateCalls)
	}
	if u.PlanCode != "personal_basic" {
		t.Fatalf("plan=%s", u.PlanCode)
	}
}

func TestCanceledPaymentDoesNotActivate(t *testing.T) {
	now := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	u := &user.User{ID: "u1", Email: "a@b.c", EntitlementSource: "trial"}
	users := &memUsersBilling{u: u}
	pays := newMemPay()
	prov := &fakeProvider{status: billing.PaymentStatusCanceled}

	svc := billing.NewService(users, pays, newMemOrders(), prov, []billing.Plan{
		{Code: "personal_basic", Title: "Basic", AmountRUB: 299, PeriodDays: 30},
	}, fixedClock{t: now}, logger.New("test", "error"))

	_ = pays.Create(context.Background(), &subscription.Payment{
		ID: "pay1", UserID: "u1", ProviderID: "prov-1",
		AmountRUB: 299, PeriodDays: 30, Status: subscription.PaymentPending,
		CreatedAt: now, Tariff: "personal_basic",
	})

	if err := svc.HandleWebhook(context.Background(), "prov-1"); err != nil {
		t.Fatal(err)
	}
	if users.activateCalls != 0 {
		t.Fatalf("activateCalls=%d want 0", users.activateCalls)
	}
}

func TestListPlansCanonicalOnly(t *testing.T) {
	svc := billing.NewService(nil, nil, nil, nil, []billing.Plan{
		{Code: "personal_basic", Title: "Basic", AmountRUB: 299, PeriodDays: 30},
		{Code: "personal_pro", Title: "Pro", AmountRUB: 499, PeriodDays: 30},
		{Code: "business", Title: "Biz", AmountRUB: 1490, PeriodDays: 30},
		{Code: "month", Title: "alias", AmountRUB: 299, PeriodDays: 30},
	}, fixedClock{t: time.Now()}, logger.New("test", "error"))

	plans := svc.ListPlans()
	if len(plans) != 3 {
		t.Fatalf("len=%d want 3", len(plans))
	}
	if plans[0].Code != "personal_basic" || plans[2].Code != "business" {
		t.Fatalf("unexpected plans: %+v", plans)
	}
}
