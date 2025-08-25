package tap

import (
	"io"
	"sync"
)

type MockReadable struct {
	buffer    []byte
	closed    bool
	mutex     sync.Mutex
	listeners map[string][]func(string, Key)
}

func NewMockReadable() *MockReadable {
	return &MockReadable{
		listeners: make(map[string][]func(string, Key)),
	}
}

func (m *MockReadable) Read(p []byte) (int, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.closed {
		return 0, io.EOF
	}

	if len(m.buffer) == 0 {
		return 0, nil
	}

	n := copy(p, m.buffer)
	m.buffer = m.buffer[n:]
	return n, nil
}

func (m *MockReadable) Close() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.closed = true
	return nil
}

func (m *MockReadable) On(event string, handler func(string, Key)) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.listeners[event] = append(m.listeners[event], handler)
}

func (m *MockReadable) EmitKeypress(char string, key Key) {
	m.mutex.Lock()
	handlers := m.listeners["keypress"]
	m.mutex.Unlock()

	for _, handler := range handlers {
		handler(char, key)
	}
}

// SendKey is a convenience method for testing
func (m *MockReadable) SendKey(char string, key Key) {
	m.EmitKeypress(char, key)
}

type MockWritable struct {
	Buffer    []string
	mutex     sync.Mutex
	listeners map[string][]func()
}

func NewMockWritable() *MockWritable {
	return &MockWritable{
		Buffer:    make([]string, 0),
		listeners: make(map[string][]func()),
	}
}

func (m *MockWritable) Write(p []byte) (int, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.Buffer = append(m.Buffer, string(p))
	return len(p), nil
}

func (m *MockWritable) On(event string, handler func()) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.listeners[event] = append(m.listeners[event], handler)
}

func (m *MockWritable) Emit(event string) {
	m.mutex.Lock()
	handlers := append([]func(){}, m.listeners[event]...)
	m.mutex.Unlock()

	for _, handler := range handlers {
		handler()
	}
}

// GetFrames returns all written frames for testing
func (m *MockWritable) GetFrames() []string {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return append([]string{}, m.Buffer...)
}
