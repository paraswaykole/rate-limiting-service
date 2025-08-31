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

func TestSlidingWindowLimiter(t *testing.T) {
	sw := &limiter.SlidingWindowLimiter{
		WindowSize:  2 * time.Second, // 2-second sliding window
		Capacity:    3,               // allow max 3 requests per window
		RequestLogs: []int64{},
	}

	// 1. Should allow first 3 requests immediately
	for i := range 3 {
		if allowed, _ := sw.Check(); !allowed {
			t.Errorf("Expected request %d to be allowed, but it was denied", i+1)
		}
	}

	// 2. Fourth request should be denied
	if allowed, _ := sw.Check(); allowed {
		t.Errorf("Expected request to be denied when limit is reached")
	}

	// 3. Wait for 2.1 seconds (window expires) and check again
	time.Sleep(2100 * time.Millisecond)
	if allowed, _ := sw.Check(); !allowed {
		t.Errorf("Expected request to be allowed after window expired")
	}
}
