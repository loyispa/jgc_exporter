package parser

import (
	"strings"
	"testing"
	"time"

	"github.com/loyispa/jgc_exporter/internal/health"
)

func TestMetricsHandlerNilEvent(t *testing.T) {
	mon := health.NewMonitor()
	h := &metricsHandler{path: "/test.log", hostname: "testhost", monitor: mon}
	h.Handle(nil)
}

func TestMetricsHandlerBasicEvent(t *testing.T) {
	mon := health.NewMonitor()
	h := &metricsHandler{path: "/test.log", hostname: "testhost", monitor: mon}

	ev := &GCEvent{
		Timestamp:     time.Now(),
		GCType:        GCTypeG1,
		Category:      "G1YoungGC",
		Duration:      0.025,
		IsPause:       true,
		HeapBeforeKB:  500000,
		HeapAfterKB:   200000,
		HeapTotalKB:   1048576,
		YoungBeforeKB: 300000,
		YoungAfterKB:  10000,
		YoungTotalKB:  350000,
		OldBeforeKB:   100000,
		OldAfterKB:    90000,
		OldTotalKB:    700000,
		MetaBeforeKB:  5000,
		MetaAfterKB:   5000,
		MetaTotalKB:   10000,
		Cause:         "Allocation Failure",
	}
	h.Handle(ev)

	d := mon.GetDashboard()
	if len(d.Files) != 1 || d.Files[0].GCRate.Last1m != 1 {
		t.Fatalf("expected 1 event in monitor, got %d files, GCRate.Last1m=%f", len(d.Files), d.Files[0].GCRate.Last1m)
	}
}

func TestMetricsHandlerFullGC(t *testing.T) {
	mon := health.NewMonitor()
	h := &metricsHandler{path: "/test.log", hostname: "testhost", monitor: mon}

	ev := &GCEvent{
		Timestamp:    time.Now(),
		GCType:       GCTypeG1,
		Category:     "G1FullGC",
		Duration:     0.5,
		IsPause:      true,
		HeapBeforeKB: 500000,
		HeapAfterKB:  100000,
		HeapTotalKB:  1048576,
	}
	h.Handle(ev)

	d := mon.GetDashboard()
	if len(d.Files) != 1 || d.Files[0].FullGCRate.Last1m != 1 {
		t.Fatalf("expected 1 full gc, got %d files, FullGCRate.Last1m=%f", len(d.Files), d.Files[0].FullGCRate.Last1m)
	}
}

func TestMetricsHandlerSystemGC(t *testing.T) {
	mon := health.NewMonitor()
	h := &metricsHandler{path: "/test.log", hostname: "testhost", monitor: mon}

	ev := &GCEvent{
		Timestamp:    time.Now(),
		GCType:       GCTypeG1,
		Category:     "G1SystemGC",
		Duration:     0.3,
		IsPause:      true,
		HeapBeforeKB: 200000,
		HeapAfterKB:  50000,
		HeapTotalKB:  512000,
	}
	h.Handle(ev)

	d := mon.GetDashboard()
	if len(d.Files) != 1 || d.Files[0].FullGCRate.Last1m != 1 {
		t.Fatalf("SystemGC should be counted as full gc, got FullGCRate.Last1m=%f", d.Files[0].FullGCRate.Last1m)
	}
}

