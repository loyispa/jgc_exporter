package config

import (
	"flag"
	"os"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	if DefaultConfig.Port != 5898 {
		t.Fatalf("expected default port 5898, got %d", DefaultConfig.Port)
	}
}

func TestResolveWithGlobPath(t *testing.T) {
	cfg, err := Resolve("/var/log/gc*.log", 0, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.GlobPath != "/var/log/gc*.log" {
		t.Fatalf("expected glob path, got %q", cfg.GlobPath)
	}
	if cfg.Port != 5898 {
		t.Fatalf("expected default port 5898, got %d", cfg.Port)
	}
}

func TestResolveNoGlobPath(t *testing.T) {
	_, err := Resolve("", 0, 0, 0)
	if err == nil {
		t.Fatal("expected error when glob-path is missing")
	}
}

func TestResolveInvalidPort(t *testing.T) {
	_, err := Resolve("/tmp/*.log", 99999, 0, 0)
	if err == nil {
		t.Fatal("expected error for invalid port")
	}
}

func TestResolveAllOverrides(t *testing.T) {
	cfg, err := Resolve("/tmp/*.log", 8080, 5000, 2000)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Port != 8080 {
		t.Fatalf("expected port 8080, got %d", cfg.Port)
	}
	if cfg.IdleTimeout != 5*time.Second {
		t.Fatalf("expected 5s idle, got %v", cfg.IdleTimeout)
	}
	if cfg.WatchInterval != 2*time.Second {
		t.Fatalf("expected 2s watch, got %v", cfg.WatchInterval)
	}
}

func TestLoadWithFlags(t *testing.T) {
	origArgs := os.Args
	origCommandLine := flag.CommandLine
	defer func() {
		os.Args = origArgs
		flag.CommandLine = origCommandLine
	}()

	os.Args = []string{"test", "--glob-path", "/tmp/gc*.log", "--port", "9999"}
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
	if cfg.GlobPath != "/tmp/gc*.log" {
		t.Fatalf("expected /tmp/gc*.log, got %q", cfg.GlobPath)
	}
	if cfg.Port != 9999 {
		t.Fatalf("expected 9999, got %d", cfg.Port)
	}
}

func TestLoadMissingGlobPath(t *testing.T) {
	origArgs := os.Args
	origCommandLine := flag.CommandLine
	defer func() {
		os.Args = origArgs
		flag.CommandLine = origCommandLine
	}()

	os.Args = []string{"test"}
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for missing glob-path")
	}
}

func TestLoadWithAllOverrides(t *testing.T) {
	origArgs := os.Args
	origCommandLine := flag.CommandLine
	defer func() {
		os.Args = origArgs
		flag.CommandLine = origCommandLine
	}()

	os.Args = []string{"test",
		"--glob-path", "/tmp/*.log",
		"--port", "8080",
		"--idle-timeout", "5000",
		"--watch-interval", "2000",
	}
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
	if cfg.Port != 8080 {
		t.Fatalf("expected 8080, got %d", cfg.Port)
	}
}
