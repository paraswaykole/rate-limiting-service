package limiter

import (
	"encoding/json"
	"sync"
)

type SlidingWindowLimiter struct {
	lock sync.Mutex `json:"-"`
	key  string     `json:"-"`
	args []string   `json:"-"`
	// To implement
}

func (b *SlidingWindowLimiter) Check() (bool, map[string]string) {
	// To implement
	return false, nil
}

func (s *SlidingWindowLimiter) Configure(configuration json.RawMessage) error {
	// To implement
	return nil
}

func (s *SlidingWindowLimiter) prepareLimiter() {
	// To implement
}

func (s *SlidingWindowLimiter) sync() {
	// To implement
}

func (b *SlidingWindowLimiter) isExpired() bool {
	// To implement
	return false
}

func (s *SlidingWindowLimiter) publishUpdate() {
	// To implement
}

func (s *SlidingWindowLimiter) subscribeUpdates() {
	// To implement
}

func (s *SlidingWindowLimiter) clear() {
	// To implement
}
