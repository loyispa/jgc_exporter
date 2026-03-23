package parser

import (
	"regexp"
	"strings"
)

type SerialParser struct {
	handler      EventHandler
	unified      bool
	currentEvent *GCEvent
	pendingBuf   pendingDetail
}

func NewSerialParser(unified bool) *SerialParser {
	return &SerialParser{unified: unified}
}

func (p *SerialParser) GCType() GCType            { return GCTypeSerial }
func (p *SerialParser) SetHandler(h EventHandler) { p.handler = h }
func (p *SerialParser) Flush()                    { p.emitPending() }

func (p *SerialParser) Feed(line string) {
	if p.handler == nil {
		return
	}
	if p.unified {
		p.feedUnified(line)
	} else {
		p.feedPreUnified(line)
	}
}

// ===================== PreUnified Serial (JDK8) =====================

// Young GC: [GC (Cause) [DefNew: young->young(total), dur secs] heap->heap(total), dur secs]
// Note: DefNew line may span multiple lines (tenuring distribution) with the actual
// DefNew header on its own line. We match the summary line containing heap totals.
var defNewRe = regexp.MustCompile(
	`:\s*([\d.]+[KMGB]?)->([\d.]+[KMGB]?)\(([\d.]+[KMGB]?)\),\s*[\d.]+\s*secs\]\s*([\d.]+[KMGB]?)->([\d.]+[KMGB]?)\(([\d.]+[KMGB]?)\),\s*([\d.]+)\s*secs\]`,
)

// DefNew start: [GC (Cause) timestamp: [DefNew
var defNewStartRe = regexp.MustCompile(`\[GC\s*(?:\(([^)]*)\))`)

// Full GC with Tenured: [Full GC (Cause) [Tenured: old->old(total), dur secs] heap->heap(total), [Metaspace: ...], dur secs]
var serialFullRe = regexp.MustCompile(
	`\[Full GC\s*(?:\((.+?)\)\s+).*?\[Tenured:\s*([\d.]+[KMGB]?)->([\d.]+[KMGB]?)\(([\d.]+[KMGB]?)\),\s*[\d.]+\s*secs\]\s*([\d.]+[KMGB]?)->([\d.]+[KMGB]?)\(([\d.]+[KMGB]?)\).*?,\s*([\d.]+)\s*secs\]`,
)

// Simple Full GC: [Full GC (Cause) heap->heap(total), dur secs]
var serialSimpleFullRe = regexp.MustCompile(
	`\[Full GC\s*(?:\((.+?)\)\s+)?([\d.]+[KMGB]?)->([\d.]+[KMGB]?)\(([\d.]+[KMGB]?)\),\s*([\d.]+)\s*secs\]`,
)

func (p *SerialParser) emitPending() {
	if p.currentEvent != nil {
		p.handler.Handle(p.currentEvent)
		p.currentEvent = nil
	}
}

func (p *SerialParser) feedPreUnified(line string) {
	ts, hasTs := parsePreUnifiedTimestamp(line)

	// Full GC with Tenured detail
	if strings.Contains(line, "[Full GC") && strings.Contains(line, "[Tenured") {
		if m := serialFullRe.FindStringSubmatch(line); len(m) > 0 {
			p.emitPending()
			ev := &GCEvent{
				GCType:       GCTypeSerial,
				IsPause:      true,
				Cause:        m[1],
				OldBeforeKB:  parseSizeKB(m[2]),
				OldAfterKB:   parseSizeKB(m[3]),
				OldTotalKB:   parseSizeKB(m[4]),
				HeapBeforeKB: parseSizeKB(m[5]),
				HeapAfterKB:  parseSizeKB(m[6]),
				HeapTotalKB:  parseSizeKB(m[7]),
				Duration:     parseFloat(m[8]),
			}
			ev.Category = classifySerialFullCause(ev.Cause)
			if hasTs {
				ev.Timestamp = ts
			}
			if pm := psMetaspaceRe.FindStringSubmatch(line); len(pm) > 0 {
				ev.MetaBeforeKB = parseSizeKB(pm[1])
				ev.MetaAfterKB = parseSizeKB(pm[2])
				ev.MetaTotalKB = parseSizeKB(pm[3])
			}
			p.currentEvent = ev
			return
		}
	}

	// Full GC without Tenured detail
	if strings.Contains(line, "[Full GC") {
		if m := serialSimpleFullRe.FindStringSubmatch(line); len(m) > 0 {
			p.emitPending()
			ev := &GCEvent{
				GCType:       GCTypeSerial,
				IsPause:      true,
				Cause:        m[1],
				HeapBeforeKB: parseSizeKB(m[2]),
				HeapAfterKB:  parseSizeKB(m[3]),
				HeapTotalKB:  parseSizeKB(m[4]),
				Duration:     parseFloat(m[5]),
			}
			ev.Category = classifySerialFullCause(ev.Cause)
			if hasTs {
				ev.Timestamp = ts
			}
			p.currentEvent = ev
			return
		}
	}

	// DefNew start line: begins a young GC event
	if strings.Contains(line, "[DefNew") {
		p.emitPending()
		cause := ""
		if m := defNewStartRe.FindStringSubmatch(line); len(m) > 0 {
			cause = m[1]
		}
		ev := &GCEvent{
			GCType:   GCTypeSerial,
			IsPause:  true,
			Cause:    cause,
			Category: classifySerialYoungCause(cause),
		}
		if hasTs {
			ev.Timestamp = ts
		}

		// The DefNew summary may appear on the same line or a later line
		if m := defNewRe.FindStringSubmatch(line); len(m) > 0 {
			ev.YoungBeforeKB = parseSizeKB(m[1])
			ev.YoungAfterKB = parseSizeKB(m[2])
			ev.YoungTotalKB = parseSizeKB(m[3])
			ev.HeapBeforeKB = parseSizeKB(m[4])
			ev.HeapAfterKB = parseSizeKB(m[5])
			ev.HeapTotalKB = parseSizeKB(m[6])
			ev.Duration = parseFloat(m[7])
		}
		p.currentEvent = ev
		return
	}

	// Continuation lines for DefNew (tenuring distribution lines, then summary)
	if p.currentEvent != nil && p.currentEvent.Duration == 0 {
		if m := defNewRe.FindStringSubmatch(line); len(m) > 0 {
			p.currentEvent.YoungBeforeKB = parseSizeKB(m[1])
			p.currentEvent.YoungAfterKB = parseSizeKB(m[2])
			p.currentEvent.YoungTotalKB = parseSizeKB(m[3])
			p.currentEvent.HeapBeforeKB = parseSizeKB(m[4])
			p.currentEvent.HeapAfterKB = parseSizeKB(m[5])
			p.currentEvent.HeapTotalKB = parseSizeKB(m[6])
			p.currentEvent.Duration = parseFloat(m[7])
		}
	}

	// [Times:] terminates the current event
	if p.currentEvent != nil && strings.Contains(line, "[Times:") {
		p.handler.Handle(p.currentEvent)
		p.currentEvent = nil
	}
}

