// Package session manages the lifecycle of an interactive terminal session and
// exposes accessors to the session-bound reader and writer. It also provides a
// simple default-session facility used by the high-level tap API.
package session

import (
	"os"
	"sync"

	"github.com/yarlson/tap/internal/core"
	"github.com/yarlson/tap/internal/terminal"
)

// Session owns a terminal and provides access to the session-bound
// core-compatible reader and writer.
type Session struct {
	term *terminal.Terminal
}

// New creates a new Session with its own terminal.
func New() (*Session, error) {
	term, err := terminal.New()
	if err != nil {
		return nil, err
	}
	return &Session{term: term}, nil
}

// Close releases resources associated with the session and restores the
// terminal state.
func (s *Session) Close() {
	if s != nil && s.term != nil {
		s.term.Close()
	}
}

// Reader returns the core-compatible reader bound to this session.
func (s *Session) Reader() core.Reader { return s.term.Reader }

// Writer returns the core-compatible writer bound to this session.
func (s *Session) Writer() core.Writer { return s.term.Writer }

var (
	defaultMu      sync.Mutex
	defaultSession *Session
)

// Init initializes and returns the default session used by the package-level
// helpers. If a default session already exists, it is returned.
func Init() (*Session, error) {
	defaultMu.Lock()
	defer defaultMu.Unlock()
	if defaultSession != nil {
		return defaultSession, nil
	}
	s, err := New()
	if err != nil {
		return nil, err
	}
	defaultSession = s
	return s, nil
}

// CloseDefault closes the default session, if any, and clears it.
func CloseDefault() {
	defaultMu.Lock()
	defer defaultMu.Unlock()
	if defaultSession != nil {
		defaultSession.Close()
		defaultSession = nil
	}
}

// Current returns the existing default session without creating a new one. It
// returns nil if no default session is present.
func Current() *Session {
	defaultMu.Lock()
	defer defaultMu.Unlock()
	return defaultSession
}

// GetOrCreateDefault returns the default session, creating it if necessary, and
// a boolean indicating whether it was created by this call.
func GetOrCreateDefault() (s *Session, created bool) {
	defaultMu.Lock()
	defer defaultMu.Unlock()

	if defaultSession != nil {
		return defaultSession, false
	}

	if ns, err := New(); err == nil {
		defaultSession = ns
		return ns, true
	}

	return nil, false
}

// RunWithDefault acquires a default session for the duration of fn and closes
// it afterwards if it was created by this call. If a session cannot be created,
// the cancel sentinel is returned.
func RunWithDefault(fn func(*Session) any) any {
	s, created := GetOrCreateDefault()
	if s == nil {
		return core.GetCancelSymbol()
	}

	defer func() {
		if created {
			CloseDefault()
		}
	}()

	return fn(s)
}

// CurrentWriter returns the writer for the current default session, or
// os.Stdout if no session is active.
func CurrentWriter() core.Writer {
	if s := Current(); s != nil {
		return s.Writer()
	}
	return newStdoutWriter()
}

// stdoutWriter satisfies core.Writer without opening the keyboard/terminal. It
// is used when no session is active.
type stdoutWriter struct {
	mu        sync.Mutex
	listeners map[string][]func()
}

func newStdoutWriter() *stdoutWriter                { return &stdoutWriter{listeners: make(map[string][]func())} }
func (w *stdoutWriter) Write(b []byte) (int, error) { return os.Stdout.Write(b) }
func (w *stdoutWriter) On(event string, handler func()) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.listeners[event] = append(w.listeners[event], handler)
}
func (w *stdoutWriter) Emit(event string) {
	w.mu.Lock()
	hs := append([]func(){}, w.listeners[event]...)
	w.mu.Unlock()
	for _, h := range hs {
		h()
	}
}
