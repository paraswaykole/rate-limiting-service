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

func (b *SlidingWindowLimiter) Check() bool {
	// To implement
	return false
}

func (s *SlidingWindowLimiter) Configure(configuration json.RawMessage) error {
	// To implement
	return nil
}

func (s *SlidingWindowLimiter) PrepareLimiter() {
	// To implement
}

func (s *SlidingWindowLimiter) Sync() {

}
