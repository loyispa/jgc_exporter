package parser

import (
	"regexp"
	"strings"
	"time"
)

type G1Parser struct {
	handler      EventHandler
	unified      bool
	currentEvent *GCEvent
}

func NewG1Parser(unified bool) *G1Parser {
	return &G1Parser{unified: unified}
}

func (p *G1Parser) GCType() GCType            { return GCTypeG1 }
func (p *G1Parser) SetHandler(h EventHandler) { p.handler = h }

func (p *G1Parser) Feed(line string) {
	if p.handler == nil {
		return
	}
	if p.unified {
		p.feedUnified(line)
	} else {
		p.feedPreUnified(line)
	}
}

// ===================== PreUnified G1 (JDK8) =====================

// GC pause: [GC pause (cause) (phase), dur secs]
// Supports multiple optional parenthesized groups:
//
//	(G1 Evacuation Pause) (young)
//	(G1 Evacuation Pause) (young) (initial-mark)
//	(G1 Evacuation Pause) (mixed)
//	(G1 Evacuation Pause) (mixed) (to-space exhausted)
//	(G1 Humongous Allocation) (young) (to-space exhausted)
//	(Metadata GC Threshold) (young) (initial-mark)
var g1PrePauseRe = regexp.MustCompile(
	`\[GC pause\s+\(([^)]+)\)\s+\(([^)]+)\)(?:\s+\(([^)]+)\))?,\s*([\d.]+)\s*secs\]`,
)

// Full GC: [Full GC (cause) heap->heap(total), dur secs]
var g1PreFullRe = regexp.MustCompile(
	`\[Full GC\s*\((.+)\)\s+([\d.]+[KMGB]?)->([\d.]+[KMGB]?)\(([\d.]+[KMGB]?)\),\s*([\d.]+)\s*secs\]`,
)

// Concurrent phases (end): [GC concurrent-X-end, dur secs]
var g1PreConcEndRe = regexp.MustCompile(
	`\[GC (concurrent-[\w-]+)-end,\s*([\d.]+)\s*secs\]`,
)

// Concurrent phases (start): [GC concurrent-X-start]
var g1PreConcStartRe = regexp.MustCompile(
	`\[GC (concurrent-[\w-]+)-start\]`,
)

// GC remark: [GC remark ... X secs]
var g1PreRemarkRe = regexp.MustCompile(
	`\[GC remark.*,\s*([\d.]+)\s*secs\]`,
)

// GC cleanup: [GC cleanup XK->XK(XK), X secs]
var g1PreCleanupRe = regexp.MustCompile(
	`\[GC cleanup\s+([\d.]+[KMGB]?)->([\d.]+[KMGB]?)\(([\d.]+[KMGB]?)\),\s*([\d.]+)\s*secs\]`,
)

// Eden/Survivor/Heap line
var g1PreEdenRe = regexp.MustCompile(
	`\[Eden:\s*([\d.]+[KMGB]?)\(([\d.]+[KMGB]?)\)->([\d.]+[KMGB]?)\(([\d.]+[KMGB]?)\)\s+Survivors:\s*([\d.]+[KMGB]?)->([\d.]+[KMGB]?)\s+Heap:\s*([\d.]+[KMGB]?)\(([\d.]+[KMGB]?)\)->([\d.]+[KMGB]?)\(([\d.]+[KMGB]?)\)\]`,
)

// Metaspace line
var g1PreMetaRe = regexp.MustCompile(
	`\[Metaspace:\s*([\d]+[KMGB]?)->([\d]+[KMGB]?)\(([\d]+[KMGB]?)\)\]`,
)

// Finalize Marking / GC ref-proc / Unloading inside remark
var g1PreFinalizeMarkRe = regexp.MustCompile(`\[Finalize Marking,\s*([\d.]+)\s*secs\]`)
var g1PreRefProcRe = regexp.MustCompile(`\[GC ref-proc,\s*([\d.]+)\s*secs\]`)
var g1PreUnloadingRe = regexp.MustCompile(`\[Unloading,\s*([\d.]+)\s*secs\]`)

