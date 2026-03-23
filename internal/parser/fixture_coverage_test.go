package parser

import (
	"os"
	"path/filepath"
	"testing"
)

// TestCommittedParserFixturesExist guards the documented JDK×GC×variant matrix:
// every listed file must be present under testdata/parser/.
func TestCommittedParserFixturesExist(t *testing.T) {
	base := filepath.Join("..", "..", "testdata", "parser")
	required := []string{
		// G1
		"jdk8-g1.log", "jdk8-g1-verbose.log", "jdk8-g1-uptimeonly.log",
		"jdk11-g1.log", "jdk11-g1-gconly.log",
		"jdk17-g1.log", "jdk21-g1.log",
		// CMS (JDK 8–11 only in production)
		"jdk8-cms.log", "jdk8-cms-verbose.log",
		"jdk11-cms.log", "jdk11-cms-gconly.log",
		// ZGC
		"jdk11-zgc.log", "jdk11-zgc-gconly.log",
		"jdk17-zgc.log", "jdk17-zgc-gconly.log",
		"jdk21-zgc-generational.log", "jdk21-zgc-gen-gconly.log", "jdk21-zgc-gen-interleaved.log",
		// Parallel
		"jdk8-parallel.log", "jdk11-parallel.log", "jdk11-parallel-psoldgen.log", "jdk17-parallel.log",
		// Serial
		"jdk8-serial.log", "jdk11-serial.log", "jdk17-serial.log",
		// Pipeline helpers
		"safepoint-preunified.log", "safepoint-unified.log",
	}
	for _, name := range required {
		p := filepath.Join(base, name)
		if _, err := os.Stat(p); err != nil {
			t.Errorf("missing committed fixture %s (%v)", name, err)
		}
	}
}
