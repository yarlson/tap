package tap

import (
	"bufio"
	"fmt"
	"io"
	"sync"
	"time"
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
	if opts.Output != nil {
		return &Stream{out: opts.Output, opts: opts}
	}

	out, _ := resolveWriter()
	opts.Output = out

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
		header := fmt.Sprintf("%s\n%s  %s\n", gray(Bar), Symbol(StateActive), message)
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

	// Prepare final message and timing

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
	// Message itself remains white to align with design language

	// Visually deactivate: repaint previously printed content lines with gray bars.
	// Move cursor up by the number of content lines we printed, then rewrite each line.
	s.mu.Lock()
	lineCount := len(s.lines)
	lines := append([]string(nil), s.lines...)
	title := s.title
	s.mu.Unlock()

	// Move up to the header (one line above first content line)
	for i := 0; i < lineCount+1; i++ {
		_, _ = out.Write([]byte(CursorUp))
	}
	// Rewrite header: inactive diamond, title stays white
	_, _ = out.Write([]byte("\r"))
	_, _ = out.Write([]byte(EraseLine))
	_, _ = fmt.Fprintf(out, "%s  %s\n", green(StepSubmit), title)

	// Repaint content lines with gray bars and dimmed text
	for i := range lineCount {
		_, _ = out.Write([]byte("\r"))
		_, _ = out.Write([]byte(EraseLine))
		_, _ = fmt.Fprintf(out, "%s  %s\n", gray(Bar), dim(lines[i]))
	}

	// Final status line with a diamond (aligned like header), white message; no bottom corner
	statusSymbol := green(StepSubmit)
	if code == 1 {
		statusSymbol = red(StepCancel)
	} else if code > 1 {
		statusSymbol = yellow(StepError)
	}

	status := fmt.Sprintf("%s  %s\n", statusSymbol, msg)
	_, _ = out.Write([]byte(status))
}
