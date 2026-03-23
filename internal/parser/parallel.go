package parser

import (
	"regexp"
	"strings"
)

type ParallelParser struct {
	handler      EventHandler
	unified      bool
	currentEvent *GCEvent
	pendingBuf   pendingDetail
}

func NewParallelParser(unified bool) *ParallelParser {
	return &ParallelParser{unified: unified}
}

func (p *ParallelParser) GCType() GCType            { return GCTypeParallel }
func (p *ParallelParser) SetHandler(h EventHandler) { p.handler = h }
func (p *ParallelParser) Flush()                    { p.emitPending() }

func (p *ParallelParser) Feed(line string) {
	if p.handler == nil {
		return
	}
	if p.unified {
		p.feedUnified(line)
	} else {
		p.feedPreUnified(line)
	}
}

// ===================== PreUnified Parallel (JDK8) =====================

// Young GC: [GC (Cause) [PSYoungGen: before->after(total)] heap->heap(total), dur secs]
var psYoungRe = regexp.MustCompile(
	`\[GC\s*(?:\(([^)]*)\)\s*)?\[PSYoungGen:\s*([\d.]+[KMGB]?)->([\d.]+[KMGB]?)\(([\d.]+[KMGB]?)\)\]\s*([\d.]+[KMGB]?)->([\d.]+[KMGB]?)\(([\d.]+[KMGB]?)\),\s*([\d.]+)\s*secs\]`,
)

// Full GC with PSYoungGen + ParOldGen:
// [Full GC (Cause) [PSYoungGen: y->y(yt)] [ParOldGen: o->o(ot)] heap->heap(total) [PSPermGen/Metaspace: m->m(mt)], dur secs]
var psFullRe = regexp.MustCompile(
	`\[Full GC\s*(?:\((.+?)\)\s+)?\[PSYoungGen:\s*[\d.]+[KMGB]?->[\d.]+[KMGB]?\([\d.]+[KMGB]?\)\]\s*\[(ParOldGen|PSOldGen):\s*([\d.]+[KMGB]?)->([\d.]+[KMGB]?)\(([\d.]+[KMGB]?)\)\]\s*([\d.]+[KMGB]?)->([\d.]+[KMGB]?)\(([\d.]+[KMGB]?)\).*?,\s*([\d.]+)\s*secs\]`,
)

// Simple Full GC without generation detail: [Full GC (Cause) heap->heap(total), dur secs]
var psSimpleFullRe = regexp.MustCompile(
	`\[Full GC\s*(?:\((.+?)\)\s+)?([\d.]+[KMGB]?)->([\d.]+[KMGB]?)\(([\d.]+[KMGB]?)\),\s*([\d.]+)\s*secs\]`,
)

var psPermGenRe = regexp.MustCompile(
	`\[PSPermGen:\s*([\d.]+[KMGB]?)->([\d.]+[KMGB]?)\(([\d.]+[KMGB]?)\)\]`,
)

var psMetaspaceRe = regexp.MustCompile(
	`\[Metaspace:\s*([\d]+[KMGB]?)->([\d]+[KMGB]?)\(([\d]+[KMGB]?)\)\]`,
)

func (p *ParallelParser) emitPending() {
	if p.currentEvent != nil {
		p.handler.Handle(p.currentEvent)
		p.currentEvent = nil
	}
}

