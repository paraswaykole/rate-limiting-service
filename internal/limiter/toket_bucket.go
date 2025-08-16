package limiter

import (
	"encoding/json"
	"math"
	"rate-limiting-service/internal/config"
	"rate-limiting-service/internal/storage"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/redis/go-redis/v9"
)

type TokenBucketLimiter struct {
	lock       sync.Mutex    `json:"-"`
	key        string        `json:"-"`
	args       []string      `json:"-"`
	sub        *redis.PubSub `json:"-"`
	Capacity   float64       `json:"capacity"`
	RefillRate float64       `json:"refillRate"`
	Tokens     float64       `json:"tokens"`
	LastRefill time.Time     `json:"lastRefill"`
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
		go b.publishUpdate()
		return true
	}
	return false
}

func (b *TokenBucketLimiter) prepareLimiter() {
	limiterKey := GetLimiterKey(TOKEN_BUCKET, b.key, b.args)
	err := storage.GetManager().GetLimiterData(limiterKey, b)
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

func (b *TokenBucketLimiter) sync() {
	limiterKey := GetLimiterKey(TOKEN_BUCKET, b.key, b.args)
	ttlSeconds := int(math.Ceil(b.Capacity/b.RefillRate)) + 2
	storage.GetManager().SetLimiterData(limiterKey, b, ttlSeconds)
}

func (b *TokenBucketLimiter) isExpired() bool {
	ttlSeconds := int(math.Ceil(b.Capacity/b.RefillRate)) + 2
	expiry := b.LastRefill.Add(time.Duration(ttlSeconds) * time.Second)
	if time.Now().UnixNano() > expiry.UnixNano() {
		return true
	}
	return false
}

func (b *TokenBucketLimiter) publishUpdate() {
	updatesKey := GetUpdatesKey(TOKEN_BUCKET, b.key, b.args)
	state := map[string]any{
		"tokens":     b.Tokens,
		"lastRefill": b.LastRefill.UnixNano(),
		"instanceId": config.RATE_LIMITING_INSTANCE_ID,
	}
	jsonData, _ := json.Marshal(state)
	storage.GetManager().PublishUpdates(updatesKey, jsonData)
}

func (b *TokenBucketLimiter) subscribeUpdates() {
	updatesKey := GetUpdatesKey(TOKEN_BUCKET, b.key, b.args)
	b.sub = storage.GetManager().SubscribeUpdates(updatesKey)
	ch := b.sub.Channel()
	go func() {
		for msg := range ch {
			var state map[string]any
			if err := json.Unmarshal([]byte(msg.Payload), &state); err != nil {
				continue
			}
			if state["instanceId"].(string) == config.RATE_LIMITING_INSTANCE_ID {
				continue
			}
			tokens, _ := state["tokens"].(float64)
			lastRefillNano := int64(state["lastRefill"].(float64))
			b.lock.Lock()
			if lastRefillNano > b.LastRefill.UnixNano() {
				lastRefill := time.Unix(0, int64(state["lastRefill"].(float64)))
				b.Tokens = tokens
				b.LastRefill = lastRefill
			}
			b.lock.Unlock()
		}
	}()
}

func (b *TokenBucketLimiter) clear() {
	b.sub.Close()
}