func (p *G1Parser) Flush() { p.emitPending() }

func (p *G1Parser) emitPending() {
	if p.currentEvent != nil {
		p.handler.Handle(p.currentEvent)
		p.currentEvent = nil
	}
}

func (p *G1Parser) feedPreUnified(line string) {
	ts, hasTs := parsePreUnifiedTimestamp(line)

	// GC pause events (young, mixed, initial-mark, to-space exhausted)
	if m := g1PrePauseRe.FindStringSubmatch(line); len(m) > 0 {
		p.emitPending()
		cause := m[1]
		phase1 := m[2]
		phase2 := m[3]
		dur := parseFloat(m[4])

		ev := &GCEvent{
			GCType:   GCTypeG1,
			IsPause:  true,
			Cause:    cause,
			Duration: dur,
		}
		if hasTs {
			ev.Timestamp = ts
		}
		ev.Category = classifyG1PrePause(cause, phase1, phase2)
		p.currentEvent = ev
		return
	}

	// Full GC
	if m := g1PreFullRe.FindStringSubmatch(line); len(m) > 0 {
		p.emitPending()
		ev := &GCEvent{
			GCType:       GCTypeG1,
			IsPause:      true,
			Cause:        m[1],
			HeapBeforeKB: parseSizeKB(m[2]),
			HeapAfterKB:  parseSizeKB(m[3]),
			HeapTotalKB:  parseSizeKB(m[4]),
			Duration:     parseFloat(m[5]),
		}
		if hasTs {
			ev.Timestamp = ts
		}
		ev.Category = classifyG1FullCause(ev.Cause)
		p.currentEvent = ev
		return
	}

	// GC remark
	if m := g1PreRemarkRe.FindStringSubmatch(line); len(m) > 0 {
		p.emitPending()
		ev := &GCEvent{
			GCType:   GCTypeG1,
			IsPause:  true,
			Category: "G1Remark",
			Duration: parseFloat(m[1]),
		}
		if hasTs {
			ev.Timestamp = ts
		}
		p.currentEvent = ev
		return
	}

	// GC cleanup
	if m := g1PreCleanupRe.FindStringSubmatch(line); len(m) > 0 {
		p.emitPending()
		ev := &GCEvent{
			GCType:       GCTypeG1,
			IsPause:      true,
			Category:     "G1Cleanup",
			HeapBeforeKB: parseSizeKB(m[1]),
			HeapAfterKB:  parseSizeKB(m[2]),
			HeapTotalKB:  parseSizeKB(m[3]),
			Duration:     parseFloat(m[4]),
		}
		if hasTs {
			ev.Timestamp = ts
		}
		p.currentEvent = ev
		return
	}

	// Concurrent phase end: [GC concurrent-mark-end, 0.032 secs]
	if m := g1PreConcEndRe.FindStringSubmatch(line); len(m) > 0 {
		phase := m[1]
		ev := &GCEvent{
			GCType:   GCTypeG1,
			IsPause:  false,
			Category: classifyG1PreConcurrent(phase),
			Duration: parseFloat(m[2]),
		}
		if hasTs {
			ev.Timestamp = ts
		}
		p.handler.Handle(ev)
		return
	}

	// Concurrent phase start: skip (no duration yet)
	if g1PreConcStartRe.MatchString(line) {
		return
	}

	// Detail lines for current pause event
	if p.currentEvent != nil {
		if m := g1PreEdenRe.FindStringSubmatch(line); len(m) > 0 {
			ev := p.currentEvent
			ev.G1EdenBeforeKB = parseSizeKB(m[1])
			ev.G1EdenTotalKB = parseSizeKB(m[2])
			ev.G1EdenAfterKB = parseSizeKB(m[3])
			ev.G1SurvivorBeforeKB = parseSizeKB(m[5])
			ev.G1SurvivorAfterKB = parseSizeKB(m[6])
			ev.HeapBeforeKB = parseSizeKB(m[7])
			ev.HeapAfterKB = parseSizeKB(m[9])
			ev.YoungBeforeKB = ev.G1EdenBeforeKB + ev.G1SurvivorBeforeKB
			ev.YoungAfterKB = ev.G1EdenAfterKB + ev.G1SurvivorAfterKB
			// m[10] is heap capacity after GC; prefer it over m[8] (before)
			// since G1 can resize the heap during collection
			if afterCap := parseSizeKB(m[10]); afterCap > 0 {
				ev.HeapTotalKB = afterCap
			} else {
				ev.HeapTotalKB = parseSizeKB(m[8])
			}
		}
		if m := g1PreMetaRe.FindStringSubmatch(line); len(m) > 0 {
			p.currentEvent.MetaBeforeKB = parseSizeKB(m[1])
			p.currentEvent.MetaAfterKB = parseSizeKB(m[2])
			p.currentEvent.MetaTotalKB = parseSizeKB(m[3])
		}
		if m := refSoftRe.FindStringSubmatch(line); len(m) > 0 {
			p.currentEvent.SoftRefCount = parseInt(m[1])
			p.currentEvent.SoftRefPauseMs = parseFloat(m[2]) * 1000.0
			return
		}
		if m := refWeakRe.FindStringSubmatch(line); len(m) > 0 {
			p.currentEvent.WeakRefCount = parseInt(m[1])
			p.currentEvent.WeakRefPauseMs = parseFloat(m[2]) * 1000.0
			return
		}
		if m := refFinalRe.FindStringSubmatch(line); len(m) > 0 {
			p.currentEvent.FinalRefCount = parseInt(m[1])
			p.currentEvent.FinalRefPauseMs = parseFloat(m[2]) * 1000.0
			return
		}
		if m := refPhantomRe.FindStringSubmatch(line); len(m) > 0 {
			p.currentEvent.PhantomRefCount = parseInt(m[1])
			if len(m) > 3 && m[2] != "" {
				p.currentEvent.PhantomRefFree = parseInt(m[2])
			}
			p.currentEvent.PhantomRefPauseMs = parseFloat(m[len(m)-1]) * 1000.0
			return
		}
		if m := refJNIWeakRe.FindStringSubmatch(line); len(m) > 0 {
			p.currentEvent.JNIWeakRefPauseMs = parseFloat(m[1]) * 1000.0
			return
		}
		if strings.Contains(line, "[Times:") {
			p.handler.Handle(p.currentEvent)
			p.currentEvent = nil
		}
	}
}

