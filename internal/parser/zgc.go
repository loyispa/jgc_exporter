package parser

import (
	"regexp"
	"strings"
)

type ZGCParser struct {
	handler      EventHandler
	events       map[string]*GCEvent // keyed by GC ID to support concurrent collections
	generational bool
}

func NewZGCParser() *ZGCParser {
	return &ZGCParser{events: make(map[string]*GCEvent)}
}

func (p *ZGCParser) GCType() GCType            { return GCTypeZGC }
func (p *ZGCParser) SetHandler(h EventHandler) { p.handler = h }

var zgcGCIDRe = regexp.MustCompile(`GC\((\d+)\)`)

func (p *ZGCParser) Feed(line string) {
	if p.handler == nil {
		return
	}
	if !strings.Contains(line, "][gc") {
		return
	}

	ts, _ := parseUnifiedTimestamp(line)

	gcID := ""
	if m := zgcGCIDRe.FindStringSubmatch(line); len(m) > 0 {
		gcID = m[1]
	}

	if m := zgcSummaryRe.FindStringSubmatch(line); len(m) > 0 {
		durMs := parseFloat(m[6])
		heapBeforeKB := parseSizeKB(m[4] + "M")
		heapAfterKB := parseSizeKB(m[5] + "M")
		summaryID := m[1]

		if ev, ok := p.events[summaryID]; ok {
			if ev.HeapBeforeKB == 0 {
				ev.HeapBeforeKB = heapBeforeKB
			}
			if ev.HeapAfterKB == 0 {
				ev.HeapAfterKB = heapAfterKB
			}
			if ev.Duration == 0 {
				ev.Duration = durMs / 1000.0
			}
			p.emitEvent(summaryID)
		} else {
			collType := ""
			cause := m[3]
			cat := classifyZGCCause(cause)
			kindLower := strings.ToLower(m[2])
			if strings.Contains(kindLower, "minor") {
				collType = "minor"
				cat = classifyGenZGCCause(collType, cause)
				p.generational = true
			} else if strings.Contains(kindLower, "major") {
				collType = "major"
				cat = classifyGenZGCCause(collType, cause)
				p.generational = true
			}
			ev := &GCEvent{
				GCType:            GCTypeZGC,
				Timestamp:         ts,
				IsPause:           true,
				Cause:             cause,
				Category:          cat,
				HeapBeforeKB:      heapBeforeKB,
				HeapAfterKB:       heapAfterKB,
				Duration:          durMs / 1000.0,
				ZGCCollectionType: collType,
			}
			p.handler.Handle(ev)
		}
		return
	}

	if m := zgcGenStartRe.FindStringSubmatch(line); len(m) > 0 {
		p.generational = true
		collType := strings.ToLower(m[2])
		cause := m[3]
		p.events[m[1]] = &GCEvent{
			GCType:            GCTypeZGC,
			Timestamp:         ts,
			IsPause:           false,
			Cause:             cause,
			Category:          classifyGenZGCCause(collType, cause),
			ZGCCollectionType: collType,
		}
		return
	}

	if m := zgcStartRe.FindStringSubmatch(line); len(m) > 0 {
		p.events[m[1]] = &GCEvent{
			GCType:    GCTypeZGC,
			Timestamp: ts,
			IsPause:   false,
			Cause:     m[2],
			Category:  classifyZGCCause(m[2]),
		}
		return
	}

	ev := p.events[gcID]
	if ev == nil {
		return
	}

	if m := zgcGenPhaseRe.FindStringSubmatch(line); len(m) > 0 {
		phase := strings.TrimSpace(m[1])
		dur := parseFloat(m[2])
		p.recordPhase(ev, phase, dur)
		return
	}

	if m := zgcPhaseRe.FindStringSubmatch(line); len(m) > 0 {
		phase := strings.TrimSpace(m[1])
		dur := parseFloat(m[2])
		p.recordPhase(ev, phase, dur)
		return
	}

	if m := zgcLoadRe.FindStringSubmatch(line); len(m) > 0 {
		ev.ZGCLoad1m = parseFloat(m[1])
		ev.ZGCLoad5m = parseFloat(m[2])
		ev.ZGCLoad15m = parseFloat(m[3])
		return
	}

	if m := zgcMMURe.FindStringSubmatch(line); len(m) > 0 {
		ev.ZGCMMU2ms = parseFloat(m[1]) / 100.0
		ev.ZGCMMU5ms = parseFloat(m[2]) / 100.0
		ev.ZGCMMU10ms = parseFloat(m[3]) / 100.0
		ev.ZGCMMU20ms = parseFloat(m[4]) / 100.0
		ev.ZGCMMU50ms = parseFloat(m[5]) / 100.0
		ev.ZGCMMU100ms = parseFloat(m[6]) / 100.0
		return
	}

	if m := zgcMetaRe.FindStringSubmatch(line); len(m) > 0 {
		ev.ZGCMetaspaceUsedKB = parseSizeKB(m[1] + "M")
		ev.ZGCMetaspaceCommitKB = parseSizeKB(m[2] + "M")
		if len(m) > 3 && m[3] != "" {
			ev.ZGCMetaspaceReservedKB = parseSizeKB(m[3] + "M")
		}
		return
	}

	if strings.Contains(line, "gc,heap") && strings.Contains(line, "Used:") {
		if m := zgcHeapUsedRe.FindStringSubmatch(line); len(m) > 0 {
			ev.HeapBeforeKB = parseSizeKB(m[1] + "M")
			ev.HeapAfterKB = parseSizeKB(m[4] + "M")
			ev.ZGCUsedMB = parseInt(m[1])
			ev.ZGCMarkStartUsedKB = parseSizeKB(m[1] + "M")
			ev.ZGCMarkEndUsedKB = parseSizeKB(m[2] + "M")
			ev.ZGCRelocateStartUsedKB = parseSizeKB(m[3] + "M")
			ev.ZGCRelocateEndUsedKB = parseSizeKB(m[4] + "M")
		}
		return
	}

	if strings.Contains(line, "gc,heap") && strings.Contains(line, "Capacity:") {
		if m := zgcCapacityRe.FindStringSubmatch(line); len(m) > 0 {
			ev.HeapTotalKB = parseSizeKB(m[1] + "M")
		}
		return
	}

	if strings.Contains(line, "gc,heap") && strings.Contains(line, "Free:") {
		if m := zgcHeapFreeRe.FindStringSubmatch(line); len(m) > 0 {
			ev.ZGCFreeMB = parseInt(m[1])
			ev.ZGCMarkStartFreeKB = parseSizeKB(m[1] + "M")
			ev.ZGCMarkEndFreeKB = parseSizeKB(m[2] + "M")
			ev.ZGCRelocateStartFreeKB = parseSizeKB(m[3] + "M")
			ev.ZGCRelocateEndFreeKB = parseSizeKB(m[4] + "M")
		} else if m := zgcFreeRe.FindStringSubmatch(line); len(m) > 0 {
			ev.ZGCFreeMB = parseInt(m[1])
		}
		if !p.generational {
			p.emitEvent(gcID)
		}
	}
}

