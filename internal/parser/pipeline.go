package parser

import (
	"log/slog"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/loyispa/jgc_exporter/internal/health"
	"github.com/loyispa/jgc_exporter/internal/metric"
)

// Pre-unified safepoint: "Total time for which application threads were stopped: X seconds, Stopping threads took: Y seconds"
var preUnifiedSafepointRe = regexp.MustCompile(
	`Total time for which application threads were stopped:\s*([\d.]+)\s*seconds,\s*Stopping threads took:\s*([\d.]+)\s*seconds`)

// Unified safepoint: "[safepoint ...] ... At safepoint: X ns, Total: Y ns"
var unifiedSafepointRe = regexp.MustCompile(
	`\[safepoint\s*\].*?At safepoint:\s*(\d+)\s*ns.*?Total:\s*(\d+)\s*ns`)

// Unified safepoint reaching: "Reaching safepoint: X ns"
var unifiedSafepointReachingRe = regexp.MustCompile(
	`Reaching safepoint:\s*(\d+)\s*ns`)

// Unified GC phase duration: "[gc,phases ] GC(N) Phase Name Xms"
var unifiedGCPhaseRe = regexp.MustCompile(
	`\[gc,phases\s*\]\s+GC\(\d+\)\s+(.+?)\s+([\d.]+)ms\s*$`)

// Unified GC worker count: "[gc,task ] GC(N) Using N workers"
var unifiedGCWorkerRe = regexp.MustCompile(
	`\[gc,task\s*\]\s+GC\(\d+\)\s+Using\s+(\d+)\s+workers`)

type detectResult struct {
	gcType  GCType
	unified bool
}

type Pipeline struct {
	parsers       map[string]Parser
	acceptCache   map[string]detectResult // path -> result from Accept, consumed by OnOpen
	unifiedByPath map[string]bool         // path -> unified JVM logging (for feed-line normalization)
	hostname      string
	monitor       *health.Monitor
	mu            sync.RWMutex
}

func NewPipeline(monitor *health.Monitor) *Pipeline {
	hostname, _ := os.Hostname()
	return &Pipeline{
		parsers:       make(map[string]Parser),
		acceptCache:   make(map[string]detectResult),
		unifiedByPath: make(map[string]bool),
		hostname:      hostname,
		monitor:       monitor,
	}
}

// Accept opens path, detects GC format; caches result and returns true only for valid GC logs.
// OnOpen will use the cached result and must not reopen the file.
func (pl *Pipeline) Accept(path string) bool {
	gcType := DetectGCType(path)
	unified := DetectUnifiedLogging(path)

	if gcType == GCTypeUnknown && !unified {
		slog.Debug("skipping non-GC log", "path", path)
		return false
	}

	pl.mu.Lock()
	pl.acceptCache[path] = detectResult{gcType: gcType, unified: unified}
	pl.mu.Unlock()
	return true
}

func (pl *Pipeline) OnOpen(path string) bool {
	pl.mu.Lock()
	dr, ok := pl.acceptCache[path]
	if ok {
		delete(pl.acceptCache, path)
	}
	pl.mu.Unlock()

	if !ok {
		slog.Warn("OnOpen called without prior Accept", "path", path)
		return false
	}

	gcType, unified := dr.gcType, dr.unified

	if gcType == GCTypeUnknown && unified {
		slog.Info("GC type unknown for unified log, registering deferred parser", "path", path)
		dp := newDeferredParser(pl, path)
		pl.mu.Lock()
		pl.parsers[path] = dp
		pl.mu.Unlock()
		return true
	}

	if gcType == GCTypeUnknown {
		return false
	}

	pl.registerParser(path, gcType, unified)
	return true
}

func (pl *Pipeline) registerParser(path string, gcType GCType, unified bool) {
	var p Parser
	switch gcType {
	case GCTypeG1:
		p = NewG1Parser(unified)
	case GCTypeCMS:
		p = NewCMSParser(unified)
	case GCTypeZGC:
		p = NewZGCParser()
	case GCTypeParallel:
		p = NewParallelParser(unified)
	case GCTypeSerial:
		p = NewSerialParser(unified)
	default:
		slog.Warn("unsupported GC type, skipping", "path", path, "type", gcType)
		return
	}

	handler := &metricsHandler{path: path, hostname: pl.hostname, monitor: pl.monitor}
	p.SetHandler(handler)
	pl.monitor.SetGCType(path, string(gcType))

	pl.mu.Lock()
	pl.parsers[path] = p
	pl.unifiedByPath[path] = unified
	pl.mu.Unlock()

	slog.Info("registered log", "path", path, "gc_type", gcType, "unified", unified)
}

