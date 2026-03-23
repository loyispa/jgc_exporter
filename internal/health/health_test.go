package health

import (
	"testing"
	"time"
)

func TestRingBuffer(t *testing.T) {
	rb := NewRingBuffer[int](3)
	if rb.Len() != 0 {
		t.Fatalf("expected empty, got %d", rb.Len())
	}
	if s := rb.Slice(); s != nil {
		t.Fatalf("expected nil slice, got %v", s)
	}

	rb.Push(1)
	rb.Push(2)
	rb.Push(3)
	if rb.Len() != 3 {
		t.Fatalf("expected 3, got %d", rb.Len())
	}
	got := rb.Slice()
	want := []int{1, 2, 3}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("index %d: got %d want %d", i, got[i], want[i])
		}
	}

	rb.Push(4)
	if rb.Len() != 3 {
		t.Fatalf("expected 3 after overflow, got %d", rb.Len())
	}
	got = rb.Slice()
	want = []int{2, 3, 4}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("index %d: got %d want %d", i, got[i], want[i])
		}
	}
}

func TestMonitorRecordEvent(t *testing.T) {
	m := NewMonitor()
	m.SetGCType("/test.log", "G1")

	now := time.Now()
	m.RecordEvent(GCEventRecord{
		Timestamp:    now,
		Path:         "/test.log",
		Category:     "G1YoungGC",
		Duration:     0.05,
		IsPause:      true,
		HeapBeforeKB: 500000,
		HeapAfterKB:  200000,
		HeapTotalKB:  1048576,
		YoungAfterKB: 10000,
		OldAfterKB:   190000,
		MetaAfterKB:  5000,
	})

	events := m.events.Slice()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Category != "G1YoungGC" {
		t.Fatalf("expected G1YoungGC, got %s", events[0].Category)
	}

	snapshots := m.heapHistory.Slice()
	if len(snapshots) != 1 {
		t.Fatalf("expected 1 heap snapshot, got %d", len(snapshots))
	}
	if snapshots[0].YoungUsedKB != 10000 {
		t.Fatalf("expected YoungUsedKB=10000, got %d", snapshots[0].YoungUsedKB)
	}
	if snapshots[0].OldUsedKB != 190000 {
		t.Fatalf("expected OldUsedKB=190000, got %d", snapshots[0].OldUsedKB)
	}
	if snapshots[0].MetaUsedKB != 5000 {
		t.Fatalf("expected MetaUsedKB=5000, got %d", snapshots[0].MetaUsedKB)
	}
}

func TestMonitorFullGCTracking(t *testing.T) {
	m := NewMonitor()
	m.SetGCType("/test.log", "G1")

	now := time.Now()
	m.RecordEvent(GCEventRecord{
		Timestamp:   now.Add(-10 * time.Second),
		Path:        "/test.log",
		Category:    "G1YoungGC",
		Duration:    0.01,
		IsPause:     true,
		HeapAfterKB: 200000,
		HeapTotalKB: 1048576,
	})
	m.RecordEvent(GCEventRecord{
		Timestamp:   now,
		Path:        "/test.log",
		Category:    "G1FullGC",
		Duration:    0.5,
		IsPause:     true,
		HeapAfterKB: 100000,
		HeapTotalKB: 1048576,
		IsFullGC:    true,
	})

	fs := m.fileSummary["/test.log"]
	if fs.totalEvents != 2 {
		t.Fatalf("expected 2 events, got %d", fs.totalEvents)
	}
	if fs.fullGCCount != 1 {
		t.Fatalf("expected 1 full gc, got %d", fs.fullGCCount)
	}
	if fs.lastFullGC == nil {
		t.Fatal("expected lastFullGC to be set")
	}
}

func TestMonitorRemoveFile(t *testing.T) {
	m := NewMonitor()
	m.SetGCType("/test.log", "G1")
	m.RecordEvent(GCEventRecord{
		Timestamp:   time.Now(),
		Path:        "/test.log",
		HeapAfterKB: 100,
		HeapTotalKB: 1000,
	})
	m.RemoveFile("/test.log")
	if _, ok := m.fileSummary["/test.log"]; ok {
		t.Fatal("expected file to be removed from fileSummary")
	}
}

func TestGetDashboardEmpty(t *testing.T) {
	m := NewMonitor()
	d := m.GetDashboard()
	if d.OverallStatus != StatusHealthy {
		t.Fatalf("expected healthy, got %s", d.OverallStatus)
	}
	if d.ThroughputRatio != 1.0 {
		t.Fatalf("expected throughput 1.0, got %f", d.ThroughputRatio)
	}
}