func (p *ZGCParser) Flush() {
	for id := range p.events {
		p.emitEvent(id)
	}
}

func (p *ZGCParser) emitEvent(gcID string) {
	ev, ok := p.events[gcID]
	if !ok || ev == nil {
		return
	}
	delete(p.events, gcID)

	totalMs := ev.ZGCPauseMarkStartMs + ev.ZGCConcurrentMarkMs + ev.ZGCPauseMarkEndMs +
		ev.ZGCConcurrentMarkFreeMs + ev.ZGCConcurrentProcessNSRMs +
		ev.ZGCConcurrentResetRelocMs + ev.ZGCConcurrentSelectRelocMs +
		ev.ZGCPauseRelocateMs + ev.ZGCConcurrentRelocMs +
		ev.ZGCConcurrentRemapRootsMs + ev.ZGCConcurrentMarkContinueMs

	if ev.Duration == 0 && totalMs > 0 {
		ev.Duration = totalMs / 1000.0
	}

	ev.IsPause = ev.ZGCPauseMarkStartMs > 0 || ev.ZGCPauseMarkEndMs > 0 || ev.ZGCPauseRelocateMs > 0
	p.handler.Handle(ev)
}

func (p *ZGCParser) recordPhase(ev *GCEvent, phase string, durMs float64) {
	lp := strings.ToLower(phase)

	switch {
	case strings.Contains(lp, "pause mark start"):
		ev.ZGCPauseMarkStartMs = durMs
	case strings.Contains(lp, "concurrent mark free"):
		ev.ZGCConcurrentMarkFreeMs = durMs
	case strings.Contains(lp, "concurrent mark continue"):
		ev.ZGCConcurrentMarkContinueMs = durMs
	case strings.Contains(lp, "concurrent mark") && !strings.Contains(lp, "free") && !strings.Contains(lp, "continue"):
		ev.ZGCConcurrentMarkMs = durMs
	case strings.Contains(lp, "pause mark end"):
		ev.ZGCPauseMarkEndMs = durMs
	case strings.Contains(lp, "concurrent process non-strong"):
		ev.ZGCConcurrentProcessNSRMs = durMs
	case strings.Contains(lp, "concurrent reset relocation set"):
		ev.ZGCConcurrentResetRelocMs = durMs
	case strings.Contains(lp, "concurrent select relocation set"):
		ev.ZGCConcurrentSelectRelocMs = durMs
	case strings.Contains(lp, "pause relocate start"):
		ev.ZGCPauseRelocateMs = durMs
	case strings.Contains(lp, "concurrent remap roots"):
		ev.ZGCConcurrentRemapRootsMs = durMs
	case strings.Contains(lp, "concurrent relocate"):
		ev.ZGCConcurrentRelocMs = durMs
	}
}