func classifyG1PrePause(cause, phase1, phase2 string) string {
	allPhases := strings.ToLower(phase1 + " " + phase2)
	causeLower := strings.ToLower(cause)

	if strings.Contains(allPhases, "to-space exhausted") || strings.Contains(allPhases, "to-space overflow") {
		return "G1ToSpaceExhausted"
	}
	// Check cause-specific categories before generic phase categories
	if strings.Contains(causeLower, "humongous") {
		return "G1HumongousAllocation"
	}
	if strings.Contains(causeLower, "metadata gc threshold") {
		return "G1MetadataGC"
	}
	if strings.Contains(allPhases, "initial-mark") {
		return "G1InitialMark"
	}
	if strings.Contains(allPhases, "mixed") {
		return "G1MixedGC"
	}
	return "G1YoungGC"
}

func classifyG1FullCause(cause string) string {
	c := strings.ToLower(cause)
	if strings.Contains(c, "system") {
		return "G1SystemGC"
	}
	if strings.Contains(c, "allocation failure") {
		return "G1FullGC"
	}
	if strings.Contains(c, "metadata") {
		return "G1MetadataGC"
	}
	return "G1FullGC"
}

func classifyG1PreConcurrent(phase string) string {
	p := strings.ToLower(phase)
	switch {
	case strings.Contains(p, "root-region-scan"):
		return "G1ConcurrentRootRegionScan"
	case strings.Contains(p, "mark"):
		return "G1ConcurrentMark"
	case strings.Contains(p, "cleanup"):
		return "G1ConcurrentCleanup"
	default:
		return "G1Concurrent"
	}
}