func TestMetricsHandlerZGC(t *testing.T) {
	mon := health.NewMonitor()
	h := &metricsHandler{path: "/test.log", hostname: "testhost", monitor: mon}

	ev := &GCEvent{
		Timestamp:            time.Now(),
		GCType:               GCTypeZGC,
		Category:             "ZGCAllocRate",
		Duration:             0.05,
		IsPause:              true,
		HeapBeforeKB:         500000,
		HeapAfterKB:          200000,
		HeapTotalKB:          1048576,
		ZGCPauseMarkStartMs:  1.5,
		ZGCConcurrentMarkMs:  10.0,
		ZGCPauseMarkEndMs:    0.5,
		ZGCPauseRelocateMs:   2.0,
		ZGCConcurrentRelocMs: 8.0,
		ZGCLoad1m:            0.5,
		ZGCLoad5m:            0.4,
		ZGCLoad15m:           0.3,
		ZGCMMU2ms:            0.95,
		ZGCMMU5ms:            0.96,
		ZGCMMU10ms:           0.97,
		ZGCMMU20ms:           0.98,
		ZGCMMU50ms:           0.99,
		ZGCMMU100ms:          0.995,
		ZGCMetaspaceUsedKB:   5000,
		ZGCMetaspaceCommitKB: 10000,
		Cause:                "Allocation Rate",
	}
	h.Handle(ev)

	d := mon.GetDashboard()
	if len(d.Files) != 1 || d.Files[0].GCRate.Last1m != 1 {
		t.Fatalf("expected 1 event, got GCRate.Last1m=%f", func() float64 {
			if len(d.Files) > 0 {
				return d.Files[0].GCRate.Last1m
			}
			return 0
		}())
	}
}

func TestMetricsHandlerNonPause(t *testing.T) {
	mon := health.NewMonitor()
	h := &metricsHandler{path: "/test.log", hostname: "testhost", monitor: mon}

	ev := &GCEvent{
		Timestamp: time.Now(),
		GCType:    GCTypeG1,
		Category:  "G1ConcurrentMark",
		Duration:  0.1,
		IsPause:   false,
	}
	h.Handle(ev)

	d := mon.GetDashboard()
	if len(d.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(d.Files))
	}
	if d.Files[0].GCRate.Last1m != 1 {
		t.Fatalf("expected 1 event, got GCRate.Last1m=%f", d.Files[0].GCRate.Last1m)
	}
}

func TestMetricsHandlerZeroHeap(t *testing.T) {
	mon := health.NewMonitor()
	h := &metricsHandler{path: "/test.log", hostname: "testhost", monitor: mon}

	ev := &GCEvent{
		Timestamp:     time.Now(),
		GCType:        GCTypeG1,
		Category:      "G1YoungGC",
		Duration:      0.01,
		IsPause:       true,
		HeapBeforeKB:  0,
		HeapAfterKB:   0,
		HeapTotalKB:   1048576,
		YoungBeforeKB: 0,
		YoungAfterKB:  0,
		YoungTotalKB:  350000,
	}
	h.Handle(ev)
}

func TestPipelineOnOpenClose(t *testing.T) {
	mon := health.NewMonitor()
	pl := NewPipeline(mon)

	path := "../../testdata/parser/jdk11-g1.log"
	if !pl.Accept(path) {
		t.Fatal("expected Accept to accept GC log")
	}
	if !pl.OnOpen(path) {
		t.Fatal("expected OnOpen to register parser")
	}

	pl.mu.RLock()
	_, ok := pl.parsers[path]
	pl.mu.RUnlock()
	if !ok {
		t.Fatal("expected parser registered")
	}

	pl.OnClose(path)
	pl.mu.RLock()
	_, ok = pl.parsers[path]
	pl.mu.RUnlock()
	if ok {
		t.Fatal("expected parser removed")
	}
}

func TestPipelineAcceptRejectsNonGC(t *testing.T) {
	mon := health.NewMonitor()
	pl := NewPipeline(mon)
	if pl.Accept("/nonexistent/file.log") {
		t.Fatal("expected Accept to reject nonexistent file")
	}
	pl.mu.RLock()
	n := len(pl.parsers)
	pl.mu.RUnlock()
	if n != 0 {
		t.Fatalf("expected no parsers, got %d", n)
	}
}

func TestPipelineOnRead(t *testing.T) {
	mon := health.NewMonitor()
	pl := NewPipeline(mon)

	path := "../../testdata/parser/jdk11-g1.log"
	pl.Accept(path)
	pl.OnOpen(path)
	pl.OnRead(path, "[2024-01-03T15:00:01.000+0800][gc          ] GC(0) Pause Young (Normal) (G1 Evacuation Pause) 50M->20M(512M) 10.123ms")
	pl.OnRead("/nonexistent", "should be ignored")
	pl.OnClose(path)
}

