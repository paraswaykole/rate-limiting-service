package limiter

import (
	"rate-limiting-service/internal/config"
	"sync"
	"time"
)

type limiterInstance struct {
	Limiter  *Limiter
	LastUsed time.Time
}

type manager struct {
	limiters   map[string]*limiterInstance
	lastSynced time.Time
	lock       *sync.Mutex
}

var instance *manager

func GetManager() *manager {
	if instance == nil {
		instance = &manager{
			limiters: map[string]*limiterInstance{},
			lock:     &sync.Mutex{},
		}
	}
	return instance
}

func (m *manager) AccessLimiter(key string, args []string) *Limiter {
	limiterType, err := GetLimiterTypeForKey(key)
	if err != nil {
		return nil
	}

	limiterKey := GetLimiterKey(limiterType, key, args)
	if instance, exists := m.limiters[limiterKey]; exists {
		instance.LastUsed = time.Now()
		return m.limiters[limiterKey].Limiter
	}

	rateLimiter := NewLimiter(key, args, limiterType)
	rateLimiter.prepareLimiter()
	rateLimiter.subscribeUpdates()
	m.lock.Lock()
	m.limiters[limiterKey] = &limiterInstance{
		LastUsed: time.Now(),
		Limiter:  &rateLimiter,
	}
	m.lock.Unlock()

	return &rateLimiter
}

func (m *manager) SyncLimiters() {
	now := time.Now()
	for key, value := range m.limiters {
		if value.LastUsed.Sub(m.lastSynced).Microseconds() > config.SYNC_LIMITER_FREQUENCY_TIME_IN_MS {
			(*value.Limiter).sync()
		}
		if (*value.Limiter).isExpired() {
			m.lock.Lock()
			delete(m.limiters, key)
			m.lock.Unlock()
			go (*value.Limiter).clear()
		}
	}
	m.lastSynced = now
}

func (m *manager) StopAll() {
	m.lock.Lock()
	for key, value := range m.limiters {
		delete(m.limiters, key)
		(*value.Limiter).clear()
	}
	m.lock.Unlock()
}
