package logger

import (
	"encoding/csv"
	"fmt"
	"os"
	"sync"
	"time"
)

type logEntry struct {
	timestamp string
	allowed   bool
	latencyMs float64
}

var (
	logFile *os.File
	writer  *csv.Writer
	logChan chan logEntry
	once    sync.Once
)

func InitLogger(path string, bufferSize int) error {
	var err error
	once.Do(func() {
		logFile, err = os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return
		}
		writer = csv.NewWriter(logFile)

		// Write header if file is new
		if info, _ := logFile.Stat(); info.Size() == 0 {
			writer.Write([]string{"timestamp", "allowed", "latency_ms"})
			writer.Flush()
		}

		logChan = make(chan logEntry, bufferSize)
		go processLogs()
	})
	return err
}

func processLogs() {
	for entry := range logChan {
		writer.Write([]string{
			entry.timestamp,
			formatBool(entry.allowed),
			formatFloat(entry.latencyMs),
		})
		writer.Flush()
	}
}

func LogRequestAsync(allowed bool, latency time.Duration) {
	logChan <- logEntry{
		timestamp: time.Now().Format(time.RFC3339Nano),
		allowed:   allowed,
		latencyMs: float64(latency.Nanoseconds()) / 1e6, // convert ns → ms with decimal
	}
}

func formatBool(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

func formatFloat(f float64) string {
	return fmt.Sprintf("%.3f", f) // 3 decimal places (e.g., 12.345 ms)
}