func TestPipelineSafepointPreUnified(t *testing.T) {
	mon := health.NewMonitor()
	pl := NewPipeline(mon)

	path := "../../testdata/parser/safepoint-preunified.log"
	pl.Accept(path)
	pl.OnOpen(path)

	lines := []string{
		`2024-03-15T10:00:00.100+0800: 0.100: [GC pause (G1 Evacuation Pause) (young), 0.0089230 secs]`,
		`Total time for which application threads were stopped: 0.0123456 seconds, Stopping threads took: 0.0001234 seconds`,
		`Total time for which application threads were stopped: 0.0067890 seconds, Stopping threads took: 0.0000567 seconds`,
	}
	for _, line := range lines {
		pl.OnRead(path, line)
	}

	// Verify safepoint metrics were recorded by checking that no panic occurred
	// (metrics are observable via /metrics endpoint in integration tests)
	pl.OnClose(path)
}

func TestPipelineSafepointUnified(t *testing.T) {
	mon := health.NewMonitor()
	pl := NewPipeline(mon)

	path := "../../testdata/parser/safepoint-unified.log"
	pl.Accept(path)
	pl.OnOpen(path)

	lines := []string{
		`[2024-03-15T10:00:00.009+0800][info][gc] Using G1`,
		`[2024-03-15T10:00:00.100+0800][info][gc,start     ] GC(0) Pause Young (Normal) (G1 Evacuation Pause)`,
		`[2024-03-15T10:00:00.108+0800][info][gc           ] GC(0) Pause Young (Normal) (G1 Evacuation Pause) 12M->1M(30M) 8.923ms`,
		`[2024-03-15T10:00:00.109+0800][info][safepoint    ] Safepoint "G1CollectForAllocation", Time since last: 123456789 ns, Reaching safepoint: 12345 ns, Cleanup: 1234 ns, At safepoint: 8923456 ns, Total: 8937035 ns`,
	}
	for _, line := range lines {
		pl.OnRead(path, line)
	}

	pl.OnClose(path)
}

func TestPipelineGCPhaseAndWorker(t *testing.T) {
	mon := health.NewMonitor()
	pl := NewPipeline(mon)

	path := "../../testdata/parser/jdk17-parallel.log"
	pl.Accept(path)
	pl.OnOpen(path)

	lines := []string{
		`[2024-03-15T10:00:11.016+0800][info][gc,phases,start] GC(3) Marking Phase`,
		`[2024-03-15T10:00:11.028+0800][info][gc,phases     ] GC(3) Marking Phase 12.345ms`,
		`[2024-03-15T10:00:11.029+0800][info][gc,phases,start] GC(3) Summary Phase`,
		`[2024-03-15T10:00:11.029+0800][info][gc,phases     ] GC(3) Summary Phase 0.015ms`,
		`[2024-03-15T10:00:11.035+0800][info][gc,phases,start] GC(3) Compaction Phase`,
		`[2024-03-15T10:00:11.059+0800][info][gc,phases     ] GC(3) Compaction Phase 24.567ms`,
		`[2024-03-15T10:00:10.400+0800][info][gc,task      ] GC(1) Using 4 workers`,
	}
	for _, line := range lines {
		pl.OnRead(path, line)
	}

	pl.OnClose(path)
}

func TestPipelineSafepointRegexPreUnified(t *testing.T) {
	// Verify the pre-unified safepoint regex matches correctly
	m := preUnifiedSafepointRe.FindStringSubmatch(
		"Total time for which application threads were stopped: 0.0123456 seconds, Stopping threads took: 0.0001234 seconds")
	if len(m) != 3 {
		t.Fatalf("expected 3 groups, got %d: %v", len(m), m)
	}
	if m[1] != "0.0123456" {
		t.Errorf("total=%q, want 0.0123456", m[1])
	}
	if m[2] != "0.0001234" {
		t.Errorf("stop=%q, want 0.0001234", m[2])
	}
}

