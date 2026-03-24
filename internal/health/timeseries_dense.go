package health

import (
	"sort"
	"time"
)

// sortedFilePaths returns lexicographically sorted paths from fileSummary (stable UI / chart series order).
func sortedFilePaths(paths map[string]*fileStat) []string {
	out := make([]string, 0, len(paths))
	for p := range paths {
		out = append(out, p)
	}
	sort.Strings(out)
	return out
}

func utcMinute(t time.Time) time.Time {
	return t.UTC().Truncate(time.Minute)
}

// denseMinuteRange returns inclusive UTC-aligned minute timestamps from start through end.
func denseMinuteRange(start, end time.Time) []time.Time {
	if end.Before(start) {
		start, end = end, start
	}
	start = utcMinute(start)
	end = utcMinute(end)
	var ts []time.Time
	for t := start; !t.After(end); t = t.Add(time.Minute) {
		ts = append(ts, t)
	}
	if len(ts) == 0 {
		return []time.Time{end}
	}
	return ts
}

// densifyGCEventsBuckets emits one bucket per minute in [start, end] with merged counts (empty map => zero).
func densifyGCEventsBuckets(bucketCounts map[time.Time]map[string]int, start, end time.Time) []GCEventsBucket {
	var out []GCEventsBucket
	for _, t := range denseMinuteRange(start, end) {
		src := bucketCounts[t]
		if len(src) == 0 {
			out = append(out, GCEventsBucket{Timestamp: t, Counts: map[string]int{}})
			continue
		}
		cp := make(map[string]int, len(src))
		for k, v := range src {
			cp[k] = v
		}
		out = append(out, GCEventsBucket{Timestamp: t, Counts: cp})
	}
	return out
}

// densifyGCDurationBuckets emits one bucket per minute with per-category summed duration (seconds).
func densifyGCDurationBuckets(sums map[time.Time]map[string]float64, start, end time.Time) []GCDurationBucket {
	var out []GCDurationBucket
	for _, t := range denseMinuteRange(start, end) {
		src := sums[t]
		if len(src) == 0 {
			out = append(out, GCDurationBucket{Timestamp: t, Seconds: map[string]float64{}})
			continue
		}
		cp := make(map[string]float64, len(src))
		for k, v := range src {
			cp[k] = v
		}
		out = append(out, GCDurationBucket{Timestamp: t, Seconds: cp})
	}
	return out
}

type pathMinuteKey struct {
	path   string
	minute time.Time
}

// densifyThroughputHistory emits one ThroughputPoint per (path, UTC minute). Ratio = 1 - pauseSum/wallSeconds, clamped [0,1].
// Wall time is 60s for completed minutes; for refMin (current minute) uses now.Sub(refMin).
func densifyThroughputHistory(pauseSumSec map[pathMinuteKey]float64, paths []string, start, end, refMin, now time.Time) []ThroughputPoint {
	curEl := now.Sub(refMin).Seconds()
	if curEl < 1 {
		curEl = 1
	}
	refMin = utcMinute(refMin)
	var out []ThroughputPoint
	for _, path := range paths {
		for _, t := range denseMinuteRange(start, end) {
			sum := pauseSumSec[pathMinuteKey{path, t}]
			denom := 60.0
			if t.Equal(refMin) {
				denom = curEl
			}
			r := 1.0 - sum/denom
			if r < 0 {
				r = 0
			}
			if r > 1 {
				r = 1
			}
			out = append(out, ThroughputPoint{Timestamp: t, Path: path, Ratio: r})
		}
	}
	return out
}

// densifyAllocationRateHistory returns one point per (path, minute); value is the last sample in that minute (0 if none).
func densifyAllocationRateHistory(sparse []AllocationRatePoint, paths []string, start, end time.Time) []AllocationRatePoint {
	lastInMin := make(map[pathMinuteKey]float64)
	for _, p := range sparse {
		if p.Path == "" {
			continue
		}
		mt := utcMinute(p.Timestamp)
		lastInMin[pathMinuteKey{p.Path, mt}] = p.RateMBPerSec
	}
	var out []AllocationRatePoint
	for _, path := range paths {
		for _, t := range denseMinuteRange(start, end) {
			v := lastInMin[pathMinuteKey{path, t}]
			out = append(out, AllocationRatePoint{
				Timestamp:    t,
				Path:         path,
				RateMBPerSec: v,
			})
		}
	}
	return out
}

// densifyHeapHistory returns one snapshot per (path, minute). Last sample within each minute updates carry;
// minutes without a new sample repeat the previous non-zero state (LOCF). Minutes before the first sample are zeros.
func densifyHeapHistory(sparse []HeapSnapshot, paths []string, start, end time.Time) []HeapSnapshot {
	byPath := make(map[string][]HeapSnapshot)
	for _, h := range sparse {
		byPath[h.Path] = append(byPath[h.Path], h)
	}
	for path := range byPath {
		sort.Slice(byPath[path], func(i, j int) bool {
			return byPath[path][i].Timestamp.Before(byPath[path][j].Timestamp)
		})
	}

	var out []HeapSnapshot
	for _, path := range paths {
		snaps := byPath[path]
		idx := 0
		var carry HeapSnapshot
		haveCarry := false

		for _, t := range denseMinuteRange(start, end) {
			nextEnd := t.Add(time.Minute)
			for idx < len(snaps) && !snaps[idx].Timestamp.Before(t) && snaps[idx].Timestamp.Before(nextEnd) {
				carry = snaps[idx]
				haveCarry = true
				idx++
			}
			if !haveCarry {
				out = append(out, HeapSnapshot{Timestamp: t, Path: path})
				continue
			}
			out = append(out, HeapSnapshot{
				Timestamp:   t,
				Path:        path,
				HeapUsedKB:  carry.HeapUsedKB,
				HeapTotalKB: carry.HeapTotalKB,
				MetaUsedKB:  carry.MetaUsedKB,
			})
		}
	}
	return out
}
