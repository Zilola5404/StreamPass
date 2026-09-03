package subscription

// Canonical product plan codes (BILLING-002). Prices live in Backend config / Plan Catalog API —
// never hard-code RUB amounts in Flutter or payment provider dashboards as SSOT.
const (
	PlanPersonalBasic = "personal_basic"
	PlanPersonalPro   = "personal_pro"
	PlanBusiness      = "business"
)

// CanonicalPlanCodes is the ordered MVP catalog exposed by GET /plans (card/SBP mode).
var CanonicalPlanCodes = []string{PlanPersonalBasic, PlanPersonalPro, PlanBusiness}

// NormalizePlanCode maps legacy aliases to canonical codes.
// Unknown non-empty codes are returned trimmed as-is (caller validates against catalog).
func NormalizePlanCode(code string) string {
	switch code {
	case "", PlanPersonalBasic, "basic", "month":
		if code == "" {
			return PlanPersonalBasic
		}
		return PlanPersonalBasic
	case PlanPersonalPro, "pro":
		return PlanPersonalPro
	case PlanBusiness:
		return PlanBusiness
	case "year":
		// Yearly SKU historically mapped to Basic entitlement.
		return PlanPersonalBasic
	default:
		return code
	}
}

// IsCanonicalPlanCode reports whether code normalizes to one of the three MVP plans.
func IsCanonicalPlanCode(code string) bool {
	switch NormalizePlanCode(code) {
	case PlanPersonalBasic, PlanPersonalPro, PlanBusiness:
		return true
	default:
		return false
	}
}

// MaxDevicesForPlan returns the MVP device limit for a plan code.
// Empty / unknown → 0 (caller keeps auth.max_devices fallback).
func MaxDevicesForPlan(planCode string) int {
	switch NormalizePlanCode(planCode) {
	case PlanPersonalBasic:
		return 2
	case PlanPersonalPro:
		return 5
	case PlanBusiness:
		return 5
	default:
		return 0
	}
}

// MaxUsersForPlan returns MVP seat limit (Business = org seats; personal = 1).
func MaxUsersForPlan(planCode string) int {
	switch NormalizePlanCode(planCode) {
	case PlanBusiness:
		return 5
	case PlanPersonalBasic, PlanPersonalPro:
		return 1
	default:
		return 0
	}
}
