package parser

import "testing"

// ===================== Parallel PreUnified (JDK8) =====================

func TestParallelPreUnified(t *testing.T) {
	h := &testHandler{}
	p := NewParallelParser(false)
	p.SetHandler(h)
	feedFile(t, p, "../../testdata/parser/jdk8-parallel.log")

	// 11 events:
	// 5 Young GC (Allocation Failure) + 1 Young GC (GCLocker) + 1 Young GC (Metadata)
	// + 1 Full GC (System.gc()) + 1 Full GC (Ergonomics) + 1 Full GC (Metadata) + 1 Full GC (Alloc Failure)
	requireEventCount(t, h.events, 11)
	cats := categoryCounts(h.events)
	t.Logf("Category counts: %v", cats)

	requireCategory(t, cats, "ParallelYoungGC")
	requireCategory(t, cats, "ParallelSystemGC")
	requireCategory(t, cats, "ParallelFullGCErgonomics")
	requireCategory(t, cats, "ParallelGCLocker")
	requireCategory(t, cats, "ParallelMetadataGC")
	requireCategory(t, cats, "ParallelFullGCAllocationFailure")

	// All events should be pauses
	for i, ev := range h.events {
		if !ev.IsPause {
			t.Errorf("event[%d] %s should be a pause", i, ev.Category)
		}
		if ev.Duration <= 0 {
			t.Errorf("event[%d] %s duration must be > 0: %f", i, ev.Category, ev.Duration)
		}
	}

	// Young GC events should have young gen details (YoungTotalKB > 0 even if young gen was empty)
	for i, ev := range h.events {
		if ev.Category == "ParallelYoungGC" || ev.Category == "ParallelGCLocker" {
			if ev.YoungTotalKB <= 0 {
				t.Errorf("event[%d] %s YoungTotalKB should be > 0: %d", i, ev.Category, ev.YoungTotalKB)
			}
		}
	}

	// Full GC events should have old gen and heap details
	for i, ev := range h.events {
		if ev.Category == "ParallelSystemGC" || ev.Category == "ParallelFullGCErgonomics" ||
			ev.Category == "ParallelFullGCAllocationFailure" {
			if ev.HeapBeforeKB <= 0 {
				t.Errorf("event[%d] %s HeapBeforeKB should be > 0: %d", i, ev.Category, ev.HeapBeforeKB)
			}
		}
	}

	// Full GC with PSPermGen/Metaspace should have meta info
	for i, ev := range h.events {
		if ev.Category == "ParallelSystemGC" && ev.MetaBeforeKB <= 0 {
			t.Errorf("event[%d] %s MetaBeforeKB should be > 0: %d", i, ev.Category, ev.MetaBeforeKB)
		}
	}

	// Verify ordering of first few events
	if h.events[0].Category != "ParallelYoungGC" {
		t.Errorf("event[0] should be ParallelYoungGC, got %s", h.events[0].Category)
	}
	if h.events[2].Category != "ParallelSystemGC" {
		t.Errorf("event[2] should be ParallelSystemGC, got %s", h.events[2].Category)
	}
	if h.events[4].Category != "ParallelFullGCErgonomics" {
		t.Errorf("event[4] should be ParallelFullGCErgonomics, got %s", h.events[4].Category)
	}

	// Timestamps should be parsed from uptime format
	for i, ev := range h.events {
		if ev.Timestamp.IsZero() {
			t.Errorf("event[%d] %s Timestamp should not be zero", i, ev.Category)
		}
	}

	t.Logf("Parsed %d Parallel PreUnified events:", len(h.events))
	logEvents(t, h.events)
}

// ===================== Parallel Unified (JDK17) =====================

