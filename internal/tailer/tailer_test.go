package tailer

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/mattn/go-zglob"
)

// waitForLines waits until the slice has at least n elements or timeout.
func waitForLines(lines *[]string, n int, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if len(*lines) >= n {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func TestTailerReadLines(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.log")
	if err := os.WriteFile(path, []byte("line1\nline2\nline3\n"), 0644); err != nil {
		t.Fatal(err)
	}

	var mu sync.Mutex
	var lines []string
	handler := func(file string, line string) {
		mu.Lock()
		lines = append(lines, line)
		mu.Unlock()
	}

	tail, err := NewTailer(path, handler, false)
	if err != nil {
		t.Fatal(err)
	}
	defer tail.Close()

	waitForLines(&lines, 3, 2*time.Second)
	mu.Lock()
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}
	if lines[0] != "line1" || lines[1] != "line2" || lines[2] != "line3" {
		t.Fatalf("unexpected lines: %v", lines)
	}
	mu.Unlock()
}

func TestTailerSeekToEnd(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.log")
	if err := os.WriteFile(path, []byte("existing\n"), 0644); err != nil {
		t.Fatal(err)
	}

	var lines []string
	handler := func(file string, line string) {
		lines = append(lines, line)
	}

	tail, err := NewTailer(path, handler, true)
	if err != nil {
		t.Fatal(err)
	}
	defer tail.Close()

	time.Sleep(50 * time.Millisecond)
	if len(lines) != 0 {
		t.Fatalf("expected 0 lines when seekToEnd (start at end), got %d", len(lines))
	}

	f, _ := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	f.WriteString("new line\n")
	f.Close()

	waitForLines(&lines, 1, 2*time.Second)
	if len(lines) != 1 {
		t.Fatalf("expected 1 new line, got %d", len(lines))
	}
	if lines[0] != "new line" {
		t.Fatalf("expected 'new line', got '%s'", lines[0])
	}
}

func TestTailerTruncation(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.log")
	if err := os.WriteFile(path, []byte("line1\nline2\nline3\n"), 0644); err != nil {
		t.Fatal(err)
	}

	var lines []string
	handler := func(file string, line string) {
		lines = append(lines, line)
	}

	tail, err := NewTailer(path, handler, false)
	if err != nil {
		t.Fatal(err)
	}
	defer tail.Close()

	waitForLines(&lines, 3, 2*time.Second)
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines initially, got %d", len(lines))
	}

	if err := os.WriteFile(path, []byte("new\n"), 0644); err != nil {
		t.Fatal(err)
	}
	waitForLines(&lines, 4, 2*time.Second)
	if len(lines) != 4 {
		t.Fatalf("expected 4 lines after truncation, got %d", len(lines))
	}
	if lines[3] != "new" {
		t.Fatalf("expected last line 'new', got '%s'", lines[3])
	}
}

func TestTailerBatchSize(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.log")
	if err := os.WriteFile(path, []byte("a\nb\nc\nd\ne\n"), 0644); err != nil {
		t.Fatal(err)
	}

	var lines []string
	handler := func(file string, line string) {
		lines = append(lines, line)
	}

	tail, err := NewTailer(path, handler, false)
	if err != nil {
		t.Fatal(err)
	}
	defer tail.Close()

	waitForLines(&lines, 5, 2*time.Second)
	if len(lines) != 5 {
		t.Fatalf("expected 5 lines, got %d", len(lines))
	}
	for i, want := range []string{"a", "b", "c", "d", "e"} {
		if lines[i] != want {
			t.Fatalf("lines[%d] want %q got %q", i, want, lines[i])
		}
	}
}

func TestTailerPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.log")
	os.WriteFile(path, []byte("x\n"), 0644)
	tail, err := NewTailer(path, func(string, string) {}, false)
	if err != nil {
		t.Fatal(err)
	}
	defer tail.Close()
	if tail.Path() != path {
		t.Fatalf("expected %s, got %s", path, tail.Path())
	}
}

func TestTailerLastModified(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.log")
	os.WriteFile(path, []byte("x\n"), 0644)
	tail, err := NewTailer(path, func(string, string) {}, false)
	if err != nil {
		t.Fatal(err)
	}
	defer tail.Close()
	lm := tail.LastModified()
	if lm.IsZero() {
		t.Fatal("expected non-zero last modified")
	}
}