// ===================== Unified G1 (JDK9+) =====================

// Pause with heap summary: GC(N) Pause ... XM->YM(ZM) Dms
var unifiedGCPauseHeapRe = regexp.MustCompile(
	`GC\((\d+)\)\s+Pause\s+(.+?)\s+(\d+[KMGB])->(\d+[KMGB])\((\d+[KMGB])\)\s+([\d.]+)ms`)

// Pause without heap: GC(N) Pause ... Dms (fallback)
var unifiedGCPauseRe = regexp.MustCompile(`GC\((\d+)\)\s+Pause\s+(.+?)\s+([\d.]+)ms`)
var unifiedGCConcRe = regexp.MustCompile(`GC\((\d+)\)\s+Concurrent\s+(.+?)\s+([\d.]+)ms`)
var unifiedHeapRe = regexp.MustCompile(`GC\(\d+\)\s+Heap\s+.*?(\d+[KMGB])\((\d+[KMGB])\)->(\d+[KMGB])\((\d+[KMGB])\)`)
var unifiedEdenRe = regexp.MustCompile(`GC\(\d+\)\s+Eden regions:\s*(\d+)->(\d+)\((\d+)\)`)
var unifiedSurvivorRe = regexp.MustCompile(`GC\(\d+\)\s+Survivor regions:\s*(\d+)->(\d+)\((\d+)\)`)
var unifiedOldRe = regexp.MustCompile(`GC\(\d+\)\s+Old regions:\s*(\d+)->(\d+)`)
var unifiedHumongousRe = regexp.MustCompile(`GC\(\d+\)\s+Humongous regions:\s*(\d+)->(\d+)`)
var unifiedArchiveRe = regexp.MustCompile(`GC\(\d+\)\s+Archive regions:\s*(\d+)->(\d+)\((\d+)\)`)
var unifiedMetaRe = regexp.MustCompile(`GC\(\d+\)\s+Metaspace:\s*(\d+[KMGB]?).*->(\d+[KMGB]?).*\((\d+[KMGB]?)\)`)

