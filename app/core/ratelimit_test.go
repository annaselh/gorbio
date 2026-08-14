package core

import (
	"testing"
	"time"
)

func TestRateLimiterAllowsUpToLimitThenBlocks(t *testing.T) {
	limiter := NewRateLimiter(3, time.Minute)

	for attempt := 1; attempt <= 3; attempt++ {
		if !limiter.Allow("10.0.0.1") {
			t.Fatalf("attempt %d should be allowed within the limit", attempt)
		}
	}

	if limiter.Allow("10.0.0.1") {
		t.Fatal("fourth attempt should be blocked")
	}
}

func TestRateLimiterIsolatesKeys(t *testing.T) {
	limiter := NewRateLimiter(1, time.Minute)

	if !limiter.Allow("10.0.0.1") {
		t.Fatal("first key should be allowed")
	}
	if limiter.Allow("10.0.0.1") {
		t.Fatal("first key should now be exhausted")
	}
	if !limiter.Allow("10.0.0.2") {
		t.Fatal("a different key must not inherit another key's count")
	}
}

func TestRateLimiterWindowExpiry(t *testing.T) {
	limiter := NewRateLimiter(1, time.Minute)
	current := time.Now()
	limiter.now = func() time.Time { return current }

	if !limiter.Allow("10.0.0.1") {
		t.Fatal("first attempt should be allowed")
	}
	if limiter.Allow("10.0.0.1") {
		t.Fatal("second attempt inside the window should be blocked")
	}

	current = current.Add(time.Minute + time.Second)
	if !limiter.Allow("10.0.0.1") {
		t.Fatal("a new window should reset the count")
	}
}

func TestRateLimiterResetForgivesFailures(t *testing.T) {
	limiter := NewRateLimiter(1, time.Minute)

	limiter.Allow("10.0.0.1")
	if limiter.Allow("10.0.0.1") {
		t.Fatal("key should be exhausted before reset")
	}

	limiter.Reset("10.0.0.1")
	if !limiter.Allow("10.0.0.1") {
		t.Fatal("reset should clear the bucket so a successful login forgives failures")
	}
}

func TestRateLimiterSweepDropsExpiredBuckets(t *testing.T) {
	limiter := NewRateLimiter(1, time.Minute)
	current := time.Now()
	limiter.now = func() time.Time { return current }

	limiter.Allow("stale")
	current = current.Add(2 * time.Minute)
	limiter.Allow("fresh")

	limiter.mu.Lock()
	_, stalePresent := limiter.buckets["stale"]
	limiter.mu.Unlock()

	if stalePresent {
		t.Fatal("expired bucket should be swept so the map cannot grow without bound")
	}
}

func TestRateLimiterZeroLimitDisablesLimiting(t *testing.T) {
	limiter := NewRateLimiter(0, time.Minute)

	for attempt := 0; attempt < 100; attempt++ {
		if !limiter.Allow("10.0.0.1") {
			t.Fatal("a zero limit should disable rate limiting entirely")
		}
	}
}
