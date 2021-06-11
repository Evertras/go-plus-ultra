package ultraio

import (
	"io/ioutil"
	"testing"
)

// Happy path where the stream has more bytes than requested
func TestPreviewReadCloserWorks(t *testing.T) {
	tests := []struct {
		name            string
		contents        string
		previewLength   int
		expectedPreview string
	}{
		{
			name:            "SufficientBytesExist",
			contents:        "Hello this is a long stream",
			previewLength:   len("Hello"),
			expectedPreview: "Hello",
		},
		{
			name:            "StreamIsSmaller",
			contents:        "Hello",
			previewLength:   len("Hello this is more than what is in the stream"),
			expectedPreview: "Hello",
		},
		{
			name:            "ZeroBytesRequested",
			contents:        "Hello",
			previewLength:   0,
			expectedPreview: "",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			mockReader := newMockReadCloser([]byte(test.contents))

			previewed, stream, err := PreviewReadCloser(mockReader, test.previewLength)

			if err != nil {
				t.Fatalf("Unexpected error from PreviewReadCloser: %v", err)
			}

			if stream == nil {
				t.Fatal("Returned stream is nil")
			}

			if string(previewed) != test.expectedPreview {
				t.Errorf("Incorrect preview: Got %q - want %q", string(previewed), test.expectedPreview)
			}

			results, err := ioutil.ReadAll(stream)

			if string(results) != test.contents {
				t.Errorf("Unexpected contents from full read: Got %q - want %q", string(results), test.contents)
			}

			if err != nil {
				t.Errorf("Unexpected error from ioutil.ReadAll: %v", err)
			}
		})
	}
}
