package limiter

type manager struct {
	limiters map[string]*Limiter
}

var instance *manager

func GetManager() *manager {
	if instance == nil {
		instance = &manager{
			limiters: map[string]*Limiter{},
		}
	}
	return instance
}

func (m *manager) GetLimiter(key string, args []string) *Limiter {
	limiterType, err := GetLimiterTypeForKey(key)
	if err != nil {
		return nil
	}

	limiterKey := GetLimiterKey(limiterType, key, args)
	if _, exists := m.limiters[limiterKey]; exists {
		return m.limiters[limiterKey]
	}

	rateLimiter := NewLimiter(key, args, limiterType)
	rateLimiter.PrepareLimiter()
	m.limiters[limiterKey] = &rateLimiter

	return &rateLimiter
}
