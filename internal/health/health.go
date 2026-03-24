package health

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

type Status string

const (
	StatusHealthy  Status = "healthy"
	StatusWarning  Status = "warning"
	StatusCritical Status = "critical"
)

type GCEventRecord struct {
	Timestamp     time.Time `json:"timestamp"`
	Path          string    `json:"path"`
	Category      string    `json:"category"`
	Duration      float64   `json:"duration"`
	IsPause       bool      `json:"is_pause"`
	HeapBeforeKB  int64     `json:"heap_before_kb"`
	HeapAfterKB   int64     `json:"heap_after_kb"`
	HeapTotalKB   int64     `json:"heap_total_kb"`
	YoungBeforeKB int64     `json:"young_before_kb,omitempty"`
	YoungAfterKB  int64     `json:"young_after_kb,omitempty"`
	OldBeforeKB   int64     `json:"old_before_kb,omitempty"`
	OldAfterKB    int64     `json:"old_after_kb,omitempty"`
	MetaBeforeKB  int64     `json:"meta_before_kb,omitempty"`
	MetaAfterKB   int64     `json:"meta_after_kb,omitempty"`
	Cause         string    `json:"cause,omitempty"`
	IsFullGC      bool      `json:"is_full_gc"`
}

// HeapSnapshot drives the dashboard heap chart: Java heap used/total and Metaspace/PermGen used (when logged).
type HeapSnapshot struct {
	Timestamp   time.Time `json:"timestamp"`
	Path        string    `json:"path"`
	HeapUsedKB  int64     `json:"heap_used_kb"`
	HeapTotalKB int64     `json:"heap_total_kb"`
	MetaUsedKB  int64     `json:"meta_used_kb,omitempty"`
}

type RawLogEntry struct {
	Path      string    `json:"path"`
	Line      string    `json:"line"`
	Timestamp time.Time `json:"timestamp"`
}

type Recommendation struct {
	Condition string `json:"condition"`
	Advice    string `json:"advice"`
}

// MultiWindowStat holds values for three rolling UTC windows ending at the current minute:
// Last1m = current minute only; Last5m / Last1h = aggregates over the last 5 / 60 minutes.
// Semantics per field differ (e.g. PauseMax uses max in-window; GCRate divides event count by 5 or 60).
type MultiWindowStat struct {
	Last1m float64 `json:"last_1m"`
	Last5m float64 `json:"last_5m"`
	Last1h float64 `json:"last_1h"`
}

type FileSummary struct {
	Path                     string         `json:"path"`
	GCType                   string         `json:"gc_type"`
	LastFullGCTime           *time.Time     `json:"last_full_gc_time,omitempty"`
	HeapUsedKB               int64          `json:"heap_used_kb"`
	HeapTotalKB              int64          `json:"heap_total_kb"`
	HeapUsageRatio           float64        `json:"heap_usage_ratio"`
	PromotionFailedCount     int            `json:"promotion_failed_count"`
	ConcurrentModeFailCount  int            `json:"concurrent_mode_failure_count"`
	ToSpaceExhaustedCount    int            `json:"to_space_exhausted_count"`
	HumongousAllocationCount int            `json:"humongous_allocation_count"`
	CategoryCounts           map[string]int `json:"category_counts"`
	Status                   Status         `json:"status"`

	// Multi-window stats (last-1m / last-5m / last-1h windows; see MultiWindowStat)
	Throughput MultiWindowStat `json:"throughput"`
	// PauseMax holds max STW pause duration (seconds) per window; only IsPause events with duration>=0.
	PauseMax MultiWindowStat `json:"pause_max"`
	GCRate   MultiWindowStat `json:"gc_rate"`
	FullGCRate MultiWindowStat `json:"full_gc_rate"`
	AllocRate  MultiWindowStat `json:"alloc_rate"`
	HeapUsage  MultiWindowStat `json:"heap_usage"`
}

// ThroughputPoint is application throughput ratio (0..1) for one path at a UTC minute: 1 - (STW pause sum / second span).
type ThroughputPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Path      string    `json:"path"`
	Ratio     float64   `json:"ratio"`
}

