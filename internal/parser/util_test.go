package parser

import (
	"strings"
	"testing"
	"time"
)

func TestParsePreUnifiedTimestamp(t *testing.T) {
	ts, ok := parsePreUnifiedTimestamp("2024-01-03T15:00:00.100+0800: 10.5: [GC")
	if !ok {
		t.Fatal("expected ok")
	}
	if ts.Year() != 2024 || ts.Month() != time.January || ts.Day() != 3 {
		t.Fatalf("unexpected date: %v", ts)
	}

	ts, ok = parsePreUnifiedTimestamp("1.234: [GC pause")
	if !ok {
		t.Fatal("expected ok for uptime format")
	}
	if ts.IsZero() {
		t.Fatal("expected non-zero time")
	}

	_, ok = parsePreUnifiedTimestamp("garbage line")
	if ok {
		t.Fatal("expected false for garbage")
	}
}

func TestStripUnifiedJVMLogPrefix(t *testing.T) {
	raw := `[2024-01-03T15:00:01.000+0800][12345][67890][info][gc,start     ] GC(0) Pause Young (Normal) (G1 Evacuation Pause) 12M->1M(30M) 8.923ms`
	payload, ok := StripUnifiedJVMLogPrefix(raw)
	if !ok {
		t.Fatal("expected ok")
	}
	want := `GC(0) Pause Young (Normal) (G1 Evacuation Pause) 12M->1M(30M) 8.923ms`
	if payload != want {
		t.Fatalf("payload = %q, want %q", payload, want)
	}
	_, ok = StripUnifiedJVMLogPrefix("not unified")
	if ok {
		t.Fatal("expected false for non-unified")
	}
}

func TestSyntheticUnifiedGCLine(t *testing.T) {
	raw := `[2024-01-03T15:00:01.000+0800][info][gc] GC(1) Pause 1ms`
	payload := `GC(1) Pause 1ms`
	synth := SyntheticUnifiedGCLine(raw, payload)
	if synth != `[2024-01-03T15:00:01.000+0800][gc] GC(1) Pause 1ms` {
		t.Fatalf("unexpected synth: %s", synth)
	}
	ts, ok := parseUnifiedTimestamp(synth)
	if !ok || ts.Year() != 2024 {
		t.Fatalf("timestamp parse failed: ok=%v ts=%v", ok, ts)
	}
	if !strings.Contains(synth, "][gc]") {
		t.Fatal("expected ][gc] for parser gate")
	}
}

func TestParseUnifiedTimestamp(t *testing.T) {
	ts, ok := parseUnifiedTimestamp("[2024-01-03T15:00:00.100+0800][gc,start] GC(0)")
	if !ok {
		t.Fatal("expected ok")
	}
	if ts.Year() != 2024 {
		t.Fatalf("unexpected year: %d", ts.Year())
	}

	_, ok = parseUnifiedTimestamp("no bracket")
	if ok {
		t.Fatal("expected false")
	}

	_, ok = parseUnifiedTimestamp("[bad-timestamp]")
	if ok {
		t.Fatal("expected false for bad format")
	}
}

func TestParseSizeKB(t *testing.T) {
	cases := []struct {
		input string
		want  int64
	}{
		{"1024K", 1024},
		{"1024k", 1024},
		{"10M", 10240},
		{"10m", 10240},
		{"1G", 1048576},
		{"1g", 1048576},
		{"1024B", 1},
		{"1024b", 1},
		{"512", 512},
		{"", 0},
		{"  ", 0},
		{"abc", 0},
		{"10.5M", 10752},
	}
	for _, c := range cases {
		got := parseSizeKB(c.input)
		if got != c.want {
			t.Errorf("parseSizeKB(%q) = %d, want %d", c.input, got, c.want)
		}
	}
}

func TestParseFloat(t *testing.T) {
	if parseFloat("1.5") != 1.5 {
		t.Fatal("expected 1.5")
	}
	if parseFloat("  2.0  ") != 2.0 {
		t.Fatal("expected 2.0")
	}
	if parseFloat("bad") != 0 {
		t.Fatal("expected 0 for bad input")
	}
}

func TestParseInt(t *testing.T) {
	if parseInt("42") != 42 {
		t.Fatal("expected 42")
	}
	if parseInt("bad") != 0 {
		t.Fatal("expected 0 for bad input")
	}
}

func TestExtractBetween(t *testing.T) {
	if extractBetween("hello [world] end", "[", "]") != "world" {
		t.Fatal("expected 'world'")
	}
	if extractBetween("no brackets", "[", "]") != "" {
		t.Fatal("expected empty for no match")
	}
	if extractBetween("hello [world", "[", "]") != "world" {
		t.Fatal("expected 'world' when no right bracket")
	}
}

func TestIsVerifyDiagnosticLine(t *testing.T) {
	cases := []struct {
		input string
		want  bool
	}{
		{"VerifyBeforeGC:[Verifying threads]", true},
		{"VerifyAfterGC:[Verifying threads]", true},
		{"[Verifying before GC]", true},
		{"[Verifying after GC]", true},
		{"  [Verifying before GC]  ", true},
		{"[GC pause (G1 Evacuation Pause) (young), 0.0089230 secs]", false},
		{"Total time for which application threads were stopped: 0.01 seconds", false},
		{"", false},
	}
	for _, c := range cases {
		got := isVerifyDiagnosticLine(c.input)
		if got != c.want {
			t.Errorf("isVerifyDiagnosticLine(%q) = %v, want %v", c.input, got, c.want)
		}
	}
}

func TestExtractAfter(t *testing.T) {
	if extractAfter("prefix: value", "prefix: ") != "value" {
		t.Fatal("expected 'value'")
	}
	if extractAfter("no match", "prefix: ") != "" {
		t.Fatal("expected empty")
	}
}
