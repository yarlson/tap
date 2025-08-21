package terminal

import (
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/eiannone/keyboard"
	"github.com/yarlson/tap/pkg/core"
)

// Reader provides terminal input functionality.
type Reader struct {
	mu        sync.Mutex
	listeners map[string][]func(string, core.Key)
}

// Writer provides terminal output functionality.
type Writer struct {
	mu        sync.Mutex
	listeners map[string][]func()
}

// Terminal manages terminal I/O operations.
type Terminal struct {
	Reader  *Reader
	Writer  *Writer
	cleanup func()
}

// New creates a new terminal instance with keyboard input and output handling.
func New() (*Terminal, error) {
	if err := keyboard.Open(); err != nil {
		return nil, err
	}

	reader := &Reader{listeners: make(map[string][]func(string, core.Key))}
	writer := &Writer{listeners: make(map[string][]func())}

	stop := make(chan struct{})

	// Keyboard input handling goroutine
	go func() {
		for {
			select {
			case <-stop:
				return
			default:
			}

			r, key, err := keyboard.GetKey()
			if err != nil {
				continue
			}

			char := string(r)
			name := ""
			ctrl := false

			switch key {
			case keyboard.KeyArrowUp:
				name = "up"
			case keyboard.KeyArrowDown:
				name = "down"
			case keyboard.KeyArrowLeft:
				name = "left"
			case keyboard.KeyArrowRight:
				name = "right"
			case keyboard.KeyEnter:
				name = "return"
			case keyboard.KeyBackspace, keyboard.KeyBackspace2:
				name = "backspace"
			case keyboard.KeyEsc:
				if char == "" {
					char = "escape"
				}
				name = "escape"
			case keyboard.KeyCtrlC:
				char = "\x03"
				name = "c"
				ctrl = true
			case keyboard.KeyDelete:
				name = "delete"
			case keyboard.KeySpace:
				char = " "
				name = "space"
			default:
				if r != 0 {
					char = string(r)
					name = strings.ToLower(string(r))
				}
			}

			k := core.Key{Name: name, Ctrl: ctrl}
			reader.emit(char, k)
		}
	}()

	// Terminal resize notifications
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGWINCH)
	go func() {
		for range sig {
			writer.Emit("resize")
		}
	}()

	cleanup := func() {
		close(stop)
		_ = keyboard.Close()
	}

	return &Terminal{
		Reader:  reader,
		Writer:  writer,
		cleanup: cleanup,
	}, nil
}

// Close releases terminal resources.
func (t *Terminal) Close() {
	if t.cleanup != nil {
		t.cleanup()
	}
}

func (r *Reader) Read(p []byte) (int, error) { return 0, nil }

func (r *Reader) On(event string, handler func(string, core.Key)) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.listeners[event] = append(r.listeners[event], handler)
}

func (r *Reader) emit(char string, key core.Key) {
	r.mu.Lock()
	hs := append([]func(string, core.Key){}, r.listeners["keypress"]...)
	r.mu.Unlock()
	for _, h := range hs {
		h(char, key)
	}
}

func (w *Writer) Write(b []byte) (int, error) { return os.Stdout.Write(b) }

func (w *Writer) On(event string, handler func()) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.listeners[event] = append(w.listeners[event], handler)
}

func (w *Writer) Emit(event string) {
	w.mu.Lock()
	hs := append([]func(){}, w.listeners[event]...)
	w.mu.Unlock()
	for _, h := range hs {
		h()
	}
}
