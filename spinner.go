package tap

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"
)

// SpinnerOptions configures the spinner behavior
type SpinnerOptions struct {
	Indicator     string   // "dots" (default) or "timer"
	Frames        []string // custom frames; defaults to unicode spinner frames
	Delay         time.Duration
	Output        Writer
	CancelMessage string
	ErrorMessage  string
}

// Spinner represents an animated spinner
type Spinner struct {
	indicator string
	frames    []string
	delay     time.Duration
	output    Writer

	mu              sync.RWMutex
	isActive        bool
	isCancelled     bool
	message         string
	startTime       time.Time
	frameIndex      int
	lastFrameLength int
	dotTick         int

	ticker *time.Ticker
	stopCh chan struct{}
}

// NewSpinner creates a new Spinner with defaults
func NewSpinner(opts SpinnerOptions) *Spinner {
	if opts.Output != nil {
		return newSpinner(opts)
	}

	out, _ := resolveWriter()
	opts.Output = out

	return newSpinner(opts)
}

// newSpinner creates a new Spinner with given options
func newSpinner(opts SpinnerOptions) *Spinner {
	indicator := opts.Indicator
	if indicator == "" {
		indicator = "dots"
	}

	frames := opts.Frames
	if len(frames) == 0 {
		frames = []string{"◒", "◐", "◓", "◑"}
	}

	delay := opts.Delay
	if delay <= 0 {
		delay = 80 * time.Millisecond
	}

	return &Spinner{
		indicator: indicator,
		frames:    frames,
		delay:     delay,
		output:    opts.Output,
		stopCh:    make(chan struct{}),
	}
}

// Start begins the spinner animation
func (s *Spinner) Start(msg string) {
	s.mu.Lock()
	if s.isActive {
		s.mu.Unlock()
		return
	}
	s.isActive = true
	s.message = removeTrailingDots(msg)
	s.frameIndex = 0
	s.dotTick = 0
	s.startTime = time.Now()
	lastLen := s.lastFrameLength
	s.mu.Unlock()

	s.ticker = time.NewTicker(s.delay)
	go s.animate()

	// OSC 9;4 indeterminate spinner
	oscSpin(s.output)

	if lastLen > 0 {
		if s.output != nil {
			_, _ = s.output.Write([]byte("\033[1A\r\033[J"))
		}
	}
	s.render()
}

// Message updates the spinner message for next frame
func (s *Spinner) Message(msg string) {
	s.mu.Lock()
	s.message = removeTrailingDots(msg)
	s.mu.Unlock()
	s.render()
}

// Stop halts the spinner and prints a final line with a status symbol
// code: 0 submit, 1 cancel, >1 error
func (s *Spinner) Stop(msg string, code int) {
	s.mu.Lock()
	if !s.isActive {
		s.mu.Unlock()
		return
	}
	s.isActive = false
	s.isCancelled = code == 1
	currentMsg := s.message
	start := s.startTime
	indicator := s.indicator
	s.mu.Unlock()

	if s.ticker != nil {
		s.ticker.Stop()
	}
	close(s.stopCh)

	// OSC 9;4 final state
	switch code {
	case 0:
		oscClear(s.output)
	case 1:
		oscPause(s.output)
	default:
		oscError(s.output)
	}

	if s.output != nil {
		if s.lastFrameLength > 0 {
			_, _ = s.output.Write([]byte("\033[1A\r\033[J"))
		}
		var symbol string
		switch code {
		case 0:
			symbol = green(StepSubmit)
		case 1:
			symbol = red(StepCancel)
		default:
			symbol = red(StepError)
		}
		finalMsg := msg
		if finalMsg == "" {
			finalMsg = currentMsg
		}
		if indicator == "timer" {
			finalMsg = fmt.Sprintf("%s %s", finalMsg, formatTimer(start))
		}
		final := fmt.Sprintf("%s\n%s  %s\n", gray(Bar), symbol, finalMsg)
		_, _ = s.output.Write([]byte(final))
	}
}

// IsCancelled reports whether Stop was called with cancel code (1)
func (s *Spinner) IsCancelled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isCancelled
}

func (s *Spinner) animate() {
	for {
		select {
		case <-s.stopCh:
			return
		case <-s.ticker.C:
			s.render()
		}
	}
}

func (s *Spinner) render() {
	if s.output == nil {
		return
	}

	s.mu.RLock()
	msg := s.message
	frame := s.frames[s.frameIndex]
	indicator := s.indicator
	start := s.startTime
	active := s.isActive
	s.mu.RUnlock()

	if !active {
		return
	}

	var displayMsg string
	if indicator == "timer" {
		displayMsg = fmt.Sprintf("%s %s", msg, formatTimer(start))
	} else {
		dots := strings.Repeat(".", s.currentDotCount())
		displayMsg = msg + dots
	}

	content := fmt.Sprintf("%s\n%s  %s", gray(Bar), cyan(frame), displayMsg)

	if s.lastFrameLength > 0 {
		_, _ = s.output.Write([]byte("\033[1A\r\033[J"))
	}

	_, _ = s.output.Write([]byte(content))

	s.mu.Lock()
	s.frameIndex = (s.frameIndex + 1) % len(s.frames)
	s.dotTick = (s.dotTick + 1) % 24 // full cycle every 24 ticks
	s.lastFrameLength = len(stripANSI(content))
	s.mu.Unlock()
}

func (s *Spinner) currentDotCount() int {
	dots := s.dotTick / 8
	if dots > 3 {
		dots = 3
	}
	return dots
}

func removeTrailingDots(in string) string {
	return regexp.MustCompile(`\.+$`).ReplaceAllString(in, "")
}

func formatTimer(start time.Time) string {
	d := time.Since(start)
	secs := int(d.Seconds())
	m := secs / 60
	sec := secs % 60
	if m > 0 {
		return fmt.Sprintf("[%dm %ds]", m, sec)
	}
	return fmt.Sprintf("[%ds]", sec)
}

// stripANSI removes ANSI color codes to get display length
func stripANSI(s string) string {
	ansiRegex := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return ansiRegex.ReplaceAllString(s, "")
}