func TestDashboardGCDurationTimeSeriesAlignedWithGCEvents(t *testing.T) {
	m := NewMonitor()
	d := m.GetDashboard()
	if len(d.GCDurationTimeSeries) != len(d.GCEventsTimeSeries) {
		t.Fatalf("gc_duration_timeseries len %d want same as gc_events_timeseries len %d",
			len(d.GCDurationTimeSeries), len(d.GCEventsTimeSeries))
	}
	for i := range d.GCEventsTimeSeries {
		if !d.GCEventsTimeSeries[i].Timestamp.Equal(d.GCDurationTimeSeries[i].Timestamp) {
			t.Fatalf("timestamp mismatch at %d: %v vs %v", i,
				d.GCEventsTimeSeries[i].Timestamp, d.GCDurationTimeSeries[i].Timestamp)
		}
	}
}

func TestDashboardGCEventsDenseMinuteGrid(t *testing.T) {
	m := NewMonitor()
	d := m.GetDashboard()
	n := len(d.GCEventsTimeSeries)
	if n < 2 || n > 120 {
		t.Fatalf("unexpected GC events series length %d (expected ~61 minute slots)", n)
	}
	for i := 1; i < len(d.GCEventsTimeSeries); i++ {
		prev := d.GCEventsTimeSeries[i-1].Timestamp
		cur := d.GCEventsTimeSeries[i].Timestamp
		if cur.Sub(prev) != time.Minute {
			t.Fatalf("non-consecutive GC buckets: %v then %v (delta %v)", prev, cur, cur.Sub(prev))
		}
	}
}

func TestSTWDurationHistoryUsesPauseEventsOnly(t *testing.T) {
	m := NewMonitor()
	min := time.Now().UTC().Truncate(time.Minute).Add(10 * time.Second)
	m.RecordEvent(GCEventRecord{
		Timestamp: min, Path: "/a.log", Category: "G1YoungGC", Duration: 0.05, IsPause: true,
		HeapAfterKB: 100, HeapTotalKB: 1000,
	})
	m.RecordEvent(GCEventRecord{
		Timestamp: min.Add(2 * time.Second), Path: "/a.log", Category: "G1YoungGC", Duration: 0.03, IsPause: true,
		HeapAfterKB: 100, HeapTotalKB: 1000,
	})
	m.RecordEvent(GCEventRecord{
		Timestamp: min.Add(4 * time.Second), Path: "/a.log", Category: "Concurrent", Duration: 1.0, IsPause: false,
		HeapAfterKB: 100, HeapTotalKB: 1000,
	})
	d := m.GetDashboard()
	wantMin := utcMinute(min)
	var found int
	for _, b := range d.STWDurationHistory {
		if b.Path != "/a.log" || !b.Timestamp.Equal(wantMin) {
			continue
		}
		found++
		if b.Count != 2 {
			t.Fatalf("expected 2 STW pauses in bucket, got %d", b.Count)
		}
		if b.SumSec < 0.079 || b.SumSec > 0.081 {
			t.Fatalf("unexpected sum_sec %v", b.SumSec)
		}
		if b.MinSec != 0.03 || b.MaxSec != 0.05 {
			t.Fatalf("min/max got %v / %v", b.MinSec, b.MaxSec)
		}
	}
	if found != 1 {
		t.Fatalf("expected exactly one matching STW bucket for path+minute, found %d", found)
	}
}

func TestGetDashboardWithEvents(t *testing.T) {
	m := NewMonitor()
	m.SetGCType("/test.log", "G1")

	// Align with dashboard "current UTC minute" window so last-1m stats see all events.
	base := time.Now().UTC().Truncate(time.Minute).Add(2 * time.Second)
	for i := 0; i < 10; i++ {
		m.RecordEvent(GCEventRecord{
			Timestamp:   base.Add(time.Duration(i) * time.Second),
			Path:        "/test.log",
			Category:    "G1YoungGC",
			Duration:    0.01,
			IsPause:     true,
			HeapAfterKB: int64(500000 - i*10000),
			HeapTotalKB: 1048576,
		})
	}

	d := m.GetDashboard()
	if len(d.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(d.Files))
	}
	if d.Files[0].GCRate.Last1m < 9 {
		t.Fatalf("expected >=9 events in last 1min, got %f", d.Files[0].GCRate.Last1m)
	}
	if d.Files[0].GCType != "G1" {
		t.Fatalf("expected G1, got %s", d.Files[0].GCType)
	}
	if d.ThroughputRatio <= 0 || d.ThroughputRatio > 1.0 {
		t.Fatalf("unexpected throughput: %f", d.ThroughputRatio)
	}
}

func TestEvaluateFileSummaryHealthHealthy(t *testing.T) {
	s := evaluateFileSummaryHealth(FileSummary{
		HeapUsedKB:     300000,
		HeapTotalKB:    1048576,
		HeapUsageRatio: 300000.0 / 1048576.0,
	})
	if s != StatusHealthy {
		t.Fatalf("expected healthy, got %s", s)
	}
}

