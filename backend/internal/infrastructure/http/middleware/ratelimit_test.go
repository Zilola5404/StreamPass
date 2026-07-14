package middleware

import (
	"testing"
	"time"
)

func TestRateLimiter_AllowsUpToLimitThenBlocks(t *testing.T) {
	rl := NewRateLimiter(3, time.Minute)

	for i := 0; i < 3; i++ {
		if !rl.allow("1.2.3.4") {
			t.Fatalf("request %d should be allowed", i)
		}
	}
	if rl.allow("1.2.3.4") {
		t.Error("4th request should be blocked")
	}
}

func TestRateLimiter_DistinctKeysHaveIndependentBudgets(t *testing.T) {
	rl := NewRateLimiter(1, time.Minute)

	if !rl.allow("1.1.1.1") {
		t.Error("first request for 1.1.1.1 should be allowed")
	}
	if !rl.allow("2.2.2.2") {
		t.Error("first request for 2.2.2.2 should be allowed, independent budget")
	}
	if rl.allow("1.1.1.1") {
		t.Error("second request for 1.1.1.1 should be blocked")
	}
}

func TestRateLimiter_ResetsAfterWindow(t *testing.T) {
	rl := NewRateLimiter(1, 10*time.Millisecond)

	if !rl.allow("3.3.3.3") {
		t.Fatal("first request should be allowed")
	}
	if rl.allow("3.3.3.3") {
		t.Fatal("second request within window should be blocked")
	}
	time.Sleep(20 * time.Millisecond)
	if !rl.allow("3.3.3.3") {
		t.Error("request after window elapses should be allowed again")
	}
}
