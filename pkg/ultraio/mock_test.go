package ultraio

import (
	"bytes"
	"io"
	"io/ioutil"
	"sync"
	"testing"
)

type mockReadCloser struct {
	mu     sync.Mutex
	closed bool
	data   io.Reader

	pendingCloseError error
	pendingReadError  error
}

func newMockReadCloser(data []byte) *mockReadCloser {
	return &mockReadCloser{
		data: bytes.NewReader(data),
	}
}

func (m *mockReadCloser) Read(p []byte) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.pendingReadError != nil {
		return 0, m.pendingReadError
	}

	return m.data.Read(p)
}

func (m *mockReadCloser) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.closed = true

	return m.pendingCloseError
}

func (m *mockReadCloser) IsClosed() bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.closed
}

// Quick sanity check for our mock with existing code that uses io.Reader
func TestMockReadCloserWorks(t *testing.T) {
	data := []byte("hello")
	m := newMockReadCloser(data)
	r, err := ioutil.ReadAll(m)

	if err != nil {
		t.Errorf("Failed to read: %w", err)
	}

	if string(r) != string(data) {
		t.Errorf("Read %q - want %q", string(r), string(data))
	}
}
