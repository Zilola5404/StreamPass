package subscription_test

import (
	"testing"
	"time"

	"streampass/backend/internal/domain/subscription"
)

func TestDeriveInfo_TrialActive(t *testing.T) {
	now := time.Date(2026, 9, 1, 14, 0, 0, 0, time.UTC)
	end := now.Add(72 * time.Hour)

	info := subscription.DeriveInfo(subscription.EntitlementInput{
		ActiveUntil: &end,
		TrialEndsAt: &end,
		Source:      "trial",
		Now:         now,
	})
	if info.Status != subscription.StatusTrial {
		t.Fatalf("status=%s want TRIAL", info.Status)
	}
	if !info.AccessAllowed {
		t.Fatal("trial must allow connect")
	}
	if info.HoursLeft < 71 || info.HoursLeft > 72 {
		t.Fatalf("hoursLeft=%d want ~72", info.HoursLeft)
	}
}

func TestDeriveInfo_TrialExpired(t *testing.T) {
	now := time.Date(2026, 9, 4, 15, 0, 0, 0, time.UTC)
	end := time.Date(2026, 9, 4, 14, 0, 0, 0, time.UTC)

	info := subscription.DeriveInfo(subscription.EntitlementInput{
		ActiveUntil: &end,
		TrialEndsAt: &end,
		Source:      "trial",
		Now:         now,
	})
	if info.Status != subscription.StatusExpired {
		t.Fatalf("status=%s want EXPIRED", info.Status)
	}
	if info.AccessAllowed {
		t.Fatal("expired trial must block connect")
	}
	if info.ErrorCode != "TRIAL_EXPIRED" {
		t.Fatalf("error_code=%q want TRIAL_EXPIRED", info.ErrorCode)
	}
}

func TestDeriveInfo_PaidActive(t *testing.T) {
	now := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	trialEnd := now.Add(-time.Hour)
	paidUntil := now.Add(30 * 24 * time.Hour)

	info := subscription.DeriveInfo(subscription.EntitlementInput{
		ActiveUntil: &paidUntil,
		TrialEndsAt: &trialEnd,
		Source:      "paid",
		PlanCode:    "personal_pro",
		Now:         now,
	})
	if info.Status != subscription.StatusActive {
		t.Fatalf("status=%s want ACTIVE", info.Status)
	}
	if !info.AccessAllowed {
		t.Fatal("paid must allow connect")
	}
	if info.PlanCode != "personal_pro" {
		t.Fatalf("plan=%s", info.PlanCode)
	}
}

func TestDeriveInfo_CanceledStillAccess(t *testing.T) {
	now := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	until := now.Add(10 * 24 * time.Hour)
	canceled := now.Add(-time.Hour)

	info := subscription.DeriveInfo(subscription.EntitlementInput{
		ActiveUntil: &until,
		Source:      "paid",
		CanceledAt:  &canceled,
		Now:         now,
	})
	if info.Status != subscription.StatusCanceled {
		t.Fatalf("status=%s want CANCELED", info.Status)
	}
	if !info.AccessAllowed {
		t.Fatal("canceled with remaining time must still allow connect")
	}
}

func TestNewInfoWithTrial(t *testing.T) {
	now := time.Date(2026, 8, 25, 12, 0, 0, 0, time.UTC)
	end := now.Add(48 * time.Hour)

	info := subscription.NewInfoWithTrial(&end, &end, "trial", now)
	if info.Status != subscription.StatusTrial {
		t.Fatalf("status=%s want TRIAL", info.Status)
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
