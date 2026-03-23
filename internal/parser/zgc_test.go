package parser

import "testing"

// ===================== ZGC (JDK11) =====================

func TestZGCJDK11(t *testing.T) {
	h := &testHandler{}
	p := NewZGCParser()
	p.SetHandler(h)
	feedFile(t, p, "../../testdata/parser/jdk11-zgc.log")

	requireEventCount(t, h.events, 7)

	wantCategories := []string{"ZGCWarmup", "ZGCAllocRate", "ZGCProactive", "ZGCTimer", "ZGCSystemGc", "ZGCAllocStall", "ZGCMetadataGCThreshold"}
	for i, ev := range h.events {
		if ev.Category != wantCategories[i] {
			t.Errorf("event[%d] category=%s, want %s", i, ev.Category, wantCategories[i])
		}
	}

	for i, ev := range h.events {
		if ev.ZGCPauseMarkStartMs <= 0 {
			t.Errorf("event[%d] ZGCPauseMarkStartMs should be > 0: %f", i, ev.ZGCPauseMarkStartMs)
		}
		if ev.ZGCConcurrentMarkMs <= 0 {
			t.Errorf("event[%d] ZGCConcurrentMarkMs should be > 0: %f", i, ev.ZGCConcurrentMarkMs)
		}
		if ev.ZGCPauseMarkEndMs <= 0 {
			t.Errorf("event[%d] ZGCPauseMarkEndMs should be > 0: %f", i, ev.ZGCPauseMarkEndMs)
		}
		if ev.ZGCPauseRelocateMs <= 0 {
			t.Errorf("event[%d] ZGCPauseRelocateMs should be > 0: %f", i, ev.ZGCPauseRelocateMs)
		}
		if ev.ZGCConcurrentRelocMs <= 0 {
			t.Errorf("event[%d] ZGCConcurrentRelocMs should be > 0: %f", i, ev.ZGCConcurrentRelocMs)
		}
		if ev.ZGCConcurrentMarkFreeMs <= 0 {
			t.Errorf("event[%d] ZGCConcurrentMarkFreeMs should be > 0: %f", i, ev.ZGCConcurrentMarkFreeMs)
		}
		if ev.ZGCConcurrentProcessNSRMs <= 0 {
			t.Errorf("event[%d] ZGCConcurrentProcessNSRMs should be > 0: %f", i, ev.ZGCConcurrentProcessNSRMs)
		}
		if ev.ZGCConcurrentSelectRelocMs <= 0 {
			t.Errorf("event[%d] ZGCConcurrentSelectRelocMs should be > 0: %f", i, ev.ZGCConcurrentSelectRelocMs)
		}
	}

	for i, ev := range h.events {
		if ev.ZGCLoad1m <= 0 {
			t.Errorf("event[%d] ZGCLoad1m should be > 0: %f", i, ev.ZGCLoad1m)
		}
		if ev.ZGCMMU2ms <= 0 {
			t.Errorf("event[%d] ZGCMMU2ms should be > 0: %f", i, ev.ZGCMMU2ms)
		}
	}

	for i, ev := range h.events {
		if ev.HeapBeforeKB <= 0 {
			t.Errorf("event[%d] HeapBeforeKB should be > 0: %d", i, ev.HeapBeforeKB)
		}
		if ev.HeapTotalKB <= 0 {
			t.Errorf("event[%d] HeapTotalKB should be > 0: %d", i, ev.HeapTotalKB)
		}
	}

	for i, ev := range h.events {
		if ev.Duration <= 0 {
			t.Errorf("event[%d] Duration should be > 0: %f", i, ev.Duration)
		}
	}

	for i, ev := range h.events {
		if !ev.IsPause {
			t.Errorf("event[%d] should have IsPause=true (has pause mark start)", i)
		}
	}

	for i, ev := range h.events {
		if ev.ZGCMetaspaceUsedKB <= 0 {
			t.Errorf("event[%d] ZGCMetaspaceUsedKB should be > 0: %d", i, ev.ZGCMetaspaceUsedKB)
		}
	}

	for i := 1; i < 5; i++ {
		if h.events[i].HeapBeforeKB <= h.events[i-1].HeapBeforeKB {
			t.Errorf("event[%d] HeapBefore (%dK) should be > event[%d] HeapBefore (%dK)",
				i, h.events[i].HeapBeforeKB, i-1, h.events[i-1].HeapBeforeKB)
		}
	}

	t.Logf("Parsed %d ZGC JDK11 events:", len(h.events))
	for i, ev := range h.events {
		t.Logf("  [%d] cat=%-18s dur=%.4fs pause_mark=%.3fms conc_mark=%.3fms "+
			"mark_free=%.3fms proc_nsr=%.3fms select_reloc=%.3fms "+
			"load=%.2f/%.2f/%.2f heap=%dK->%dK(%dK) free=%dM meta=%dK",
			i, ev.Category, ev.Duration,
			ev.ZGCPauseMarkStartMs, ev.ZGCConcurrentMarkMs,
			ev.ZGCConcurrentMarkFreeMs, ev.ZGCConcurrentProcessNSRMs,
			ev.ZGCConcurrentSelectRelocMs,
			ev.ZGCLoad1m, ev.ZGCLoad5m, ev.ZGCLoad15m,
			ev.HeapBeforeKB, ev.HeapAfterKB, ev.HeapTotalKB,
			ev.ZGCFreeMB, ev.ZGCMetaspaceUsedKB)
	}
}