func TestEvaluateFileSummaryHealthHeapWarning(t *testing.T) {
	s := evaluateFileSummaryHealth(FileSummary{
		HeapUsedKB:     860000,
		HeapTotalKB:    1048576,
		HeapUsageRatio: 860000.0 / 1048576.0,
	})
	if s != StatusWarning {
		t.Fatalf("expected warning for >80%% heap, got %s", s)
	}
}

func TestEvaluateFileSummaryHealthHeapCritical(t *testing.T) {
	s := evaluateFileSummaryHealth(FileSummary{
		HeapUsedKB:     960000,
		HeapTotalKB:    1048576,
		HeapUsageRatio: 960000.0 / 1048576.0,
	})
	if s != StatusCritical {
		t.Fatalf("expected critical for >90%% heap, got %s", s)
	}
}

func TestEvaluateFileSummaryHealthPauseCritical(t *testing.T) {
	s := evaluateFileSummaryHealth(FileSummary{
		P99Pause: MultiWindowStat{Last1m: 0.6},
	})
	if s != StatusCritical {
		t.Fatalf("expected critical for p99 >500ms, got %s", s)
	}
}

func TestEvaluateFileSummaryHealthPauseWarning(t *testing.T) {
	s := evaluateFileSummaryHealth(FileSummary{
		P99Pause: MultiWindowStat{Last1m: 0.25},
	})
	if s != StatusWarning {
		t.Fatalf("expected warning for p99 >200ms, got %s", s)
	}
}

func TestEvaluateFileSummaryHealthRecentFullGC(t *testing.T) {
	s := evaluateFileSummaryHealth(FileSummary{
		FullGCRate: MultiWindowStat{Last1m: 2},
	})
	if s != StatusCritical {
		t.Fatalf("expected critical for >=2 full GCs in window, got %s", s)
	}
}

func TestEvaluateOverallHealth(t *testing.T) {
	s, alerts := evaluateOverallHealth(nil, 0, 0)
	if s != StatusHealthy {
		t.Fatalf("expected healthy, got %s", s)
	}
	if len(alerts) != 1 || alerts[0] != "All systems healthy" {
		t.Fatalf("unexpected alerts: %v", alerts)
	}

	s, alerts = evaluateOverallHealth(nil, 1, 0)
	if s != StatusWarning {
		t.Fatalf("expected warning for 1 full gc, got %s", s)
	}

	s, _ = evaluateOverallHealth(nil, 3, 0)
	if s != StatusCritical {
		t.Fatalf("expected critical for >=2 full gcs, got %s", s)
	}

	s, _ = evaluateOverallHealth(nil, 0, 0.6)
	if s != StatusCritical {
		t.Fatalf("expected critical for p99 >500ms, got %s", s)
	}
}

func TestPercentile(t *testing.T) {
	data := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	p99 := percentile(data, 0.99)
	// floor(9 * 0.99) = 8, so index 8 = value 9 (sorted)
	if p99 != 9 {
		t.Fatalf("expected p99=9, got %f", p99)
	}
	p50 := percentile(data, 0.5)
	// floor(9 * 0.5) = 4, so index 4 = value 5
	if p50 != 5 {
		t.Fatalf("expected p50=5, got %f", p50)
	}
	p0 := percentile(nil, 0.5)
	if p0 != 0 {
		t.Fatalf("expected 0 for empty, got %f", p0)
	}
}

func TestAvgFloat(t *testing.T) {
	if avgFloat(nil) != 0 {
		t.Fatal("expected 0 for nil")
	}
	if avgFloat([]float64{2, 4}) != 3 {
		t.Fatalf("expected 3, got %f", avgFloat([]float64{2, 4}))
	}
}

func TestMaxFloat(t *testing.T) {
	if maxFloat(nil) != 0 {
		t.Fatal("expected 0 for nil")
	}
	if maxFloat([]float64{1, 5, 3}) != 5 {
		t.Fatal("expected 5")
	}
}

func TestPauseDurationsTrimming(t *testing.T) {
	m := NewMonitor()
	m.SetGCType("/test.log", "G1")
	for i := 0; i < 10005; i++ {
		m.RecordEvent(GCEventRecord{
			Timestamp:   time.Now(),
			Path:        "/test.log",
			Duration:    0.01,
			IsPause:     true,
			HeapAfterKB: 100,
			HeapTotalKB: 1000,
		})
	}
	fs := m.fileSummary["/test.log"]
	if len(fs.pauseDurations) > 5500 {
		t.Fatalf("expected pause durations trimmed, got %d", len(fs.pauseDurations))
	}
}
