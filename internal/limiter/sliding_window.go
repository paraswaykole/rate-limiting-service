package limiter

import (
	"encoding/json"
)

type SlidingWindowLimiter struct {
	// To implement
}

func (b *SlidingWindowLimiter) Check() bool {
	// To implement
	return false
}

func (s *SlidingWindowLimiter) Configure(data []byte) {
	err := json.Unmarshal(data, &s)
	if err != nil {
		panic(err)
	}
}
