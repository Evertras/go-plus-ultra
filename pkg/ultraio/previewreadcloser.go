package ultraio

import (
	"bytes"
	"io"
	"io/ioutil"
)

// PreviewReadCloser will read ahead the given number of bytes and return both
// the previewed data and a new io.ReadCloser which will act as if it was never
// read in the first place, allowing you to still consume it from the beginning
func PreviewReadCloser(stream io.ReadCloser, n int) ([]byte, io.ReadCloser, error) {
	if n == 0 {
		return nil, stream, nil
	}

	previewed := make([]byte, n)

	// Note that we don't check the error immediately here; Read says:
	//
	// Callers should always process the n > 0 bytes returned before
	// considering the error err. Doing so correctly handles I/O errors...
	n, err := stream.Read(previewed)

	// This is a special case where the stream we were given is already
	// at the end, so just return an empty preview and let the next Read
	// discover the same error
	if err == io.EOF {
		return []byte{}, stream, nil
	}

	// Trim extra nulls at the end if we didn't read the full requested amount
	previewed = previewed[:n]

	previewedStream := ioutil.NopCloser(bytes.NewReader(previewed))
	resetStream := MultiReadCloser(previewedStream, stream)

	return previewed, resetStream, err
}

// PreviewReadAllCloser will read the entire stream into memory, but not close it.
// Reading from the stream will use the in-memory copy, and calling Close will
// close the underlying original stream.
func PreviewReadAllCloser(stream io.ReadCloser) ([]byte, io.ReadCloser, error) {
	previewed, err := ioutil.ReadAll(stream)

	// This is a special case where the stream we were given is already
	// at the end, so just return an empty preview and let the next Read
	// discover the same error
	if err == io.EOF {
		return []byte{}, stream, nil
	}

	previewedStream := ioutil.NopCloser(bytes.NewReader(previewed))
	resetStream := MultiReadCloser(previewedStream, stream)

	return previewed, resetStream, err
}