// ===================== ZGC (JDK17) =====================

func TestZGCJDK17(t *testing.T) {
	h := &testHandler{}
	p := NewZGCParser()
	p.SetHandler(h)
	feedFile(t, p, "../../testdata/parser/jdk17-zgc.log")

	requireEventCount(t, h.events, 7)

	wantCategories := []string{"ZGCWarmup", "ZGCAllocRate", "ZGCSystemGc", "ZGCProactive", "ZGCTimer", "ZGCAllocStall", "ZGCMetadataGCThreshold"}
	for i, ev := range h.events {
		if ev.Category != wantCategories[i] {
			t.Errorf("event[%d] category=%s, want %s", i, ev.Category, wantCategories[i])
		}
	}

	for i, ev := range h.events {
		if ev.ZGCPauseMarkStartMs <= 0 {
			t.Errorf("event[%d] ZGCPauseMarkStartMs should be > 0", i)
		}
		if ev.ZGCConcurrentMarkMs <= 0 {
			t.Errorf("event[%d] ZGCConcurrentMarkMs should be > 0", i)
		}
		if ev.Duration <= 0 {
			t.Errorf("event[%d] Duration should be > 0", i)
		}
		if ev.HeapBeforeKB <= 0 {
			t.Errorf("event[%d] HeapBeforeKB should be > 0", i)
		}
		if ev.ZGCLoad1m <= 0 {
			t.Errorf("event[%d] ZGCLoad1m should be > 0", i)
		}
		if ev.ZGCMMU2ms <= 0 {
			t.Errorf("event[%d] ZGCMMU2ms should be > 0", i)
		}
		if ev.ZGCFreeMB <= 0 {
			t.Errorf("event[%d] ZGCFreeMB should be > 0", i)
		}
	}

	t.Logf("Parsed %d ZGC JDK17 events:", len(h.events))
	logEvents(t, h.events)
}

// ===================== ZGC Generational (JDK21) =====================

