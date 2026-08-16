package tunbridge

import "testing"

func TestIsControlPlaneDest(t *testing.T) {
	cases := []struct {
		host, ip, relay string
		want            bool
	}{
		{"212-43-156-33.nip.io", "212.43.156.33", "212.43.156.33", true},
		{"212-43-156-33.nip.io", "212.43.156.33", "pl-warsaw-1", true}, // nip.io alone
		{"api.example.com", "1.2.3.4", "212.43.156.33", false},
		{"", "212.43.156.33", "212.43.156.33", true},
		{"", "1.2.3.4", "212.43.156.33", false},
		{"localhost", "127.0.0.1", "", true},
		{"chatgpt.com", "8.6.112.0", "212.43.156.33", false},
		{"ya.ru", "87.250.250.242", "212.43.156.33", false},
	}
	for _, c := range cases {
		got := isControlPlaneDest(c.host, c.ip, c.relay)
		if got != c.want {
			t.Fatalf("host=%q ip=%q relay=%q got %v want %v", c.host, c.ip, c.relay, got, c.want)
		}
	}
}
