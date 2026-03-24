package parser

import "testing"

// ===================== G1 PreUnified (JDK8) =====================

func TestG1PreUnified(t *testing.T) {
	h := &testHandler{}
	p := NewG1Parser(false)
	p.SetHandler(h)
	feedFile(t, p, "../../testdata/parser/jdk8-g1.log")

	requireEventCount(t, h.events, 13)
	cats := categoryCounts(h.events)

	requireCategory(t, cats, "G1SystemGC")
	requireCategory(t, cats, "G1YoungGC")
	requireCategory(t, cats, "G1InitialMark")
	requireCategory(t, cats, "G1MixedGC")
	requireCategory(t, cats, "G1FullGC")
	requireCategory(t, cats, "G1Remark")
	requireCategory(t, cats, "G1Cleanup")
	requireCategory(t, cats, "G1ConcurrentRootRegionScan")
	requireCategory(t, cats, "G1ConcurrentMark")
	requireCategory(t, cats, "G1ToSpaceExhausted")
	requireCategory(t, cats, "G1HumongousAllocation")
	requireCategory(t, cats, "G1MetadataGC")

	ev0 := h.events[0]
	if ev0.Category != "G1SystemGC" {
		t.Errorf("event[0] category=%s, want G1SystemGC", ev0.Category)
	}
	if ev0.HeapBeforeKB <= 0 || ev0.HeapAfterKB <= 0 || ev0.HeapTotalKB <= 0 {
		t.Errorf("event[0] heap fields must be > 0: before=%d after=%d total=%d",
			ev0.HeapBeforeKB, ev0.HeapAfterKB, ev0.HeapTotalKB)
	}
	if ev0.Duration <= 0 {
		t.Errorf("event[0] duration must be > 0: %f", ev0.Duration)
	}
	if !ev0.IsPause {
		t.Errorf("event[0] Full GC must be a pause event")
	}

	ev1 := h.events[1]
	if ev1.Category != "G1YoungGC" {
		t.Errorf("event[1] category=%s, want G1YoungGC", ev1.Category)
	}
	if ev1.G1EdenBeforeKB <= 0 {
		t.Errorf("event[1] Eden before should be > 0: %d", ev1.G1EdenBeforeKB)
	}
	if ev1.HeapBeforeKB <= 0 || ev1.HeapAfterKB <= 0 {
		t.Errorf("event[1] heap fields should be > 0")
	}

	for _, ev := range h.events {
		if ev.Category == "G1ConcurrentRootRegionScan" || ev.Category == "G1ConcurrentMark" {
			if ev.IsPause {
				t.Errorf("concurrent event %s should not be a pause", ev.Category)
			}
			if ev.Duration <= 0 {
				t.Errorf("concurrent event %s duration must be > 0", ev.Category)
			}
		}
	}

	pauseCategories := map[string]bool{
		"G1SystemGC": true, "G1YoungGC": true, "G1InitialMark": true,
		"G1MixedGC": true, "G1FullGC": true, "G1Remark": true,
		"G1Cleanup": true, "G1ToSpaceExhausted": true,
		"G1HumongousAllocation": true, "G1MetadataGC": true,
	}
	for _, ev := range h.events {
		if pauseCategories[ev.Category] && !ev.IsPause {
			t.Errorf("event %s should be a pause but IsPause=false", ev.Category)
		}
	}

	for i, ev := range h.events {
		if ev.Duration <= 0 {
			t.Errorf("event[%d] %s duration must be > 0: %f", i, ev.Category, ev.Duration)
		}
	}

	t.Logf("Parsed %d G1 PreUnified events:", len(h.events))
	logEvents(t, h.events)
}

// ===================== G1 Unified (JDK11) =====================

