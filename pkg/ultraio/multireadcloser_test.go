package ultraio

import (
	"fmt"
	"io"
	"io/ioutil"
	"testing"
)

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
