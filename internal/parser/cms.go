package parser

import (
	"regexp"
	"strings"
)

type CMSParser struct {
	handler      EventHandler
	unified      bool
	currentEvent *GCEvent
}

func NewCMSParser(unified bool) *CMSParser {
	return &CMSParser{unified: unified}
}

func (p *CMSParser) GCType() GCType            { return GCTypeCMS }
func (p *CMSParser) SetHandler(h EventHandler) { p.handler = h }

func (p *CMSParser) Feed(line string) {
	if p.handler == nil {
		return
	}
	if p.unified {
		p.feedUnified(line)
	} else {
		p.feedPreUnified(line)
	}
}

// ===================== PreUnified CMS (JDK8) =====================

// ParNew: [GC (Cause) [ParNew: young->young(total), dur secs] heap->heap(total), dur secs]
// With -XX:+PrintGCDateStamps -XX:+PrintGCTimeStamps JDK may emit an inner timestamp before [ParNew:].
var cmsParNewRe = regexp.MustCompile(
	`\[GC\s*(?:\(([^)]*)\)\s*)?(?:\s*\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{3}[+-]\d{4}:\s*[\d.]+:\s*)?\[ParNew:\s*([\d.]+[KMGB]?)->([\d.]+[KMGB]?)\(([\d.]+[KMGB]?)\),\s*([\d.]+)\s*secs\]\s*([\d.]+[KMGB]?)->([\d.]+[KMGB]?)\(([\d.]+[KMGB]?)\),\s*([\d.]+)\s*secs\]`,
)

// ParNew promotion failed: [GC (cause) [ParNew (promotion failed): young->young(total), dur secs]
var cmsPromotionFailedRe = regexp.MustCompile(
	`\[GC\s*(?:\(([^)]*)\)\s*)?\[ParNew\s*\(promotion failed\):\s*([\d.]+[KMGB]?)->([\d.]+[KMGB]?)\(([\d.]+[KMGB]?)\),\s*([\d.]+)\s*secs\]`,
)

// CMS Initial Mark: [GC (CMS Initial Mark) [1 CMS-initial-mark: old(total)] heap(total), dur secs]
var cmsInitialMarkRe = regexp.MustCompile(
	`CMS-initial-mark:\s*([\d.]+[KMGB]?)\(([\d.]+[KMGB]?)\)\]\s*([\d.]+[KMGB]?)\(([\d.]+[KMGB]?)\),\s*([\d.]+)\s*secs\]`,
)

// CMS Remark: [GC (CMS Final Remark) [...][1 CMS-remark: old(total)] heap(total), dur secs]
var cmsRemarkRe = regexp.MustCompile(
	`CMS-remark:\s*([\d.]+[KMGB]?)\(([\d.]+[KMGB]?)\)\]\s*([\d.]+[KMGB]?)\(([\d.]+[KMGB]?)\),\s*([\d.]+)\s*secs\]`,
)

// Concurrent phases: [CMS-concurrent-X: elapsed/wall secs]
var cmsConcurrentRe = regexp.MustCompile(
	`\[(CMS-concurrent-[\w-]+):\s*([\d.]+)/([\d.]+)\s*secs\]`,
)

// Full GC: [Full GC (Cause) heap->heap(total), dur secs]
var cmsFullRe = regexp.MustCompile(
	`\[Full GC\s*(?:\((.+)\)\s+)?([\d.]+[KMGB]?)->([\d.]+[KMGB]?)\(([\d.]+[KMGB]?)\),\s*([\d.]+)\s*secs\]`,
)

// Concurrent mode failure: (concurrent mode failure): old->old(total), dur secs]
var cmsConcModeFailRe = regexp.MustCompile(
	`\(concurrent mode failure\):\s*([\d.]+[KMGB]?)->([\d.]+[KMGB]?)\(([\d.]+[KMGB]?)\),\s*([\d.]+)\s*secs\]`,
)

// Simple GC line (no ParNew detail): [GC (cause)  heap->heap(total), dur secs]
// Used by -verbose:gc format where ParNew detail is omitted
var cmsSimpleGCRe = regexp.MustCompile(
	`\[GC\s*\(([^)]+)\)\s+([\d.]+[KMGB]?)->([\d.]+[KMGB]?)\(([\d.]+[KMGB]?)\),\s*([\d.]+)\s*secs\]`,
)