func TestG1UnifiedJDK11(t *testing.T) {
	h := &testHandler{}
	p := NewG1Parser(true)
	p.SetHandler(h)
	feedFile(t, p, "../../testdata/parser/jdk11-g1.log")

	requireEventCount(t, h.events, 18)
	cats := categoryCounts(h.events)

	// Unified [gc,heap] lines precede Pause summary; merge must attach region counts.
	ev0 := h.events[0]
	if ev0.G1SurvivorRegionAfter != 2 {
		t.Errorf("event[0] G1SurvivorRegionAfter want 2 got %d", ev0.G1SurvivorRegionAfter)
	}

	requireCategory(t, cats, "G1YoungGC")
	requireCategory(t, cats, "G1ConcurrentStart")
	requireCategory(t, cats, "G1PrepareMixed")
	requireCategory(t, cats, "G1MixedGC")
	requireCategory(t, cats, "G1SystemGC")
	requireCategory(t, cats, "G1FullGC")
	requireCategory(t, cats, "G1Remark")
	requireCategory(t, cats, "G1Cleanup")
	requireCategory(t, cats, "G1ToSpaceExhausted")
	requireCategory(t, cats, "G1ConcurrentScanRootRegions")
	requireCategory(t, cats, "G1ConcurrentMark")
	requireCategory(t, cats, "G1ConcurrentMarkFromRoots")
	requireCategory(t, cats, "G1ConcurrentPreclean")
	requireCategory(t, cats, "G1ConcurrentRebuildRememberedSets")
	requireCategory(t, cats, "G1ConcurrentCleanupforNextMark")
	requireCategory(t, cats, "G1ConcurrentMarkCycle")

	pauseCategories := map[string]bool{
		"G1YoungGC": true, "G1ConcurrentStart": true, "G1PrepareMixed": true,
		"G1MixedGC": true, "G1SystemGC": true, "G1FullGC": true,
		"G1Remark": true, "G1Cleanup": true, "G1ToSpaceExhausted": true,
	}
	for _, ev := range h.events {
		if pauseCategories[ev.Category] {
			if !ev.IsPause {
				t.Errorf("pause event %s should have IsPause=true", ev.Category)
			}
		} else {
			if ev.IsPause {
				t.Errorf("concurrent event %s should have IsPause=false", ev.Category)
			}
		}
		if ev.Duration <= 0 {
			t.Errorf("event %s duration must be > 0", ev.Category)
		}
	}

	for _, ev := range h.events {
		if ev.IsPause && ev.HeapBeforeKB <= 0 {
			t.Errorf("pause event %s HeapBeforeKB should be > 0: %d", ev.Category, ev.HeapBeforeKB)
		}
	}

	t.Logf("Parsed %d G1 Unified JDK11 events:", len(h.events))
	logEvents(t, h.events)
}

// ===================== G1 Unified (JDK17) =====================

func TestG1UnifiedJDK17(t *testing.T) {
	h := &testHandler{}
	p := NewG1Parser(true)
	p.SetHandler(h)
	feedFile(t, p, "../../testdata/parser/jdk17-g1.log")

	cats := categoryCounts(h.events)

	requireCategory(t, cats, "G1YoungGC")
	requireCategory(t, cats, "G1ConcurrentStart")
	requireCategory(t, cats, "G1PrepareMixed")
	requireCategory(t, cats, "G1MixedGC")
	requireCategory(t, cats, "G1SystemGC")
	requireCategory(t, cats, "G1FullGC")
	requireCategory(t, cats, "G1ToSpaceExhausted")
	requireCategory(t, cats, "G1Remark")
	requireCategory(t, cats, "G1Cleanup")
	requireCategory(t, cats, "G1ConcurrentScanRootRegions")
	requireCategory(t, cats, "G1ConcurrentMark")
	requireCategory(t, cats, "G1ConcurrentMarkCycle")

	if len(h.events) < 15 {
		t.Errorf("expected at least 15 events, got %d", len(h.events))
	}

	for i, ev := range h.events {
		if ev.Duration <= 0 {
			t.Errorf("event[%d] %s duration must be > 0: %f", i, ev.Category, ev.Duration)
		}
	}

	t.Logf("Parsed %d G1 Unified JDK17 events:", len(h.events))
	logEvents(t, h.events)
}

// ===================== G1 Pre-Unified Verbose (JDK8) =====================

