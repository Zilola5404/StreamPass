package subscription_test

import (
	"testing"
	"time"

	"streampass/backend/internal/domain/subscription"
)

func TestNewInfoWithTrial(t *testing.T) {
	now := time.Date(2026, 8, 25, 12, 0, 0, 0, time.UTC)
	end := now.Add(48 * time.Hour)

	info := subscription.NewInfoWithTrial(&end, &end, "trial", now)
	if info.Status != subscription.StatusTrial {
		t.Fatalf("status=%s want TRIAL", info.Status)
	}
	if info.DaysLeft < 1 {
		t.Fatalf("daysLeft=%d", info.DaysLeft)
	}

	paidUntil := now.Add(30 * 24 * time.Hour)
	info = subscription.NewInfoWithTrial(&paidUntil, &end, "paid", now)
	if info.Status != subscription.StatusActive {
		t.Fatalf("status=%s want ACTIVE", info.Status)
	}

	past := now.Add(-time.Hour)
	info = subscription.NewInfoWithTrial(&past, &past, "trial", now)
	if info.Status != subscription.StatusExpired {
		t.Fatalf("status=%s want EXPIRED", info.Status)
	}
}
