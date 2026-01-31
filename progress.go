package tap

import (
	"fmt"
	"math"
	"regexp"
	"strings"
	"sync"
	"time"
)

// ProgressOptions configures the progress bar
type ProgressOptions struct {
	Style  string // "light", "heavy", "block"
	Max    int    // maximum value (default 100)
	Size   int    // bar width in characters (default 40)
	Output Writer
}

// Progress represents a progress bar that wraps spinner functionality
type Progress struct {
	style    string
	max      int
	size     int
	output   Writer
	ticker   *time.Ticker
	stopChan chan struct{}
	frames   []string

	// Protected by mutex
	mu              sync.RWMutex
	value           int
	isActive        bool
	previousMsg     string
	frameIndex      int
	lastFrameLength int
	lastPct         int
}

// Progress bar character styles
var progressChars = map[string]string{
	"light": "─",
	"heavy": "━",
	"block": "█",
}

// NewProgress creates a new progress bar
func NewProgress(opts ProgressOptions) *Progress {
	if opts.Output != nil {
		return newProgress(opts)
	}

	out, _ := resolveWriter()
	opts.Output = out

	return newProgress(opts)
}

// newProgress creates a new progress bar with given options
func newProgress(opts ProgressOptions) *Progress {
	style := opts.Style
	if style == "" {
		style = "heavy"
	}

	max := opts.Max
	if max <= 0 {
		max = 100
	}

	size := opts.Size
	if size <= 0 {
		size = 40
	}

	return &Progress{
		style:      style,
		max:        max,
		size:       size,
		value:      0,
		output:     opts.Output,
		stopChan:   make(chan struct{}),
		frames:     []string{"◒", "◐", "◓", "◑"},
		frameIndex: 0,
		lastPct:    -1,
	}
}

// Start begins the progress bar animation
func (p *Progress) Start(msg string) {
	p.mu.Lock()

	if p.isActive {
		p.mu.Unlock()
		return
	}

	p.isActive = true
	p.previousMsg = msg
	p.lastFrameLength = 0 // Reset for new progress bar
	p.mu.Unlock()

	// Start animation
	p.ticker = time.NewTicker(80 * time.Millisecond)
	go p.animate()

	// Initial render
	p.render(msg)
}

// Advance updates progress by the given step and optionally updates message
func (p *Progress) Advance(step int, msg string) {
	p.mu.Lock()

	if !p.isActive {
		p.mu.Unlock()
		return
	}

	if step > 0 {
		p.value = int(math.Min(float64(p.max), float64(p.value+step)))
	}

	if msg != "" {
		p.previousMsg = msg
	}

	renderMsg := p.previousMsg
	p.mu.Unlock()

	p.render(renderMsg)
}

// Message updates the message without advancing progress
func (p *Progress) Message(msg string) {
	p.Advance(0, msg)
}

// Stop halts the progress bar and shows final state
func (p *Progress) Stop(msg string, code int, opts ...StopOptions) {
	p.mu.Lock()

	if !p.isActive {
		p.mu.Unlock()
		return
	}

	p.isActive = false
	lastLength := p.lastFrameLength
	p.mu.Unlock()

	// Stop animation
	if p.ticker != nil {
		p.ticker.Stop()
	}

	close(p.stopChan)

	oscClear(p.output)

	// Final render with state symbol
	var symbol string

	switch code {
	case 0:
		symbol = green(StepSubmit)
	case 1:
		symbol = red(StepCancel)
	default:
		symbol = red(StepError)
	}

	if p.output != nil {
		// Clear the current progress frame (3 lines)
		if lastLength > 0 {
			_, _ = p.output.Write([]byte("\033[2A\r\033[J"))
		}

		var hint string
		if len(opts) > 0 {
			hint = opts[0].Hint
		}

		// Write final state following clack pattern
		var finalMsg string
		if hint != "" {
			finalMsg = fmt.Sprintf("%s\n%s  %s\n%s  %s\n", gray(Bar), symbol, msg, gray(Bar), gray(hint))
		} else {
			finalMsg = fmt.Sprintf("%s\n%s  %s\n%s\n", gray(Bar), symbol, msg, gray(Bar))
		}

		_, _ = p.output.Write([]byte(finalMsg))
	}
}

// animate runs the animation loop
func (p *Progress) animate() {
	for {
		select {
		case <-p.stopChan:
			return
		case <-p.ticker.C:
			p.mu.Lock()

			if p.isActive {
				p.frameIndex = (p.frameIndex + 1) % len(p.frames)
				renderMsg := p.previousMsg
				p.mu.Unlock()
				p.render(renderMsg)
			} else {
				p.mu.Unlock()
			}
		}
	}
}

// render draws the current progress bar frame
func (p *Progress) render(msg string) {
	if p.output == nil {
		return
	}

	// Read current state
	p.mu.Lock()
	progress := float64(p.value) / float64(p.max)
	filled := int(progress * float64(p.size))
	frame := p.frames[p.frameIndex]
	isActive := p.isActive
	lastLength := p.lastFrameLength
	p.mu.Unlock()

	// Get progress character
	char, exists := progressChars[p.style]
	if !exists {
		char = "━" // fallback to heavy
	}

	// Build progress bar
	filledBar := strings.Repeat(char, filled)
	emptyBar := strings.Repeat(char, p.size-filled)

	// Color the progress bar based on state
	var coloredBar string
	if isActive {
		coloredBar = fmt.Sprintf("%s%s",
			cyan(filledBar), // active progress in cyan
			dim(emptyBar)) // remaining progress dimmed
	} else {
		coloredBar = fmt.Sprintf("%s%s",
			green(filledBar), // completed progress in green
			dim(emptyBar))
	}

	// Build frame following the clack visual pattern
	output := fmt.Sprintf("%s\n%s  %s\n%s  %s", gray(Bar), cyan(frame), msg, cyan(Bar), coloredBar)

	// Emit OSC 9;4 set if percent changed, before writing frame
	p.mu.Lock()

	pct := int(progress * 100.0)

	emitSet := pct != p.lastPct
	if emitSet {
		p.lastPct = pct
	}

	p.mu.Unlock()

	if emitSet {
		oscSet(p.output, pct)
	}

	// Clear previous frame if this is not the first render
	if lastLength > 0 {
		// Progress bar has 3 lines, move up 2 and clear down
		_, _ = p.output.Write([]byte("\033[2A\r\033[J"))
	}

	// Write new frame
	_, _ = p.output.Write([]byte(output))

	// Update frame length for next clear
	p.mu.Lock()
	p.lastFrameLength = len(removeAnsiCodes(output))
	p.mu.Unlock()
}

// removeAnsiCodes removes ANSI color codes to get actual display length
func removeAnsiCodes(s string) string {
	// Simple regex to remove ANSI escape sequences
	ansiRegex := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return ansiRegex.ReplaceAllString(s, "")
}
