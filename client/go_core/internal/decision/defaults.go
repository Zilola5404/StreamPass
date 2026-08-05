package decision

// DefaultDirectRules are always prepended before backend rules (ТЗ §6: *.ru, банки, гос).
// Ensures Russian traffic stays DIRECT even when the published rule set is sparse.
var DefaultDirectRules = []Rule{
	{Kind: KindDomain, Pattern: "*.ru", Mode: ModeDirect},
	{Kind: KindDomain, Pattern: "*.su", Mode: ModeDirect},
	{Kind: KindDomain, Pattern: "*.xn--p1ai", Mode: ModeDirect},
	{Kind: KindDomain, Pattern: "*.рф", Mode: ModeDirect},
	{Kind: KindDomain, Pattern: "*.mos.ru", Mode: ModeDirect},
	{Kind: KindDomain, Pattern: "*.gov.ru", Mode: ModeDirect},
	{Kind: KindDomain, Pattern: "*.edu.ru", Mode: ModeDirect},
	{Kind: KindDomain, Pattern: "*.mil.ru", Mode: ModeDirect},
}

// MergeWithDefaults prepends built-in DIRECT rules for Russian TLDs.
func MergeWithDefaults(rules []Rule) []Rule {
	if len(DefaultDirectRules) == 0 {
		return rules
	}
	out := make([]Rule, 0, len(DefaultDirectRules)+len(rules))
	out = append(out, DefaultDirectRules...)
	out = append(out, rules...)
	return out
}