func (p *ParallelParser) feedPreUnified(line string) {
	ts, hasTs := parsePreUnifiedTimestamp(line)

	// Full GC with generation detail
	if m := psFullRe.FindStringSubmatch(line); len(m) > 0 {
		p.emitPending()
		ev := &GCEvent{
			GCType:       GCTypeParallel,
			IsPause:      true,
			Cause:        m[1],
			OldBeforeKB:  parseSizeKB(m[3]),
			OldAfterKB:   parseSizeKB(m[4]),
			OldTotalKB:   parseSizeKB(m[5]),
			HeapBeforeKB: parseSizeKB(m[6]),
			HeapAfterKB:  parseSizeKB(m[7]),
			HeapTotalKB:  parseSizeKB(m[8]),
			Duration:     parseFloat(m[9]),
		}
		ev.Category = classifyParallelFullCause(ev.Cause)
		if hasTs {
			ev.Timestamp = ts
		}
		if pm := psPermGenRe.FindStringSubmatch(line); len(pm) > 0 {
			ev.MetaBeforeKB = parseSizeKB(pm[1])
			ev.MetaAfterKB = parseSizeKB(pm[2])
			ev.MetaTotalKB = parseSizeKB(pm[3])
		}
		if pm := psMetaspaceRe.FindStringSubmatch(line); len(pm) > 0 {
			ev.MetaBeforeKB = parseSizeKB(pm[1])
			ev.MetaAfterKB = parseSizeKB(pm[2])
			ev.MetaTotalKB = parseSizeKB(pm[3])
		}
		p.currentEvent = ev
		return
	}

	// Young GC
	if m := psYoungRe.FindStringSubmatch(line); len(m) > 0 {
		p.emitPending()
		ev := &GCEvent{
			GCType:        GCTypeParallel,
			IsPause:       true,
			Cause:         m[1],
			YoungBeforeKB: parseSizeKB(m[2]),
			YoungAfterKB:  parseSizeKB(m[3]),
			YoungTotalKB:  parseSizeKB(m[4]),
			HeapBeforeKB:  parseSizeKB(m[5]),
			HeapAfterKB:   parseSizeKB(m[6]),
			HeapTotalKB:   parseSizeKB(m[7]),
			Duration:      parseFloat(m[8]),
		}
		ev.Category = classifyParallelYoungCause(ev.Cause)
		if hasTs {
			ev.Timestamp = ts
		}
		p.currentEvent = ev
		return
	}

	// Simple Full GC (no generation detail)
	if m := psSimpleFullRe.FindStringSubmatch(line); len(m) > 0 {
		p.emitPending()
		ev := &GCEvent{
			GCType:       GCTypeParallel,
			IsPause:      true,
			Cause:        m[1],
			HeapBeforeKB: parseSizeKB(m[2]),
			HeapAfterKB:  parseSizeKB(m[3]),
			HeapTotalKB:  parseSizeKB(m[4]),
			Duration:     parseFloat(m[5]),
		}
		ev.Category = classifyParallelFullCause(ev.Cause)
		if hasTs {
			ev.Timestamp = ts
		}
		p.currentEvent = ev
		return
	}

	// [Times:] terminates the current event
	if p.currentEvent != nil {
		if strings.Contains(line, "[Times:") {
			p.handler.Handle(p.currentEvent)
			p.currentEvent = nil
		}
	}
}

func classifyParallelYoungCause(cause string) string {
	c := strings.ToLower(cause)
	if strings.Contains(c, "system") {
		return "ParallelSystemGC"
	}
	if strings.Contains(c, "metadata") {
		return "ParallelMetadataGC"
	}
	if strings.Contains(c, "ergonomics") {
		return "ParallelErgonomics"
	}
	if strings.Contains(c, "gclocker") {
		return "ParallelGCLocker"
	}
	return "ParallelYoungGC"
}

func classifyParallelFullCause(cause string) string {
	c := strings.ToLower(cause)
	if strings.Contains(c, "system") {
		return "ParallelSystemGC"
	}
	if strings.Contains(c, "metadata") {
		return "ParallelMetadataGC"
	}
	if strings.Contains(c, "ergonomics") {
		return "ParallelFullGCErgonomics"
	}
	if strings.Contains(c, "allocation") {
		return "ParallelFullGCAllocationFailure"
	}
	return "ParallelFullGC"
}

// ===================== Unified Parallel (JDK9+) =====================

// PSYoungGen detail: GC(N) PSYoungGen: XK(YK)->ZK(WK) Eden: ...
var unifiedPSYoungRe = regexp.MustCompile(
	`GC\((\d+)\)\s+PSYoungGen:\s*([\d]+[KMGB]?)\(([\d]+[KMGB]?)\)->([\d]+[KMGB]?)\(([\d]+[KMGB]?)\)`)

// ParOldGen / PSOldGen detail (JDK may label old gen as PSOldGen in unified logs)
var unifiedParOldRe = regexp.MustCompile(
	`GC\((\d+)\)\s+(?:ParOldGen|PSOldGen):\s*([\d]+[KMGB]?)\(([\d]+[KMGB]?)\)->([\d]+[KMGB]?)\(([\d]+[KMGB]?)\)`)

// Metaspace detail with GC number
var unifiedMetaWithGCRe = regexp.MustCompile(
	`GC\((\d+)\)\s+Metaspace:\s*(\d+[KMGB]?).*?->(\d+[KMGB]?).*?\((\d+[KMGB]?)\)`)

// pendingDetail buffers heap detail lines that arrive before the summary line
type pendingDetail struct {
	gcNum                                     string
	youngBeforeKB, youngAfterKB, youngTotalKB int64
	oldBeforeKB, oldAfterKB, oldTotalKB       int64
	metaBeforeKB, metaAfterKB, metaTotalKB    int64
}

