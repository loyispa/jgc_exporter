package parser

import (
	"os"
	"path/filepath"
	"testing"
)

type testHandler struct {
	events []*GCEvent
}

func (h *testHandler) Handle(ev *GCEvent) {
	if ev != nil {
		h.events = append(h.events, ev)
	}
}

func feedFile(t *testing.T, p Parser, path string) {
	t.Helper()
	abs, err := filepath.Abs(path)
	if err != nil {
		t.Fatalf("abs path: %v", err)
	}
	data, err := os.ReadFile(abs)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	lines := splitLines(string(data))
	for _, line := range lines {
		p.Feed(line)
	}
	p.Flush()
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i <= len(s); i++ {
		if i == len(s) || s[i] == '\n' {
			line := s[start:i]
			if len(line) > 0 && line[len(line)-1] == '\r' {
				line = line[:len(line)-1]
			}
			lines = append(lines, line)
			start = i + 1
		}
	}
	return lines
}

func requireEventCount(t *testing.T, events []*GCEvent, want int) {
	t.Helper()
	if got := len(events); got != want {
		t.Errorf("event count: got %d, want %d", got, want)
	}
}

func categoryCounts(events []*GCEvent) map[string]int {
	m := make(map[string]int)
	for _, ev := range events {
		if ev != nil && ev.Category != "" {
			m[ev.Category]++
		}
	}
	return m
}

func requireCategory(t *testing.T, cats map[string]int, name string) {
	t.Helper()
	if cats[name] == 0 {
		t.Errorf("expected at least one event with category %q", name)
	}
}

func logEvents(t *testing.T, events []*GCEvent) {
	t.Helper()
	for i, ev := range events {
		if ev != nil {
			t.Logf("event[%d]: category=%s duration=%.3f isPause=%v heap=%d->%d",
				i, ev.Category, ev.Duration, ev.IsPause, ev.HeapBeforeKB, ev.HeapAfterKB)
		}
	}
}
