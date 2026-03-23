package tailer

import (
	"io"
	"strings"
	"sync"
	"time"

	"github.com/nxadm/tail"
)

// LineHandler is called for each line read from the tailed file.
type LineHandler func(file string, line string)

// Tailer wraps github.com/nxadm/tail to tail a single file. Rotation and
// truncation are handled by the library (ReOpen: true); the caller is only
// responsible for idle timeout and shutdown via Close().
type Tailer struct {
	path    string
	inner   *tail.Tail
	lastMod time.Time
	mu      sync.Mutex
	done    chan struct{}
	handler LineHandler
}

// NewTailer starts tailing path. When seekToEnd is true, starts at end of file
// (avoid reading existing content); after rotation the library reopens the
// new file and reads from the beginning.
func NewTailer(path string, handler LineHandler, seekToEnd bool) (*Tailer, error) {
	cfg := tail.Config{
		Follow:        true,
		ReOpen:        true, // library handles move/delete/truncation (tail -F)
		MustExist:     true,
		Poll:          true, // use polling instead of inotify; avoids watch limits and event storms
		CompleteLines: true, // buffer until full line (ends with \n); JVM may flush long lines in chunks
		Logger:        tail.DiscardingLogger,
	}
	if seekToEnd {
		cfg.Location = &tail.SeekInfo{Whence: io.SeekEnd}
	}

	inner, err := tail.TailFile(path, cfg)
	if err != nil {
		return nil, err
	}

	t := &Tailer{
		path:    path,
		inner:   inner,
		handler: handler,
		done:    make(chan struct{}),
		lastMod: time.Now(), // avoid idle timeout before first line
	}

	linesCh := inner.Lines
	go t.readLoop(linesCh)
	return t, nil
}

func (t *Tailer) readLoop(linesCh <-chan *tail.Line) {
	defer close(t.done)
	for line := range linesCh {
		if line == nil {
			continue
		}
		text := strings.TrimSuffix(line.Text, "\n")
		text = strings.TrimSuffix(text, "\r")
		if text != "" {
			t.handler(t.path, text)
		}
		t.mu.Lock()
		t.lastMod = time.Now()
		t.mu.Unlock()
	}
}

// LastModified returns the last time a line was received (for idle timeout).
func (t *Tailer) LastModified() time.Time {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.lastMod
}

// Path returns the tailed file path.
func (t *Tailer) Path() string {
	return t.path
}

// Close stops the tail and waits for the read goroutine to exit.
func (t *Tailer) Close() error {
	t.mu.Lock()
	inner := t.inner
	t.inner = nil
	t.mu.Unlock()

	if inner == nil {
		return nil
	}

	inner.Kill(nil)
	<-t.done
	return nil
}