func (p *ParallelParser) feedUnified(line string) {
	if !strings.Contains(line, "][gc") {
		return
	}
	ts, _ := parseUnifiedTimestamp(line)

	// Buffer detail lines (arrive before summary)
	if m := unifiedPSYoungRe.FindStringSubmatch(line); len(m) > 0 {
		p.pendingBuf.gcNum = m[1]
		p.pendingBuf.youngBeforeKB = parseSizeKB(m[2])
		p.pendingBuf.youngTotalKB = parseSizeKB(m[3])
		p.pendingBuf.youngAfterKB = parseSizeKB(m[4])
		return
	}
	if m := unifiedParOldRe.FindStringSubmatch(line); len(m) > 0 {
		p.pendingBuf.gcNum = m[1]
		p.pendingBuf.oldBeforeKB = parseSizeKB(m[2])
		p.pendingBuf.oldTotalKB = parseSizeKB(m[3])
		p.pendingBuf.oldAfterKB = parseSizeKB(m[4])
		return
	}
	if m := unifiedMetaWithGCRe.FindStringSubmatch(line); len(m) > 0 {
		p.pendingBuf.gcNum = m[1]
		p.pendingBuf.metaBeforeKB = parseSizeKB(m[2])
		p.pendingBuf.metaAfterKB = parseSizeKB(m[3])
		p.pendingBuf.metaTotalKB = parseSizeKB(m[4])
		return
	}

	// Pause with heap summary: GC(N) Pause Young/Full ... XM->YM(ZM) Dms
	if m := unifiedGCPauseHeapRe.FindStringSubmatch(line); len(m) > 0 {
		p.emitPending()
		ev := &GCEvent{
			GCType:       GCTypeParallel,
			Timestamp:    ts,
			IsPause:      true,
			Category:     classifyUnifiedParallelPause(m[2]),
			HeapBeforeKB: parseSizeKB(m[3]),
			HeapAfterKB:  parseSizeKB(m[4]),
			HeapTotalKB:  parseSizeKB(m[5]),
			Duration:     parseFloat(m[6]) / 1000.0,
		}
		if cm := unifiedPauseCauseRe.FindStringSubmatch(line); len(cm) > 0 {
			ev.Cause = cm[1]
		}
		if p.pendingBuf.gcNum == m[1] {
			p.applyPendingDetail(ev)
		}
		p.currentEvent = ev
		return
	}

	// Pause without heap (fallback)
	if m := unifiedGCPauseRe.FindStringSubmatch(line); len(m) > 0 {
		p.emitPending()
		ev := &GCEvent{
			GCType:    GCTypeParallel,
			Timestamp: ts,
			IsPause:   true,
			Category:  classifyUnifiedParallelPause(m[2]),
			Duration:  parseFloat(m[3]) / 1000.0,
		}
		if cm := unifiedPauseCauseRe.FindStringSubmatch(line); len(cm) > 0 {
			ev.Cause = cm[1]
		}
		if p.pendingBuf.gcNum == m[1] {
			p.applyPendingDetail(ev)
		}
		p.currentEvent = ev
		return
	}
}

func (p *ParallelParser) applyPendingDetail(ev *GCEvent) {
	ev.YoungBeforeKB = p.pendingBuf.youngBeforeKB
	ev.YoungAfterKB = p.pendingBuf.youngAfterKB
	ev.YoungTotalKB = p.pendingBuf.youngTotalKB
	ev.OldBeforeKB = p.pendingBuf.oldBeforeKB
	ev.OldAfterKB = p.pendingBuf.oldAfterKB
	ev.OldTotalKB = p.pendingBuf.oldTotalKB
	ev.MetaBeforeKB = p.pendingBuf.metaBeforeKB
	ev.MetaAfterKB = p.pendingBuf.metaAfterKB
	ev.MetaTotalKB = p.pendingBuf.metaTotalKB
	p.pendingBuf = pendingDetail{}
}

func classifyUnifiedParallelPause(desc string) string {
	d := strings.ToLower(desc)
	if strings.Contains(d, "full") {
		if strings.Contains(d, "system") {
			return "ParallelSystemGC"
		}
		if strings.Contains(d, "metadata") {
			return "ParallelMetadataGC"
		}
		if strings.Contains(d, "ergonomics") {
			return "ParallelFullGCErgonomics"
		}
		if strings.Contains(d, "allocation") {
			return "ParallelFullGCAllocationFailure"
		}
		return "ParallelFullGC"
	}
	if strings.Contains(d, "young") {
		if strings.Contains(d, "system") {
			return "ParallelSystemGC"
		}
		if strings.Contains(d, "metadata") {
			return "ParallelMetadataGC"
		}
		return "ParallelYoungGC"
	}
	return "ParallelUnknown"
}

// shared cause extraction regex for unified format
var unifiedPauseCauseRe = regexp.MustCompile(`Pause\s+\w+\s+\((.+?)\)\s`)

var _ Parser = (*ParallelParser)(nil)
