package limiter

type LimiterType int

const (
	TOKEN_BUCKET = iota
	SLIDING_WINDOW
)

type Limiter interface {
	Check() bool
	Configure([]byte)
}

func NewLimiter(limiterType LimiterType) Limiter {
	switch limiterType {
	case TOKEN_BUCKET:
		return &TokenBucketLimiter{}
	case SLIDING_WINDOW:
		return &SlidingWindowLimiter{}
	}
	panic("unknown limiter type")
}
