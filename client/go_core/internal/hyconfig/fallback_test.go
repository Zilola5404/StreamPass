package hyconfig

import "testing"

func TestFallbackCandidates_primaryFirst(t *testing.T) {
	got := FallbackCandidates("relay.example.com", 443)
	if len(got) != 5 {
		t.Fatalf("len=%d want 5: %#v", len(got), got)
	}
	if got[0].Port != 443 || got[0].Network != "udp" || got[0].Host != "relay.example.com:443" {
		t.Fatalf("first=%#v", got[0])
	}
	if got[1].Port != 8443 || got[1].Network != "udp" {
		t.Fatalf("udp2=%#v", got[1])
	}
	if got[2].Port != 24443 || got[2].Network != "udp" {
		t.Fatalf("udp3=%#v", got[2])
	}
	if got[3].Network != "tcp" || got[3].Port != 8443 {
		t.Fatalf("tcp1=%#v", got[3])
	}
	if got[4].Network != "tcp" || got[4].Port != 24443 {
		t.Fatalf("tcp2=%#v", got[4])
	}
}

func TestFallbackCandidates_dedupePrimary(t *testing.T) {
	got := FallbackCandidates("10.0.0.1", 8443)
	// udp/8443, udp/24443, tcp/8443, tcp/24443
	if len(got) != 4 {
		t.Fatalf("len=%d want 4: %#v", len(got), got)
	}
	if got[0].Port != 8443 || got[0].Network != "udp" {
		t.Fatalf("got %#v", got[0])
	}
}

func TestFallbackCandidates_hostPortInHost(t *testing.T) {
	got := FallbackCandidates("relay.example.com:443", 0)
	if got[0].Port != 443 || got[0].Network != "udp" {
		t.Fatalf("got %#v", got[0])
	}
}

func TestDialCandidateString(t *testing.T) {
	c := DialCandidate{Network: "tcp", Port: 8443}
	if c.String() != "tcp/8443" {
		t.Fatalf("got %s", c.String())
	}
}