// parserFeedLine normalizes unified JVM log lines (strip decorator brackets, reattach timestamp+[gc]) for GC parsers.
func (pl *Pipeline) parserFeedLine(path, line string) string {
	pl.mu.RLock()
	unified := pl.unifiedByPath[path]
	pl.mu.RUnlock()
	if !unified {
		return line
	}
	payload, ok := StripUnifiedJVMLogPrefix(line)
	if !ok {
		return line
	}
	return SyntheticUnifiedGCLine(line, payload)
}

// OnClose is called when a file becomes idle (service stopped) or the
// exporter is shutting down. Flushes pending events, removes the parser,
// and cleans up all Prometheus metrics for this path.
func (pl *Pipeline) OnClose(path string) {
	pl.mu.Lock()
	p := pl.parsers[path]
	delete(pl.parsers, path)
	pl.mu.Unlock()

	if p != nil {
		p.Flush()
	}

	pathLabel := prometheus.Labels{"path": path}
	metric.EventDuration.DeletePartialMatch(pathLabel)
	metric.EventPauseDuration.DeletePartialMatch(pathLabel)
	metric.EventLastMinDuration.DeletePartialMatch(pathLabel)
	metric.EventLastMinPauseDuration.DeletePartialMatch(pathLabel)
	metric.HeapUsedBeforeCollection.DeletePartialMatch(pathLabel)
	metric.HeapUsedAfterCollection.DeletePartialMatch(pathLabel)
	metric.HeapSizeBeforeCollection.DeletePartialMatch(pathLabel)
	metric.HeapSizeAfterCollection.DeletePartialMatch(pathLabel)
	metric.MetaspaceUsedBeforeCollection.DeletePartialMatch(pathLabel)
	metric.MetaspaceUsedAfterCollection.DeletePartialMatch(pathLabel)
	metric.MetaspaceSizeBeforeCollection.DeletePartialMatch(pathLabel)
	metric.MetaspaceSizeAfterCollection.DeletePartialMatch(pathLabel)
	metric.ZGCPauseMarkStartDuration.DeletePartialMatch(pathLabel)
	metric.ZGCConcurrentMarkDuration.DeletePartialMatch(pathLabel)
	metric.ZGCPauseMarkEndDuration.DeletePartialMatch(pathLabel)
	metric.ZGCPauseRelocateStartDuration.DeletePartialMatch(pathLabel)
	metric.ZGCConcurrentRelocateDuration.DeletePartialMatch(pathLabel)
	metric.ZGCLoad1m.DeletePartialMatch(pathLabel)
	metric.ZGCLoad5m.DeletePartialMatch(pathLabel)
	metric.ZGCLoad15m.DeletePartialMatch(pathLabel)
	metric.ZGCMMU2ms.DeletePartialMatch(pathLabel)
	metric.ZGCMMU5ms.DeletePartialMatch(pathLabel)
	metric.ZGCMMU10ms.DeletePartialMatch(pathLabel)
	metric.ZGCMMU20ms.DeletePartialMatch(pathLabel)
	metric.ZGCMMU50ms.DeletePartialMatch(pathLabel)
	metric.ZGCMMU100ms.DeletePartialMatch(pathLabel)
	metric.ZGCMetaspaceUsed.DeletePartialMatch(pathLabel)
	metric.ZGCMetaspaceCommitted.DeletePartialMatch(pathLabel)
	metric.SafepointDuration.DeletePartialMatch(pathLabel)
	metric.SafepointStopThreadsDuration.DeletePartialMatch(pathLabel)
	metric.GCWorkers.DeletePartialMatch(pathLabel)
	metric.CMSSymbolTableProcessDuration.DeletePartialMatch(pathLabel)
	metric.CMSStringTableProcessDuration.DeletePartialMatch(pathLabel)
	metric.CMSClassUnloadingProcessDuration.DeletePartialMatch(pathLabel)
	metric.CMSSymbolAndStringTableProcessDuration.DeletePartialMatch(pathLabel)
	metric.ZGCConcurrentMarkFreeDuration.DeletePartialMatch(pathLabel)
	metric.ZGCProcessNonStrongReferencesDuration.DeletePartialMatch(pathLabel)
	metric.ZGCConcurrentResetRelocationsetDuration.DeletePartialMatch(pathLabel)
	metric.ZGCConcurrentSelectRelocationsetDuration.DeletePartialMatch(pathLabel)
	metric.G1EdenBeforeCollectionRegions.DeletePartialMatch(pathLabel)
	metric.G1EdenAfterCollectionRegions.DeletePartialMatch(pathLabel)
	metric.G1EdenAssignRegions.DeletePartialMatch(pathLabel)
	metric.G1SurvivorBeforeCollectionRegions.DeletePartialMatch(pathLabel)
	metric.G1SurvivorAfterCollectionRegions.DeletePartialMatch(pathLabel)
	metric.G1SurvivorAssignRegions.DeletePartialMatch(pathLabel)
	metric.G1OldBeforeCollectionRegions.DeletePartialMatch(pathLabel)
	metric.G1OldAfterCollectionRegions.DeletePartialMatch(pathLabel)
	metric.G1HumongousBeforeCollectionRegions.DeletePartialMatch(pathLabel)
	metric.G1HumongousAfterCollectionRegions.DeletePartialMatch(pathLabel)
	metric.G1ArchiveBeforeCollectionRegions.DeletePartialMatch(pathLabel)
	metric.G1ArchiveAfterCollectionRegions.DeletePartialMatch(pathLabel)
	metric.G1ArchiveAssignRegions.DeletePartialMatch(pathLabel)
	metric.G1SoftReferences.DeletePartialMatch(pathLabel)
	metric.G1SoftReferencePauseDuration.DeletePartialMatch(pathLabel)
	metric.CMSSoftReferences.DeletePartialMatch(pathLabel)
	metric.CMSSoftReferencePauseDuration.DeletePartialMatch(pathLabel)
	metric.G1WeakReferences.DeletePartialMatch(pathLabel)
	metric.G1WeakReferencePauseDuration.DeletePartialMatch(pathLabel)
	metric.CMSWeakReferences.DeletePartialMatch(pathLabel)
	metric.CMSWeakReferencePauseDuration.DeletePartialMatch(pathLabel)
	metric.G1FinalReferences.DeletePartialMatch(pathLabel)
	metric.G1FinalReferencePauseDuration.DeletePartialMatch(pathLabel)
	metric.CMSFinalReferences.DeletePartialMatch(pathLabel)
	metric.CMSFinalReferencePauseDuration.DeletePartialMatch(pathLabel)
	metric.G1PhantomReferences.DeletePartialMatch(pathLabel)
	metric.G1FreePhantomReferences.DeletePartialMatch(pathLabel)
	metric.G1PhantomReferencePauseDuration.DeletePartialMatch(pathLabel)
	metric.CMSPhantomReferences.DeletePartialMatch(pathLabel)
	metric.CMSFreePhantomReferences.DeletePartialMatch(pathLabel)
	metric.CMSPhantomReferencePauseDuration.DeletePartialMatch(pathLabel)
	metric.G1JNIWeakReferences.DeletePartialMatch(pathLabel)
	metric.G1JNIWeakReferencePauseDuration.DeletePartialMatch(pathLabel)
	metric.CMSJNIWeakReferences.DeletePartialMatch(pathLabel)
	metric.CMSJNIWeakReferencePauseDuration.DeletePartialMatch(pathLabel)
	metric.ZGCMarkStartUsedBytes.DeletePartialMatch(pathLabel)
	metric.ZGCMarkStartFreeBytes.DeletePartialMatch(pathLabel)
	metric.ZGCMarkEndUsedBytes.DeletePartialMatch(pathLabel)
	metric.ZGCMarkEndFreeBytes.DeletePartialMatch(pathLabel)
	metric.ZGCRelocateStartUsedBytes.DeletePartialMatch(pathLabel)
	metric.ZGCRelocateStartFreeBytes.DeletePartialMatch(pathLabel)
	metric.ZGCRelocateEndUsedBytes.DeletePartialMatch(pathLabel)
	metric.ZGCRelocateEndFreeBytes.DeletePartialMatch(pathLabel)
	metric.ZGCMetaspaceReservedBytes.DeletePartialMatch(pathLabel)
	metric.ThroughputGauge.DeletePartialMatch(pathLabel)
	metric.AllocationRateGauge.DeletePartialMatch(pathLabel)
	metric.LogLines.DeletePartialMatch(pathLabel)

	pl.mu.Lock()
	delete(pl.unifiedByPath, path)
	pl.mu.Unlock()

	pl.monitor.RemoveFile(path)
	slog.Info("unregistered log", "path", path)
}

