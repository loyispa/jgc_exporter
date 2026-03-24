package health

import (
	"fmt"
	"strings"
)

// computeRecommendations builds tuning hints from multi-window FileSummary stats and recent Full GC causes.
// It prefers Last5m / Last1h over Last1m alone to reduce UI flicker and avoid generic one-size advice.
func computeRecommendations(d Dashboard) []Recommendation {
	var recs []Recommendation
	files := d.Files
	if len(files) == 0 {
		return recs
	}

	maxSTW5m := 0.0
	maxSTW1h := 0.0
	for _, f := range files {
		if f.PauseMax.Last5m > maxSTW5m {
			maxSTW5m = f.PauseMax.Last5m
		}
		if f.PauseMax.Last1h > maxSTW1h {
			maxSTW1h = f.PauseMax.Last1h
		}
	}
	// STW: require sustained pressure (5m or 1h max), with 1m dashboard max as tie-breaker for spikes.
	stwRef := maxSTW5m
	if stwRef < maxSTW1h {
		stwRef = maxSTW1h
	}
	if d.MaxSTWPause1mSec > stwRef {
		stwRef = d.MaxSTWPause1mSec
	}

	if stwRef > 0.5 {
		recs = append(recs, Recommendation{
			Condition: fmt.Sprintf("Max STW pause ~%.0fms (5m/1h/1m blended; worst across files)", stwRef*1000),
			Advice:    pauseAdviceForFiles(files, stwRef),
		})
	} else if stwRef > 0.2 {
		recs = append(recs, Recommendation{
			Condition: fmt.Sprintf("Max STW pause ~%.0fms exceeds 200ms (multi-window)", stwRef*1000),
			Advice:    pauseAdviceForFiles(files, stwRef),
		})
	}

	// Global throughput (1m) + per-file 5m: avoid single-minute noise.
	if d.ThroughputRatio < 0.9 && d.ThroughputRatio > 0 {
		low5m := false
		for _, f := range files {
			if f.Throughput.Last5m < 0.9 && f.Throughput.Last5m > 0 {
				low5m = true
				break
			}
		}
		if low5m || d.ThroughputRatio < 0.85 {
			recs = append(recs, Recommendation{
				Condition: fmt.Sprintf("Throughput %.1f%% (1m); at least one file below 90%% in 5m window", d.ThroughputRatio*100),
				Advice:      throughputAdvice(files, d.ThroughputRatio),
			})
		}
	}

	// Full GC: sustained pressure (5m rate >= 0.2/min implies ≥1 Full GC in 5m) or elevated 1h rate.
	for _, f := range files {
		path := shortPath(f.Path)
		gc := strings.ToLower(strings.TrimSpace(f.GCType))
		switch {
		case f.FullGCRate.Last5m >= 0.2:
			recs = append(recs, Recommendation{
				Condition: fmt.Sprintf("%s [%s]: Full GC pressure in last 5m (%.2f/min avg)", path, f.GCType, f.FullGCRate.Last5m),
				Advice:      fullGCAdvice(gc, f),
			})
		case f.FullGCRate.Last1h >= 0.05:
			recs = append(recs, Recommendation{
				Condition: fmt.Sprintf("%s [%s]: elevated Full GC rate in last 1h (%.3f/min avg)", path, f.GCType, f.FullGCRate.Last1h),
				Advice:      fullGCAdvice(gc, f),
			})
		}
	}

	hasAllocationFailure := false
	hasSystemGC := false
	for _, fgc := range d.RecentFullGCs5m {
		causeLower := strings.ToLower(fgc.Cause)
		if strings.Contains(causeLower, "allocation failure") {
			hasAllocationFailure = true
		}
		if strings.Contains(causeLower, "system.gc") || strings.Contains(causeLower, "system") {
			hasSystemGC = true
		}
	}
	if hasAllocationFailure {
		recs = append(recs, Recommendation{
			Condition: "Full GC with Allocation Failure (seen in last 5m)",
			Advice:    allocationFailureAdvice(files),
		})
	}
	if hasSystemGC {
		recs = append(recs, Recommendation{
			Condition: "Full GC involving System.gc (last 5m)",
			Advice:    "Find explicit System.gc() calls (frameworks, JVM tools). Consider -XX:+DisableExplicitGC or -XX:+ExplicitGCInvokesConcurrent where appropriate; fix callers if possible.",
		})
	}

	for _, f := range files {
		path := shortPath(f.Path)
		gc := strings.ToLower(strings.TrimSpace(f.GCType))
		heap5 := f.HeapUsage.Last5m
		heap1h := f.HeapUsage.Last1h
		alloc5 := f.AllocRate.Last5m

		if heap5 > 0.9 || heap1h > 0.9 {
			recs = append(recs, Recommendation{
				Condition: fmt.Sprintf("%s: Java heap usage high (5m ~%.0f%%, 1h ~%.0f%%)", path, heap5*100, heap1h*100),
				Advice:    heapHighAdvice(f, gc, alloc5, heap5),
			})
		} else if heap5 > 0.8 || heap1h > 0.8 {
			recs = append(recs, Recommendation{
				Condition: fmt.Sprintf("%s: Java heap usage elevated (5m ~%.0f%%, 1h ~%.0f%%)", path, heap5*100, heap1h*100),
				Advice:    heapElevatedAdvice(f, gc, alloc5),
			})
		}

		if alloc5 > 30 && heap5 < 0.65 && f.GCRate.Last5m > 6 {
			recs = append(recs, Recommendation{
				Condition: fmt.Sprintf("%s: high allocation (~%.0f MB/s 5m) with moderate heap (~%.0f%% 5m)", path, alloc5, heap5*100),
				Advice:    "Focus on allocation churn (object reuse, fewer short-lived allocations) before raising -Xmx. If GC/min stays high, tune young generation (collector-specific); avoid -Xmn with G1.",
			})
		} else if alloc5 > 20 && heap5 > 0.82 {
			recs = append(recs, Recommendation{
				Condition: fmt.Sprintf("%s: high allocation (~%.0f MB/s) with high heap headroom pressure (~%.0f%% 5m)", path, alloc5, heap5*100),
				Advice:    "Likely need more Java heap (-Xmx) or less live set; correlate with promotion/FGC if any. Profile retained memory and allocation hot spots.",
			})
		}

		if f.PromotionFailedCount > 0 {
			recs = append(recs, Recommendation{
				Condition: fmt.Sprintf("%s: %d Promotion Failure(s) (cumulative)", path, f.PromotionFailedCount),
				Advice:    "ParNew/CMS: old gen cannot hold promotions; increase heap or -XX:CMSInitiatingOccupancyFraction / survivor sizing per collector.",
			})
		}
		if f.ConcurrentModeFailCount > 0 {
			recs = append(recs, Recommendation{
				Condition: fmt.Sprintf("%s: %d Concurrent Mode Failure(s)", path, f.ConcurrentModeFailCount),
				Advice:    "CMS: concurrent cycle too late; lower initiating occupancy fraction or increase heap so CMS finishes before old gen fills.",
			})
		}
		if f.ToSpaceExhaustedCount > 0 {
			recs = append(recs, Recommendation{
				Condition: fmt.Sprintf("%s: %d To-Space Exhausted event(s)", path, f.ToSpaceExhaustedCount),
				Advice:    "G1: survivor/evacuation pressure; increase -XX:G1ReservePercent or overall heap; check for oversized objects in young collection.",
			})
		}
		if f.HumongousAllocationCount > 0 {
			recs = append(recs, Recommendation{
				Condition: fmt.Sprintf("%s: %d Humongous Allocation(s)", path, f.HumongousAllocationCount),
				Advice:    "G1: humongous objects; align region size (-XX:G1HeapRegionSize) with allocation pattern or reduce huge object churn.",
			})
		}
		if f.GCRate.Last5m > 8 {
			recs = append(recs, Recommendation{
				Condition: fmt.Sprintf("%s: ~%.1f GC events/min (5m avg, sustained)", path, f.GCRate.Last5m),
				Advice:    gcFrequencyAdvice(gc, f),
			})
		}
	}

	return recs
}

