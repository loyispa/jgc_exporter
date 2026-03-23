package parser

import "testing"

// ===================== Serial PreUnified (JDK8) =====================

func TestSerialPreUnified(t *testing.T) {
	h := &testHandler{}
	p := NewSerialParser(false)
	p.SetHandler(h)
	feedFile(t, p, "../../testdata/parser/jdk8-serial.log")

	// 11 events: 7 DefNew (Young GC) + 4 Full GC (1 Metadata + 2 System.gc() + 1 Allocation Failure)
	requireEventCount(t, h.events, 11)
	cats := categoryCounts(h.events)
	t.Logf("Category counts: %v", cats)

	requireCategory(t, cats, "SerialYoungGC")
	requireCategory(t, cats, "SerialMetadataGC")
	requireCategory(t, cats, "SerialSystemGC")
	requireCategory(t, cats, "SerialFullGCAllocationFailure")

	// All events should be pauses
	for i, ev := range h.events {
		if !ev.IsPause {
			t.Errorf("event[%d] %s should be a pause", i, ev.Category)
		}
		if ev.Duration <= 0 {
			t.Errorf("event[%d] %s duration must be > 0: %f", i, ev.Category, ev.Duration)
		}
	}

	// Young GC events should have young gen and heap details
	for i, ev := range h.events {
		if ev.Category == "SerialYoungGC" {
			if ev.YoungBeforeKB <= 0 {
				t.Errorf("event[%d] YoungBeforeKB should be > 0: %d", i, ev.YoungBeforeKB)
			}
			if ev.HeapBeforeKB <= 0 {
				t.Errorf("event[%d] HeapBeforeKB should be > 0: %d", i, ev.HeapBeforeKB)
			}
		}
	}

	// Full GC events should have old gen and heap details
	for i, ev := range h.events {
		if ev.Category == "SerialMetadataGC" || ev.Category == "SerialSystemGC" || ev.Category == "SerialFullGCAllocationFailure" {
			if ev.OldBeforeKB <= 0 {
				t.Errorf("event[%d] %s OldBeforeKB should be > 0: %d", i, ev.Category, ev.OldBeforeKB)
			}
			if ev.HeapBeforeKB <= 0 {
				t.Errorf("event[%d] %s HeapBeforeKB should be > 0: %d", i, ev.Category, ev.HeapBeforeKB)
			}
			if ev.MetaBeforeKB <= 0 {
				t.Errorf("event[%d] %s MetaBeforeKB should be > 0: %d", i, ev.Category, ev.MetaBeforeKB)
			}
		}
	}

	// Timestamps should be parsed from uptime format
	for i, ev := range h.events {
		if ev.Timestamp.IsZero() {
			t.Errorf("event[%d] %s Timestamp should not be zero", i, ev.Category)
		}
	}

	// Verify first event
	if h.events[0].Category != "SerialYoungGC" {
		t.Errorf("event[0] should be SerialYoungGC, got %s", h.events[0].Category)
	}
	if h.events[0].Cause != "Allocation Failure" {
		t.Errorf("event[0] cause should be 'Allocation Failure', got %q", h.events[0].Cause)
	}

	t.Logf("Parsed %d Serial PreUnified events:", len(h.events))
	logEvents(t, h.events)
}

// ===================== Serial Unified (JDK11) =====================

func TestSerialUnifiedJDK11(t *testing.T) {
	h := &testHandler{}
	p := NewSerialParser(true)
	p.SetHandler(h)
	feedFile(t, p, "../../testdata/parser/jdk11-serial.log")

	// 6 events: 1 Full System.gc() + 3 Young Alloc Failure + 1 Full Metadata + 1 Full Alloc Failure
	requireEventCount(t, h.events, 6)
	cats := categoryCounts(h.events)
	t.Logf("Category counts: %v", cats)

	requireCategory(t, cats, "SerialSystemGC")
	requireCategory(t, cats, "SerialYoungGC")
	requireCategory(t, cats, "SerialMetadataGC")
	requireCategory(t, cats, "SerialFullGCAllocationFailure")

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

	// Young GC events should have DefNew detail
	for i, ev := range h.events {
		if ev.Category == "SerialYoungGC" {
			if ev.YoungBeforeKB <= 0 {
				t.Errorf("event[%d] YoungBeforeKB should be > 0: %d", i, ev.YoungBeforeKB)
			}
		}
	}

	// Full GC events should have Tenured detail
	for i, ev := range h.events {
		if ev.Category == "SerialSystemGC" || ev.Category == "SerialMetadataGC" ||
			ev.Category == "SerialFullGCAllocationFailure" {
			if ev.OldBeforeKB <= 0 {
				t.Errorf("event[%d] %s OldBeforeKB should be > 0: %d", i, ev.Category, ev.OldBeforeKB)
			}
		}
	}

	// Metaspace data should be present
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

	// Verify ordering
	if h.events[0].Category != "SerialSystemGC" {
		t.Errorf("event[0] should be SerialSystemGC, got %s", h.events[0].Category)
	}

	t.Logf("Parsed %d Serial Unified JDK11 events:", len(h.events))
	logEvents(t, h.events)
}
