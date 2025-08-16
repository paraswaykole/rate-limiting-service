package limiter

import (
	"encoding/json"
	"errors"
	"fmt"
	"rate-limiting-service/internal/storage"
	"strings"
	"sync"
)

type LimiterType int

const (
	TOKEN_BUCKET   = 10
	SLIDING_WINDOW = 20
)

type Limiter interface {
	Check() bool
	Configure(json.RawMessage) error
	prepareLimiter()
	sync()
	publishUpdate()
	subscribeUpdates()
	isExpired() bool
	clear()
}

func NewLimiter(key string, args []string, limiterType LimiterType) Limiter {
	switch limiterType {
	case TOKEN_BUCKET:
		return &TokenBucketLimiter{
			lock: sync.Mutex{},
			key:  key,
			args: args,
		}
	case SLIDING_WINDOW:
		return &SlidingWindowLimiter{
			lock: sync.Mutex{},
			key:  key,
			args: args,
		}
	}
	panic("unknown limiter type")
}

var KeyLimiterTypeMap = map[string]LimiterType{}

func GetLimiterTypeForKey(key string) (LimiterType, error) {
	if limiterType, exists := KeyLimiterTypeMap[key]; exists {
		return limiterType, nil
	}
	ltype, err := storage.GetManager().GetConfigureType(key)
	if err == nil {
		KeyLimiterTypeMap[key] = LimiterType(ltype)
		return LimiterType(ltype), nil
	}
	if err.Error() == storage.ErrDataNotFound {
		return -1, errors.New("key not configured")
	}
	panic(err)
}

func GetLimiterKey(limType LimiterType, key string, args []string) string {
	var limiterKey string
	switch limType {
	case TOKEN_BUCKET:
		limiterKey = fmt.Sprintf("limiter:tbl:%s", key)
	case SLIDING_WINDOW:
		limiterKey = fmt.Sprintf("limiter:sw:%s", key)
	}
	if len(args) > 0 {
		limiterKey = fmt.Sprintf("%s:%s", limiterKey, strings.Join(args, ":"))
	}
	return limiterKey
}

func GetUpdatesKey(limType LimiterType, key string, args []string) string {
	var limiterKey string
	switch limType {
	case TOKEN_BUCKET:
		limiterKey = fmt.Sprintf("updates:tbl:%s", key)
	case SLIDING_WINDOW:
		limiterKey = fmt.Sprintf("updates:sw:%s", key)
	}
	if len(args) > 0 {
		limiterKey = fmt.Sprintf("%s:%s", limiterKey, strings.Join(args, ":"))
	}
	return limiterKey
}