type mockListener struct {
	mu     sync.Mutex
	opened []string
	closed []string
	lines  map[string][]string
}

func newMockListener() *mockListener {
	return &mockListener{lines: make(map[string][]string)}
}

func (l *mockListener) Accept(path string) bool {
	return true
}

func (l *mockListener) OnOpen(path string) bool {
	l.mu.Lock()
	l.opened = append(l.opened, path)
	l.mu.Unlock()
	return true
}

func (l *mockListener) OnClose(path string) {
	l.mu.Lock()
	l.closed = append(l.closed, path)
	l.mu.Unlock()
}

func (l *mockListener) OnRead(path string, line string) {
	l.mu.Lock()
	l.lines[path] = append(l.lines[path], line)
	l.mu.Unlock()
}

func TestGlobExpandDoubleStar(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "gc.log"), []byte("x"), 0644)
	os.MkdirAll(filepath.Join(dir, "svc1"), 0755)
	os.WriteFile(filepath.Join(dir, "svc1", "gc.log"), []byte("x"), 0644)
	os.MkdirAll(filepath.Join(dir, "svc2", "nested"), 0755)
	os.WriteFile(filepath.Join(dir, "svc2", "nested", "gc.log"), []byte("x"), 0644)

	pattern := filepath.Join(dir, "**", "gc.log")
	matches, err := zglob.Glob(pattern)
	if err != nil {
		t.Fatalf("zglob.Glob: %v", err)
	}
	if len(matches) != 3 {
		t.Fatalf("expected 3 matches for **/gc.log, got %d: %v", len(matches), matches)
	}
	got := make(map[string]bool)
	for _, m := range matches {
		got[filepath.Clean(m)] = true
	}
	for _, p := range []string{filepath.Join(dir, "gc.log"), filepath.Join(dir, "svc1", "gc.log"), filepath.Join(dir, "svc2", "nested", "gc.log")} {
		if !got[p] {
			t.Errorf("missing match: %s", p)
		}
	}
}

func TestManagerScanAndStop(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "gc.log")
	os.WriteFile(path, []byte("test line\n"), 0644)

	listener := newMockListener()
	mgr := NewManager(
		filepath.Join(dir, "*.log"),
		time.Hour,
		50*time.Millisecond,
		listener,
	)
	mgr.Start()
	time.Sleep(200 * time.Millisecond)
	mgr.Stop()

	listener.mu.Lock()
	defer listener.mu.Unlock()
	if len(listener.opened) == 0 {
		t.Fatal("expected at least one OnOpen call")
	}
	if len(listener.closed) == 0 {
		t.Fatal("expected OnClose on Stop")
	}
}

// TestTailerRotationHandledByLibrary verifies that with ReOpen: true the
// library reopens the file after rotation and we keep receiving lines.
func TestTailerRotationHandledByLibrary(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.log")
	if err := os.WriteFile(path, []byte("first\n"), 0644); err != nil {
		t.Fatal(err)
	}

	var lines []string
	var mu sync.Mutex
	handler := func(file string, line string) {
		mu.Lock()
		lines = append(lines, line)
		mu.Unlock()
	}

	tail, err := NewTailer(path, handler, false)
	if err != nil {
		t.Fatal(err)
	}
	defer tail.Close()

	waitForLines(&lines, 1, 2*time.Second)
	mu.Lock()
	if len(lines) != 1 || lines[0] != "first" {
		t.Fatalf("expected [first], got %v", lines)
	}
	mu.Unlock()

	// Rotate: move away and create new file at same path
	if err := os.Rename(path, filepath.Join(dir, "test.log.1")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("after_rotate\n"), 0644); err != nil {
		t.Fatal(err)
	}

	waitForLines(&lines, 2, 3*time.Second)
	mu.Lock()
	defer mu.Unlock()
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines after rotation, got %d: %v", len(lines), lines)
	}
	if lines[1] != "after_rotate" {
		t.Fatalf("expected second line 'after_rotate', got %q", lines[1])
	}
}