func TestPipelineSafepointRegexUnified(t *testing.T) {
	line := `[2024-03-15T10:00:00.109+0800][info][safepoint    ] Safepoint "G1CollectForAllocation", Time since last: 123456789 ns, Reaching safepoint: 12345 ns, Cleanup: 1234 ns, At safepoint: 8923456 ns, Total: 8937035 ns`
	m := unifiedSafepointRe.FindStringSubmatch(line)
	if len(m) != 3 {
		t.Fatalf("expected 3 groups, got %d: %v", len(m), m)
	}
	if m[1] != "8923456" {
		t.Errorf("at_safepoint=%q, want 8923456", m[1])
	}
	if m[2] != "8937035" {
		t.Errorf("total=%q, want 8937035", m[2])
	}
	rm := unifiedSafepointReachingRe.FindStringSubmatch(line)
	if len(rm) != 2 {
		t.Fatalf("expected 2 groups for reaching, got %d", len(rm))
	}
	if rm[1] != "12345" {
		t.Errorf("reaching=%q, want 12345", rm[1])
	}
}

func TestPipelineGCPhaseRegex(t *testing.T) {
	cases := []struct {
		line  string
		phase string
		durMs string
	}{
		{`[2024-03-15T10:00:11.028+0800][info][gc,phases     ] GC(3) Marking Phase 12.345ms`, "Marking Phase", "12.345"},
		{`[2024-03-15T10:00:11.029+0800][info][gc,phases     ] GC(3) Summary Phase 0.015ms`, "Summary Phase", "0.015"},
		{`[2024-03-15T10:00:11.059+0800][info][gc,phases     ] GC(3) Compaction Phase 24.567ms`, "Compaction Phase", "24.567"},
	}
	for _, c := range cases {
		m := unifiedGCPhaseRe.FindStringSubmatch(c.line)
		if len(m) != 3 {
			t.Errorf("line %q: expected 3 groups, got %d", c.line, len(m))
			continue
		}
		if strings.TrimSpace(m[1]) != c.phase {
			t.Errorf("phase=%q, want %q", strings.TrimSpace(m[1]), c.phase)
		}
		if m[2] != c.durMs {
			t.Errorf("dur=%q, want %q", m[2], c.durMs)
		}
	}

	// Phase start lines should NOT match
	startLine := `[2024-03-15T10:00:11.016+0800][info][gc,phases,start] GC(3) Marking Phase`
	if unifiedGCPhaseRe.FindStringSubmatch(startLine) != nil {
		t.Error("phase start line should not match")
	}
}

func TestPipelineGCWorkerRegex(t *testing.T) {
	m := unifiedGCWorkerRe.FindStringSubmatch(
		`[2024-03-15T10:00:10.400+0800][info][gc,task      ] GC(1) Using 4 workers`)
	if len(m) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(m))
	}
	if m[1] != "4" {
		t.Errorf("workers=%q, want 4", m[1])
	}
}

func TestPipelineVerifyLineFiltered(t *testing.T) {
	mon := health.NewMonitor()
	pl := NewPipeline(mon)

	path := "../../testdata/parser/safepoint-preunified.log"
	pl.Accept(path)
	pl.OnOpen(path)

	verifyLines := []string{
		"VerifyBeforeGC:[Verifying threads]",
		"VerifyAfterGC:[Verifying threads]",
		"[Verifying before GC]",
		"[Verifying after GC]",
	}
	for _, line := range verifyLines {
		pl.OnRead(path, line)
	}

	pl.OnClose(path)
}

func TestPipelineCMSStringSymbolTable(t *testing.T) {
	mon := health.NewMonitor()
	h := &metricsHandler{path: "/test.log", hostname: "testhost", monitor: mon}

	ev := &GCEvent{
		Timestamp:        time.Now(),
		GCType:           GCTypeCMS,
		Category:         "CMSRemark",
		Duration:         0.05,
		IsPause:          true,
		HeapBeforeKB:     100000,
		HeapAfterKB:      50000,
		HeapTotalKB:      512000,
		CMSSymbolTableMs: 1.234,
		CMSStringTableMs: 0.567,
	}
	h.Handle(ev)
}