func classifySerialYoungCause(cause string) string {
	c := strings.ToLower(cause)
	if strings.Contains(c, "system") {
		return "SerialSystemGC"
	}
	if strings.Contains(c, "metadata") {
		return "SerialMetadataGC"
	}
	return "SerialYoungGC"
}

func classifySerialFullCause(cause string) string {
	c := strings.ToLower(cause)
	if strings.Contains(c, "system") {
		return "SerialSystemGC"
	}
	if strings.Contains(c, "metadata") {
		return "SerialMetadataGC"
	}
	if strings.Contains(c, "allocation") {
		return "SerialFullGCAllocationFailure"
	}
	if strings.Contains(c, "ergonomics") {
		return "SerialFullGCErgonomics"
	}
	return "SerialFullGC"
}

// ===================== Unified Serial (JDK9+) =====================

// DefNew detail: GC(N) DefNew: XK(YK)->ZK(WK) Eden: ...
var unifiedDefNewRe = regexp.MustCompile(
	`GC\((\d+)\)\s+DefNew:\s*([\d]+[KMGB]?)\(([\d]+[KMGB]?)\)->([\d]+[KMGB]?)\(([\d]+[KMGB]?)\)`)

// Tenured detail: GC(N) Tenured: XK(YK)->ZK(WK)
var unifiedTenuredRe = regexp.MustCompile(
	`GC\((\d+)\)\s+Tenured:\s*([\d]+[KMGB]?)\(([\d]+[KMGB]?)\)->([\d]+[KMGB]?)\(([\d]+[KMGB]?)\)`)

func (p *SerialParser) feedUnified(line string) {
	if !strings.Contains(line, "][gc") {
		return
	}
	ts, _ := parseUnifiedTimestamp(line)

	// Buffer detail lines (arrive before summary in unified format)
	if m := unifiedDefNewRe.FindStringSubmatch(line); len(m) > 0 {
		p.pendingBuf.gcNum = m[1]
		p.pendingBuf.youngBeforeKB = parseSizeKB(m[2])
		p.pendingBuf.youngTotalKB = parseSizeKB(m[3])
		p.pendingBuf.youngAfterKB = parseSizeKB(m[4])
		return
	}
	if m := unifiedTenuredRe.FindStringSubmatch(line); len(m) > 0 {
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

	// Pause with heap summary
	if m := unifiedGCPauseHeapRe.FindStringSubmatch(line); len(m) > 0 {
		p.emitPending()
		ev := &GCEvent{
			GCType:       GCTypeSerial,
			Timestamp:    ts,
			IsPause:      true,
			Category:     classifyUnifiedSerialPause(m[2]),
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

	// Pause without heap
	if m := unifiedGCPauseRe.FindStringSubmatch(line); len(m) > 0 {
		p.emitPending()
		ev := &GCEvent{
			GCType:    GCTypeSerial,
			Timestamp: ts,
			IsPause:   true,
			Category:  classifyUnifiedSerialPause(m[2]),
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

func (p *SerialParser) applyPendingDetail(ev *GCEvent) {
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

func classifyUnifiedSerialPause(desc string) string {
	d := strings.ToLower(desc)
	if strings.Contains(d, "full") {
		if strings.Contains(d, "system") {
			return "SerialSystemGC"
		}
		if strings.Contains(d, "metadata") {
			return "SerialMetadataGC"
		}
		if strings.Contains(d, "allocation") {
			return "SerialFullGCAllocationFailure"
		}
		return "SerialFullGC"
	}
	if strings.Contains(d, "young") {
		if strings.Contains(d, "system") {
			return "SerialSystemGC"
		}
		return "SerialYoungGC"
	}
	return "SerialUnknown"
}

var _ Parser = (*SerialParser)(nil)
