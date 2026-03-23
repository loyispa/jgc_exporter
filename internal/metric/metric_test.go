package metric

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func TestRegistryNotNil(t *testing.T) {
	if Registry == nil {
		t.Fatal("expected non-nil registry")
	}
}

func TestAllVecMetricsCount(t *testing.T) {
	all := AllVecMetrics()
	if len(all) == 0 {
		t.Fatal("expected non-empty AllVecMetrics")
	}
}

func TestMetricsRegistered(t *testing.T) {
	StartupTimestamp.Set(1)
	VersionInfo.WithLabelValues("test", "host").Set(1)
	LogLines.WithLabelValues("/test", "host", "G1").Inc()
	defer func() {
		VersionInfo.DeleteLabelValues("test", "host")
		LogLines.DeleteLabelValues("/test", "host", "G1")
	}()

	families, err := Registry.Gather()
	if err != nil {
		t.Fatalf("gather failed: %v", err)
	}
	names := make(map[string]bool)
	for _, f := range families {
		names[f.GetName()] = true
	}

	required := []string{
		"jgc_startup_timestamp_seconds",
		"jgc_exporter_version_info",
		"jgc_log_lines_total",
	}
	for _, name := range required {
		if !names[name] {
			t.Errorf("metric %q not found in registry", name)
		}
	}
}

func TestLogLinesLabels(t *testing.T) {
	LogLines.WithLabelValues("/test", "host", "G1").Inc()
	defer LogLines.DeleteLabelValues("/test", "host", "G1")
}

func TestHeapSizeSet(t *testing.T) {
	HeapSizeBeforeCollection.WithLabelValues("/test", "host").Set(1048576)
	defer HeapSizeBeforeCollection.DeletePartialMatch(prometheus.Labels{"path": "/test"})
}