// GCEventsBucket is a time bucket with per-category event counts (for last 1h).
type GCEventsBucket struct {
	Timestamp time.Time      `json:"timestamp"`
	Counts    map[string]int `json:"counts"`
}

// GCDurationBucket is a time bucket with per-category summed GC pause duration in seconds (UTC minute).
type GCDurationBucket struct {
	Timestamp time.Time          `json:"timestamp"`
	Seconds   map[string]float64 `json:"seconds"`
}

// AllocationRatePoint is the allocation rate (MB/s) at a given time for a path.
type AllocationRatePoint struct {
	Timestamp    time.Time `json:"timestamp"`
	Path         string    `json:"path"`
	RateMBPerSec float64   `json:"rate_mb_per_sec"`
}

type Dashboard struct {
	OverallStatus         Status                `json:"overall_status"`
	Alerts                []string              `json:"alerts"`
	Files                 []FileSummary         `json:"files"`
	RecentFullGCs         []GCEventRecord       `json:"recent_full_gcs"`
	// RecentFullGCs5m is Full GC events in the last 5 UTC minutes (tuning advice; less noisy than recent_full_gcs).
	RecentFullGCs5m       []GCEventRecord       `json:"recent_full_gcs_5m,omitempty"`
	HeapHistory           []HeapSnapshot        `json:"heap_history"`
	CategoryCounts        map[string]int        `json:"category_counts"`
	Recommendations       []Recommendation      `json:"recommendations"`
	ThroughputRatio       float64               `json:"throughput_ratio"`
	MaxSTWPause1mSec      float64               `json:"max_stw_pause_1m_sec"`
	GCEventsTimeSeries    []GCEventsBucket      `json:"gc_events_timeseries"`
	GCDurationTimeSeries  []GCDurationBucket    `json:"gc_duration_timeseries"`
	ThroughputHistory     []ThroughputPoint     `json:"throughput_history"`
	AllocationRateHistory []AllocationRatePoint `json:"allocation_rate_history"`
}

type Monitor struct {
	mu          sync.RWMutex
	events      *RingBuffer[GCEventRecord]
	heapHistory *RingBuffer[HeapSnapshot]
	rawLogLines *RingBuffer[RawLogEntry]
	fileSummary map[string]*fileStat
}

const (
	maxEvents      = 10000
	maxHeapHistory = 2000
	maxFullGCs     = 100
	maxRawLines    = 500
)

func NewMonitor() *Monitor {
	return &Monitor{
		events:      NewRingBuffer[GCEventRecord](maxEvents),
		heapHistory: NewRingBuffer[HeapSnapshot](maxHeapHistory),
		rawLogLines: NewRingBuffer[RawLogEntry](maxRawLines),
		fileSummary: make(map[string]*fileStat),
	}
}

type fileStat struct {
	gcType                   string
	totalEvents              int
	fullGCCount              int
	lastFullGC               *time.Time
	heapUsedKB               int64
	heapTotalKB              int64
	pauseDurations           []float64
	promotionFailedCount     int
	concurrentModeFailCount  int
	toSpaceExhaustedCount    int
	humongousAllocationCount int
	firstEventTime           *time.Time
	lastEventTime            *time.Time
	lastFullGCCause          string
	categoryCounts           map[string]int
	totalPauseTime           float64
}

