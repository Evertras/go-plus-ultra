package ultraio

import (
	"bytes"
	"fmt"
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

// Happy path
func TestMultiReadCloserReadsSequentiallyThenClosesAll(t *testing.T) {
	data := [][]byte{
		[]byte("hello"),
		[]byte("some"),
		[]byte("data"),
	}
	const expected = "hellosomedata"

	mockReaders := []*mockReadCloser{}
	readers := []io.ReadCloser{}

	for _, d := range data {
		mockReader := newMockReadCloser(d)
		mockReaders = append(mockReaders, mockReader)
		readers = append(readers, mockReader)
	}

	// Can't use mockReaders here, unfortunately, so we created readers too
	combined := MultiReadCloser(readers...)

	result, err := ioutil.ReadAll(combined)

	if err != nil {
		t.Fatalf("Failed to read: %v", err)
	}

	if string(result) != expected {
		t.Errorf("Got %q - want %q", string(result), expected)
	}

	err = combined.Close()

	if err != nil {
		t.Errorf("Failed to close: %v", err)
	}

	for i, r := range mockReaders {
		if !r.IsClosed() {
			t.Errorf("ReadCloser #%d was not closed", i)
		}
	}
}

// Sad path - Read fails
func TestMultiReadCloserErrorsOnReadWhenAnyReaderErrors(t *testing.T) {
	data := [][]byte{
		[]byte("hello"),
		[]byte("some"),
		[]byte("data"),
	}

	for errIndex := range data {
		t.Run(fmt.Sprintf("Error%d", errIndex), func(t *testing.T) {
			mockReaders := []*mockReadCloser{}
			readers := []io.ReadCloser{}

			for i, d := range data {
				mockReader := newMockReadCloser(d)

				if i == errIndex {
					mockReader.pendingReadError = fmt.Errorf("boom explosions")
				}

				mockReaders = append(mockReaders, mockReader)
				readers = append(readers, mockReader)
			}

			combined := MultiReadCloser(readers...)

			_, err := ioutil.ReadAll(combined)

			if err == nil {
				t.Error("Expected error but it read everything")
			}
		})
	}
}

// Sad path - Close fails
func TestMultiReadCloserErrorsOnCloseWhenAnyReaderErrors(t *testing.T) {
	data := [][]byte{
		[]byte("hello"),
		[]byte("some"),
		[]byte("data"),
	}

	for errIndex := range data {
		t.Run(fmt.Sprintf("Error%d", errIndex), func(t *testing.T) {
			mockReaders := []*mockReadCloser{}
			readers := []io.ReadCloser{}

			for i, d := range data {
				mockReader := newMockReadCloser(d)

				if i == errIndex {
					mockReader.pendingCloseError = fmt.Errorf("boom explosions")
				}

				mockReaders = append(mockReaders, mockReader)
				readers = append(readers, mockReader)
			}

			combined := MultiReadCloser(readers...)

			_, err := ioutil.ReadAll(combined)

			if err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			err = combined.Close()

			if err == nil {
				t.Error("Expected error, but got none")
			}
		})
	}
}
