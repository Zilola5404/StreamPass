package tunbridge

import "testing"

func TestSameUDPEndpoint(t *testing.T) {
	cases := []struct {
		a, b string
		want bool
	}{
		{"1.1.1.1:53", "1.1.1.1:53", true},
		{"1.1.1.1:53", "1.1.1.1:54", false},
		{"8.8.8.8:53", "1.1.1.1:53", false},
		{"127.0.0.1:443", "127.0.0.1:443", true},
		{"[::1]:53", "[::1]:53", true},
		{"not-an-addr", "1.1.1.1:53", false},
	}
	for _, tc := range cases {
		if got := sameUDPEndpoint(tc.a, tc.b); got != tc.want {
			t.Fatalf("sameUDPEndpoint(%q,%q)=%v want %v", tc.a, tc.b, got, tc.want)
		}
	}
}
