package ContentTypes

import (
	"bytes"
	"encoding/base64"
	"testing"
)

func TestGetFileFromRequest(t *testing.T) {
	originalContent := "Hello, World!"
	encodedContent := base64.StdEncoding.EncodeToString([]byte(originalContent))

	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{"Encoded content", encodedContent, originalContent},
		{"Plain content", originalContent, originalContent},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rf := RegularFile{Buf: bytes.NewBufferString(tt.content)}
			result := rf.GetFileFromRequest()
			if result.String() != tt.expected {
				t.Errorf("got %v, want %v", result.String(), tt.expected)
			}
		})
	}
}

func TestBodyAfterScanning(t *testing.T) {
	originalContent := "Hello, World!"
	encodedContent := base64.StdEncoding.EncodeToString([]byte(originalContent))

	tests := []struct {
		name     string
		encoded  bool
		content  string
		expected string
	}{
		{"Encoded content", true, originalContent, encodedContent},
		{"Plain content", false, originalContent, originalContent},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rf := RegularFile{Encoded: tt.encoded}
			result := rf.BodyAfterScanning([]byte(tt.content))
			if result != tt.expected {
				t.Errorf("got %v, want %v", result, tt.expected)
			}
		})
	}
}
