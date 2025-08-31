package limiter

import (
	"encoding/json"
	"fmt"
	"rate-limiting-service/internal/config"
	"rate-limiting-service/internal/storage"
	"rate-limiting-service/internal/utils"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/redis/go-redis/v9"
)

type SlidingWindowLimiter struct {
	lock        sync.Mutex       `json:"-"`
	key         string           `json:"-"`
	args        []string         `json:"-"`
	sub         *redis.PubSub    `json:"-"`
	syncmap     map[string]int64 `json:"-"`
	Capacity    int              `json:"capacity"`
	WindowSize  time.Duration    `json:"windowSize"`
	RequestLogs []int64          `json:"requestLog"`
	LastUpdated time.Time        `json:"lastUpdated"`
}

func (s *SlidingWindowLimiter) Check() (bool, map[string]string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	now := time.Now()
	cutoff := now.Add(-s.WindowSize)
	filtered := s.RequestLogs[:0]
	for _, t := range s.RequestLogs {
		if time.Unix(0, t).After(cutoff) {
			filtered = append(filtered, t)
		}
	}
	s.RequestLogs = filtered

	allowed := len(s.RequestLogs) < s.Capacity
	if allowed {
		s.RequestLogs = append(s.RequestLogs, now.UnixNano())
		go s.publishUpdate()
	}
	s.LastUpdated = now

	remaining := max(s.Capacity-len(s.RequestLogs), 0)
	reset := 0.0
	if len(s.RequestLogs) > 0 {
		reset = s.WindowSize.Seconds() - now.Sub(time.Unix(0, s.RequestLogs[0])).Seconds()
	}
	headers := map[string]string{
		"X-RateLimit-Limit":     fmt.Sprintf("%d", s.Capacity),
		"X-RateLimit-Remaining": fmt.Sprintf("%d", remaining),
		"X-RateLimit-Reset":     fmt.Sprintf("%.0f", reset),
	}
	return allowed, headers
}

func (s *SlidingWindowLimiter) Configure(configuration json.RawMessage) error {
	var configurationData struct {
		Capacity         int `json:"capacity" validate:"required" message:"capacity is required"`
		WindowSizeInSecs int `json:"windowSize" validate:"required" message:"windowSize in seconds is required"`
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

	s.Capacity = configurationData.Capacity
	s.WindowSize = time.Second * time.Duration(configurationData.WindowSizeInSecs)
	s.RequestLogs = []int64{}
	s.LastUpdated = time.Now()
	storage.GetManager().SetConfigureData(s.key, SLIDING_WINDOW, s)
	return nil
}

func (s *SlidingWindowLimiter) prepareLimiter() {
	limiterKey := GetLimiterKey(SLIDING_WINDOW, s.key, s.args)
	err := storage.GetManager().GetLimiterData(limiterKey, s)
	if err != nil {
		if err.Error() == storage.ErrDataNotFound {
			err = storage.GetManager().GetConfigureData(s.key, s)
			if err != nil && err.Error() == storage.ErrDataNotFound {
				panic("rate limiter not configured")
			}
		} else {
			panic(err)
		}
	}
}

func (s *SlidingWindowLimiter) sync() {
	limiterKey := GetLimiterKey(SLIDING_WINDOW, s.key, s.args)
	previouslastUpdated, _ := storage.GetManager().GetLimiterField(limiterKey, "lastUpdated")
	if previouslastUpdated != "" {
		previouslastUpdatedTime, _ := time.Parse(time.RFC3339Nano, previouslastUpdated)
		if previouslastUpdatedTime.UnixNano() >= s.LastUpdated.UnixNano() {
			return
		}
	}
	ttlSeconds := int(s.WindowSize)/int(time.Second)*2 + 2
	storage.GetManager().SetLimiterData(limiterKey, s, ttlSeconds)
}

func (s *SlidingWindowLimiter) isExpired() bool {
	return time.Since(s.LastUpdated) > s.WindowSize*2
}

func (s *SlidingWindowLimiter) publishUpdate() {
	updatesKey := GetUpdatesKey(SLIDING_WINDOW, s.key, s.args)
	state := map[string]any{
		"requestLogs": s.RequestLogs,
		"lastUpdated": s.LastUpdated.UnixNano(),
		"instanceId":  config.RATE_LIMITING_INSTANCE_ID,
	}
	jsonData, _ := json.Marshal(state)
	storage.GetManager().PublishUpdates(updatesKey, jsonData)
}

func (s *SlidingWindowLimiter) subscribeUpdates() {
	updatesKey := GetUpdatesKey(SLIDING_WINDOW, s.key, s.args)
	s.sub = storage.GetManager().SubscribeUpdates(updatesKey)
	s.syncmap = map[string]int64{}
	ch := s.sub.Channel()
	go func() {
		for msg := range ch {
			var state map[string]any
			if err := json.Unmarshal([]byte(msg.Payload), &state); err != nil {
				continue
			}
			instanceId := state["instanceId"].(string)
			if instanceId == config.RATE_LIMITING_INSTANCE_ID {
				continue
			}
			reqlogs, err := utils.GetFloat64Slice(state, "requestLogs")
			if err != nil {
				panic("request log invalid")
			}
			lastUpdatedNano := int64(state["lastUpdated"].(float64))
			s.lock.Lock()
			if lastUpdatedNano > s.LastUpdated.UnixNano() {
				lastUpdated := time.Unix(0, lastUpdatedNano)
				instanceLastSyncedNano := s.syncmap[instanceId]
				cutoffTime := time.Unix(0, instanceLastSyncedNano)

				newestReqLogs := []int64{}
				for _, t := range reqlogs {
					if time.Unix(0, int64(t)).After(cutoffTime) {
						newestReqLogs = append(newestReqLogs, int64(t))
					}
				}
				s.RequestLogs = append(s.RequestLogs, newestReqLogs...)
				s.LastUpdated = lastUpdated
				s.syncmap[instanceId] = lastUpdatedNano
			}
			s.lock.Unlock()
		}
	}()
}

func (s *SlidingWindowLimiter) clear() {
	s.sub.Close()
}