func pauseAdviceForFiles(files []FileSummary, stwRef float64) string {
	var gcs []string
	for _, f := range files {
		if t := strings.TrimSpace(f.GCType); t != "" {
			gcs = append(gcs, t)
		}
	}
	base := "Prioritize: (1) reduce live set / allocation; (2) collector-specific pause tuning."
	if stwRef > 0.5 {
		base = "Critical pause time. " + base
	}
	if len(gcs) == 0 {
		return base + " For G1, revisit -XX:MaxGCPauseMillis and region sizing; for ZGC/Shenandoah, verify sufficient heap and allocation rate; for Parallel/CMS, consider heap size and young/old balance."
	}
	return base + " Observed GC types: " + strings.Join(uniqueStrings(gcs), ", ") + "."
}

func throughputAdvice(files []FileSummary, global1m float64) string {
	var parts []string
	for _, f := range files {
		if f.Throughput.Last5m >= 0.9 {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s 5m=%.0f%%", shortPath(f.Path), f.Throughput.Last5m*100))
	}
	s := "GC time fraction is high vs wall clock. "
	if len(parts) > 0 {
		s += "Per file (5m): " + strings.Join(parts, "; ") + ". "
	}
	s += "Use allocation rate + heap usage cards together: if allocation is high but heap headroom is low, increase -Xmx or reduce retention; if heap is moderate but GC/min is high, reduce churn or tune young-gen behavior (collector-specific)."
	return s
}

func fullGCAdvice(gc string, f FileSummary) string {
	heap := f.HeapUsage.Last5m
	if heap == 0 {
		heap = f.HeapUsageRatio
	}
	alloc := f.AllocRate.Last5m
	s := fmt.Sprintf("Java heap after GC ~%.0f%% (5m avg of samples); allocation ~%.0f MB/s (5m). ", heap*100, alloc)
	switch {
	case strings.Contains(gc, "g1"):
		s += "G1: review IHOP (-XX:InitiatingHeapOccupancyPercent), heap size, and humongous objects; avoid manual -Xmn."
	case strings.Contains(gc, "cms"):
		s += "CMS: old-gen saturation or promotion failure risk; increase heap or lower CMS initiating occupancy."
	case strings.Contains(gc, "parallel") || strings.Contains(gc, "ps"):
		s += "Parallel: full GC often means old-gen or metaspace pressure; increase -Xmx or -XX:MaxMetaspaceSize if class metadata is growing."
	case strings.Contains(gc, "zgc"):
		s += "ZGC: rare Full GC; check allocation stall / metadata / external triggers."
	default:
		s += "Correlate with cause lines and heap chart; increase heap or reduce retained set if old gen is saturated."
	}
	return s
}

func allocationFailureAdvice(files []FileSummary) string {
	var b strings.Builder
	b.WriteString("Old generation cannot satisfy promotion or allocation. ")
	for _, f := range files {
		b.WriteString(fmt.Sprintf("%s: heap ~%.0f%% (5m), alloc ~%.0f MB/s; ", shortPath(f.Path), f.HeapUsage.Last5m*100, f.AllocRate.Last5m))
	}
	b.WriteString("Increase -Xmx if headroom is low, or reduce long-lived object volume; inspect promotion rate and old-gen usage.")
	return b.String()
}

func heapHighAdvice(f FileSummary, gc string, alloc5, heap5 float64) string {
	s := fmt.Sprintf("Heap headroom low (5m ~%.0f%%). Allocation ~%.0f MB/s (5m). ", heap5*100, alloc5)
	if alloc5 > 25 {
		s += "High allocation with high occupancy: prioritize larger -Xmx or lower allocation/retention before micro-tuning."
	} else {
		s += "Lower allocation pressure or memory leak investigation (heap dump, dominator tree) if occupancy climbs without traffic growth."
	}
	if strings.Contains(gc, "g1") {
		s += " G1: check IHOP; avoid -Xmn."
	}
	return s
}

func heapElevatedAdvice(f FileSummary, gc string, alloc5 float64) string {
	return fmt.Sprintf("Monitor trend; allocation ~%.0f MB/s (5m). If climbing toward 90%%, plan capacity. %s", alloc5, collectorHeapHint(gc))
}

func collectorHeapHint(gc string) string {
	switch {
	case strings.Contains(gc, "g1"):
		return "G1: watch concurrent cycle start and old-gen regions."
	case strings.Contains(gc, "cms"):
		return "CMS: watch concurrent mode failure counters and old-gen occupancy."
	case strings.Contains(gc, "zgc"):
		return "ZGC: single-heap view; high usage increases GC pressure."
	default:
		return "Parallel/Serial: correlate Full GC causes with heap used vs total and Metaspace lines."
	}
}

func gcFrequencyAdvice(gc string, f FileSummary) string {
	alloc := f.AllocRate.Last5m
	heap := f.HeapUsage.Last5m
	s := fmt.Sprintf("Sustained GC churn (5m). Allocation ~%.0f MB/s, heap usage ~%.0f%% (5m). ", alloc, heap*100)
	if strings.Contains(gc, "g1") {
		s += "G1: do not set -Xmn; prefer -XX:MaxGCPauseMillis and region sizing; reduce allocation if heap is healthy but GC count is high."
	} else if strings.Contains(gc, "zgc") {
		s += "ZGC: frequent cycles often track allocation rate; ensure heap is large enough for allocation spikes."
	} else {
		s += "Consider larger young generation or lower allocation rate; verify old-gen is not driving full collections."
	}
	return s
}

func uniqueStrings(in []string) []string {
	seen := make(map[string]struct{})
	var out []string
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}
