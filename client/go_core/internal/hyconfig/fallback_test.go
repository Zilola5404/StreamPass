package hyconfig

import "testing"

func TestFallbackCandidates_primaryFirst(t *testing.T) {
	got := FallbackCandidates("relay.example.com", 443)
	if len(got) != 3 {
		t.Fatalf("len=%d want 3: %#v", len(got), got)
	}
	if got[0].Port != 443 || got[0].Host != "relay.example.com:443" {
		t.Fatalf("first=%#v", got[0])
	}
	if got[1].Port != 8443 || got[2].Port != 24443 {
		t.Fatalf("order=%#v", got)
	}
}

func TestFallbackCandidates_dedupePrimary(t *testing.T) {
	got := FallbackCandidates("10.0.0.1", 8443)
	if len(got) != 2 {
		t.Fatalf("len=%d want 2", len(got))
	}
	if got[0].Port != 8443 || got[1].Port != 24443 {
		t.Fatalf("got %#v", got)
	}
}

func TestFallbackCandidates_hostPortInHost(t *testing.T) {
	got := FallbackCandidates("relay.example.com:443", 0)
	if got[0].Port != 443 {
		t.Fatalf("got %#v", got[0])
	}
}
