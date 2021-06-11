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

	// Trim extra nulls at the end if we didn't read the full requested amount
	previewed = previewed[:n]

	previewedStream := ioutil.NopCloser(bytes.NewReader(previewed))
	resetStream := MultiReadCloser(previewedStream, stream)

	return previewed, resetStream, err
}
