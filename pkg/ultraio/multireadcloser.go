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
	closers []io.ReadCloser
	reader  io.Reader
	current int
	mu      sync.Mutex
}

// MultiReadCloser is similar to io.MultiReader, but instead it works
// with io.ReadCloser.  It will read all io.ReadClosers sequentially.
//
// This will close all the io.ReadClosers at once when Close is called.
// Nothing will be closed until Close is called.  Even if some closers
// return an error, all closers will attempt to be closed.
func MultiReadCloser(readClosers ...io.ReadCloser) io.ReadCloser {
	readers := make([]io.Reader, len(readClosers))

	for i, readCloser := range readClosers {
		readers[i] = readCloser
	}

	return &multiReadCloser{
		closers: readClosers,
		reader:  io.MultiReader(readers...),
	}
}

// Close attempts to close all the underlying io.ReadClosers, even if some error
func (m *multiReadCloser) Close() error {
	errs := []string{}

	for _, r := range m.closers {
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

// Read will read from the io.ReadClosers sequentially
func (m *multiReadCloser) Read(p []byte) (int, error) {
	return m.reader.Read(p)
}
