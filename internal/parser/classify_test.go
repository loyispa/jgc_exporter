package parser

import "testing"

func TestClassifyG1FullCause(t *testing.T) {
	tests := []struct {
		cause string
		want  string
	}{
		{"System.gc()", "G1SystemGC"},
		{"Allocation Failure", "G1FullGC"},
		{"Metadata GC Threshold", "G1MetadataGC"},
		{"Unknown Cause", "G1FullGC"},
	}
	for _, tt := range tests {
		got := classifyG1FullCause(tt.cause)
		if got != tt.want {
			t.Errorf("classifyG1FullCause(%q) = %q, want %q", tt.cause, got, tt.want)
		}
	}
}

func TestClassifyG1PreConcurrent(t *testing.T) {
	tests := []struct {
		line string
		want string
	}{
		{"[GC concurrent-root-region-scan-start]", "G1ConcurrentRootRegionScan"},
		{"[GC concurrent-mark-start]", "G1ConcurrentMark"},
		{"[GC concurrent-cleanup-start]", "G1ConcurrentCleanup"},
		{"[GC concurrent-unknown]", "G1Concurrent"},
	}
	for _, tt := range tests {
		got := classifyG1PreConcurrent(tt.line)
		if got != tt.want {
			t.Errorf("classifyG1PreConcurrent(%q) = %q, want %q", tt.line, got, tt.want)
		}
	}
}

func TestClassifyCMSConcurrent(t *testing.T) {
	tests := []struct {
		phase string
		want  string
	}{
		{"CMS-concurrent-mark", "CMSConcurrentMark"},
		{"CMS-concurrent-preclean", "CMSConcurrentPreClean"},
		{"CMS-concurrent-sweep", "CMSConcurrentSweep"},
		{"CMS-concurrent-reset", "CMSConcurrentReset"},
		{"CMS-concurrent-abortable-preclean", "CMSAbortablePreClean"},
		{"CMS-concurrent-unknown", "CMSConcurrent"},
	}
	for _, tt := range tests {
		got := classifyCMSConcurrent(tt.phase)
		if got != tt.want {
			t.Errorf("classifyCMSConcurrent(%q) = %q, want %q", tt.phase, got, tt.want)
		}
	}
}

func TestClassifyCMSSimpleCause(t *testing.T) {
	tests := []struct {
		cause string
		want  string
	}{
		{"GCLocker Initiated GC", "ParNew"},
		{"Allocation Failure", "ParNew"},
		{"System.gc()", "ParNew"},
	}
	for _, tt := range tests {
		got := classifyCMSSimpleCause(tt.cause)
		if got != tt.want {
			t.Errorf("classifyCMSSimpleCause(%q) = %q, want %q", tt.cause, got, tt.want)
		}
	}
}

func TestClassifyCMSFullCause(t *testing.T) {
	tests := []struct {
		cause string
		want  string
	}{
		{"Allocation Failure", "FullGC"},
		{"System.gc()", "SystemGC"},
		{"Metadata GC Threshold", "MetadataGC"},
		{"Unknown", "FullGC"},
	}
	for _, tt := range tests {
		got := classifyCMSFullCause(tt.cause)
		if got != tt.want {
			t.Errorf("classifyCMSFullCause(%q) = %q, want %q", tt.cause, got, tt.want)
		}
	}
}

func TestClassifySerialYoungCause(t *testing.T) {
	tests := []struct {
		cause string
		want  string
	}{
		{"System.gc()", "SerialSystemGC"},
		{"Metadata GC Threshold", "SerialMetadataGC"},
		{"Allocation Failure", "SerialYoungGC"},
		{"", "SerialYoungGC"},
	}
	for _, tt := range tests {
		got := classifySerialYoungCause(tt.cause)
		if got != tt.want {
			t.Errorf("classifySerialYoungCause(%q) = %q, want %q", tt.cause, got, tt.want)
		}
	}
}

func TestClassifySerialFullCause(t *testing.T) {
	tests := []struct {
		cause string
		want  string
	}{
		{"System.gc()", "SerialSystemGC"},
		{"Metadata GC Threshold", "SerialMetadataGC"},
		{"Allocation Failure", "SerialFullGCAllocationFailure"},
		{"Ergonomics", "SerialFullGCErgonomics"},
		{"Unknown", "SerialFullGC"},
	}
	for _, tt := range tests {
		got := classifySerialFullCause(tt.cause)
		if got != tt.want {
			t.Errorf("classifySerialFullCause(%q) = %q, want %q", tt.cause, got, tt.want)
		}
	}
}

