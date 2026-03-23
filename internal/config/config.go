package config

import (
	"flag"
	"fmt"
	"time"
)

// Config holds runtime options from CLI flags only.
type Config struct {
	GlobPath      string
	Port          int
	IdleTimeout   time.Duration
	WatchInterval time.Duration
}

// DefaultConfig is used when flags omit values.
var DefaultConfig = Config{
	Port:          5898,
	IdleTimeout:   time.Hour,
	WatchInterval: 10 * time.Second,
}

// Load parses flags and returns validated config.
func Load() (*Config, error) {
	var globPath string
	var port int
	var idleTimeoutMs int
	var watchIntervalMs int

	flag.StringVar(&globPath, "glob-path", "", "Glob pattern for GC log files (required)")
	flag.IntVar(&port, "port", 0, "HTTP port for /metrics and /ui (default 5898)")
	flag.IntVar(&idleTimeoutMs, "idle-timeout", 0, "Idle file close timeout in milliseconds (default 3600000)")
	flag.IntVar(&watchIntervalMs, "watch-interval", 0, "File scan interval in milliseconds (default 10000)")
	flag.Parse()

	return Resolve(globPath, port, idleTimeoutMs, watchIntervalMs)
}

// Resolve applies flag values on top of defaults and validates.
func Resolve(globPath string, port, idleTimeoutMs, watchIntervalMs int) (*Config, error) {
	cfg := DefaultConfig

	if globPath != "" {
		cfg.GlobPath = globPath
	}
	if port > 0 {
		cfg.Port = port
	}
	if idleTimeoutMs > 0 {
		cfg.IdleTimeout = time.Duration(idleTimeoutMs) * time.Millisecond
	}
	if watchIntervalMs > 0 {
		cfg.WatchInterval = time.Duration(watchIntervalMs) * time.Millisecond
	}

	if cfg.GlobPath == "" {
		return nil, fmt.Errorf("glob-path is required")
	}
	if cfg.Port < 1 || cfg.Port > 65535 {
		return nil, fmt.Errorf("port must be between 1 and 65535, got %d", cfg.Port)
	}

	return &cfg, nil
}