var cmsMetaRe = regexp.MustCompile(
	`\[Metaspace:\s*([\d]+[KMGB]?)->([\d]+[KMGB]?)\(([\d]+[KMGB]?)\)\]`,
)

// Class unloading inside remark
var cmsClassUnloadingRe = regexp.MustCompile(`\[class unloading,\s*([\d.]+)\s*secs\]`)
var cmsSymbolTableRe = regexp.MustCompile(`\[scrub symbol table,\s*([\d.]+)\s*secs\]`)
var cmsStringTableRe = regexp.MustCompile(`\[scrub string table,\s*([\d.]+)\s*secs\]`)

func (p *CMSParser) Flush() { p.emitPending() }

func (p *CMSParser) emitPending() {
	if p.currentEvent != nil {
		p.handler.Handle(p.currentEvent)
		p.currentEvent = nil
	}
}

func (p *CMSParser) feedPreUnified(line string) {
	ts, hasTs := parsePreUnifiedTimestamp(line)

	// Concurrent mode failure (appears inside a GC line, must check before Full GC)
	if strings.Contains(line, "concurrent mode failure") {
		if m := cmsConcModeFailRe.FindStringSubmatch(line); len(m) > 0 {
			p.emitPending()
			ev := &GCEvent{
				GCType:      GCTypeCMS,
				IsPause:     true,
				Category:    "ConcurrentModeFailure",
				OldBeforeKB: parseSizeKB(m[1]),
				OldAfterKB:  parseSizeKB(m[2]),
				OldTotalKB:  parseSizeKB(m[3]),
				Duration:    parseFloat(m[4]),
			}
			if hasTs {
				ev.Timestamp = ts
			}
			p.currentEvent = ev
			return
		}
	}

	// Promotion failed (must check before normal ParNew)
	if strings.Contains(line, "promotion failed") {
		if m := cmsPromotionFailedRe.FindStringSubmatch(line); len(m) > 0 {
			p.emitPending()
			ev := &GCEvent{
				GCType:        GCTypeCMS,
				IsPause:       true,
				Category:      "PromotionFailed",
				Cause:         m[1],
				YoungBeforeKB: parseSizeKB(m[2]),
				YoungAfterKB:  parseSizeKB(m[3]),
				YoungTotalKB:  parseSizeKB(m[4]),
				Duration:      parseFloat(m[5]),
			}
			if hasTs {
				ev.Timestamp = ts
			}
			p.currentEvent = ev
			return
		}
	}

	// ParNew
	if m := cmsParNewRe.FindStringSubmatch(line); len(m) > 0 {
		p.emitPending()
		ev := &GCEvent{
			GCType:        GCTypeCMS,
			IsPause:       true,
			Category:      "ParNew",
			Cause:         m[1],
			YoungBeforeKB: parseSizeKB(m[2]),
			YoungAfterKB:  parseSizeKB(m[3]),
			YoungTotalKB:  parseSizeKB(m[4]),
			HeapBeforeKB:  parseSizeKB(m[6]),
			HeapAfterKB:   parseSizeKB(m[7]),
			HeapTotalKB:   parseSizeKB(m[8]),
			Duration:      parseFloat(m[9]),
		}
		if hasTs {
			ev.Timestamp = ts
		}
		p.currentEvent = ev
		return
	}

	// Simple GC line: [GC (cause) heap->heap(total), dur secs]
	// Fallback for -verbose:gc format without [ParNew:] detail
	if m := cmsSimpleGCRe.FindStringSubmatch(line); len(m) > 0 {
		cause := m[1]
		if !strings.Contains(cause, "CMS") {
			p.emitPending()
			ev := &GCEvent{
				GCType:       GCTypeCMS,
				IsPause:      true,
				Cause:        cause,
				Category:     classifyCMSSimpleCause(cause),
				HeapBeforeKB: parseSizeKB(m[2]),
				HeapAfterKB:  parseSizeKB(m[3]),
				HeapTotalKB:  parseSizeKB(m[4]),
				Duration:     parseFloat(m[5]),
			}
			if hasTs {
				ev.Timestamp = ts
			}
			p.currentEvent = ev
			return
		}
	}

	// CMS Initial Mark
	if m := cmsInitialMarkRe.FindStringSubmatch(line); len(m) > 0 {
		p.emitPending()
		ev := &GCEvent{
			GCType:       GCTypeCMS,
			IsPause:      true,
			Category:     "CMSInitialMark",
			OldBeforeKB:  parseSizeKB(m[1]),
			OldTotalKB:   parseSizeKB(m[2]),
			HeapBeforeKB: parseSizeKB(m[3]),
			HeapTotalKB:  parseSizeKB(m[4]),
			Duration:     parseFloat(m[5]),
		}
		if hasTs {
			ev.Timestamp = ts
		}
		p.currentEvent = ev
		return
	}

	// CMS Remark
	if m := cmsRemarkRe.FindStringSubmatch(line); len(m) > 0 {
		p.emitPending()
		ev := &GCEvent{
			GCType:       GCTypeCMS,
			IsPause:      true,
			Category:     "CMSRemark",
			OldBeforeKB:  parseSizeKB(m[1]),
			OldTotalKB:   parseSizeKB(m[2]),
			HeapBeforeKB: parseSizeKB(m[3]),
			HeapTotalKB:  parseSizeKB(m[4]),
			Duration:     parseFloat(m[5]),
		}
		if hasTs {
			ev.Timestamp = ts
		}
		// Extract class unloading details if present on same line
		if cm := cmsClassUnloadingRe.FindStringSubmatch(line); len(cm) > 0 {
			ev.CMSClassUnloadingMs = parseFloat(cm[1]) * 1000.0
		}
		if cm := cmsSymbolTableRe.FindStringSubmatch(line); len(cm) > 0 {
			ev.CMSSymbolTableMs = parseFloat(cm[1]) * 1000.0
		}
		if cm := cmsStringTableRe.FindStringSubmatch(line); len(cm) > 0 {
			ev.CMSStringTableMs = parseFloat(cm[1]) * 1000.0
		}
		p.currentEvent = ev
		return
	}

	// Concurrent phases
	if m := cmsConcurrentRe.FindStringSubmatch(line); len(m) > 0 {
		phase := m[1]
		ev := &GCEvent{
			GCType:   GCTypeCMS,
			IsPause:  false,
			Category: classifyCMSConcurrent(phase),
			Duration: parseFloat(m[2]),
		}
		if hasTs {
			ev.Timestamp = ts
		}
		p.handler.Handle(ev)
		return
	}

	// Full GC
	if m := cmsFullRe.FindStringSubmatch(line); len(m) > 0 {
		p.emitPending()
		ev := &GCEvent{
			GCType:       GCTypeCMS,
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
		ev.Category = classifyCMSFullCause(ev.Cause)
		p.currentEvent = ev
		return
	}

	// Detail lines for current event
	if p.currentEvent != nil {
		if m := cmsMetaRe.FindStringSubmatch(line); len(m) > 0 {
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

func classifyCMSConcurrent(phase string) string {
	switch {
	case strings.Contains(phase, "abortable-preclean"):
		return "CMSAbortablePreClean"
	case strings.Contains(phase, "preclean"):
		return "CMSConcurrentPreClean"
	case strings.Contains(phase, "mark"):
		return "CMSConcurrentMark"
	case strings.Contains(phase, "sweep"):
		return "CMSConcurrentSweep"
	case strings.Contains(phase, "reset"):
		return "CMSConcurrentReset"
	default:
		return "CMSConcurrent"
	}
}

func classifyCMSSimpleCause(cause string) string {
	c := strings.ToLower(cause)
	if strings.Contains(c, "allocation failure") || strings.Contains(c, "allocation") {
		return "ParNew"
	}
	if strings.Contains(c, "gclocker") {
		return "ParNew"
	}
	return "ParNew"
}

func classifyCMSFullCause(cause string) string {
	c := strings.ToLower(cause)
	if strings.Contains(c, "system") {
		return "SystemGC"
	}
	if strings.Contains(c, "metadata") {
		return "MetadataGC"
	}
	return "FullGC"
}

// ===================== Unified CMS (JDK9-13) =====================

// Pause with heap: GC(N) Pause ... XM->YM(ZM) Dms
var unifiedCMSPauseHeapRe = regexp.MustCompile(
	`GC\((\d+)\)\s+Pause\s+(.+?)\s+(\d+[KMGB])->(\d+[KMGB])\((\d+[KMGB])\)\s+([\d.]+)ms`)

// Pause without heap (fallback)
var unifiedCMSPauseRe = regexp.MustCompile(`GC\((\d+)\)\s+Pause\s+(.+?)\s+([\d.]+)ms`)
var unifiedCMSConcRe = regexp.MustCompile(`GC\((\d+)\)\s+Concurrent\s+(.+?)\s+([\d.]+)ms`)

// Extract cause from parenthesized text BEFORE heap data: Pause ... (Cause) XM->YM(ZM)
var unifiedCMSCauseRe = regexp.MustCompile(`Pause\s+\w+\s+\(([^)]+)\)`)

func (p *CMSParser) feedUnified(line string) {
	if !strings.Contains(line, "][gc") {
		return
	}

	ts, _ := parseUnifiedTimestamp(line)

	// Try pause with heap first
	if m := unifiedCMSPauseHeapRe.FindStringSubmatch(line); len(m) > 0 {
		if p.currentEvent != nil {
			p.handler.Handle(p.currentEvent)
		}
		category := classifyUnifiedCMSPause(m[2])
		ev := &GCEvent{
			GCType:       GCTypeCMS,
			Timestamp:    ts,
			IsPause:      true,
			Category:     category,
			HeapBeforeKB: parseSizeKB(m[3]),
			HeapAfterKB:  parseSizeKB(m[4]),
			HeapTotalKB:  parseSizeKB(m[5]),
			Duration:     parseFloat(m[6]) / 1000.0,
		}
		if cm := unifiedCMSCauseRe.FindStringSubmatch(line); len(cm) > 0 {
			ev.Cause = cm[1]
		}
		p.currentEvent = ev
		return
	}

	// Fallback: pause without heap
	if m := unifiedCMSPauseRe.FindStringSubmatch(line); len(m) > 0 {
		if p.currentEvent != nil {
			p.handler.Handle(p.currentEvent)
		}
		category := classifyUnifiedCMSPause(m[2])
		ev := &GCEvent{
			GCType:    GCTypeCMS,
			Timestamp: ts,
			IsPause:   true,
			Category:  category,
			Duration:  parseFloat(m[3]) / 1000.0,
		}
		if cm := unifiedCMSCauseRe.FindStringSubmatch(line); len(cm) > 0 {
			ev.Cause = cm[1]
		}
		p.currentEvent = ev
		return
	}

	if m := unifiedCMSConcRe.FindStringSubmatch(line); len(m) > 0 {
		ev := &GCEvent{
			GCType:    GCTypeCMS,
			Timestamp: ts,
			IsPause:   false,
			Category:  classifyUnifiedCMSConcurrent(m[2]),
			Duration:  parseFloat(m[3]) / 1000.0,
		}
		p.handler.Handle(ev)
		return
	}

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
	}
}

func classifyUnifiedCMSPause(desc string) string {
	d := strings.ToLower(desc)
	if strings.Contains(d, "full") {
		if strings.Contains(d, "system") {
			return "SystemGC"
		}
		if strings.Contains(d, "concurrent mode failure") {
			return "ConcurrentModeFailure"
		}
		if strings.Contains(d, "metadata") {
			return "MetadataGC"
		}
		return "FullGC"
	}
	if strings.Contains(d, "initial mark") {
		return "CMSInitialMark"
	}
	if strings.Contains(d, "remark") {
		return "CMSRemark"
	}
	if strings.Contains(d, "young") {
		return "ParNew"
	}
	return "CMSUnknown"
}

func classifyUnifiedCMSConcurrent(desc string) string {
	d := strings.ToLower(desc)
	switch {
	case strings.Contains(d, "abortable preclean") || strings.Contains(d, "abortable-preclean"):
		return "CMSAbortablePreClean"
	case strings.Contains(d, "preclean"):
		return "CMSConcurrentPreClean"
	case strings.Contains(d, "mark"):
		return "CMSConcurrentMark"
	case strings.Contains(d, "sweep"):
		return "CMSConcurrentSweep"
	case strings.Contains(d, "reset"):
		return "CMSConcurrentReset"
	default:
		return "CMSConcurrent"
	}
}

var _ Parser = (*CMSParser)(nil)
