package dnscache_test

import (
	"testing"

	"streampass/go_core/internal/dnscache"
)

// TestTrafficDNSMatrix documents which DNS path each site category uses.
func TestTrafficDNSMatrix(t *testing.T) {
	cases := []struct {
		host string
		want string
		note string
	}{
		{"gosuslugi.ru", "yandex", "RU gov — correct geo DNS"},
		{"yandex.ru", "yandex", "RU search"},
		{"sberbank.ru", "yandex", "RU bank web"},
		{"2ip.ru", "yandex", "geo IP check"},
		{"youtube.com", "doh", "foreign — Cloudflare DoH"},
		{"instagram.com", "doh", "foreign — Cloudflare DoH"},
		{"google.com", "doh", "foreign default"},
		{"example.com", "doh", "generic foreign"},
	}
	for _, tc := range cases {
		if got := dnscache.PreferredDNSRoute(tc.host); got != tc.want {
			t.Errorf("PreferredDNSRoute(%q)=%q want %q (%s)", tc.host, got, tc.want, tc.note)
		}
	}
}
