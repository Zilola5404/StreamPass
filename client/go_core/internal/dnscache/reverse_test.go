package dnscache

import "testing"

func TestRememberIP_HostForIP(t *testing.T) {
	RememberIP("Ya.Ru.", "77.88.8.8")
	if got := HostForIP("77.88.8.8"); got != "ya.ru" {
		t.Fatalf("HostForIP= %q want ya.ru", got)
	}
	RememberIP("youtube.com", "142.251.208.118")
	if got := HostForIP("142.251.208.118"); got != "youtube.com" {
		t.Fatalf("HostForIP= %q", got)
	}
}
