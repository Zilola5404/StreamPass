package dnscache_test

import (
	"testing"

	"streampass/go_core/internal/dnscache"
)

func TestIsRussianDomain(t *testing.T) {
	cases := []struct {
		host string
		want bool
	}{
		{"gosuslugi.ru", true},
		{"www.sberbank.ru.", true},
		{"example.su", true},
		{"site.xn--p1ai", true},
		{"youtube.com", false},
		{"instagram.com", false},
		{"", false},
	}
	for _, tc := range cases {
		if got := dnscache.IsRussianDomain(tc.host); got != tc.want {
			t.Errorf("IsRussianDomain(%q)=%v want %v", tc.host, got, tc.want)
		}
	}
}