func truncateForLog(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

func (pl *Pipeline) OnRead(path string, line string) {
	pl.mu.RLock()
	_, ok := pl.parsers[path]
	pl.mu.RUnlock()

	if !ok {
		slog.Warn("dropping line, no parser registered", "path", path, "line_preview", truncateForLog(line, 120))
		return
	}

	slog.Debug("consumed line", "path", path, "line_preview", truncateForLog(line, 100))

	gcType := pl.monitor.GetGCType(path)
	if gcType == "" {
		gcType = "unknown"
	}
	metric.LogLines.WithLabelValues(path, pl.hostname, gcType).Inc()

	// Skip VerifyBeforeGC/VerifyAfterGC diagnostic output
	if isVerifyDiagnosticLine(line) {
		slog.Debug("parser skipping verify diagnostic line", "path", path, "gc_type", gcType)
		return
	}

	labels := []string{path, pl.hostname}

	// Safepoint (pre-unified): appears in all GC logs with -XX:+PrintGCApplicationStoppedTime
	if m := preUnifiedSafepointRe.FindStringSubmatch(line); len(m) > 0 {
		slog.Debug("parser handled pre-unified safepoint line", "path", path, "gc_type", gcType)
		totalSec := parseFloat(m[1])
		stopSec := parseFloat(m[2])
		metric.SafepointDuration.WithLabelValues(labels...).Observe(totalSec)
		metric.SafepointStopThreadsDuration.WithLabelValues(labels...).Observe(stopSec)
		return
	}

	// Safepoint (unified): [safepoint] tag
	if strings.Contains(line, "[safepoint]") {
		slog.Debug("parser handled unified safepoint line", "path", path, "gc_type", gcType)
		if m := unifiedSafepointRe.FindStringSubmatch(line); len(m) > 0 {
			totalNs := parseFloat(m[2])
			metric.SafepointDuration.WithLabelValues(labels...).Observe(totalNs / 1e9)
			if rm := unifiedSafepointReachingRe.FindStringSubmatch(line); len(rm) > 0 {
				reachingNs := parseFloat(rm[1])
				metric.SafepointStopThreadsDuration.WithLabelValues(labels...).Observe(reachingNs / 1e9)
			}
		}
		return
	}

	// GC Worker count (unified): [gc,task] lines
	if strings.Contains(line, "[gc,task") {
		if m := unifiedGCWorkerRe.FindStringSubmatch(line); len(m) > 0 {
			workers := parseFloat(m[1])
			metric.GCWorkers.WithLabelValues(labels...).Set(workers)
		}
	}

	pl.monitor.RecordRawLine(path, line)

	// Feed to collector-specific parser
	pl.mu.RLock()
	p, ok := pl.parsers[path]
	pl.mu.RUnlock()
	if ok {
		p.Feed(pl.parserFeedLine(path, line))
	}
}

const maxDeferredLines = 200

// deferredParser buffers lines for unified logs where GC type could not be
// determined from the file header (e.g. rotated log files missing "Using ..."
// preamble). Once a distinguishing line is seen, it promotes to the real parser
// and replays the buffer.
type deferredParser struct {
	pipeline *Pipeline
	path     string
	buffer   []string
	resolved bool
}

func newDeferredParser(pl *Pipeline, path string) *deferredParser {
	return &deferredParser{pipeline: pl, path: path}
}

func (d *deferredParser) GCType() GCType            { return GCTypeUnknown }
func (d *deferredParser) SetHandler(_ EventHandler) {}
func (d *deferredParser) Flush()                    {}

func (d *deferredParser) Feed(line string) {
	if d.resolved {
		return
	}

	d.buffer = append(d.buffer, line)

	if gcType := detectFromLine(line); gcType != GCTypeUnknown {
		d.resolve(gcType)
		return
	}

	if len(d.buffer) >= maxDeferredLines {
		slog.Warn("deferred parser exhausted buffer without detecting GC type, defaulting to Parallel",
			"path", d.path, "lines", len(d.buffer))
		d.resolve(GCTypeParallel)
	}
}

func (d *deferredParser) resolve(gcType GCType) {
	d.resolved = true
	slog.Info("deferred parser resolved GC type", "path", d.path, "gc_type", gcType)

	d.pipeline.registerParser(d.path, gcType, true)

	d.pipeline.mu.RLock()
	p, ok := d.pipeline.parsers[d.path]
	d.pipeline.mu.RUnlock()

	if ok {
		for _, line := range d.buffer {
			p.Feed(d.pipeline.parserFeedLine(d.path, line))
		}
	}
	d.buffer = nil
}

type metricsHandler struct {
	path     string
	hostname string
	monitor  *health.Monitor
}

func (h *metricsHandler) Handle(ev *GCEvent) {
	if ev == nil {
		return
	}

	labels := []string{h.path, h.hostname}

	metric.EventDuration.WithLabelValues(h.path, h.hostname, ev.Category).Observe(ev.Duration)
	metric.EventLastMinDuration.WithLabelValues(labels...).Observe(ev.Duration)

	if ev.IsPause {
		metric.EventPauseDuration.WithLabelValues(h.path, h.hostname, ev.Category).Observe(ev.Duration)
		metric.EventLastMinPauseDuration.WithLabelValues(labels...).Observe(ev.Duration)
	}

	if ev.HeapTotalKB > 0 {
		metric.HeapUsedBeforeCollection.WithLabelValues(labels...).Set(float64(ev.HeapBeforeKB) * 1024)
		metric.HeapUsedAfterCollection.WithLabelValues(labels...).Set(float64(ev.HeapAfterKB) * 1024)
		metric.HeapSizeBeforeCollection.WithLabelValues(labels...).Set(float64(ev.HeapTotalKB) * 1024)
		metric.HeapSizeAfterCollection.WithLabelValues(labels...).Set(float64(ev.HeapTotalKB) * 1024)
	}

	if ev.MetaTotalKB > 0 {
		metric.MetaspaceUsedBeforeCollection.WithLabelValues(labels...).Set(float64(ev.MetaBeforeKB) * 1024)
		metric.MetaspaceUsedAfterCollection.WithLabelValues(labels...).Set(float64(ev.MetaAfterKB) * 1024)
		metric.MetaspaceSizeBeforeCollection.WithLabelValues(labels...).Set(float64(ev.MetaTotalKB) * 1024)
		metric.MetaspaceSizeAfterCollection.WithLabelValues(labels...).Set(float64(ev.MetaTotalKB) * 1024)
	}

	isFullGC := strings.Contains(strings.ToLower(ev.Category), "full") ||
		strings.Contains(strings.ToLower(ev.Category), "systemgc")
	metaAfter := ev.MetaAfterKB
	if metaAfter == 0 && ev.ZGCMetaspaceUsedKB > 0 {
		metaAfter = ev.ZGCMetaspaceUsedKB
	}
	h.monitor.RecordEvent(health.GCEventRecord{
		Timestamp:     ev.Timestamp,
		Path:          h.path,
		Category:      ev.Category,
		Duration:      ev.Duration,
		IsPause:       ev.IsPause,
		HeapBeforeKB:  ev.HeapBeforeKB,
		HeapAfterKB:   ev.HeapAfterKB,
		HeapTotalKB:   ev.HeapTotalKB,
		YoungBeforeKB: ev.YoungBeforeKB,
		YoungAfterKB:  ev.YoungAfterKB,
		OldBeforeKB:   ev.OldBeforeKB,
		OldAfterKB:    ev.OldAfterKB,
		MetaBeforeKB:  ev.MetaBeforeKB,
		MetaAfterKB:   metaAfter,
		Cause:         ev.Cause,
		IsFullGC:      isFullGC,
	})

	// CMS String/Symbol Table durations
	if ev.CMSSymbolTableMs > 0 {
		metric.CMSSymbolTableProcessDuration.WithLabelValues(labels...).Observe(ev.CMSSymbolTableMs / 1000.0)
	}
	if ev.CMSStringTableMs > 0 {
		metric.CMSStringTableProcessDuration.WithLabelValues(labels...).Observe(ev.CMSStringTableMs / 1000.0)
	}
	if ev.CMSClassUnloadingMs > 0 {
		metric.CMSClassUnloadingProcessDuration.WithLabelValues(labels...).Observe(ev.CMSClassUnloadingMs / 1000.0)
	}
	if ev.CMSSymbolTableMs > 0 || ev.CMSStringTableMs > 0 {
		metric.CMSSymbolAndStringTableProcessDuration.WithLabelValues(labels...).Observe((ev.CMSSymbolTableMs + ev.CMSStringTableMs) / 1000.0)
	}

	// ZGC-specific
	if ev.GCType == GCTypeZGC {
		if ev.ZGCPauseMarkStartMs > 0 {
			metric.ZGCPauseMarkStartDuration.WithLabelValues(labels...).Observe(ev.ZGCPauseMarkStartMs / 1000.0)
		}
		if ev.ZGCConcurrentMarkMs > 0 {
			metric.ZGCConcurrentMarkDuration.WithLabelValues(labels...).Observe(ev.ZGCConcurrentMarkMs / 1000.0)
		}
		if ev.ZGCPauseMarkEndMs > 0 {
			metric.ZGCPauseMarkEndDuration.WithLabelValues(labels...).Observe(ev.ZGCPauseMarkEndMs / 1000.0)
		}
		if ev.ZGCPauseRelocateMs > 0 {
			metric.ZGCPauseRelocateStartDuration.WithLabelValues(labels...).Observe(ev.ZGCPauseRelocateMs / 1000.0)
		}
		if ev.ZGCConcurrentRelocMs > 0 {
			metric.ZGCConcurrentRelocateDuration.WithLabelValues(labels...).Observe(ev.ZGCConcurrentRelocMs / 1000.0)
		}
		if ev.ZGCConcurrentMarkFreeMs > 0 {
			metric.ZGCConcurrentMarkFreeDuration.WithLabelValues(labels...).Observe(ev.ZGCConcurrentMarkFreeMs / 1000.0)
		}
		if ev.ZGCConcurrentProcessNSRMs > 0 {
			metric.ZGCProcessNonStrongReferencesDuration.WithLabelValues(labels...).Observe(ev.ZGCConcurrentProcessNSRMs / 1000.0)
		}
		if ev.ZGCConcurrentResetRelocMs > 0 {
			metric.ZGCConcurrentResetRelocationsetDuration.WithLabelValues(labels...).Observe(ev.ZGCConcurrentResetRelocMs / 1000.0)
		}
		if ev.ZGCConcurrentSelectRelocMs > 0 {
			metric.ZGCConcurrentSelectRelocationsetDuration.WithLabelValues(labels...).Observe(ev.ZGCConcurrentSelectRelocMs / 1000.0)
		}
		if ev.ZGCLoad1m > 0 {
			metric.ZGCLoad1m.WithLabelValues(labels...).Set(ev.ZGCLoad1m)
		}
		if ev.ZGCLoad5m > 0 {
			metric.ZGCLoad5m.WithLabelValues(labels...).Set(ev.ZGCLoad5m)
		}
		if ev.ZGCLoad15m > 0 {
			metric.ZGCLoad15m.WithLabelValues(labels...).Set(ev.ZGCLoad15m)
		}
		metric.ZGCMMU2ms.WithLabelValues(labels...).Set(ev.ZGCMMU2ms)
		metric.ZGCMMU5ms.WithLabelValues(labels...).Set(ev.ZGCMMU5ms)
		metric.ZGCMMU10ms.WithLabelValues(labels...).Set(ev.ZGCMMU10ms)
		metric.ZGCMMU20ms.WithLabelValues(labels...).Set(ev.ZGCMMU20ms)
		metric.ZGCMMU50ms.WithLabelValues(labels...).Set(ev.ZGCMMU50ms)
		metric.ZGCMMU100ms.WithLabelValues(labels...).Set(ev.ZGCMMU100ms)
		if ev.ZGCMetaspaceUsedKB > 0 {
			metric.ZGCMetaspaceUsed.WithLabelValues(labels...).Set(float64(ev.ZGCMetaspaceUsedKB) * 1024)
		}
		if ev.ZGCMetaspaceCommitKB > 0 {
			metric.ZGCMetaspaceCommitted.WithLabelValues(labels...).Set(float64(ev.ZGCMetaspaceCommitKB) * 1024)
		}
		if ev.ZGCMetaspaceReservedKB > 0 {
			metric.ZGCMetaspaceReservedBytes.WithLabelValues(labels...).Set(float64(ev.ZGCMetaspaceReservedKB) * 1024)
		}
		if ev.ZGCMarkStartUsedKB > 0 {
			metric.ZGCMarkStartUsedBytes.WithLabelValues(labels...).Set(float64(ev.ZGCMarkStartUsedKB) * 1024)
		}
		if ev.ZGCMarkStartFreeKB > 0 {
			metric.ZGCMarkStartFreeBytes.WithLabelValues(labels...).Set(float64(ev.ZGCMarkStartFreeKB) * 1024)
		}
		if ev.ZGCMarkEndUsedKB > 0 {
			metric.ZGCMarkEndUsedBytes.WithLabelValues(labels...).Set(float64(ev.ZGCMarkEndUsedKB) * 1024)
		}
		if ev.ZGCMarkEndFreeKB > 0 {
			metric.ZGCMarkEndFreeBytes.WithLabelValues(labels...).Set(float64(ev.ZGCMarkEndFreeKB) * 1024)
		}
		if ev.ZGCRelocateStartUsedKB > 0 {
			metric.ZGCRelocateStartUsedBytes.WithLabelValues(labels...).Set(float64(ev.ZGCRelocateStartUsedKB) * 1024)
		}
		if ev.ZGCRelocateStartFreeKB > 0 {
			metric.ZGCRelocateStartFreeBytes.WithLabelValues(labels...).Set(float64(ev.ZGCRelocateStartFreeKB) * 1024)
		}
		if ev.ZGCRelocateEndUsedKB > 0 {
			metric.ZGCRelocateEndUsedBytes.WithLabelValues(labels...).Set(float64(ev.ZGCRelocateEndUsedKB) * 1024)
		}
		if ev.ZGCRelocateEndFreeKB > 0 {
			metric.ZGCRelocateEndFreeBytes.WithLabelValues(labels...).Set(float64(ev.ZGCRelocateEndFreeKB) * 1024)
		}
	}

	// G1-specific: region counts (unified log); no generic young/old gauges.
	if ev.GCType == GCTypeG1 {
		if ev.G1EdenRegionBefore > 0 || ev.G1EdenRegionAfter > 0 {
			metric.G1EdenBeforeCollectionRegions.WithLabelValues(labels...).Set(float64(ev.G1EdenRegionBefore))
			metric.G1EdenAfterCollectionRegions.WithLabelValues(labels...).Set(float64(ev.G1EdenRegionAfter))
			metric.G1EdenAssignRegions.WithLabelValues(labels...).Set(float64(ev.G1EdenRegionAssign))
		}
		if ev.G1SurvivorRegionBefore > 0 || ev.G1SurvivorRegionAfter > 0 {
			metric.G1SurvivorBeforeCollectionRegions.WithLabelValues(labels...).Set(float64(ev.G1SurvivorRegionBefore))
			metric.G1SurvivorAfterCollectionRegions.WithLabelValues(labels...).Set(float64(ev.G1SurvivorRegionAfter))
			metric.G1SurvivorAssignRegions.WithLabelValues(labels...).Set(float64(ev.G1SurvivorRegionAssign))
		}
		if ev.G1OldRegionBefore > 0 || ev.G1OldRegionAfter > 0 {
			metric.G1OldBeforeCollectionRegions.WithLabelValues(labels...).Set(float64(ev.G1OldRegionBefore))
			metric.G1OldAfterCollectionRegions.WithLabelValues(labels...).Set(float64(ev.G1OldRegionAfter))
		}
		if ev.G1HumongousRegionBefore > 0 || ev.G1HumongousRegionAfter > 0 {
			metric.G1HumongousBeforeCollectionRegions.WithLabelValues(labels...).Set(float64(ev.G1HumongousRegionBefore))
			metric.G1HumongousAfterCollectionRegions.WithLabelValues(labels...).Set(float64(ev.G1HumongousRegionAfter))
		}
		if ev.G1ArchiveRegionBefore > 0 || ev.G1ArchiveRegionAfter > 0 {
			metric.G1ArchiveBeforeCollectionRegions.WithLabelValues(labels...).Set(float64(ev.G1ArchiveRegionBefore))
			metric.G1ArchiveAfterCollectionRegions.WithLabelValues(labels...).Set(float64(ev.G1ArchiveRegionAfter))
			metric.G1ArchiveAssignRegions.WithLabelValues(labels...).Set(float64(ev.G1ArchiveRegionAssign))
		}
	}

	emitReferenceProcessingMetrics(ev, labels)
}

// emitReferenceProcessingMetrics records G1- or CMS-specific reference metrics (Remark + PrintReferenceGC).
func emitReferenceProcessingMetrics(ev *GCEvent, labels []string) {
	if ev.GCType != GCTypeG1 && ev.GCType != GCTypeCMS {
		return
	}
	isG1 := ev.GCType == GCTypeG1

	if ev.SoftRefPauseMs > 0 || ev.SoftRefCount > 0 {
		sec := ev.SoftRefPauseMs / 1000.0
		if isG1 {
			metric.G1SoftReferences.WithLabelValues(labels...).Set(float64(ev.SoftRefCount))
			metric.G1SoftReferencePauseDuration.WithLabelValues(labels...).Observe(sec)
		} else {
			metric.CMSSoftReferences.WithLabelValues(labels...).Set(float64(ev.SoftRefCount))
			metric.CMSSoftReferencePauseDuration.WithLabelValues(labels...).Observe(sec)
		}
	}
	if ev.WeakRefPauseMs > 0 || ev.WeakRefCount > 0 {
		sec := ev.WeakRefPauseMs / 1000.0
		if isG1 {
			metric.G1WeakReferences.WithLabelValues(labels...).Set(float64(ev.WeakRefCount))
			metric.G1WeakReferencePauseDuration.WithLabelValues(labels...).Observe(sec)
		} else {
			metric.CMSWeakReferences.WithLabelValues(labels...).Set(float64(ev.WeakRefCount))
			metric.CMSWeakReferencePauseDuration.WithLabelValues(labels...).Observe(sec)
		}
	}
	if ev.FinalRefPauseMs > 0 || ev.FinalRefCount > 0 {
		sec := ev.FinalRefPauseMs / 1000.0
		if isG1 {
			metric.G1FinalReferences.WithLabelValues(labels...).Set(float64(ev.FinalRefCount))
			metric.G1FinalReferencePauseDuration.WithLabelValues(labels...).Observe(sec)
		} else {
			metric.CMSFinalReferences.WithLabelValues(labels...).Set(float64(ev.FinalRefCount))
			metric.CMSFinalReferencePauseDuration.WithLabelValues(labels...).Observe(sec)
		}
	}
	if ev.PhantomRefPauseMs > 0 || ev.PhantomRefCount > 0 || ev.PhantomRefFree > 0 {
		sec := ev.PhantomRefPauseMs / 1000.0
		if isG1 {
			metric.G1PhantomReferences.WithLabelValues(labels...).Set(float64(ev.PhantomRefCount))
			metric.G1FreePhantomReferences.WithLabelValues(labels...).Set(float64(ev.PhantomRefFree))
			metric.G1PhantomReferencePauseDuration.WithLabelValues(labels...).Observe(sec)
		} else {
			metric.CMSPhantomReferences.WithLabelValues(labels...).Set(float64(ev.PhantomRefCount))
			metric.CMSFreePhantomReferences.WithLabelValues(labels...).Set(float64(ev.PhantomRefFree))
			metric.CMSPhantomReferencePauseDuration.WithLabelValues(labels...).Observe(sec)
		}
	}
	if ev.JNIWeakRefPauseMs > 0 {
		sec := ev.JNIWeakRefPauseMs / 1000.0
		if isG1 {
			metric.G1JNIWeakReferencePauseDuration.WithLabelValues(labels...).Observe(sec)
		} else {
			metric.CMSJNIWeakReferencePauseDuration.WithLabelValues(labels...).Observe(sec)
		}
	}
}