func (m *Monitor) RecordEvent(rec GCEventRecord) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.events.Push(rec)

	fs, ok := m.fileSummary[rec.Path]
	if !ok {
		fs = &fileStat{categoryCounts: make(map[string]int)}
		m.fileSummary[rec.Path] = fs
	}

	fs.totalEvents++
	t := rec.Timestamp
	if fs.firstEventTime == nil {
		fs.firstEventTime = &t
	}
	fs.lastEventTime = &t

	if rec.IsFullGC {
		fs.fullGCCount++
		fs.lastFullGC = &t
		fs.lastFullGCCause = rec.Cause
	}
	if rec.HeapAfterKB > 0 {
		fs.heapUsedKB = rec.HeapAfterKB
	}
	if rec.HeapTotalKB > 0 {
		fs.heapTotalKB = rec.HeapTotalKB
	}
	if rec.IsPause && rec.Duration > 0 {
		fs.pauseDurations = append(fs.pauseDurations, rec.Duration)
		fs.totalPauseTime += rec.Duration
		if len(fs.pauseDurations) > 10000 {
			fs.pauseDurations = fs.pauseDurations[len(fs.pauseDurations)-5000:]
		}
	}

	cat := rec.Category
	if cat != "" {
		fs.categoryCounts[cat]++
	}
	if strings.Contains(cat, "PromotionFailed") {
		fs.promotionFailedCount++
	}
	if strings.Contains(cat, "ConcurrentModeFailure") {
		fs.concurrentModeFailCount++
	}
	if strings.Contains(cat, "ToSpaceExhausted") {
		fs.toSpaceExhaustedCount++
	}
	if strings.Contains(cat, "HumongousAllocation") {
		fs.humongousAllocationCount++
	}

	if rec.HeapAfterKB > 0 || rec.HeapTotalKB > 0 {
		m.heapHistory.Push(HeapSnapshot{
			Timestamp:   rec.Timestamp,
			Path:        rec.Path,
			HeapUsedKB:  rec.HeapAfterKB,
			HeapTotalKB: rec.HeapTotalKB,
			MetaUsedKB:  rec.MetaAfterKB,
		})
	}
}

func (m *Monitor) RecordRawLine(path, line string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rawLogLines.Push(RawLogEntry{
		Path:      path,
		Line:      line,
		Timestamp: time.Now(),
	})
}

func (m *Monitor) SetGCType(path, gcType string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	fs, ok := m.fileSummary[path]
	if !ok {
		fs = &fileStat{categoryCounts: make(map[string]int)}
		m.fileSummary[path] = fs
	}
	fs.gcType = gcType
}

// GetGCType returns the GC type for the given path, or empty string if unknown.
func (m *Monitor) GetGCType(path string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if fs, ok := m.fileSummary[path]; ok {
		return fs.gcType
	}
	return ""
}

func (m *Monitor) RemoveFile(path string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.fileSummary, path)
}

const (
	window1m = 1 * time.Minute
	window5m = 5 * time.Minute
	window1h = 1 * time.Hour
)

