package services

import (
	"sync"
	"time"
)

type Metrics struct {
	TotalRequests int64   `json:"total_requests"`
	Allowed       int64   `json:"allowed"`
	Blocked       int64   `json:"blocked"`
	AvgLatencyMs  float64 `json:"avg_latency_ms"`
}

var (
	metricsData = struct {
		sync.Mutex
		Data Metrics
	}{}
)

func GetMetrics() Metrics {
	metricsData.Lock()
	defer metricsData.Unlock()
	return metricsData.Data
}

func ResetMetrics() {
	metricsData.Lock()
	defer metricsData.Unlock()
	metricsData.Data = Metrics{}
}

func UpdateMetrics(allowed bool, latency time.Duration) {
	metricsData.Lock()
	defer metricsData.Unlock()

	metricsData.Data.TotalRequests++

	if allowed {
		metricsData.Data.Allowed++
	} else {
		metricsData.Data.Blocked++
	}
	total := float64(metricsData.Data.TotalRequests)
	metricsData.Data.AvgLatencyMs = ((metricsData.Data.AvgLatencyMs * (total - 1)) + (float64(latency.Nanoseconds()) / 1e6)) / total
}
