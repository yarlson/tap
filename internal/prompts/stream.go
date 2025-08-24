package prompts

import (
	"bufio"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/yarlson/tap/internal/core"
)

// StreamOptions configure the styled stream renderer
type StreamOptions struct {
	Output Writer
	// If true, show elapsed time on finalize line
	ShowTimer bool
}

// Stream renders a live stream area with clack-like styling
// Use Start to begin, WriteLine/Pipe to add content, and Stop to finalize.
type Stream struct {
	out   Writer
	mu    sync.Mutex
	open  bool
	lines []string
	start time.Time
	opts  StreamOptions
	title string
}

// NewStream creates a Stream
func NewStream(opts StreamOptions) *Stream {
	return &Stream{out: opts.Output, opts: opts}
}

// Start prints the header and prepares to receive lines
func (s *Stream) Start(message string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.open {
		return
	}
	s.open = true
	s.start = time.Now()
	s.title = message
	if s.out != nil {
		header := fmt.Sprintf("%s\n%s  %s\n", gray(Bar), Symbol(core.StateActive), message)
		_, _ = s.out.Write([]byte(header))
	}
}

// WriteLine appends a single line into the stream area
func (s *Stream) WriteLine(line string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.open || s.out == nil {
		return
	}
	content := fmt.Sprintf("%s  %s\n", cyan(Bar), line)
	_, _ = s.out.Write([]byte(content))
	s.lines = append(s.lines, line)
}

// Pipe reads from r line-by-line and writes to the stream area
func (s *Stream) Pipe(r io.Reader) {
	s.mu.Lock()
	open := s.open
	s.mu.Unlock()
	if !open {
		return
	}

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		s.WriteLine(scanner.Text())
	}
}

// Stop finalizes the stream with a status symbol and optional timer
// code: 0 submit, 1 cancel, >1 error
func (s *Stream) Stop(finalMessage string, code int) {
	s.mu.Lock()
	if !s.open {
		s.mu.Unlock()
		return
	}
	s.open = false
	start := s.start
	showTimer := s.opts.ShowTimer
	out := s.out
	s.mu.Unlock()

	if out == nil {
		return
	}

	// Colorize final message subtly to distinguish from stream content

	msg := finalMessage
	if msg == "" {
		msg = ""
	}
	if showTimer {
		d := time.Since(start)
		secs := int(d.Seconds())
		m := secs / 60
		sec := secs % 60
		if m > 0 {
			msg = fmt.Sprintf("%s [%dm %ds]", msg, m, sec)
		} else {
			msg = fmt.Sprintf("%s [%ds]", msg, sec)
		}
	}
	// Apply color by status: green for success, red for cancel/error
	if code == 0 {
		msg = green(msg)
	} else {
		msg = red(msg)
	}

	// Visually deactivate: repaint previously printed content lines with gray bars.
	// Move cursor up by the number of content lines we printed, then rewrite each line.
	s.mu.Lock()
	lineCount := len(s.lines)
	lines := append([]string(nil), s.lines...)
	title := s.title
	s.mu.Unlock()

	// Move up to the header (one line above first content line)
	for i := 0; i < lineCount+1; i++ {
		_, _ = out.Write([]byte(core.CursorUp))
	}
	// Rewrite header: green dimmed title to indicate completion (no diamond)
	_, _ = out.Write([]byte("\r"))
	_, _ = out.Write([]byte(core.EraseLine))
	_, _ = out.Write([]byte(fmt.Sprintf("%s  %s\n", gray(Bar), dim(title))))

	// Repaint content lines with gray bars
	for i := range lineCount {
		_, _ = out.Write([]byte("\r"))
		_, _ = out.Write([]byte(core.EraseLine))
		_, _ = out.Write([]byte(fmt.Sprintf("%s  %s\n", gray(Bar), lines[i])))
	}

	// Integrate status inside the block without any diamond symbol.
	// Do not render a bottom corner on submit/cancel/error to match other primitives' submit state.
	status := fmt.Sprintf("%s  %s\n", gray(Bar), msg)
	_, _ = out.Write([]byte(status))
}
