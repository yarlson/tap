package core

import (
	"bufio"
	"os"
	"sync"
	"time"

	"golang.org/x/term"
)

type defaultTTYInput struct {
	f         *os.File
	oldState  *term.State
	mu        sync.Mutex
	listeners map[string][]func(string, Key)
	stopCh    chan struct{}
}

type defaultTTYOutput struct {
	f         *os.File
	mu        sync.Mutex
	listeners map[string][]func()
}

func newDefaultTTYInput(f *os.File) (*defaultTTYInput, func(), error) {
	old, err := term.MakeRaw(int(f.Fd()))
	if err != nil {
		return nil, nil, err
	}
	d := &defaultTTYInput{
		f:         f,
		oldState:  old,
		listeners: make(map[string][]func(string, Key)),
		stopCh:    make(chan struct{}),
	}
	go d.readLoop()
	restore := func() { _ = term.Restore(int(f.Fd()), old) }
	return d, restore, nil
}

func (d *defaultTTYInput) Read(p []byte) (int, error) { return d.f.Read(p) }

func (d *defaultTTYInput) On(event string, handler func(string, Key)) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.listeners[event] = append(d.listeners[event], handler)
}

func (d *defaultTTYInput) emitKey(char string, key Key) {
	d.mu.Lock()
	hs := append([]func(string, Key){}, d.listeners["keypress"]...)
	d.mu.Unlock()
	for _, h := range hs {
		h(char, key)
	}
}

func (d *defaultTTYInput) readLoop() {
	br := bufio.NewReader(d.f)
	for {
		select {
		case <-d.stopCh:
			return
		default:
		}
		_ = d.f.SetReadDeadline(time.Now().Add(50 * time.Millisecond))
		b, err := br.ReadByte()
		if err != nil {
			continue
		}
		switch b {
		case 0x03:
			d.emitKey("\x03", Key{Name: "c", Ctrl: true})
		case 0x1b:
			b1, _ := br.Peek(2)
			if len(b1) == 2 && b1[0] == '[' {
				_, _ = br.ReadByte()
				dir, _ := br.ReadByte()
				switch dir {
				case 'A':
					d.emitKey("", Key{Name: "up"})
				case 'B':
					d.emitKey("", Key{Name: "down"})
				case 'C':
					d.emitKey("", Key{Name: "right"})
				case 'D':
					d.emitKey("", Key{Name: "left"})
				default:
					d.emitKey("escape", Key{Name: "escape"})
				}
			} else {
				d.emitKey("escape", Key{Name: "escape"})
			}
		case '\r', '\n':
			d.emitKey("", Key{Name: "return"})
		default:
			c := rune(b)
			if c >= 'A' && c <= 'Z' {
				c = c - 'A' + 'a'
			}
			ch := string(c)
			d.emitKey(ch, Key{Name: ch})
		}
	}
}

func newDefaultTTYOutput(f *os.File) *defaultTTYOutput {
	return &defaultTTYOutput{f: f, listeners: make(map[string][]func())}
}

func (o *defaultTTYOutput) Write(p []byte) (int, error) {
	// Map LF to CRLF to reset column when printing lines while in raw mode
	buf := make([]byte, 0, len(p)*2)
	for _, b := range p {
		if b == '\n' {
			buf = append(buf, '\r', '\n')
		} else {
			buf = append(buf, b)
		}
	}

	return o.f.Write(buf)
}

func (o *defaultTTYOutput) On(event string, handler func()) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.listeners[event] = append(o.listeners[event], handler)
}

func (o *defaultTTYOutput) Emit(event string) {
	o.mu.Lock()
	hs := append([]func(){}, o.listeners[event]...)
	o.mu.Unlock()
	for _, h := range hs {
		h()
	}
}