func (m *Monitor) GetDashboard() Dashboard {
	m.mu.RLock()
	defer m.mu.RUnlock()

	d := Dashboard{}
	now := time.Now()
	// All dashboard windows use UTC wall-clock minutes to avoid boundary drift / double-counting.
	refMin := utcMinute(now)
	chartEndMin := refMin
	chartStartMin := refMin.Add(-59 * time.Minute) // inclusive 60 UTC minutes
	fiveMinStartMin := refMin.Add(-4 * time.Minute)

	type pathWindowStats struct {
		events1m       int
		events5m       int
		events1h       int
		fullGC1m       int
		fullGC5m       int
		fullGC1h       int
		pause1m        []float64
		pause5m        []float64
		pause1h        []float64
		totalPause1m   float64
		totalPause5m   float64
		totalPause1h   float64
		categoryCounts map[string]int
	}
	pathWindows := make(map[string]*pathWindowStats)

	var recentFullGCs []GCEventRecord
	var recentFullGCs5m []GCEventRecord
	var allPauses1m []float64
	totalPauseTime1m := 0.0
	recentFullGCCount := 0

	categoryCounts1h := make(map[string]int)
	pathMinPauseSum := make(map[pathMinuteKey]float64)

	events := m.events.Slice()
	var events1h []GCEventRecord
	for _, ev := range events {
		em := utcMinute(ev.Timestamp)
		if em.Before(chartStartMin) || em.After(chartEndMin) {
			continue
		}
		events1h = append(events1h, ev)

		if ev.IsPause && ev.Duration >= 0 && ev.Path != "" {
			pathMinPauseSum[pathMinuteKey{ev.Path, em}] += ev.Duration
		}
		if ev.Category != "" {
			categoryCounts1h[ev.Category]++
		}

		pw, ok := pathWindows[ev.Path]
		if !ok {
			pw = &pathWindowStats{categoryCounts: make(map[string]int)}
			pathWindows[ev.Path] = pw
		}
		if ev.Category != "" {
			pw.categoryCounts[ev.Category]++
		}

		if em.Equal(refMin) {
			pw.events1m++
			if ev.IsPause && ev.Duration >= 0 {
				pw.pause1m = append(pw.pause1m, ev.Duration)
				pw.totalPause1m += ev.Duration
				allPauses1m = append(allPauses1m, ev.Duration)
				totalPauseTime1m += ev.Duration
			}
			if ev.IsFullGC {
				pw.fullGC1m++
				recentFullGCs = append(recentFullGCs, ev)
				recentFullGCCount++
			}
		}
		if !em.Before(fiveMinStartMin) && !em.After(refMin) {
			pw.events5m++
			if ev.IsPause && ev.Duration >= 0 {
				pw.pause5m = append(pw.pause5m, ev.Duration)
				pw.totalPause5m += ev.Duration
			}
			if ev.IsFullGC {
				pw.fullGC5m++
				recentFullGCs5m = append(recentFullGCs5m, ev)
			}
		}
		pw.events1h++
		if ev.IsPause && ev.Duration >= 0 {
			pw.pause1h = append(pw.pause1h, ev.Duration)
			pw.totalPause1h += ev.Duration
		}
		if ev.IsFullGC {
			pw.fullGC1h++
		}
	}

	if len(recentFullGCs) > maxFullGCs {
		recentFullGCs = recentFullGCs[len(recentFullGCs)-maxFullGCs:]
	}
	if len(recentFullGCs5m) > maxFullGCs {
		recentFullGCs5m = recentFullGCs5m[len(recentFullGCs5m)-maxFullGCs:]
	}
	d.RecentFullGCs = recentFullGCs
	d.RecentFullGCs5m = recentFullGCs5m
	d.CategoryCounts = categoryCounts1h

	curMinElapsed := now.Sub(refMin).Seconds()
	if curMinElapsed < 1 {
		curMinElapsed = 1
	}
	if totalPauseTime1m > 0 {
		d.ThroughputRatio = 1.0 - (totalPauseTime1m / curMinElapsed)
		if d.ThroughputRatio < 0 {
			d.ThroughputRatio = 0
		}
	} else {
		d.ThroughputRatio = 1.0
	}
	if len(allPauses1m) > 0 {
		d.MaxSTWPause1mSec = maxFloat(allPauses1m)
	}

	heapSlice := m.heapHistory.Slice()
	var sparseHeap []HeapSnapshot
	for _, h := range heapSlice {
		hm := utcMinute(h.Timestamp)
		if hm.Before(chartStartMin) || hm.After(chartEndMin) {
			continue
		}
		sparseHeap = append(sparseHeap, h)
	}

	bucketCounts := make(map[time.Time]map[string]int)
	durationBucketSums := make(map[time.Time]map[string]float64)
	for _, ev := range events1h {
		if ev.Category == "" {
			continue
		}
		bucket := utcMinute(ev.Timestamp)
		if bucketCounts[bucket] == nil {
			bucketCounts[bucket] = make(map[string]int)
		}
		bucketCounts[bucket][ev.Category]++
		if durationBucketSums[bucket] == nil {
			durationBucketSums[bucket] = make(map[string]float64)
		}
		if ev.Duration > 0 {
			durationBucketSums[bucket][ev.Category] += ev.Duration
		}
	}
	fiveMinElapsed := now.Sub(fiveMinStartMin).Seconds()
	if fiveMinElapsed < 1 {
		fiveMinElapsed = 1
	}
	chartSpanElapsed := now.Sub(chartStartMin).Seconds()
	if chartSpanElapsed < 1 {
		chartSpanElapsed = 1
	}

	byPath := make(map[string][]GCEventRecord)
	for _, ev := range events1h {
		if !ev.IsPause || ev.HeapBeforeKB <= 0 && ev.HeapAfterKB <= 0 {
			continue
		}
		byPath[ev.Path] = append(byPath[ev.Path], ev)
	}
	for path, pathEvents := range byPath {
		sort.Slice(pathEvents, func(i, j int) bool { return pathEvents[i].Timestamp.Before(pathEvents[j].Timestamp) })
		for i := 1; i < len(pathEvents); i++ {
			prev, cur := pathEvents[i-1], pathEvents[i]
			dt := cur.Timestamp.Sub(prev.Timestamp).Seconds()
			if dt <= 0 {
				continue
			}
			deltaKB := cur.HeapBeforeKB - prev.HeapAfterKB
			if deltaKB < 0 {
				deltaKB = 0
			}
			rateMBPerSec := float64(deltaKB) / 1024.0 / dt
			d.AllocationRateHistory = append(d.AllocationRateHistory, AllocationRatePoint{
				Timestamp:    cur.Timestamp,
				Path:         path,
				RateMBPerSec: rateMBPerSec,
			})
		}
	}
	sparseAlloc := d.AllocationRateHistory
	d.AllocationRateHistory = nil
	sort.Slice(sparseAlloc, func(i, j int) bool {
		return sparseAlloc[i].Timestamp.Before(sparseAlloc[j].Timestamp)
	})

	allocRateByPath := make(map[string][]AllocationRatePoint)
	for _, p := range sparseAlloc {
		allocRateByPath[p.Path] = append(allocRateByPath[p.Path], p)
	}

	for path, fs := range m.fileSummary {
		pw := pathWindows[path]
		summary := FileSummary{
			Path:                     path,
			GCType:                   fs.gcType,
			LastFullGCTime:           fs.lastFullGC,
			HeapUsedKB:               fs.heapUsedKB,
			HeapTotalKB:              fs.heapTotalKB,
			PromotionFailedCount:     fs.promotionFailedCount,
			ConcurrentModeFailCount:  fs.concurrentModeFailCount,
			ToSpaceExhaustedCount:    fs.toSpaceExhaustedCount,
			HumongousAllocationCount: fs.humongousAllocationCount,
		}
		if fs.heapTotalKB > 0 {
			summary.HeapUsageRatio = float64(fs.heapUsedKB) / float64(fs.heapTotalKB)
		}

		if pw != nil {
			cc := make(map[string]int, len(pw.categoryCounts))
			for k, v := range pw.categoryCounts {
				cc[k] = v
			}
			summary.CategoryCounts = cc

			// Throughput: 1 - (pauseTime / wall time spanned by aligned windows)
			summary.Throughput.Last1m = 1.0
			if pw.totalPause1m > 0 {
				summary.Throughput.Last1m = 1.0 - (pw.totalPause1m / curMinElapsed)
				if summary.Throughput.Last1m < 0 {
					summary.Throughput.Last1m = 0
				}
			}
			summary.Throughput.Last5m = 1.0
			if pw.totalPause5m > 0 {
				summary.Throughput.Last5m = 1.0 - (pw.totalPause5m / fiveMinElapsed)
				if summary.Throughput.Last5m < 0 {
					summary.Throughput.Last5m = 0
				}
			}
			summary.Throughput.Last1h = 1.0
			if pw.totalPause1h > 0 {
				summary.Throughput.Last1h = 1.0 - (pw.totalPause1h / chartSpanElapsed)
				if summary.Throughput.Last1h < 0 {
					summary.Throughput.Last1h = 0
				}
			}

			// Max STW pause per window (same event set as throughput: IsPause && duration>=0)
			if len(pw.pause1m) > 0 {
				summary.PauseMax.Last1m = maxFloat(pw.pause1m)
			}
			if len(pw.pause5m) > 0 {
				summary.PauseMax.Last5m = maxFloat(pw.pause5m)
			}
			if len(pw.pause1h) > 0 {
				summary.PauseMax.Last1h = maxFloat(pw.pause1h)
			}

			// GC Rate: events per minute (last 1m = count in 1min, 5m = total/5, 1h = total/60)
			summary.GCRate.Last1m = float64(pw.events1m)
			summary.GCRate.Last5m = float64(pw.events5m) / 5.0
			summary.GCRate.Last1h = float64(pw.events1h) / 60.0

			// Full GC Rate
			summary.FullGCRate.Last1m = float64(pw.fullGC1m)
			summary.FullGCRate.Last5m = float64(pw.fullGC5m) / 5.0
			summary.FullGCRate.Last1h = float64(pw.fullGC1h) / 60.0

			// Alloc Rate: average MB/s in each window
			allocPoints := allocRateByPath[path]
			var sum1m, sum5m, sum1h float64
			var n1m, n5m, n1h int
			for _, p := range allocPoints {
				pm := utcMinute(p.Timestamp)
				if pm.Equal(refMin) {
					sum1m += p.RateMBPerSec
					n1m++
				}
				if !pm.Before(fiveMinStartMin) && !pm.After(refMin) {
					sum5m += p.RateMBPerSec
					n5m++
				}
				if !pm.Before(chartStartMin) && !pm.After(chartEndMin) {
					sum1h += p.RateMBPerSec
					n1h++
				}
			}
			if n1m > 0 {
				summary.AllocRate.Last1m = sum1m / float64(n1m)
			}
			if n5m > 0 {
				summary.AllocRate.Last5m = sum5m / float64(n5m)
			}
			if n1h > 0 {
				summary.AllocRate.Last1h = sum1h / float64(n1h)
			}
		} else {
			summary.Throughput.Last1m = 1.0
			summary.Throughput.Last5m = 1.0
			summary.Throughput.Last1h = 1.0
		}

		// Heap Usage (Last1m / Last5m / Last1h): mean of per-sample Java heap usage ratios in each UTC window.
		//
		// Each sample is r_i = HeapUsedKB / HeapTotalKB from HeapSnapshot, filled from GCEventRecord.HeapAfterKB /
		// HeapTotalKB when RecordEvent pushes to heapHistory. Those two fields are whatever the collector-specific
		// parser extracted as the JVM's total heap used and total heap capacity after that GC (same definition the
		// log line uses for the Java heap as a whole). Generational splits (eden, survivor, old, humongous, etc.)
		// are collector-dependent; we do not re-sum regions for this card — that would risk double-counting or
		// mismatching the JVM's own "heap" total. MetaUsedKB (Metaspace or PermGen in logs) is for the chart only;
		// metaspace is outside the Java heap in HotSpot, so it is not included in HeapUsedKB/HeapTotalKB when
		// parsers follow standard log semantics.
		var sum1m, sum5m, sum1h float64
		var n1m, n5m, n1h int
		for _, h := range sparseHeap {
			if h.Path != path || h.HeapTotalKB <= 0 {
				continue
			}
			ratio := float64(h.HeapUsedKB) / float64(h.HeapTotalKB)
			hm := utcMinute(h.Timestamp)
			if hm.Equal(refMin) {
				sum1m += ratio
				n1m++
			}
			if !hm.Before(fiveMinStartMin) && !hm.After(refMin) {
				sum5m += ratio
				n5m++
			}
			if !hm.Before(chartStartMin) && !hm.After(chartEndMin) {
				sum1h += ratio
				n1h++
			}
		}
		if n1m > 0 {
			summary.HeapUsage.Last1m = sum1m / float64(n1m)
		}
		if n5m > 0 {
			summary.HeapUsage.Last5m = sum5m / float64(n5m)
		}
		if n1h > 0 {
			summary.HeapUsage.Last1h = sum1h / float64(n1h)
		}
		// Fallback to current ratio when no history
		if n1m == 0 && summary.HeapTotalKB > 0 {
			summary.HeapUsage.Last1m = summary.HeapUsageRatio
		}
		if n5m == 0 && summary.HeapTotalKB > 0 {
			summary.HeapUsage.Last5m = summary.HeapUsageRatio
		}
		if n1h == 0 && summary.HeapTotalKB > 0 {
			summary.HeapUsage.Last1h = summary.HeapUsageRatio
		}

		summary.Status = evaluateFileSummaryHealth(summary)
		d.Files = append(d.Files, summary)
	}

	pathList := sortedFilePaths(m.fileSummary)
	d.GCEventsTimeSeries = densifyGCEventsBuckets(bucketCounts, chartStartMin, chartEndMin)
	d.GCDurationTimeSeries = densifyGCDurationBuckets(durationBucketSums, chartStartMin, chartEndMin)
	d.ThroughputHistory = densifyThroughputHistory(pathMinPauseSum, pathList, chartStartMin, chartEndMin, refMin, now)
	d.AllocationRateHistory = densifyAllocationRateHistory(sparseAlloc, pathList, chartStartMin, chartEndMin)
	d.HeapHistory = densifyHeapHistory(sparseHeap, pathList, chartStartMin, chartEndMin)

	d.OverallStatus, d.Alerts = evaluateOverallHealth(d.Files, recentFullGCCount, d.MaxSTWPause1mSec)
	d.Recommendations = computeRecommendations(d)
	return d
}