func TestClassifyParallelYoungCause(t *testing.T) {
	tests := []struct {
		cause string
		want  string
	}{
		{"System.gc()", "ParallelSystemGC"},
		{"Metadata GC Threshold", "ParallelMetadataGC"},
		{"Ergonomics", "ParallelErgonomics"},
		{"GCLocker Initiated GC", "ParallelGCLocker"},
		{"Allocation Failure", "ParallelYoungGC"},
	}
	for _, tt := range tests {
		got := classifyParallelYoungCause(tt.cause)
		if got != tt.want {
			t.Errorf("classifyParallelYoungCause(%q) = %q, want %q", tt.cause, got, tt.want)
		}
	}
}

func TestClassifyParallelFullCause(t *testing.T) {
	tests := []struct {
		cause string
		want  string
	}{
		{"System.gc()", "ParallelSystemGC"},
		{"Metadata GC Threshold", "ParallelMetadataGC"},
		{"Ergonomics", "ParallelFullGCErgonomics"},
		{"Allocation Failure", "ParallelFullGCAllocationFailure"},
		{"Unknown", "ParallelFullGC"},
	}
	for _, tt := range tests {
		got := classifyParallelFullCause(tt.cause)
		if got != tt.want {
			t.Errorf("classifyParallelFullCause(%q) = %q, want %q", tt.cause, got, tt.want)
		}
	}
}

func TestClassifyZGCCause(t *testing.T) {
	tests := []struct {
		cause string
		want  string
	}{
		{"Warmup", "ZGCWarmup"},
		{"Allocation Rate", "ZGCAllocRate"},
		{"Proactive", "ZGCProactive"},
		{"Timer", "ZGCTimer"},
		{"System.gc()", "ZGCSystemGc"},
		{"Allocation Stall", "ZGCAllocStall"},
		{"Metadata GC Threshold", "ZGCMetadataGCThreshold"},
		{"Unknown", "ZGCUnknown"},
	}
	for _, tt := range tests {
		got := classifyZGCCause(tt.cause)
		if got != tt.want {
			t.Errorf("classifyZGCCause(%q) = %q, want %q", tt.cause, got, tt.want)
		}
	}
}

func TestClassifyGenZGCCause(t *testing.T) {
	tests := []struct {
		collType string
		cause    string
		want     string
	}{
		{"minor", "Warmup", "ZGCMinorWarmup"},
		{"minor", "Allocation Rate", "ZGCMinorAllocRate"},
		{"minor", "Timer", "ZGCMinorTimer"},
		{"minor", "Allocation Stall", "ZGCMinorAllocStall"},
		{"minor", "Proactive", "ZGCMinorProactive"},
		{"minor", "Metadata GC Threshold", "ZGCMinorMetadataGC"},
		{"minor", "System.gc()", "ZGCMinorSystemGc"},
		{"minor", "Unknown Cause", "ZGCMinorUnknown"},
		{"major", "Proactive", "ZGCMajorProactive"},
		{"major", "System.gc()", "ZGCMajorSystemGc"},
		{"major", "Metadata GC Threshold", "ZGCMajorMetadataGC"},
		{"major", "Allocation Stall", "ZGCMajorAllocStall"},
		{"major", "Unknown Cause", "ZGCMajorUnknown"},
		// empty collType defaults to "minor" prefix
		{"", "Warmup", "ZGCMinorWarmup"},
	}
	for _, tt := range tests {
		got := classifyGenZGCCause(tt.collType, tt.cause)
		if got != tt.want {
			t.Errorf("classifyGenZGCCause(%q, %q) = %q, want %q", tt.collType, tt.cause, got, tt.want)
		}
	}
}