func TestZGCGenerationalJDK21(t *testing.T) {
	h := &testHandler{}
	p := NewZGCParser()
	p.SetHandler(h)
	feedFile(t, p, "../../testdata/parser/jdk21-zgc-generational.log")

	requireEventCount(t, h.events, 8)

	wantCategories := []string{
		"ZGCMinorWarmup", "ZGCMinorAllocRate", "ZGCMinorTimer",
		"ZGCMajorProactive", "ZGCMajorMetadataGC", "ZGCMajorSystemGc",
		"ZGCMinorAllocStall", "ZGCMajorAllocStall",
	}
	for i, ev := range h.events {
		if ev.Category != wantCategories[i] {
			t.Errorf("event[%d] category=%s, want %s", i, ev.Category, wantCategories[i])
		}
	}

	for i := 0; i < 3; i++ {
		if h.events[i].ZGCCollectionType != "minor" {
			t.Errorf("event[%d] ZGCCollectionType=%q, want \"minor\"", i, h.events[i].ZGCCollectionType)
		}
	}
	for i := 3; i < 6; i++ {
		if h.events[i].ZGCCollectionType != "major" {
			t.Errorf("event[%d] ZGCCollectionType=%q, want \"major\"", i, h.events[i].ZGCCollectionType)
		}
	}
	if h.events[6].ZGCCollectionType != "minor" {
		t.Errorf("event[6] ZGCCollectionType=%q, want \"minor\"", h.events[6].ZGCCollectionType)
	}
	if h.events[7].ZGCCollectionType != "major" {
		t.Errorf("event[7] ZGCCollectionType=%q, want \"major\"", h.events[7].ZGCCollectionType)
	}

	for i, ev := range h.events {
		if ev.Duration <= 0 {
			t.Errorf("event[%d] Duration should be > 0: %f", i, ev.Duration)
		}
		if ev.HeapBeforeKB <= 0 {
			t.Errorf("event[%d] HeapBeforeKB should be > 0: %d", i, ev.HeapBeforeKB)
		}
		if ev.HeapAfterKB <= 0 {
			t.Errorf("event[%d] HeapAfterKB should be > 0: %d", i, ev.HeapAfterKB)
		}
	}

	for _, i := range []int{0, 1, 2, 6} {
		ev := h.events[i]
		if ev.ZGCPauseMarkStartMs <= 0 {
			t.Errorf("minor event[%d] ZGCPauseMarkStartMs should be > 0", i)
		}
		if ev.ZGCConcurrentMarkMs <= 0 {
			t.Errorf("minor event[%d] ZGCConcurrentMarkMs should be > 0", i)
		}
		if ev.ZGCPauseMarkEndMs <= 0 {
			t.Errorf("minor event[%d] ZGCPauseMarkEndMs should be > 0", i)
		}
		if ev.ZGCPauseRelocateMs <= 0 {
			t.Errorf("minor event[%d] ZGCPauseRelocateMs should be > 0", i)
		}
		if ev.ZGCConcurrentRelocMs <= 0 {
			t.Errorf("minor event[%d] ZGCConcurrentRelocMs should be > 0", i)
		}
		if ev.ZGCConcurrentSelectRelocMs <= 0 {
			t.Errorf("minor event[%d] ZGCConcurrentSelectRelocMs should be > 0", i)
		}
	}

	for _, i := range []int{3, 4, 5, 7} {
		ev := h.events[i]
		if ev.ZGCConcurrentMarkMs <= 0 {
			t.Errorf("major event[%d] ZGCConcurrentMarkMs should be > 0", i)
		}
		if ev.ZGCPauseMarkEndMs <= 0 {
			t.Errorf("major event[%d] ZGCPauseMarkEndMs should be > 0", i)
		}
		if ev.ZGCConcurrentProcessNSRMs <= 0 {
			t.Errorf("major event[%d] ZGCConcurrentProcessNSRMs should be > 0", i)
		}
		if ev.ZGCConcurrentRemapRootsMs <= 0 {
			t.Errorf("major event[%d] ZGCConcurrentRemapRootsMs should be > 0", i)
		}
		if ev.ZGCConcurrentRelocMs <= 0 {
			t.Errorf("major event[%d] ZGCConcurrentRelocMs should be > 0", i)
		}
	}

	if h.events[2].ZGCConcurrentMarkContinueMs <= 0 {
		t.Errorf("event[2] ZGCConcurrentMarkContinueMs should be > 0")
	}

	for i, ev := range h.events {
		if ev.ZGCLoad1m <= 0 {
			t.Errorf("event[%d] ZGCLoad1m should be > 0", i)
		}
		if ev.ZGCMMU2ms <= 0 {
			t.Errorf("event[%d] ZGCMMU2ms should be > 0", i)
		}
		if ev.ZGCMetaspaceUsedKB <= 0 {
			t.Errorf("event[%d] ZGCMetaspaceUsedKB should be > 0", i)
		}
	}

	for i, ev := range h.events {
		if !ev.IsPause {
			t.Errorf("event[%d] %s should have IsPause=true", i, ev.Category)
		}
	}

	t.Logf("Parsed %d ZGC Generational JDK21 events:", len(h.events))
	logEvents(t, h.events)
}

