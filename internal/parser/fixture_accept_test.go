package parser

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/loyispa/jgc_exporter/internal/health"
)

// TestFixtureDetectAcceptAndScan runs the pipeline over each checked-in *.log under testdata/parser
// (detect → accept → open → scan → close). Ensures no fixture regresses to Unknown GC type or reject Accept.
func TestFixtureDetectAcceptAndScan(t *testing.T) {
	pattern := filepath.Join("..", "..", "testdata", "parser", "*.log")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	if len(matches) == 0 {
		t.Fatalf("no fixtures under testdata/parser")
	}

	for _, path := range matches {
		base := filepath.Base(path)
		if strings.HasPrefix(base, ".") {
			continue
		}
		gc := DetectGCType(path)
		if gc == GCTypeUnknown {
			t.Errorf("%s: DetectGCType=Unknown", path)
			continue
		}

		mon := health.NewMonitor()
		pl := NewPipeline(mon)
		if !pl.Accept(path) {
			t.Errorf("%s: Accept=false (gc=%s)", path, gc)
			continue
		}
		if !pl.OnOpen(path) {
			t.Errorf("%s: OnOpen=false", path)
			continue
		}

		f, err := os.Open(path)
		if err != nil {
			t.Errorf("%s: open: %v", path, err)
			pl.OnClose(path)
			continue
		}
		sc := bufio.NewScanner(f)
		const maxScan = 512 * 1024
		buf := make([]byte, maxScan)
		sc.Buffer(buf, maxScan)
		n := 0
		for sc.Scan() {
			pl.OnRead(path, sc.Text())
			n++
		}
		_ = f.Close()
		if err := sc.Err(); err != nil {
			t.Errorf("%s: scan: %v", path, err)
		}
		if n == 0 {
			t.Errorf("%s: empty log", path)
		}
		pl.OnClose(path)
	}
}
