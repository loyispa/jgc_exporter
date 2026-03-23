package metric

import "github.com/prometheus/client_golang/prometheus"

var Registry = prometheus.NewRegistry()

var (
	StartupTimestamp = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "jgc_startup_timestamp_seconds",
		Help: "Timestamp of exporter startup",
	})

	VersionInfo = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "jgc_exporter_version_info",
		Help: "Version of jgc_exporter",
	}, []string{"version", "host"})

	LogLines = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "jgc_log_lines_total",
		Help: "Total number of processed log lines",
	}, []string{"path", "host", "gc_type"})
)

func init() {
	Registry.MustRegister(StartupTimestamp)
	Registry.MustRegister(VersionInfo)
	Registry.MustRegister(LogLines)
}