func TestG1PreUnifiedVerbose(t *testing.T) {
	h := &testHandler{}
	p := NewG1Parser(false)
	p.SetHandler(h)
	feedFile(t, p, "../../testdata/parser/jdk8-g1-verbose.log")

	requireEventCount(t, h.events, 7)

	wantCategories := []string{
		"G1YoungGC", "G1YoungGC", "G1InitialMark", "G1MixedGC", "G1ToSpaceExhausted",
		"G1SystemGC", "G1FullGC",
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
		if !ev.IsPause {
			t.Errorf("event[%d] IsPause should be true for verbose format", i)
		}
	}

	for _, idx := range []int{5, 6} {
		ev := h.events[idx]
		if ev.HeapBeforeKB <= 0 {
			t.Errorf("event[%d] HeapBeforeKB should be > 0: %d", idx, ev.HeapBeforeKB)
		}
		if ev.HeapAfterKB <= 0 {
			t.Errorf("event[%d] HeapAfterKB should be > 0: %d", idx, ev.HeapAfterKB)
		}
	}

	for i := 0; i < 5; i++ {
		if h.events[i].HeapBeforeKB != 0 {
			t.Errorf("event[%d] HeapBeforeKB should be 0 in verbose format: %d", i, h.events[i].HeapBeforeKB)
		}
	}

	t.Logf("Parsed %d G1 Pre-Unified Verbose events:", len(h.events))
	logEvents(t, h.events)
}

// ===================== G1 Unified Minimal (JDK11) =====================

func TestG1UnifiedMinimal(t *testing.T) {
	h := &testHandler{}
	p := NewG1Parser(true)
	p.SetHandler(h)
	feedFile(t, p, "../../testdata/parser/jdk11-g1-gconly.log")

	cats := categoryCounts(h.events)
	t.Logf("Category counts: %v", cats)

	requireCategory(t, cats, "G1YoungGC")
	requireCategory(t, cats, "G1SystemGC")
	requireCategory(t, cats, "G1Remark")
	requireCategory(t, cats, "G1Cleanup")

	for i, ev := range h.events {
		if ev.Duration <= 0 {
			t.Errorf("event[%d] %s Duration should be > 0: %f", i, ev.Category, ev.Duration)
		}
	}

	for i, ev := range h.events {
		if ev.IsPause && ev.HeapBeforeKB <= 0 {
			t.Errorf("event[%d] %s HeapBeforeKB should be > 0: %d", i, ev.Category, ev.HeapBeforeKB)
		}
	}

	t.Logf("Parsed %d G1 Unified Minimal events:", len(h.events))
	logEvents(t, h.events)
}

// ===================== G1 Pre-Unified Uptime-Only (JDK8) =====================

func TestG1PreUnifiedUptimeOnly(t *testing.T) {
	h := &testHandler{}
	p := NewG1Parser(false)
	p.SetHandler(h)
	feedFile(t, p, "../../testdata/parser/jdk8-g1-uptimeonly.log")

	cats := categoryCounts(h.events)
	t.Logf("Category counts: %v", cats)

	requireCategory(t, cats, "G1YoungGC")
	requireCategory(t, cats, "G1InitialMark")
	requireCategory(t, cats, "G1MixedGC")
	requireCategory(t, cats, "G1Remark")
	requireCategory(t, cats, "G1Cleanup")
	requireCategory(t, cats, "G1SystemGC")

	for i, ev := range h.events {
		if ev.Duration <= 0 {
			t.Errorf("event[%d] %s Duration should be > 0: %f", i, ev.Category, ev.Duration)
		}
	}

	for i, ev := range h.events {
		if ev.Timestamp.IsZero() {
			t.Errorf("event[%d] %s Timestamp should not be zero", i, ev.Category)
		}
	}

	for i, ev := range h.events {
		if ev.Category == "G1YoungGC" || ev.Category == "G1InitialMark" || ev.Category == "G1MixedGC" {
			if ev.HeapBeforeKB <= 0 {
				t.Errorf("event[%d] %s HeapBeforeKB should be > 0: %d", i, ev.Category, ev.HeapBeforeKB)
			}
		}
	}

	t.Logf("Parsed %d G1 Pre-Unified Uptime-Only events:", len(h.events))
	logEvents(t, h.events)
}
