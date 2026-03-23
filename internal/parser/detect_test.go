package parser

import "testing"

func TestDetectGCType(t *testing.T) {
	tests := []struct {
		file     string
		expected GCType
	}{
		// G1
		{"../../testdata/parser/jdk8-g1.log", GCTypeG1},
		{"../../testdata/parser/jdk8-g1-verbose.log", GCTypeG1},
		{"../../testdata/parser/jdk8-g1-uptimeonly.log", GCTypeG1},
		{"../../testdata/parser/jdk11-g1.log", GCTypeG1},
		{"../../testdata/parser/jdk11-g1-gconly.log", GCTypeG1},
		{"../../testdata/parser/jdk17-g1.log", GCTypeG1},
		{"../../testdata/parser/jdk21-g1.log", GCTypeG1},
		// CMS
		{"../../testdata/parser/jdk8-cms.log", GCTypeCMS},
		{"../../testdata/parser/jdk8-cms-verbose.log", GCTypeCMS},
		{"../../testdata/parser/jdk11-cms.log", GCTypeCMS},
		{"../../testdata/parser/jdk11-cms-gconly.log", GCTypeCMS},
		// ZGC
		{"../../testdata/parser/jdk11-zgc.log", GCTypeZGC},
		{"../../testdata/parser/jdk11-zgc-gconly.log", GCTypeZGC},
		{"../../testdata/parser/jdk17-zgc.log", GCTypeZGC},
		{"../../testdata/parser/jdk17-zgc-gconly.log", GCTypeZGC},
		{"../../testdata/parser/jdk21-zgc-generational.log", GCTypeZGC},
		{"../../testdata/parser/jdk21-zgc-gen-gconly.log", GCTypeZGC},
		// Parallel
		{"../../testdata/parser/jdk8-parallel.log", GCTypeParallel},
		{"../../testdata/parser/jdk11-parallel.log", GCTypeParallel},
		{"../../testdata/parser/jdk11-parallel-psoldgen.log", GCTypeParallel},
		// Serial
		{"../../testdata/parser/jdk8-serial.log", GCTypeSerial},
		{"../../testdata/parser/jdk11-serial.log", GCTypeSerial},
		{"../../testdata/parser/jdk17-serial.log", GCTypeSerial},
	}
	for _, tt := range tests {
		t.Run(tt.file, func(t *testing.T) {
			got := DetectGCType(tt.file)
			if got != tt.expected {
				t.Errorf("DetectGCType(%s) = %s, want %s", tt.file, got, tt.expected)
			}
		})
	}
}

func TestDetectGCTypeNonexistent(t *testing.T) {
	got := DetectGCType("/nonexistent/path.log")
	if got != GCTypeUnknown {
		t.Errorf("expected Unknown, got %s", got)
	}
}

func TestDetectUnifiedLogging(t *testing.T) {
	tests := []struct {
		file string
		want bool
	}{
		{"../../testdata/parser/jdk8-g1.log", false},
		{"../../testdata/parser/jdk11-g1.log", true},
		{"../../testdata/parser/jdk11-zgc.log", true},
		{"../../testdata/parser/jdk8-cms.log", false},
		{"../../testdata/parser/jdk11-cms.log", true},
		{"../../testdata/parser/jdk8-parallel.log", false},
		{"../../testdata/parser/jdk11-parallel.log", true},
		{"../../testdata/parser/jdk8-serial.log", false},
		{"../../testdata/parser/jdk11-serial.log", true},
	}
	for _, tt := range tests {
		t.Run(tt.file, func(t *testing.T) {
			got := DetectUnifiedLogging(tt.file)
			if got != tt.want {
				t.Errorf("DetectUnifiedLogging(%s) = %v, want %v", tt.file, got, tt.want)
			}
		})
	}
}

func TestDetectUnifiedLoggingNonexistent(t *testing.T) {
	if DetectUnifiedLogging("/nonexistent") {
		t.Error("expected false for nonexistent file")
	}
}

func TestDetectGenerationalZGC(t *testing.T) {
	tests := []struct {
		file string
		want bool
	}{
		{"../../testdata/parser/jdk21-zgc-generational.log", true},
		{"../../testdata/parser/jdk21-zgc-gen-gconly.log", true},
		{"../../testdata/parser/jdk11-zgc.log", false},
		{"../../testdata/parser/jdk17-zgc.log", false},
	}
	for _, tt := range tests {
		t.Run(tt.file, func(t *testing.T) {
			got := DetectGenerationalZGC(tt.file)
			if got != tt.want {
				t.Errorf("DetectGenerationalZGC(%s) = %v, want %v", tt.file, got, tt.want)
			}
		})
	}
}

func TestDetectGenerationalZGCNonexistent(t *testing.T) {
	if DetectGenerationalZGC("/nonexistent") {
		t.Error("expected false for nonexistent file")
	}
}

func TestDetectFromLine(t *testing.T) {
	tests := []struct {
		line string
		want GCType
	}{
		{"Using G1", GCTypeG1},
		{"-XX:+UseG1GC", GCTypeG1},
		{"Using Concurrent Mark Sweep", GCTypeCMS},
		{"-XX:+UseConcMarkSweepGC", GCTypeCMS},
		{"[ParNew: 1000K->500K(2000K)", GCTypeCMS},
		{"Using Parallel", GCTypeParallel},
		{"-XX:+UseParallelGC", GCTypeParallel},
		{"[PSYoungGen: 1000K->500K", GCTypeParallel},
		{"PSYoungGen: data", GCTypeParallel},
		{"Using Serial", GCTypeSerial},
		{"-XX:+UseSerialGC", GCTypeSerial},
		{"[DefNew: data", GCTypeSerial},
		{"DefNew: data", GCTypeSerial},
		{"Using the Z Garbage Collector", GCTypeZGC},
		{"-XX:+UseZGC", GCTypeZGC},
		{"random log line", GCTypeUnknown},
	}
	for _, tt := range tests {
		got := detectFromLine(tt.line)
		if got != tt.want {
			t.Errorf("detectFromLine(%q) = %s, want %s", tt.line, got, tt.want)
		}
	}
}