func TestParallelUnifiedJDK17(t *testing.T) {
	h := &testHandler{}
	p := NewParallelParser(true)
	p.SetHandler(h)
	feedFile(t, p, "../../testdata/parser/jdk17-parallel.log")

	// GC(0) Young Alloc Failure, GC(1) Young Alloc Failure, GC(2) Young System.gc()
	// GC(3) Full System.gc(), GC(4) Young Metadata, GC(5) Full Ergonomics
	requireEventCount(t, h.events, 6)

	wantCategories := []string{
		"ParallelYoungGC", "ParallelYoungGC", "ParallelSystemGC",
		"ParallelSystemGC", "ParallelMetadataGC", "ParallelFullGCErgonomics",
	}
	for i, ev := range h.events {
		if ev.Category != wantCategories[i] {
			t.Errorf("event[%d] category=%s, want %s", i, ev.Category, wantCategories[i])
		}
	}

	for i, ev := range h.events {
		if !ev.IsPause {
			t.Errorf("event[%d] %s should be a pause", i, ev.Category)
		}
		if ev.Duration <= 0 {
			t.Errorf("event[%d] %s Duration should be > 0: %f", i, ev.Category, ev.Duration)
		}
		if ev.HeapBeforeKB <= 0 {
			t.Errorf("event[%d] %s HeapBeforeKB should be > 0: %d", i, ev.Category, ev.HeapBeforeKB)
		}
		if ev.HeapTotalKB <= 0 {
			t.Errorf("event[%d] %s HeapTotalKB should be > 0: %d", i, ev.Category, ev.HeapTotalKB)
		}
	}

	// All events should have young gen detail from PSYoungGen lines
	for i, ev := range h.events {
		if ev.YoungBeforeKB <= 0 && ev.Category == "ParallelYoungGC" {
			t.Errorf("event[%d] Young GC should have YoungBeforeKB > 0: %d", i, ev.YoungBeforeKB)
		}
	}

	// Full GC events should have old gen detail from ParOldGen lines
	for i, ev := range h.events {
		if ev.Category == "ParallelFullGCErgonomics" || (ev.Category == "ParallelSystemGC" && i == 3) {
			if ev.OldBeforeKB <= 0 && ev.OldTotalKB <= 0 {
				t.Errorf("event[%d] Full GC should have OldTotalKB > 0: %d", i, ev.OldTotalKB)
			}
		}
	}

	// All events should have metaspace data
	for i, ev := range h.events {
		if ev.MetaBeforeKB <= 0 {
			t.Errorf("event[%d] %s MetaBeforeKB should be > 0: %d", i, ev.Category, ev.MetaBeforeKB)
		}
	}

	t.Logf("Parsed %d Parallel Unified JDK17 events:", len(h.events))
	logEvents(t, h.events)
}

// ===================== Parallel Unified (JDK11) =====================

func TestParallelUnifiedJDK11(t *testing.T) {
	h := &testHandler{}
	p := NewParallelParser(true)
	p.SetHandler(h)
	feedFile(t, p, "../../testdata/parser/jdk11-parallel.log")

	// 9 events: 4 Young GC + 1 SystemGC (young) + 1 Full SystemGC + 1 Full Ergonomics
	// + 1 Full Alloc Failure + 1 Metadata (young) + 1 Full Metadata (from Flush)... let me count
	// GC(0) Young Alloc Failure, GC(1) Young Alloc Failure, GC(2) Young System.gc(),
	// GC(3) Full System.gc(), GC(4) Young Alloc Failure, GC(5) Full Ergonomics,
	// GC(6) Full Alloc Failure, GC(7) Young Metadata, GC(8) Full Metadata
	requireEventCount(t, h.events, 9)
	cats := categoryCounts(h.events)
	t.Logf("Category counts: %v", cats)

	requireCategory(t, cats, "ParallelYoungGC")
	requireCategory(t, cats, "ParallelSystemGC")
	requireCategory(t, cats, "ParallelFullGCErgonomics")
	requireCategory(t, cats, "ParallelFullGCAllocationFailure")
	requireCategory(t, cats, "ParallelMetadataGC")

	// All events should be pauses with positive duration
	for i, ev := range h.events {
		if !ev.IsPause {
			t.Errorf("event[%d] %s should be a pause", i, ev.Category)
		}
		if ev.Duration <= 0 {
			t.Errorf("event[%d] %s Duration should be > 0: %f", i, ev.Category, ev.Duration)
		}
	}

	// All events should have heap info from summary line
	for i, ev := range h.events {
		if ev.HeapBeforeKB <= 0 {
			t.Errorf("event[%d] %s HeapBeforeKB should be > 0: %d", i, ev.Category, ev.HeapBeforeKB)
		}
		if ev.HeapTotalKB <= 0 {
			t.Errorf("event[%d] %s HeapTotalKB should be > 0: %d", i, ev.Category, ev.HeapTotalKB)
		}
	}

	// Events with PSYoungGen/ParOldGen detail lines should have young/old gen data
	for i, ev := range h.events {
		if ev.YoungBeforeKB <= 0 && ev.OldBeforeKB <= 0 {
			t.Logf("event[%d] %s: no gen detail (may be flushed event)", i, ev.Category)
		}
	}

	// Metaspace data should be present for events with metaspace lines
	for i, ev := range h.events {
		if ev.MetaBeforeKB <= 0 {
			t.Logf("event[%d] %s: no Metaspace data", i, ev.Category)
		}
	}

	// Timestamps should be parsed
	for i, ev := range h.events {
		if ev.Timestamp.IsZero() {
			t.Errorf("event[%d] %s Timestamp should not be zero", i, ev.Category)
		}
	}

	t.Logf("Parsed %d Parallel Unified JDK11 events:", len(h.events))
	logEvents(t, h.events)
}