func TestTailerNewTailerError(t *testing.T) {
	_, err := NewTailer("/nonexistent/file.log", func(string, string) {}, false)
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestTailerReadAfterClose(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.log")
	os.WriteFile(path, []byte("data\n"), 0644)

	tail, err := NewTailer(path, func(string, string) {}, false)
	if err != nil {
		t.Fatal(err)
	}
	tail.Close()
	// After Close, Path and LastModified should not panic.
	_ = tail.Path()
	_ = tail.LastModified()
}

func TestTailerCloseIdempotent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.log")
	os.WriteFile(path, []byte("x\n"), 0644)
	tail, err := NewTailer(path, func(string, string) {}, false)
	if err != nil {
		t.Fatal(err)
	}
	tail.Close()
	tail.Close()
}

func TestManagerIdleTimeout(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "gc.log")
	os.WriteFile(path, []byte("old content\n"), 0644)

	listener := newMockListener()
	mgr := NewManager(
		filepath.Join(dir, "*.log"),
		1*time.Millisecond,
		50*time.Millisecond,
		listener,
	)
	mgr.scan()
	time.Sleep(10 * time.Millisecond)
	mgr.scan()

	listener.mu.Lock()
	defer listener.mu.Unlock()
	if len(listener.closed) < 1 {
		t.Fatal("expected idle close")
	}
}

func TestManagerGlobError(t *testing.T) {
	listener := newMockListener()
	mgr := NewManager(
		"[invalid-glob",
		time.Hour,
		50*time.Millisecond,
		listener,
	)
	mgr.scan()
}

func TestManagerMultiPath(t *testing.T) {
	dirA := t.TempDir()
	dirB := t.TempDir()
	os.WriteFile(filepath.Join(dirA, "a.log"), []byte("a\n"), 0644)
	os.WriteFile(filepath.Join(dirB, "b.log"), []byte("b\n"), 0644)

	listener := newMockListener()
	pattern := filepath.Join(dirA, "*.log") + "," + filepath.Join(dirB, "*.log")
	mgr := NewManager(pattern, time.Hour, 50*time.Millisecond, listener)
	mgr.scan()

	listener.mu.Lock()
	opened := listener.opened
	listener.mu.Unlock()
	if len(opened) != 2 {
		t.Fatalf("expected 2 files opened from comma-separated patterns, got %d: %v", len(opened), opened)
	}
}

// TestManagerRotation verifies that after file rotation the same Tailer keeps
// running (library reopens) and new content is still delivered.
func TestManagerRotation(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "gc.log")
	if err := os.WriteFile(path, []byte("original\n"), 0644); err != nil {
		t.Fatal(err)
	}

	listener := newMockListener()
	mgr := NewManager(
		filepath.Join(dir, "*.log"),
		time.Hour,
		50*time.Millisecond,
		listener,
	)

	mgr.scan()

	listener.mu.Lock()
	openCount := len(listener.opened)
	listener.mu.Unlock()
	if openCount != 1 {
		t.Fatalf("expected 1 open, got %d", openCount)
	}

	// Rotate file
	if err := os.Rename(path, filepath.Join(dir, "gc.log.1")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("rotated content\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Library reopens internally; wait for "rotated content" to be delivered
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		mgr.scan()
		listener.mu.Lock()
		pathLines := listener.lines[listener.opened[0]]
		listener.mu.Unlock()
		for _, l := range pathLines {
			if l == "rotated content" {
				goto done
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
done:

	listener.mu.Lock()
	openedPath := listener.opened[0]
	pathLines := listener.lines[openedPath]
	listener.mu.Unlock()
	// Manager opens with seekToEnd: true so we may not see "original"; after
	// rotation the library reopens and reads the new file, so we must see "rotated content".
	hasRotated := false
	for _, l := range pathLines {
		if l == "rotated content" {
			hasRotated = true
			break
		}
	}
	if !hasRotated {
		t.Errorf("expected to receive 'rotated content' after rotation (library ReOpen), got lines: %v", pathLines)
	}
	// Still only one OnOpen (we don't close and re-add on rotation)
	listener.mu.Lock()
	if len(listener.opened) != 1 {
		t.Errorf("expected 1 OnOpen total (no second open on rotation), got %d", len(listener.opened))
	}
	listener.mu.Unlock()
}