// ===================== ZGC Generational Minimal (JDK21) =====================

func TestZGCGenMinimalJDK21(t *testing.T) {
	h := &testHandler{}
	p := NewZGCParser()
	p.SetHandler(h)
	feedFile(t, p, "../../testdata/parser/jdk21-zgc-gen-gconly.log")

	requireEventCount(t, h.events, 8)

	wantCategories := []string{
		"ZGCMinorWarmup", "ZGCMinorAllocRate", "ZGCMinorTimer",
		"ZGCMajorProactive", "ZGCMajorMetadataGC", "ZGCMajorSystemGc",
		"ZGCMinorAllocStall", "ZGCMajorAllocStall",
	}
	for i, ev := range h.events {
		if ev.Category != wantCategories[i] {
			t.Errorf("event[%d] category=%s, want %s", i, ev.Category, wantCategories[i])
		}
	}

	for i, ev := range h.events {
		if ev.Duration <= 0 {
			t.Errorf("event[%d] Duration should be > 0: %f", i, ev.Duration)
		}
		if ev.HeapBeforeKB <= 0 {
			t.Errorf("event[%d] HeapBeforeKB should be > 0: %d", i, ev.HeapBeforeKB)
		}
		if ev.HeapAfterKB <= 0 {
			t.Errorf("event[%d] HeapAfterKB should be > 0: %d", i, ev.HeapAfterKB)
		}
	}

	for i := 0; i < 3; i++ {
		if h.events[i].ZGCCollectionType != "minor" {
			t.Errorf("event[%d] ZGCCollectionType=%q, want \"minor\"", i, h.events[i].ZGCCollectionType)
		}
	}
	for i := 3; i < 6; i++ {
		if h.events[i].ZGCCollectionType != "major" {
			t.Errorf("event[%d] ZGCCollectionType=%q, want \"major\"", i, h.events[i].ZGCCollectionType)
		}
	}
	if h.events[6].ZGCCollectionType != "minor" {
		t.Errorf("event[6] ZGCCollectionType=%q, want \"minor\"", h.events[6].ZGCCollectionType)
	}
	if h.events[7].ZGCCollectionType != "major" {
		t.Errorf("event[7] ZGCCollectionType=%q, want \"major\"", h.events[7].ZGCCollectionType)
	}

	for i, ev := range h.events {
		if ev.ZGCPauseMarkStartMs != 0 {
			t.Errorf("event[%d] ZGCPauseMarkStartMs should be 0 in minimal format", i)
		}
	}

	t.Logf("Parsed %d ZGC Generational Minimal events:", len(h.events))
	logEvents(t, h.events)
}

// ===================== ZGC Generational Interleaved (JDK21) =====================

