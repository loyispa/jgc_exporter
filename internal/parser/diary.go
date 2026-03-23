package parser

import (
	"bufio"
	"os"
	"strings"
)

const maxDetectLines = 500

func DetectGCType(path string) GCType {
	f, err := os.Open(path)
	if err != nil {
		return GCTypeUnknown
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	count := 0
	for scanner.Scan() && count < maxDetectLines {
		line := scanner.Text()
		count++

		if gcType := detectFromLine(line); gcType != GCTypeUnknown {
			return gcType
		}
	}

	return GCTypeUnknown
}

func detectFromLine(line string) GCType {
	upper := strings.ToUpper(line)

	// ZGC: most distinctive markers first
	if strings.Contains(upper, "USING THE Z GARBAGE COLLECTOR") ||
		strings.Contains(upper, "+USEZGC") ||
		strings.Contains(line, "UseZGC") {
		return GCTypeZGC
	}
	// Generational/non-generational ZGC event lines (require GC(N) prefix to
	// avoid false positives from unrelated log content containing these words)
	if strings.Contains(line, "Minor Collection") ||
		strings.Contains(line, "Major Collection") {
		return GCTypeZGC
	}
	if strings.Contains(line, "Garbage Collection") && strings.Contains(line, "GC(") {
		return GCTypeZGC
	}

	// G1
	if strings.Contains(upper, "USING G1") ||
		strings.Contains(upper, "+USEG1GC") ||
		strings.Contains(line, "UseG1GC") {
		return GCTypeG1
	}
	if strings.Contains(line, "[GC pause") && strings.Contains(line, "G1") {
		return GCTypeG1
	}
	// Unified G1 markers (JDK9+)
	if strings.Contains(line, "G1 Evacuation Pause") ||
		strings.Contains(line, "G1 Humongous Allocation") ||
		strings.Contains(line, "G1 Preventive Collection") ||
		strings.Contains(line, "Humongous regions:") ||
		strings.Contains(line, "Eden regions:") {
		return GCTypeG1
	}

	// CMS (check before Parallel/Serial since CMS has unique markers)
	if strings.Contains(upper, "USING CONCURRENT MARK SWEEP") ||
		strings.Contains(upper, "+USECONCMARKSWEEPGC") ||
		strings.Contains(line, "UseConcMarkSweepGC") {
		return GCTypeCMS
	}
	if strings.Contains(line, "CMS-initial-mark") ||
		strings.Contains(line, "CMS-concurrent") ||
		strings.Contains(line, "[CMS") ||
		strings.Contains(line, "[ParNew") {
		return GCTypeCMS
	}

	// Parallel (pre-unified: [PSYoungGen, unified: PSYoungGen:/ParOldGen:/PSOldGen:)
	if strings.Contains(upper, "USING PARALLEL") ||
		strings.Contains(upper, "+USEPARALLELGC") ||
		strings.Contains(line, "UseParallelGC") {
		return GCTypeParallel
	}
	if strings.Contains(line, "[PSYoungGen") || strings.Contains(line, "PSYoungGen:") ||
		strings.Contains(line, "ParOldGen:") || strings.Contains(line, "PSOldGen:") {
		return GCTypeParallel
	}

	// Serial (pre-unified: [DefNew, unified: DefNew:/Tenured:)
	if strings.Contains(upper, "USING SERIAL") ||
		strings.Contains(upper, "+USESERIALGC") ||
		strings.Contains(line, "UseSerialGC") {
		return GCTypeSerial
	}
	if strings.Contains(line, "[DefNew") || strings.Contains(line, "DefNew:") ||
		strings.Contains(line, "Tenured:") {
		return GCTypeSerial
	}

	return GCTypeUnknown
}

func DetectUnifiedLogging(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for i := 0; i < 20 && scanner.Scan(); i++ {
		line := scanner.Text()
		if len(line) > 0 && line[0] == '[' && (strings.Contains(line, "][gc") || strings.Contains(line, "][gc]")) {
			return true
		}
	}
	return false
}

func DetectGenerationalZGC(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for i := 0; i < maxDetectLines && scanner.Scan(); i++ {
		line := scanner.Text()
		if strings.Contains(line, "Minor Collection") || strings.Contains(line, "Major Collection") {
			return true
		}
	}
	return false
}
