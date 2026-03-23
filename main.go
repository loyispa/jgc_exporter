package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/loyispa/jgc_exporter/internal/config"
	"github.com/loyispa/jgc_exporter/internal/health"
	"github.com/loyispa/jgc_exporter/internal/metric"
	"github.com/loyispa/jgc_exporter/internal/parser"
	"github.com/loyispa/jgc_exporter/internal/server"
	"github.com/loyispa/jgc_exporter/internal/tailer"
)

var version = "dev"

func main() {
	printBanner()

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	hostname, _ := os.Hostname()
	metric.StartupTimestamp.SetToCurrentTime()
	metric.VersionInfo.WithLabelValues(version, hostname).Set(1)

	slog.Info("jgc_exporter starting",
		"version", version,
		"port", cfg.Port,
		"glob_path", cfg.GlobPath,
		"idle_timeout", cfg.IdleTimeout,
		"watch_interval", cfg.WatchInterval,
	)

	monitor := health.NewMonitor()
	pipeline := parser.NewPipeline(monitor)

	mgr := tailer.NewManager(
		cfg.GlobPath,
		cfg.IdleTimeout,
		cfg.WatchInterval,
		pipeline,
	)
	mgr.Start()

	srv := server.New(cfg.Port, metric.Registry, monitor)

	// Update throughput and allocation rate gauges periodically from dashboard.
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			d := monitor.GetDashboard()
			for _, f := range d.Files {
				metric.ThroughputGauge.WithLabelValues(f.Path, hostname).Set(f.Throughput.Last1m)
				metric.AllocationRateGauge.WithLabelValues(f.Path, hostname).Set(f.AllocRate.Last1m)
			}
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		slog.Info("shutting down...")
		mgr.Stop()
		os.Exit(0)
	}()

	if err := srv.ListenAndServe(); err != nil {
		slog.Error("HTTP server failed", "error", err)
		os.Exit(1)
	}
}

func printBanner() {
	banner := `
     _ ____  ____                                _
    (_/ ___|/ ___|   _____  ___ __   ___  _ __| |_ ___ _ __
    | | |  _| |      / _ \ \/ / '_ \ / _ \| '__| __/ _ \ '__|
    | | |_| | |___  |  __/>  <| |_) | (_) | |  | ||  __/ |
   _/ |\____|\____|  \___/_/\_\ .__/ \___/|_|   \__\___|_|
  |__/                        |_|`
	fmt.Println(banner)
	fmt.Printf("%55s\n\n", version)
}
