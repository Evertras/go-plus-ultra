package ultraio

import (
	"fmt"
	"io"
	"strings"
	"sync"
)

// multiReadCloser is the underlying implementation that handles multiple
// io.ReadClosers sequentially and in a thread-safe way
type multiReadCloser struct {
	readClosers []io.ReadCloser
	current     int
	mu          sync.Mutex
}

// MultiReadCloser is similar to io.MultiReader, but instead it works
// with io.ReadCloser.  It will read all io.ReadClosers sequentially.
//
// This will close all the io.ReadClosers at once when Close is called.
// Nothing will be closed until Close is called.  Even if some closers
// return an error, all closers will attempt to be closed.
//
// It will only read from a single ReadCloser at a time, so multiple
// calls to Read are required for each reader.  Any code using io.Reader
// or io.ReadCloser should already naturally handle this, but noting it
// here anyway.
func MultiReadCloser(readClosers ...io.ReadCloser) io.ReadCloser {
	return &multiReadCloser{
		readClosers: readClosers,
	}
}

// Close attempts to close all the underlying io.ReadClosers, even if some error
func (m *multiReadCloser) Close() error {
	errs := []string{}

	for _, r := range m.readClosers {
		err := r.Close()

		if err != nil {
			errs = append(errs, err.Error())
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("at least one underlying io.ReadCloser failed to close: %s", strings.Join(errs, ", "))
	}

	return nil
}

// Read will read from the current io.ReadCloser, then move on to the next once
// the current one is finished.
func (m *multiReadCloser) Read(p []byte) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	currentReader := m.readClosers[m.current]

	n, err := currentReader.Read(p)

	// If we finish a ReadCloser, move on to the next
	if err == io.EOF {
		m.current = m.current + 1

		// We're done
		if m.current == len(m.readClosers) {
			return n, io.EOF
		}

		// Otherwise, we'll pick up on the next Read call
		return n, nil
	}

	return n, err
}
