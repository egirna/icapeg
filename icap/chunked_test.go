package icap

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
)

// TestNewChunkedReader tests the creation of a new chunkedReader
func TestNewChunkedReader(t *testing.T) {
	data := "4\r\nWiki\r\n5\r\npedia\r\nE\r\n in\r\n\r\nchunks.\r\n0\r\n\r\n"
	reader := newChunkedReader(bytes.NewReader([]byte(data)))

	expected := "Wikipedia in\r\n\r\nchunks."
	buf := new(bytes.Buffer)
	_, err := io.Copy(buf, reader)
	if err != nil {
		t.Errorf("Error reading chunked data: %v", err)
	}

	if buf.String() != expected {
		t.Errorf("Expected %q, got %q", expected, buf.String())
	}
}

// TestNewChunkedWriter tests the creation of a new chunkedWriter
func TestNewChunkedWriter(t *testing.T) {
	var buf bytes.Buffer
	writer := NewChunkedWriter(&buf)

	data := []byte("Wikipedia in\r\n\r\nchunks.")
	_, err := writer.Write(data)
	if err != nil {
		t.Errorf("Error writing chunked data: %v", err)
	}
	writer.Close()

	// The length of "Wikipedia in\r\n\r\nchunks." is 23 bytes, so the chunk length should be "17\r\n" in hexadecimal
	expected := "17\r\nWikipedia in\r\n\r\nchunks.\r\n0\r\n\r\n"

	// Trim trailing newlines for comparison
	expected = strings.TrimSpace(expected)
	actual := strings.TrimSpace(buf.String())

	if actual != expected {
		t.Errorf("Expected %q, got %q", expected, actual)
	}
}

// TestParseHexUint tests the parsing of a hexadecimal number
func TestParseHexUint(t *testing.T) {
	tests := []struct {
		input    []byte
		expected uint64
		err      error
	}{
		{[]byte("4"), 4, nil},
		{[]byte("1A"), 26, nil},
		{[]byte("1a"), 26, nil},
		{[]byte("G"), 0, errors.New("invalid chunk length: 'G'")},
	}

	for _, test := range tests {
		result, err := parseHexUint(test.input)
		if result != test.expected || (err != nil && err.Error() != test.err.Error()) {
			t.Errorf("parseHexUint(%q) = %d, %v; want %d, %v", test.input, result, err, test.expected, test.err)
		}
	}
}
