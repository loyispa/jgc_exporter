package config

import (
	"flag"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	GlobPath      string        `yaml:"fileGlobPattern"`
	Port          int           `yaml:"port"`
	HostPort      string        `yaml:"hostPort"`
	IdleTimeout   time.Duration `yaml:"idleTimeout"`
	WatchInterval time.Duration `yaml:"watchInterval"`
}

var DefaultConfig = Config{
	Port:          5898,
	IdleTimeout:   time.Hour,
	WatchInterval: 10 * time.Second,
}

func Load() (*Config, error) {
	var configPath string
	var globPath string
	var port int
	var idleTimeoutMs int
	var watchIntervalMs int

	flag.StringVar(&configPath, "config", "", "Path to YAML config file")
	flag.StringVar(&globPath, "glob-path", "", "Glob pattern for GC log files (required if no config)")
	flag.IntVar(&port, "port", 0, "HTTP port for /metrics and /ui")
	flag.IntVar(&idleTimeoutMs, "idle-timeout", 0, "Idle file close timeout in milliseconds")
	flag.IntVar(&watchIntervalMs, "watch-interval", 0, "File scan interval in milliseconds")
	flag.Parse()

	return Resolve(configPath, globPath, port, idleTimeoutMs, watchIntervalMs)
}

func Resolve(configPath, globPath string, port, idleTimeoutMs, watchIntervalMs int) (*Config, error) {
	cfg := DefaultConfig

	if configPath != "" {
		if err := loadYAML(configPath, &cfg); err != nil {
			return nil, fmt.Errorf("load config %s: %w", configPath, err)
		}
	}

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
		return nil, fmt.Errorf("glob-path is required (via --glob-path or fileGlobPattern in config)")
	}
	if cfg.Port < 1 || cfg.Port > 65535 {
		return nil, fmt.Errorf("port must be between 1 and 65535, got %d", cfg.Port)
	}

	return &cfg, nil
}

func loadYAML(path string, cfg *Config) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var raw yamlConfig
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return err
	}

	if raw.FileGlobPattern != "" {
		cfg.GlobPath = raw.FileGlobPattern
	}
	if raw.HostPort != "" {
		cfg.HostPort = raw.HostPort
		if p := parsePort(raw.HostPort); p > 0 {
			cfg.Port = p
		}
	}
	if raw.IdleTimeout > 0 {
		cfg.IdleTimeout = time.Duration(raw.IdleTimeout) * time.Millisecond
	}
	if raw.WatchInterval > 0 {
		cfg.WatchInterval = time.Duration(raw.WatchInterval) * time.Millisecond
	}

	return nil
}

type yamlConfig struct {
	FileGlobPattern string `yaml:"fileGlobPattern"`
	HostPort        string `yaml:"hostPort"`
	IdleTimeout     int    `yaml:"idleTimeout"`
	WatchInterval   int    `yaml:"watchInterval"`
}

func parsePort(hostPort string) int {
	for i := len(hostPort) - 1; i >= 0; i-- {
		if hostPort[i] == ':' {
			port := 0
			for _, c := range hostPort[i+1:] {
				if c < '0' || c > '9' {
					return 0
				}
				port = port*10 + int(c-'0')
			}
			return port
		}
	}
	return 0
}
