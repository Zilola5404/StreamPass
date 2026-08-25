package dnscache

import "testing"

func TestInvalidateAfterIdleClearsCacheAndPins(t *testing.T) {
	c := New(16)
	c.PutRaw(1, "example.com.", []byte{1, 2, 3}, 0)
	if _, ok := c.GetRaw(1, "example.com."); !ok {
		t.Fatal("expected cache hit before clear")
	}
	c.Clear()
	if _, ok := c.GetRaw(1, "example.com."); ok {
		t.Fatal("expected empty cache after Clear")
	}

	RememberIP("yt.example", "1.2.3.4")
	PinRelayIP("yt.example", "1.2.3.4")
	PinDirectIP("2ip.ru", "5.6.7.8")
	ClearSessionMaps()
	if HostForIP("1.2.3.4") != "" {
		t.Fatalf("reverse map should be cleared, got %q", HostForIP("1.2.3.4"))
	}
	if RelayHostNearIP("1.2.3.4") != "" {
		t.Fatal("relay pin should be cleared")
	}
	if DirectHostNearIP("5.6.7.8") != "" {
		t.Fatal("direct pin should be cleared")
	}
}
