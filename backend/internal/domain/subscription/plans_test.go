package subscription_test

import (
	"testing"

	"streampass/backend/internal/domain/subscription"
)

func TestNormalizePlanCode(t *testing.T) {
	cases := map[string]string{
		"":               subscription.PlanPersonalBasic,
		"basic":          subscription.PlanPersonalBasic,
		"month":          subscription.PlanPersonalBasic,
		"year":           subscription.PlanPersonalBasic,
		"personal_basic": subscription.PlanPersonalBasic,
		"pro":            subscription.PlanPersonalPro,
		"personal_pro":   subscription.PlanPersonalPro,
		"business":       subscription.PlanBusiness,
	}
	for in, want := range cases {
		if got := subscription.NormalizePlanCode(in); got != want {
			t.Fatalf("NormalizePlanCode(%q)=%q want %q", in, got, want)
		}
	}
}

func TestMaxDevicesForPlan(t *testing.T) {
	if subscription.MaxDevicesForPlan("personal_basic") != 2 {
		t.Fatal("basic devices")
	}
	if subscription.MaxDevicesForPlan("pro") != 5 {
		t.Fatal("pro devices")
	}
	if subscription.MaxDevicesForPlan("business") != 5 {
		t.Fatal("business devices")
	}
	if subscription.MaxDevicesForPlan("unknown_sku") != 0 {
		t.Fatal("unknown")
	}
}
