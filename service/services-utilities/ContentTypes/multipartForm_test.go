package ContentTypes

import (
	"net/http"
	"strings"
	"testing"
)

func TestMultipartForm_GetFileFromRequest(t *testing.T) {
	content := "Hello, World!"
	formPart := FormPart{Content: []byte(content)}
	mf := MultipartForm{theFile: formPart}

	result := mf.GetFileFromRequest()
	if result.String() != content {
		t.Errorf("got %v, want %v", result.String(), content)
	}
}

func TestParsingRequest(t *testing.T) {
	body := `--boundary
Content-Disposition: form-data; name="file"; filename="hello.txt"
Content-Type: text/plain

Hello, World!
--boundary--`
	req, err := http.NewRequest("POST", "http://example.com/upload", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "multipart/form-data; boundary=boundary")

	formParts, theFile, boundary := ParsingRequest(req)
	if len(formParts) == 0 || boundary != "boundary" || string(theFile.Content) != "Hello, World!" {
		t.Errorf("ParsingRequest did not parse the request correctly")
	}
}