func TestClassifyUnifiedG1Pause(t *testing.T) {
	tests := []struct {
		desc string
		want string
	}{
		{"Pause Young (Normal) (G1 Evacuation Pause)", "G1YoungGC"},
		{"Pause Young (Concurrent Start)", "G1ConcurrentStart"},
		{"Pause Young (Prepare Mixed)", "G1PrepareMixed"},
		{"Pause Mixed", "G1MixedGC"},
		{"Pause Full (G1 Compaction Pause)", "G1FullGC"},
		{"Pause Full (System.gc())", "G1SystemGC"},
		{"Pause Full (Metadata GC Threshold)", "G1FullGC"},
		{"Pause Remark", "G1Remark"},
		{"Pause Cleanup", "G1Cleanup"},
		{"Unknown Phase", "G1Unknown"},
	}
	for _, tt := range tests {
		got := classifyUnifiedG1Pause(tt.desc)
		if got != tt.want {
			t.Errorf("classifyUnifiedG1Pause(%q) = %q, want %q", tt.desc, got, tt.want)
		}
	}
}

func TestClassifyUnifiedCMSPause(t *testing.T) {
	tests := []struct {
		desc string
		want string
	}{
		{"Pause Young (Allocation Failure)", "ParNew"},
		{"Pause Young (GCLocker Initiated GC)", "ParNew"},
		{"Pause Initial Mark", "CMSInitialMark"},
		{"Pause Remark", "CMSRemark"},
		{"Pause Full (Allocation Failure)", "FullGC"},
		{"Pause Full (System.gc())", "SystemGC"},
		{"Pause Full (Metadata GC Threshold)", "MetadataGC"},
		{"Unknown Phase", "CMSUnknown"},
	}
	for _, tt := range tests {
		got := classifyUnifiedCMSPause(tt.desc)
		if got != tt.want {
			t.Errorf("classifyUnifiedCMSPause(%q) = %q, want %q", tt.desc, got, tt.want)
		}
	}
}

func TestClassifyUnifiedCMSConcurrent(t *testing.T) {
	tests := []struct {
		desc string
		want string
	}{
		{"Concurrent Mark", "CMSConcurrentMark"},
		{"Concurrent Preclean", "CMSConcurrentPreClean"},
		{"Concurrent Sweep", "CMSConcurrentSweep"},
		{"Concurrent Reset", "CMSConcurrentReset"},
		{"Concurrent Abortable Preclean", "CMSAbortablePreClean"},
		{"Unknown Phase", "CMSConcurrent"},
	}
	for _, tt := range tests {
		got := classifyUnifiedCMSConcurrent(tt.desc)
		if got != tt.want {
			t.Errorf("classifyUnifiedCMSConcurrent(%q) = %q, want %q", tt.desc, got, tt.want)
		}
	}
}

func TestClassifyUnifiedParallelPause(t *testing.T) {
	tests := []struct {
		desc string
		want string
	}{
		{"Pause Young (Allocation Failure)", "ParallelYoungGC"},
		{"Pause Young (System.gc())", "ParallelSystemGC"},
		{"Pause Young (Metadata GC Threshold)", "ParallelMetadataGC"},
		{"Pause Young (GCLocker Initiated GC)", "ParallelYoungGC"},
		{"Pause Full (Ergonomics)", "ParallelFullGCErgonomics"},
		{"Pause Full (System.gc())", "ParallelSystemGC"},
		{"Pause Full (Metadata GC Threshold)", "ParallelMetadataGC"},
		{"Pause Full (Allocation Failure)", "ParallelFullGCAllocationFailure"},
		{"Unknown Phase", "ParallelUnknown"},
	}
	for _, tt := range tests {
		got := classifyUnifiedParallelPause(tt.desc)
		if got != tt.want {
			t.Errorf("classifyUnifiedParallelPause(%q) = %q, want %q", tt.desc, got, tt.want)
		}
	}
}

func TestClassifyUnifiedSerialPause(t *testing.T) {
	tests := []struct {
		desc string
		want string
	}{
		{"Pause Young (Allocation Failure)", "SerialYoungGC"},
		{"Pause Young (System.gc())", "SerialSystemGC"},
		{"Pause Young (Metadata GC Threshold)", "SerialYoungGC"},
		{"Pause Full (System.gc())", "SerialSystemGC"},
		{"Pause Full (Metadata GC Threshold)", "SerialMetadataGC"},
		{"Pause Full (Allocation Failure)", "SerialFullGCAllocationFailure"},
		{"Pause Full (Ergonomics)", "SerialFullGC"},
		{"Unknown Phase", "SerialUnknown"},
	}
	for _, tt := range tests {
		got := classifyUnifiedSerialPause(tt.desc)
		if got != tt.want {
			t.Errorf("classifyUnifiedSerialPause(%q) = %q, want %q", tt.desc, got, tt.want)
		}
	}
}
