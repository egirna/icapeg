package utils

import (
	"net/http"
	"testing"
)

func TestContentType(t *testing.T) {
	type testSample struct {
		r           *http.Response
		contentType string
	}

	const headerContentType = "Content-Type"

	sampleTable := []testSample{
		{
			r: &http.Response{
				Header: http.Header{headerContentType: []string{"text/plain"}},
			},
			contentType: "text/plain",
		},
		{
			r: &http.Response{
				Header: http.Header{headerContentType: []string{"text/html"}},
			},
			contentType: "text/html",
		},
		{
			r: &http.Response{
				Header: http.Header{headerContentType: []string{"application/octet-stream"}},
			},
			contentType: "application/octet-stream",
		},
		{
			r: &http.Response{
				Header: http.Header{headerContentType: []string{"image/png"}},
			},
			contentType: "image/png",
		},
	}

	for _, sample := range sampleTable {
		got := GetContentType(sample.r)
		want := sample.contentType

		if got != want {
			t.Errorf("GetContentType Failed for %s , wanted: %s got: %s", want, want, got)
		}
	}

}

func TestMimeExtension(t *testing.T) {
	type testSample struct {
		data []byte
		ext  string
	}

	sampleTable := []testSample{
		{
			data: []byte{0x42, 0x5A, 0x68},
			ext:  "bz2",
		},
		{
			data: []byte{0x78, 0xDA},
			ext:  "dmg",
		},
		{
			data: []byte{0x58, 0x35},
			ext:  "com",
		},
		{
			data: []byte{0xFF, 0xD8, 0xFF},
			ext:  "jpg",
		},
		{
			data: []byte{0x4D, 0x5A},
			ext:  "exe",
		},
		{
			data: []byte{0x25, 0x50, 0x44, 0x46, 0x2d},
			ext:  "pdf",
		},
		{
			data: []byte{0x52, 0x61, 0x72, 0x21, 0x1A, 0x07, 0x00},
			ext:  "rar",
		},
		{
			data: []byte{},
			ext:  "unknown",
		},
	}

	for _, sample := range sampleTable {
		got := GetMimeExtension(sample.data)
		want := sample.ext

		if got != want {
			t.Errorf("GetMimeExtension Failed for %s , wanted: %s got: %s", want, want, got)
		}
	}

}

func TestFileExtension(t *testing.T) {

	type testSample struct {
		req *http.Request
		ext string
	}

	sampleTable := []testSample{
		{
			req: &http.Request{RequestURI: "http://somehost.com/somefile.pdf?someparam=someval"},
			ext: "pdf",
		},
		{
			req: &http.Request{RequestURI: "http://somehost.com/somefile.exe"},
			ext: "exe",
		},
		{
			req: &http.Request{RequestURI: "http://somehost.com/"},
			ext: "",
		},
	}

	for _, sample := range sampleTable {
		got := GetFileExtension(sample.req)
		want := sample.ext

		if got != want {
			t.Errorf("GetFileExtension Failed for %s , wanted: %s got: %s", sample.req.RequestURI, want, got)
		}
	}

}

func TestInStringSlice(t *testing.T) {
	type testSample struct {
		stringSlice []string
		str         string
		exists      bool
	}

	sampleTable := []testSample{
		{
			stringSlice: []string{"Hello", "World"},
			str:         "hello",
			exists:      false,
		},
		{
			stringSlice: []string{"", "testing", "something else"},
			str:         "testing",
			exists:      true,
		},
	}

	for _, sample := range sampleTable {
		got := InStringSlice(sample.str, sample.stringSlice)
		want := sample.exists

		if got != want {
			t.Errorf("InStringSlice Failed for %v , wanted: %v got: %v", sample.stringSlice, want, got)
		}

	}

}

func TestByteToMegaBytes(t *testing.T) {
	type testSample struct {
		byteNum     int
		megaByteNum float64
	}

	sampleTable := []testSample{
		{
			byteNum:     1000,
			megaByteNum: 0.001,
		},
		{
			byteNum:     3500000,
			megaByteNum: 3.5,
		},
	}

	for _, sample := range sampleTable {
		got := ByteToMegaBytes(sample.byteNum)
		want := sample.megaByteNum

		if got != want {
			t.Errorf("ByteToMegaBytes Failed for %d bytes , wanted: %v mb got: %v mb", sample.byteNum, want, got)
		}

	}
}

func TestBreakURL(t *testing.T) {
	type testSample struct {
		originalURL string
		brokenURL   string
	}

	sampleTable := []testSample{
		{
			originalURL: "http://somehost.com/somefile.exe",
			brokenURL:   "hxxp://somehost.com/somefile.exe",
		},
		{
			originalURL: "https://somehost.com/somefile.exe",
			brokenURL:   "hxxps://somehost.com/somefile.exe",
		},
	}

	for _, sample := range sampleTable {
		got := BreakHTTPURL(sample.originalURL)
		want := sample.brokenURL

		if got != want {
			t.Errorf("BreakHTTPURL Failed for %s , wanted: %s  got: %s", sample.originalURL, want, got)
		}

	}
}
