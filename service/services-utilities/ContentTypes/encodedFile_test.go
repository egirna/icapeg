package ContentTypes

import (
	"encoding/base64"
	"testing"
)

func TestEncodedFile_GetFileFromRequest(t *testing.T) {
	originalContent := "Hello, World!"
	encodedContent := base64.StdEncoding.EncodeToString([]byte(originalContent))
	ef := EncodedFile{data: map[string]interface{}{"Base64": encodedContent}}

	result := ef.GetFileFromRequest()
	if result.String() != originalContent {
		t.Errorf("got %v, want %v", result.String(), originalContent)
	}
}

func TestEncodedFile_BodyAfterScanning(t *testing.T) {
	originalContent := "Hello, World!"
	encodedContent := base64.StdEncoding.EncodeToString([]byte(originalContent))
	ef := EncodedFile{data: make(map[string]interface{})}

	result := ef.BodyAfterScanning([]byte(originalContent))
	if ef.data["Base64"] != encodedContent {
		t.Errorf("got %v, want %v", ef.data["Base64"], encodedContent)
	}

	if result == "" {
		t.Errorf("expected non-empty result")
	}
}
