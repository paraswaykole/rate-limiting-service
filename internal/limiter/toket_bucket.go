package limiter

import (
	"encoding/json"
	"fmt"
	"math"
	"rate-limiting-service/internal/storage"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"
)

type TokenBucketLimiter struct {
	lock       sync.Mutex `json:"-"`
	key        string     `json:"-"`
	args       []string   `json:"-"`
	Capacity   float64    `json:"capacity"`
	RefillRate float64    `json:"refillRate"`
	Tokens     float64    `json:"tokens"`
	LastRefill time.Time  `json:"lastRefill"`
}

func (b *TokenBucketLimiter) Configure(configuration json.RawMessage) error {
	var configurationData struct {
		Capacity   float64 `json:"capacity" validate:"required" message:"capacity is required"`
		RefillRate float64 `json:"refillRate" validate:"required" message:"refillRate is required"`
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	err := json.Unmarshal(configuration, &configurationData)
	if err != nil {
		return err
	}
	err = validate.Struct(configurationData)
	if err != nil {
		return err
	}

	b.Capacity = configurationData.Capacity
	b.RefillRate = configurationData.RefillRate
	storage.GetManager().SetConfigureData(b.key, TOKEN_BUCKET, b)
	return nil
}

func (b *TokenBucketLimiter) Check() bool {
	b.lock.Lock()
	defer b.lock.Unlock()
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

func (b *TokenBucketLimiter) PrepareLimiter() {
	limiterKey := GetLimiterKey(TOKEN_BUCKET, b.key, b.args)
	data, err := storage.GetManager().GetLimiterData(limiterKey)
	if err == nil {
		err = json.Unmarshal([]byte(data), &b)
		if err != nil {
			panic(err)
		}
	}
	if err != nil {
		if err.Error() == storage.ErrDataNotFound {
			err = storage.GetManager().GetConfigureData(b.key, b)
			if err != nil && err.Error() == storage.ErrDataNotFound {
				panic("rate limiter not configured")
			}
		} else {
			panic(err)
		}
	}
}

func (b *TokenBucketLimiter) Sync() {
	limiterKey := GetLimiterKey(TOKEN_BUCKET, b.key, b.args)
	ttlSeconds := int(math.Ceil(b.Capacity/b.RefillRate)) + 2
	data, err := json.Marshal(b)
	if err != nil {
		fmt.Println("error storing limiter data", err)
		return
	}
	storage.GetManager().SetLimiterData(limiterKey, data, ttlSeconds)
}
