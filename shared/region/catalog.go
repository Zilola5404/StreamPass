// Package region defines the canonical StreamPass relay geography
// (ТЗ §9 / Этап 6): Frankfurt, Amsterdam, Warsaw, Helsinki.
package region

import "strings"

// Info is one entry in the operator/client region catalog.
type Info struct {
	Code    string `json:"code"`
	City    string `json:"city"`
	Country string `json:"country"`
	Label   string `json:"label"`
}

// Catalog is the ordered list shown in Admin and client pickers.
var Catalog = []Info{
	{Code: "de", City: "Frankfurt", Country: "Germany", Label: "Frankfurt (DE)"},
	{Code: "nl", City: "Amsterdam", Country: "Netherlands", Label: "Amsterdam (NL)"},
	{Code: "pl", City: "Warsaw", Country: "Poland", Label: "Warsaw (PL)"},
	{Code: "fi", City: "Helsinki", Country: "Finland", Label: "Helsinki (FI)"},
}

// ByCode returns catalog info for a canonical code, or false if unknown.
func ByCode(code string) (Info, bool) {
	code = strings.ToLower(strings.TrimSpace(code))
	for _, info := range Catalog {
		if info.Code == code {
			return info, true
		}
	}
	return Info{}, false
}

// Normalize maps free-form operator input (aliases, city names, legacy
// uppercase codes like "NL") onto a canonical catalog code. Unknown
// values are lowercased and trimmed so existing custom regions keep
// working, but IsKnown reports false for them.
func Normalize(raw string) string {
	s := strings.ToLower(strings.TrimSpace(raw))
	if s == "" {
		return ""
	}
	if info, ok := ByCode(s); ok {
		return info.Code
	}
	switch {
	case containsAny(s, "frankfurt", "germany", "deutschland"):
		return "de"
	case containsAny(s, "amsterdam", "netherlands", "holland"):
		return "nl"
	case containsAny(s, "warsaw", "warszawa", "poland", "polska"):
		return "pl"
	case containsAny(s, "helsinki", "finland", "suomi"):
		return "fi"
	default:
		return s
	}
}

// IsKnown reports whether Normalize(raw) is a catalog code.
func IsKnown(raw string) bool {
	_, ok := ByCode(Normalize(raw))
	return ok
}

// LabelOf returns a human-readable label for a region code, or the raw
// code itself when unknown.
func LabelOf(raw string) string {
	code := Normalize(raw)
	if info, ok := ByCode(code); ok {
		return info.Label
	}
	if code == "" {
		return raw
	}
	return code
}

func containsAny(s string, parts ...string) bool {
	for _, p := range parts {
		if strings.Contains(s, p) {
			return true
		}
	}
	return false
}