func TestZGCGenerationalInterleaved(t *testing.T) {
	h := &testHandler{}
	p := NewZGCParser()
	p.SetHandler(h)
	feedFile(t, p, "../../testdata/parser/jdk21-zgc-gen-interleaved.log")

	requireEventCount(t, h.events, 3)

	// GC(0) Minor Warmup emits first (sequential), then GC(2) Minor completes
	// before GC(1) Major because GC(2)'s summary line appears earlier in the log.
	if h.events[0].Category != "ZGCMinorWarmup" {
		t.Errorf("event[0] category=%s, want ZGCMinorWarmup", h.events[0].Category)
	}
	if h.events[1].Category != "ZGCMinorAllocRate" {
		t.Errorf("event[1] category=%s, want ZGCMinorAllocRate", h.events[1].Category)
	}
	if h.events[2].Category != "ZGCMajorProactive" {
		t.Errorf("event[2] category=%s, want ZGCMajorProactive", h.events[2].Category)
	}

	// Verify GC(1) Major was NOT lost due to interleaving
	majorEv := h.events[2]
	if majorEv.ZGCConcurrentMarkMs <= 0 {
		t.Errorf("major event ZGCConcurrentMarkMs should be > 0: %f", majorEv.ZGCConcurrentMarkMs)
	}
	if majorEv.ZGCConcurrentRemapRootsMs <= 0 {
		t.Errorf("major event ZGCConcurrentRemapRootsMs should be > 0: %f", majorEv.ZGCConcurrentRemapRootsMs)
	}
	if majorEv.ZGCPauseRelocateMs <= 0 {
		t.Errorf("major event ZGCPauseRelocateMs should be > 0: %f", majorEv.ZGCPauseRelocateMs)
	}
	if majorEv.HeapBeforeKB <= 0 {
		t.Errorf("major event HeapBeforeKB should be > 0: %d", majorEv.HeapBeforeKB)
	}

	// Verify GC(2) Minor also has correct data
	minorEv := h.events[1]
	if minorEv.ZGCPauseMarkStartMs <= 0 {
		t.Errorf("minor event ZGCPauseMarkStartMs should be > 0: %f", minorEv.ZGCPauseMarkStartMs)
	}
	if minorEv.ZGCConcurrentMarkMs <= 0 {
		t.Errorf("minor event ZGCConcurrentMarkMs should be > 0: %f", minorEv.ZGCConcurrentMarkMs)
	}

	t.Logf("Parsed %d ZGC interleaved events:", len(h.events))
	logEvents(t, h.events)
}

// ===================== ZGC Unified Minimal (JDK11) =====================

func TestZGCUnifiedMinimal(t *testing.T) {
	h := &testHandler{}
	p := NewZGCParser()
	p.SetHandler(h)
	feedFile(t, p, "../../testdata/parser/jdk11-zgc-gconly.log")

	requireEventCount(t, h.events, 7)

	wantCategories := []string{
		"ZGCWarmup", "ZGCAllocRate", "ZGCProactive", "ZGCTimer", "ZGCSystemGc",
		"ZGCAllocStall", "ZGCMetadataGCThreshold",
	}
	for i, ev := range h.events {
		if ev.Category != wantCategories[i] {
			t.Errorf("event[%d] category=%s, want %s", i, ev.Category, wantCategories[i])
		}
	}

	for i, ev := range h.events {
		if ev.Duration <= 0 {
			t.Errorf("event[%d] Duration should be > 0: %f", i, ev.Duration)
		}
		if ev.HeapBeforeKB <= 0 {
			t.Errorf("event[%d] HeapBeforeKB should be > 0: %d", i, ev.HeapBeforeKB)
		}
		if ev.HeapAfterKB <= 0 {
			t.Errorf("event[%d] HeapAfterKB should be > 0: %d", i, ev.HeapAfterKB)
		}
	}

	for i, ev := range h.events {
		if ev.ZGCPauseMarkStartMs != 0 {
			t.Errorf("event[%d] ZGCPauseMarkStartMs should be 0 in minimal format", i)
		}
	}

	t.Logf("Parsed %d ZGC Unified Minimal events:", len(h.events))
	logEvents(t, h.events)
}