func (p *G1Parser) feedUnified(line string) {
	if !strings.Contains(line, "][gc") {
		return
	}

	ts, _ := parseUnifiedTimestamp(line)

	// Pause with heap summary: GC(N) Pause ... XM->YM(ZM) Dms
	if m := unifiedGCPauseHeapRe.FindStringSubmatch(line); len(m) > 0 {
		if p.currentEvent != nil {
			p.handler.Handle(p.currentEvent)
		}
		ev := &GCEvent{
			GCType:       GCTypeG1,
			Timestamp:    ts,
			IsPause:      true,
			Category:     classifyUnifiedG1Pause(m[2]),
			HeapBeforeKB: parseSizeKB(m[3]),
			HeapAfterKB:  parseSizeKB(m[4]),
			HeapTotalKB:  parseSizeKB(m[5]),
			Duration:     parseFloat(m[6]) / 1000.0,
		}
		p.currentEvent = ev
		return
	}

	// Pause without heap (fallback)
	if m := unifiedGCPauseRe.FindStringSubmatch(line); len(m) > 0 {
		if p.currentEvent != nil {
			p.handler.Handle(p.currentEvent)
		}
		ev := &GCEvent{
			GCType:    GCTypeG1,
			Timestamp: ts,
			IsPause:   true,
			Category:  classifyUnifiedG1Pause(m[2]),
			Duration:  parseFloat(m[3]) / 1000.0,
		}
		p.currentEvent = ev
		return
	}

	// Concurrent phases with duration
	if m := unifiedGCConcRe.FindStringSubmatch(line); len(m) > 0 {
		ev := &GCEvent{
			GCType:    GCTypeG1,
			Timestamp: ts,
			IsPause:   false,
			Category:  "G1Concurrent" + sanitizeCategory(m[2]),
			Duration:  parseFloat(m[3]) / 1000.0,
		}
		p.handler.Handle(ev)
		return
	}

	// Detail lines for current pause
	if p.currentEvent != nil {
		if m := unifiedHeapRe.FindStringSubmatch(line); len(m) > 0 {
			p.currentEvent.HeapBeforeKB = parseSizeKB(m[1])
			p.currentEvent.HeapTotalKB = parseSizeKB(m[2])
			p.currentEvent.HeapAfterKB = parseSizeKB(m[3])
		}
		if m := unifiedMetaRe.FindStringSubmatch(line); len(m) > 0 {
			p.currentEvent.MetaBeforeKB = parseSizeKB(m[1])
			p.currentEvent.MetaAfterKB = parseSizeKB(m[2])
			p.currentEvent.MetaTotalKB = parseSizeKB(m[3])
		}
		if m := unifiedEdenRe.FindStringSubmatch(line); len(m) > 0 {
			p.currentEvent.G1EdenRegionBefore = parseInt(m[1])
			p.currentEvent.G1EdenRegionAfter = parseInt(m[2])
			p.currentEvent.G1EdenRegionAssign = parseInt(m[3])
		}
		if m := unifiedSurvivorRe.FindStringSubmatch(line); len(m) > 0 {
			p.currentEvent.G1SurvivorRegionBefore = parseInt(m[1])
			p.currentEvent.G1SurvivorRegionAfter = parseInt(m[2])
			p.currentEvent.G1SurvivorRegionAssign = parseInt(m[3])
		}
		if m := unifiedOldRe.FindStringSubmatch(line); len(m) > 0 {
			p.currentEvent.G1OldRegionBefore = parseInt(m[1])
			p.currentEvent.G1OldRegionAfter = parseInt(m[2])
		}
		if m := unifiedHumongousRe.FindStringSubmatch(line); len(m) > 0 {
			p.currentEvent.G1HumongousRegionBefore = parseInt(m[1])
			p.currentEvent.G1HumongousRegionAfter = parseInt(m[2])
		}
		if m := unifiedArchiveRe.FindStringSubmatch(line); len(m) > 0 {
			p.currentEvent.G1ArchiveRegionBefore = parseInt(m[1])
			p.currentEvent.G1ArchiveRegionAfter = parseInt(m[2])
			p.currentEvent.G1ArchiveRegionAssign = parseInt(m[3])
		}
	}
}

func classifyUnifiedG1Pause(desc string) string {
	d := strings.ToLower(desc)
	if strings.Contains(d, "full") {
		if strings.Contains(d, "system") {
			return "G1SystemGC"
		}
		if strings.Contains(d, "compaction") {
			return "G1FullGC"
		}
		return "G1FullGC"
	}
	if strings.Contains(d, "cleanup") {
		return "G1Cleanup"
	}
	if strings.Contains(d, "remark") {
		return "G1Remark"
	}
	if strings.Contains(d, "to-space exhausted") {
		return "G1ToSpaceExhausted"
	}
	if strings.Contains(d, "concurrent start") {
		return "G1ConcurrentStart"
	}
	if strings.Contains(d, "prepare mixed") {
		return "G1PrepareMixed"
	}
	if strings.Contains(d, "mixed") {
		return "G1MixedGC"
	}
	if strings.Contains(d, "young") {
		return "G1YoungGC"
	}
	return "G1Unknown"
}

func sanitizeCategory(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, " ", "")
	return s
}

var _ Parser = (*G1Parser)(nil)
var _ = time.Now
