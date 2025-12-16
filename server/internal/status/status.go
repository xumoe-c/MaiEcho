package status

import (
	"runtime"
	"time"

	"github.com/xumoe-c/maiecho/server/internal/logger"
)

type SystemStatus struct {
	Uptime       time.Duration `json:"uptime"`
	Goroutines   int           `json:"goroutines"`
	MemoryUsage  uint64        `json:"memory_usage_mb"`
	ActiveTasks  int           `json:"active_tasks"`
	LastLogEntry string        `json:"last_log_entry"`
}

var (
	startTime time.Time
)

func Init() {
	startTime = time.Now()
}

func GetSystemStatus() SystemStatus {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return SystemStatus{
		Uptime:       time.Since(startTime),
		Goroutines:   runtime.NumGoroutine(),
		MemoryUsage:  m.Alloc / 1024 / 1024,
		ActiveTasks:  0, // TODO: Connect to task manager
		LastLogEntry: logger.GetLastLog(),
	}
}

func LogStatus() {
	status := GetSystemStatus()
	logger.Info("系统状态",
		"module", "status",
		"uptime", status.Uptime,
		"goroutines", status.Goroutines,
		"memory_mb", status.MemoryUsage,
	)
}
