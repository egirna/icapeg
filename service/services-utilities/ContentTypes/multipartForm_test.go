package ContentTypes

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParsingRequest(t *testing.T) {

	TestReq := createhttpRequest("./body.txt")

	_, f, _ := ParsingRequest(TestReq)
	//t.Error(mf)
	fmt.Print(string(f.Content))

	os.WriteFile(f.FileName, f.Content, 0644)
	//t.Error(d)

}
func createhttpRequest(filePath string) (r *http.Request) {

	dat, _ := os.ReadFile(filePath)
	r, _ = http.NewRequest("POST", "http://example.com", nil)
	r.Body = io.NopCloser(bytes.NewBuffer(dat))
	//Content-Type: multipart/form-data; boundary=----WebKitFormBoundaryBzWBc6Z4Ap8XAi3w
	r.Header.Set("Content-Type", "multipart/form-data; boundary=----WebKitFormBoundaryjRM1P6kX6JD71yBA")

	return
}

func FileSave(r *http.Request) string {
	// left shift 32 << 20 which results in 32*2^20 = 33554432
	// x << y, results in x*2^y
	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		return ""
	}
	n := r.Form.Get("name")
	// Retrieve the file from form data
	f, h, err := r.FormFile("file")
	if err != nil {
		return ""
	}
	defer f.Close()
	path := filepath.Join(".", "files")
	_ = os.MkdirAll(path, os.ModePerm)
	fullPath := path + "/" + n
	file, err := os.OpenFile(fullPath, os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		return ""
	}
	defer file.Close()
	// Copy the file to the destination path
	_, err = io.Copy(file, f)
	if err != nil {
		return ""
	}
	return n + filepath.Ext(h.Filename)
}
func Upload(client *http.Client, url string, values map[string]io.Reader) (err error) {
	// Prepare a form that you will submit to that URL.
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	for key, r := range values {
		var fw io.Writer
		if x, ok := r.(io.Closer); ok {
			defer x.Close()
		}
		// Add an image file
		if x, ok := r.(*os.File); ok {
			if fw, err = w.CreateFormFile(key, x.Name()); err != nil {
				return
			}
		} else {
			// Add other fields
			if fw, err = w.CreateFormField(key); err != nil {
				return
			}
		}
		if _, err = io.Copy(fw, r); err != nil {
			return err
		}

	}
	// Don't forget to close the multipart writer.
	// If you don't close it, your request will be missing the terminating boundary.
	w.Close()

	// Now that you have a form, you can submit it to your handler.
	req, err := http.NewRequest("POST", url, &b)
	if err != nil {
		return
	}
	// Don't forget to set the content type, this will contain the boundary.
	req.Header.Set("Content-Type", w.FormDataContentType())

	return
}

func mustOpen(f string) *os.File {
	r, err := os.Open(f)
	if err != nil {
		panic(err)
	}
	return r
}

func ExampleNewReader(req *http.Request) {
	/*
		msg := &mail.Message{
			Header: map[string][]string{
				"Content-Type": {"multipart/mixed; boundary=foo"},
			},
			Body: strings.NewReader(
				"--foo\r\nFoo: one\r\n\r\nA section\r\n" +
					"--foo\r\nFoo: two\r\n\r\nAnd another\r\n" +
					"--foo--\r\n"),
		}*/
	mediaType, params, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
	if err != nil {
		log.Fatal(err)
	}

	if strings.HasPrefix(mediaType, "multipart/") {
		mr := multipart.NewReader(req.Body, params["boundary"])
		for {
			p, err := mr.NextPart()
			if err == io.EOF {
				return
			}
			if err != nil {
				log.Fatal(err)
			}
			slurp, err := io.ReadAll(p)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("Part %q: %q\n", p.Header.Get("Foo"), slurp)
		}
	}

	// Output:
	// Part "one": "A section"
	// Part "two": "And another"
}
