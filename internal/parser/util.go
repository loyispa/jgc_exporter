package parser

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

// StripUnifiedJVMLogPrefix removes leading JVM unified-log bracket fields (timestamp, pid, tid, level, tags, ...)
// and returns the message payload. ok is false if the line does not start with a parsable unified ISO timestamp.
func StripUnifiedJVMLogPrefix(line string) (payload string, ok bool) {
	s := strings.TrimSpace(line)
	if len(s) < 20 || s[0] != '[' {
		return "", false
	}
	if _, tsOK := parseUnifiedTimestamp(s); !tsOK {
		return "", false
	}
	for {
		s = strings.TrimSpace(s)
		if len(s) == 0 || s[0] != '[' {
			break
		}
		closeIdx := strings.Index(s, "]")
		if closeIdx < 0 {
			return "", false
		}
		s = s[closeIdx+1:]
	}
	s = strings.TrimSpace(s)
	if s == "" {
		return "", false
	}
	return s, true
}

// SyntheticUnifiedGCLine rebuilds a minimal unified line: first timestamp bracket from original + "[gc] " + payload
// so parseUnifiedTimestamp and parsers that gate on "][gc" keep working after prefix strip.
func SyntheticUnifiedGCLine(original, payload string) string {
	orig := strings.TrimSpace(original)
	closeIdx := strings.Index(orig, "]")
	if closeIdx > 0 && orig[0] == '[' {
		return orig[:closeIdx+1] + "[gc] " + strings.TrimSpace(payload)
	}
	return "[1970-01-01T00:00:00.000+0000][gc] " + strings.TrimSpace(payload)
}

var preUnifiedTimestampRe = regexp.MustCompile(`^(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{3}[+-]\d{4}):\s*[\d.]+:\s*`)
var preUptimeOnlyRe = regexp.MustCompile(`^([\d.]+):\s*`)
var unifiedTimestampRe = regexp.MustCompile(`^\[(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{3}[+-]\d{4})\]`)

func parsePreUnifiedTimestamp(line string) (time.Time, bool) {
	if m := preUnifiedTimestampRe.FindStringSubmatch(line); len(m) >= 2 {
		t, err := time.Parse("2006-01-02T15:04:05.000-0700", m[1])
		if err != nil {
			return time.Time{}, false
		}
		return t, true
	}
	// Uptime-only: "1.234: [GC ..."
	if m := preUptimeOnlyRe.FindStringSubmatch(line); len(m) >= 2 {
		secs := parseFloat(m[1])
		if secs > 0 {
			return time.Unix(int64(secs), int64((secs-float64(int64(secs)))*1e9)), true
		}
	}
	return time.Time{}, false
}

func parseUnifiedTimestamp(line string) (time.Time, bool) {
	m := unifiedTimestampRe.FindStringSubmatch(line)
	if len(m) < 2 {
		return time.Time{}, false
	}
	t, err := time.Parse("2006-01-02T15:04:05.000-0700", m[1])
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}

func parseSizeKB(s string) int64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}

	multiplier := int64(1)
	if strings.HasSuffix(s, "M") || strings.HasSuffix(s, "m") {
		multiplier = 1024
		s = s[:len(s)-1]
	} else if strings.HasSuffix(s, "G") || strings.HasSuffix(s, "g") {
		multiplier = 1024 * 1024
		s = s[:len(s)-1]
	} else if strings.HasSuffix(s, "K") || strings.HasSuffix(s, "k") {
		multiplier = 1
		s = s[:len(s)-1]
	} else if strings.HasSuffix(s, "B") || strings.HasSuffix(s, "b") {
		s = s[:len(s)-1]
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return 0
		}
		return int64(v / 1024)
	}

	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return int64(v * float64(multiplier))
}

func parseFloat(s string) float64 {
	s = strings.TrimSpace(s)
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

func parseInt(s string) int64 {
	s = strings.TrimSpace(s)
	v, _ := strconv.ParseInt(s, 10, 64)
	return v
}

func extractBetween(s, left, right string) string {
	start := strings.Index(s, left)
	if start < 0 {
		return ""
	}
	start += len(left)
	end := strings.Index(s[start:], right)
	if end < 0 {
		return s[start:]
	}
	return s[start : start+end]
}

// isVerifyDiagnosticLine returns true if the line is a JVM diagnostic output
// from -XX:+VerifyBeforeGC or -XX:+VerifyAfterGC. These lines break GC event
// regex matching and should be skipped.
func isVerifyDiagnosticLine(line string) bool {
	if strings.HasPrefix(line, "VerifyBeforeGC:") || strings.HasPrefix(line, "VerifyAfterGC:") {
		return true
	}
	trimmed := strings.TrimSpace(line)
	return trimmed == "[Verifying before GC]" || trimmed == "[Verifying after GC]"
}

func extractAfter(s, prefix string) string {
	idx := strings.Index(s, prefix)
	if idx < 0 {
		return ""
	}
	return strings.TrimSpace(s[idx+len(prefix):])
}

// Reference processing regexes (shared by G1 and CMS)
var refSoftRe = regexp.MustCompile(`\[SoftReference,\s*(\d+)\s*refs,\s*([\d.]+)\s*secs\]`)
var refWeakRe = regexp.MustCompile(`\[WeakReference,\s*(\d+)\s*refs,\s*([\d.]+)\s*secs\]`)
var refFinalRe = regexp.MustCompile(`\[FinalReference,\s*(\d+)\s*refs,\s*([\d.]+)\s*secs\]`)
var refPhantomRe = regexp.MustCompile(`\[PhantomReference,\s*(\d+)\s*refs(?:,\s*(\d+)\s*free)?,\s*([\d.]+)\s*secs\]`)
var refJNIWeakRe = regexp.MustCompile(`\[JNI Weak Reference,\s*([\d.]+)\s*secs\]`)
