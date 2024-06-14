package ContentTypes

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func TestGetContentType(t *testing.T) {
	multipartBody := `--boundary
Content-Disposition: form-data; name="file"; filename="hello.txt"
Content-Type: text/plain

Hello, World!
--boundary--`

	jsonBody := map[string]interface{}{
		"file": base64.StdEncoding.EncodeToString([]byte("Hello, World!")),
	}
	jsonBodyBytes, _ := json.Marshal(jsonBody)

	tests := []struct {
		name         string
		contentType  string
		body         string
		expectedType string
	}{
		{"Multipart form", "multipart/form-data; boundary=boundary", multipartBody, "ContentTypes.MultipartForm"},
		{"JSON with base64 file", "application/json", string(jsonBodyBytes), "ContentTypes.RegularFile"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "http://example.com/upload", strings.NewReader(tt.body))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", tt.contentType)

			contentTypeInstance := GetContentType(req)
			if contentTypeInstance == nil {
				t.Fatalf("expected non-nil contentTypeInstance")
			}

			if contentType := fmt.Sprintf("%T", contentTypeInstance); contentType != tt.expectedType {
				t.Errorf("expected %v, got %v", tt.expectedType, contentType)
			}
		})
	}
}
