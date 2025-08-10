package limiter

import (
	"encoding/json"
	"math"
	"time"
)

type TokenBucketLimiter struct {
	Capacity   float64   `json:"capacity"`
	RefillRate float64   `json:"refillRate"`
	Tokens     float64   `json:"tokens"`
	LastRefill time.Time `json:"lastRefill"`
}

func (b *TokenBucketLimiter) Check() bool {
	now := time.Now()
	elapsed := now.Sub(b.LastRefill).Seconds()
	b.Tokens = math.Min(b.Capacity, b.Tokens+elapsed*b.RefillRate)
	b.LastRefill = now
	if b.Tokens >= 1 {
		b.Tokens -= 1
		return true
	}
	return false
}

func (b *TokenBucketLimiter) Configure(data []byte) {
	err := json.Unmarshal(data, &b)
	if err != nil {
		panic(err)
	}
}