func evaluateFileSummaryHealth(s FileSummary) Status {
	fullGC1m := int(s.FullGCRate.Last1m)
	if fullGC1m >= 2 {
		return StatusCritical
	}
	if fullGC1m > 0 {
		return StatusWarning
	}
	if s.HeapTotalKB > 0 && s.HeapUsageRatio > 0.9 {
		return StatusCritical
	}
	if s.HeapTotalKB > 0 && s.HeapUsageRatio > 0.8 {
		return StatusWarning
	}
	if s.PauseMax.Last1m > 0.5 {
		return StatusCritical
	}
	if s.PauseMax.Last1m > 0.2 {
		return StatusWarning
	}
	return StatusHealthy
}

func evaluateOverallHealth(files []FileSummary, recentFullGCCount int, maxSTWPause1mSec float64) (Status, []string) {
	status := StatusHealthy
	var alerts []string

	if recentFullGCCount >= 2 {
		status = StatusCritical
		alerts = append(alerts, fmt.Sprintf("Critical: %d Full GCs in the last 1min", recentFullGCCount))
	} else if recentFullGCCount > 0 {
		if status != StatusCritical {
			status = StatusWarning
		}
		alerts = append(alerts, fmt.Sprintf("Warning: %d Full GC in the last 1min", recentFullGCCount))
	}

	if maxSTWPause1mSec > 0.5 {
		status = StatusCritical
		alerts = append(alerts, fmt.Sprintf("Critical: max STW pause %.0fms exceeds 500ms", maxSTWPause1mSec*1000))
	} else if maxSTWPause1mSec > 0.2 {
		if status != StatusCritical {
			status = StatusWarning
		}
		alerts = append(alerts, fmt.Sprintf("Warning: max STW pause %.0fms exceeds 200ms", maxSTWPause1mSec*1000))
	}

	for _, f := range files {
		if f.HeapUsageRatio > 0.9 {
			status = StatusCritical
			alerts = append(alerts, fmt.Sprintf("Critical: %s heap usage %.0f%%", f.Path, f.HeapUsageRatio*100))
		} else if f.HeapUsageRatio > 0.8 {
			if status != StatusCritical {
				status = StatusWarning
			}
			alerts = append(alerts, fmt.Sprintf("Warning: %s heap usage %.0f%%", f.Path, f.HeapUsageRatio*100))
		}
	}

	if len(alerts) == 0 {
		alerts = append(alerts, "All systems healthy")
	}
	return status, alerts
}

func percentile(data []float64, p float64) float64 {
	if len(data) == 0 {
		return 0
	}
	sorted := make([]float64, len(data))
	copy(sorted, data)
	sort.Float64s(sorted)
	idx := int(float64(len(sorted)-1) * p)
	return sorted[idx]
}

func avgFloat(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range data {
		sum += v
	}
	return sum / float64(len(data))
}

func maxFloat(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}
	m := data[0]
	for _, v := range data[1:] {
		if v > m {
			m = v
		}
	}
	return m
}

func shortPath(p string) string {
	parts := strings.Split(p, "/")
	if len(parts) > 2 {
		return ".../" + strings.Join(parts[len(parts)-2:], "/")
	}
	return p
}
