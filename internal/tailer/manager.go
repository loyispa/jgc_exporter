package tailer

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/mattn/go-zglob"
)

// Listener receives lifecycle and line events.
type Listener interface {
	// Accept is called before creating a Tailer: open path, detect GC format; return true only
	// for valid GC logs. If true, implementation should cache result for OnOpen to use without reopening.
	Accept(path string) bool

	// OnOpen is called after Tailer is created; use cached result from Accept to register parser.
	// Return false to close the Tailer immediately.
	OnOpen(path string) bool

	// OnClose is called when a file is idle (no new lines for idleTimeout) or when the Manager is shutting down.
	OnClose(path string)

	OnRead(path string, line string)
}

// Manager discovers files via glob and runs a Tailer per file. Idle timeout
// (no new lines for idleTimeout) triggers OnClose; rotation is handled by the
// tail library and does not close the Tailer.
type Manager struct {
	globPattern   string
	idleTimeout   time.Duration
	watchInterval time.Duration
	listener      Listener
	tailers       map[string]*Tailer
	mu            sync.Mutex
	stopCh        chan struct{}
}

func NewManager(
	globPattern string,
	idleTimeout, watchInterval time.Duration,
	listener Listener,
) *Manager {
	return &Manager{
		globPattern:   globPattern,
		idleTimeout:   idleTimeout,
		watchInterval: watchInterval,
		listener:      listener,
		tailers:       make(map[string]*Tailer),
		stopCh:        make(chan struct{}),
	}
}

func (m *Manager) Start() {
	go m.watchLoop()
}

func (m *Manager) Stop() {
	close(m.stopCh)
	m.mu.Lock()
	defer m.mu.Unlock()
	for path, t := range m.tailers {
		t.Close()
		m.listener.OnClose(path)
		delete(m.tailers, path)
	}
}

func (m *Manager) watchLoop() {
	m.scan()
	ticker := time.NewTicker(m.watchInterval)
	defer ticker.Stop()
	for {
		select {
		case <-m.stopCh:
			return
		case <-ticker.C:
			m.scan()
		}
	}
}

func (m *Manager) scan() {
	patterns := strings.Split(m.globPattern, ",")
	var matches []string
	for _, p := range patterns {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		m, err := zglob.Glob(p)
		if err != nil {
			slog.Error("glob match failed", "pattern", p, "error", err)
			continue
		}
		matches = append(matches, m...)
	}

	matchSet := make(map[string]struct{}, len(matches))
	for _, match := range matches {
		abs, err := filepath.Abs(match)
		if err != nil {
			slog.Debug("glob match abs failed", "match", match, "error", err)
			continue
		}
		info, err := os.Stat(abs)
		if err != nil {
			slog.Debug("glob match stat failed", "path", abs, "error", err)
			continue
		}
		if info.IsDir() {
			continue
		}
		matchSet[abs] = struct{}{}
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	for path, t := range m.tailers {
		if now.Sub(t.LastModified()) > m.idleTimeout {
			slog.Info("file idle, closing tailer", "path", path)
			t.Close()
			m.listener.OnClose(path)
			delete(m.tailers, path)
		}
	}

	for abs := range matchSet {
		if _, exists := m.tailers[abs]; exists {
			continue
		}
		if !m.listener.Accept(abs) {
			slog.Info("skipping non-GC log", "path", abs)
			continue
		}

		handler := func(file string, line string) {
			m.listener.OnRead(file, line)
		}
		t, err := NewTailer(abs, handler, true)
		if err != nil {
			slog.Error("open tailer failed", "path", abs, "error", err)
			continue
		}

		if !m.listener.OnOpen(abs) {
			t.Close()
			continue
		}

		m.tailers[abs] = t
	}
}
