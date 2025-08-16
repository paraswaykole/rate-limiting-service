package limiter

import (
	"rate-limiting-service/internal/limiter"
	"testing"
	"time"
)

func TestTokenBucketLimiter(t *testing.T) {
	// Create limiter with capacity 5, refill rate 1 token/sec
	tb := &limiter.TokenBucketLimiter{
		Capacity:   5,
		RefillRate: 1,
		Tokens:     5, // Start full
		LastRefill: time.Now(),
	}

	// 1. Should allow first 5 requests immediately
	for i := range 5 {
		if allowed, _ := tb.Check(); !allowed {
			t.Errorf("Expected request %d to be allowed, but it was denied", i+1)
		}
	}

	// 2. Next request should be denied (empty bucket)
	if allowed, _ := tb.Check(); allowed {
		t.Errorf("Expected request to be denied when bucket is empty")
	}

	// 3. Wait for 2.5 seconds, should get ~2 tokens back
	time.Sleep(2500 * time.Millisecond)
	allowedCount := 0
	for range 3 {
		if allowed, _ := tb.Check(); allowed {
			allowedCount++
		}
	}

	// We should have allowed exactly 2 requests after waiting
	if allowedCount != 2 {
		t.Errorf("Expected 2 requests allowed after refill, got %d", allowedCount)
	}
}
