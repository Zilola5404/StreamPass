package mobile_test

import (
	"testing"

	"streampass/go_core/mobile"
)

// TestTrafficMatrix_publicAPI exposes the same routing contract through gomobile exports.
func TestTrafficMatrix_publicAPI(t *testing.T) {
	rules := `{"version":4,"rules":[
		{"kind":"DOMAIN","pattern":"youtube.com","mode":"RELAY"},
		{"kind":"DOMAIN","pattern":"*.youtube.com","mode":"RELAY"},
		{"kind":"DOMAIN","pattern":"instagram.com","mode":"RELAY"}
	]}`

	cases := []struct {
		host string
		want string
	}{
		{"yandex.ru", "DIRECT"},
		{"gosuslugi.ru", "DIRECT"},
		{"www.youtube.com", "RELAY"},
		{"instagram.com", "RELAY"},
		{"google.com", "DIRECT"},
	}

	for _, tc := range cases {
		got := mobile.DecideRoute(rules, "", tc.host, "")
		if got != tc.want {
			t.Errorf("DecideRoute(%q)=%q want %q", tc.host, got, tc.want)
		}
	}
}
