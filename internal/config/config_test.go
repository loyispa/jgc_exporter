package config

import (
	"flag"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestParsePort(t *testing.T) {
	cases := []struct {
		input string
		want  int
	}{
		{":5898", 5898},
		{"0.0.0.0:8080", 8080},
		{"localhost:443", 443},
		{"noport", 0},
		{":abc", 0},
		{"", 0},
	}
	for _, c := range cases {
		got := parsePort(c.input)
		if got != c.want {
			t.Errorf("parsePort(%q) = %d, want %d", c.input, got, c.want)
		}
	}
}

func TestLoadYAML(t *testing.T) {
	dir := t.TempDir()
	yamlContent := `
fileGlobPattern: "/var/log/gc*.log"
hostPort: ":9090"
idleTimeout: 30000
watchInterval: 5000
`
	yamlPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := DefaultConfig
	if err := loadYAML(yamlPath, &cfg); err != nil {
		t.Fatalf("loadYAML failed: %v", err)
	}

	if cfg.GlobPath != "/var/log/gc*.log" {
		t.Fatalf("expected glob path '/var/log/gc*.log', got %q", cfg.GlobPath)
	}
	if cfg.Port != 9090 {
		t.Fatalf("expected port 9090, got %d", cfg.Port)
	}
	if cfg.IdleTimeout != 30*time.Second {
		t.Fatalf("expected idle timeout 30s, got %v", cfg.IdleTimeout)
	}
	if cfg.WatchInterval != 5*time.Second {
		t.Fatalf("expected watch interval 5s, got %v", cfg.WatchInterval)
	}
}

func TestLoadYAMLMultiPath(t *testing.T) {
	dir := t.TempDir()
	yamlContent := `
fileGlobPattern: "/tmp/a/*.log,/tmp/b/*.log"
hostPort: ":9090"
`
	yamlPath := filepath.Join(dir, "multipath.yaml")
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := DefaultConfig
	if err := loadYAML(yamlPath, &cfg); err != nil {
		t.Fatalf("loadYAML failed: %v", err)
	}

	want := "/tmp/a/*.log,/tmp/b/*.log"
	if cfg.GlobPath != want {
		t.Fatalf("expected glob path %q, got %q", want, cfg.GlobPath)
	}
}

func TestLoadYAMLPartial(t *testing.T) {
	dir := t.TempDir()
	yamlContent := `
fileGlobPattern: "/tmp/gc.log"
`
	yamlPath := filepath.Join(dir, "partial.yaml")
	os.WriteFile(yamlPath, []byte(yamlContent), 0644)

	cfg := DefaultConfig
	if err := loadYAML(yamlPath, &cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.GlobPath != "/tmp/gc.log" {
		t.Fatalf("expected /tmp/gc.log, got %q", cfg.GlobPath)
	}
	if cfg.Port != DefaultConfig.Port {
		t.Fatalf("expected default port %d, got %d", DefaultConfig.Port, cfg.Port)
	}
}

func TestLoadYAMLInvalidPath(t *testing.T) {
	cfg := DefaultConfig
	err := loadYAML("/nonexistent/config.yaml", &cfg)
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestLoadYAMLInvalidContent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yaml")
	os.WriteFile(path, []byte(":::invalid:::yaml"), 0644)

	cfg := DefaultConfig
	err := loadYAML(path, &cfg)
	if err == nil {
		t.Fatal("expected error for invalid yaml")
	}
}

func TestDefaultConfig(t *testing.T) {
	if DefaultConfig.Port != 5898 {
		t.Fatalf("expected default port 5898, got %d", DefaultConfig.Port)
	}
}

func TestResolveWithGlobPath(t *testing.T) {
	cfg, err := Resolve("", "/var/log/gc*.log", 0, 0, 0)
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
	_, err := Resolve("", "", 0, 0, 0)
	if err == nil {
		t.Fatal("expected error when glob-path is missing")
	}
}

func TestResolveInvalidPort(t *testing.T) {
	_, err := Resolve("", "/tmp/*.log", 99999, 0, 0)
	if err == nil {
		t.Fatal("expected error for invalid port")
	}
}

func TestResolveAllOverrides(t *testing.T) {
	cfg, err := Resolve("", "/tmp/*.log", 8080, 5000, 2000)
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

func TestResolveWithYAML(t *testing.T) {
	dir := t.TempDir()
	yamlContent := `
fileGlobPattern: "/from/yaml/*.log"
hostPort: ":7070"
`
	yamlPath := filepath.Join(dir, "config.yaml")
	os.WriteFile(yamlPath, []byte(yamlContent), 0644)

	cfg, err := Resolve(yamlPath, "", 0, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.GlobPath != "/from/yaml/*.log" {
		t.Fatalf("expected yaml glob, got %q", cfg.GlobPath)
	}
	if cfg.Port != 7070 {
		t.Fatalf("expected port 7070 from yaml, got %d", cfg.Port)
	}
}

func TestResolveCLIOverridesYAML(t *testing.T) {
	dir := t.TempDir()
	yamlContent := `
fileGlobPattern: "/from/yaml/*.log"
hostPort: ":7070"
`
	yamlPath := filepath.Join(dir, "config.yaml")
	os.WriteFile(yamlPath, []byte(yamlContent), 0644)

	cfg, err := Resolve(yamlPath, "/from/cli/*.log", 9090, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.GlobPath != "/from/cli/*.log" {
		t.Fatalf("expected CLI glob override, got %q", cfg.GlobPath)
	}
	if cfg.Port != 9090 {
		t.Fatalf("expected CLI port override 9090, got %d", cfg.Port)
	}
}

func TestResolveInvalidYAMLPath(t *testing.T) {
	_, err := Resolve("/nonexistent.yaml", "", 0, 0, 0)
	if err == nil {
		t.Fatal("expected error for nonexistent yaml")
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

func TestLoadWithConfig(t *testing.T) {
	dir := t.TempDir()
	yamlContent := `fileGlobPattern: "/var/log/*.log"`
	yamlPath := filepath.Join(dir, "config.yaml")
	os.WriteFile(yamlPath, []byte(yamlContent), 0644)

	origArgs := os.Args
	origCommandLine := flag.CommandLine
	defer func() {
		os.Args = origArgs
		flag.CommandLine = origCommandLine
	}()

	os.Args = []string{"test", "--config", yamlPath}
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
	if cfg.GlobPath != "/var/log/*.log" {
		t.Fatalf("expected /var/log/*.log, got %q", cfg.GlobPath)
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