// Non-generational: "GC(N) Garbage Collection (Cause)"
var zgcStartRe = regexp.MustCompile(`GC\((\d+)\)\s+Garbage Collection\s*\((.+)\)`)

// Generational: "GC(N) Minor/Major Collection (Cause)"
var zgcGenStartRe = regexp.MustCompile(`GC\((\d+)\)\s+(Minor|Major) Collection\s*\((.+)\)`)

// Summary line: "GC(N) Garbage/Minor/Major Collection (cause) XM(X%)->YM(Y%) Dms"
var zgcSummaryRe = regexp.MustCompile(
	`GC\((\d+)\)\s+((?:Garbage|Minor|Major) Collection)\s*\((.+?)\)\s+(\d+)M\(\d+%\)->(\d+)M\(\d+%\)\s+([\d.]+)ms`)

// Generational phases: "GC(N) y:/o: Pause/Concurrent Phase Xms"
var zgcGenPhaseRe = regexp.MustCompile(`GC\(\d+\)\s+[yo]:\s+((?:Pause|Concurrent)\s+[\w\s-]+?)\s+([\d.]+)ms`)

// Non-generational phases
var zgcPhaseRe = regexp.MustCompile(`GC\(\d+\)\s+((?:Pause|Concurrent)\s+[\w\s-]+?)\s+([\d.]+)ms`)

var zgcLoadRe = regexp.MustCompile(`GC\(\d+\)\s+Load:\s*([\d.]+)/([\d.]+)/([\d.]+)`)
var zgcMMURe = regexp.MustCompile(`GC\(\d+\)\s+MMU:\s*2ms/([\d.]+)%,\s*5ms/([\d.]+)%,\s*10ms/([\d.]+)%,\s*20ms/([\d.]+)%,\s*50ms/([\d.]+)%,\s*100ms/([\d.]+)%`)
var zgcMetaRe = regexp.MustCompile(`GC\(\d+\)\s+Metaspace:\s*(\d+)M\s+used,\s*(\d+)M\s+committed(?:,\s*(\d+)M\s+reserved)?`)
var zgcCapacityRe = regexp.MustCompile(`Capacity:\s*(\d+)M`)
var zgcHeapUsedRe = regexp.MustCompile(`Used:\s*(\d+)M\s*\(\d+%\)\s+(\d+)M\s*\(\d+%\)\s+(\d+)M\s*\(\d+%\)\s+(\d+)M\s*\(\d+%\)\s+(\d+)M\s*\(\d+%\)\s+(\d+)M`)
var zgcHeapFreeRe = regexp.MustCompile(`Free:\s*(\d+)M\s*\(\d+%\)\s+(\d+)M\s*\(\d+%\)\s+(\d+)M\s*\(\d+%\)\s+(\d+)M\s*\(\d+%\)\s+(\d+)M\s*\(\d+%\)\s+(\d+)M`)
var zgcFreeRe = regexp.MustCompile(`Free:\s*(\d+)M`)

func classifyZGCCause(cause string) string {
	c := strings.ToLower(cause)
	switch {
	case strings.Contains(c, "warmup"):
		return "ZGCWarmup"
	case strings.Contains(c, "timer"):
		return "ZGCTimer"
	case strings.Contains(c, "allocation rate") || strings.Contains(c, "alloc rate"):
		return "ZGCAllocRate"
	case strings.Contains(c, "allocation stall") || strings.Contains(c, "alloc stall"):
		return "ZGCAllocStall"
	case strings.Contains(c, "proactive"):
		return "ZGCProactive"
	case strings.Contains(c, "metadata"):
		return "ZGCMetadataGCThreshold"
	case strings.Contains(c, "system"):
		return "ZGCSystemGc"
	default:
		return "ZGCUnknown"
	}
}

func classifyGenZGCCause(collType, cause string) string {
	prefix := "ZGCMinor"
	if collType == "major" {
		prefix = "ZGCMajor"
	}
	c := strings.ToLower(cause)
	switch {
	case strings.Contains(c, "warmup"):
		return prefix + "Warmup"
	case strings.Contains(c, "timer"):
		return prefix + "Timer"
	case strings.Contains(c, "allocation rate") || strings.Contains(c, "alloc rate"):
		return prefix + "AllocRate"
	case strings.Contains(c, "allocation stall") || strings.Contains(c, "alloc stall"):
		return prefix + "AllocStall"
	case strings.Contains(c, "proactive"):
		return prefix + "Proactive"
	case strings.Contains(c, "metadata"):
		return prefix + "MetadataGC"
	case strings.Contains(c, "system"):
		return prefix + "SystemGc"
	default:
		return prefix + "Unknown"
	}
}

var _ Parser = (*ZGCParser)(nil)
