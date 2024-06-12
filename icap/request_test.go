// File: icap/icap_test.go

package icap

import (
	"bufio"
	"bytes"
	"strings"
	"testing"
)

// Helper function to create a bufio.ReadWriter from a string
func createReadWriter(s string) *bufio.ReadWriter {
	reader := strings.NewReader(s)
	r := bufio.NewReader(reader)
	w := bufio.NewWriter(&bytes.Buffer{})
	return bufio.NewReadWriter(r, w)
}

// Test valid ICAP request with REQMOD method
func TestReadRequest_REQMOD(t *testing.T) {
	data := "REQMOD icap://icap-server.net/server?arg=87 ICAP/1.0\r\n" +
		"Host: icap-server.net\r\n" +
		"Encapsulated: req-hdr=0, req-body=51\r\n" + // Correct length of req-body
		"\r\n" +
		"GET / HTTP/1.1\r\n" +
		"Host: www.origin-server.com\r\n" +
		"\r\n" +
		"4\r\nWiki\r\n" +
		"0\r\n" // Correct end of chunked body

	rw := createReadWriter(data)
	req, err := ReadRequest(rw)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if req.Method != "REQMOD" {
		t.Errorf("expected method REQMOD, got %s", req.Method)
	}
	if req.URL.String() != "icap://icap-server.net/server?arg=87" {
		t.Errorf("expected URL icap://icap-server.net/server?arg=87, got %s", req.URL.String())
	}
	if req.Proto != "ICAP/1.0" {
		t.Errorf("expected proto ICAP/1.0, got %s", req.Proto)
	}
	if req.Header.Get("Host") != "icap-server.net" {
		t.Errorf("expected header Host to be icap-server.net, got %s", req.Header.Get("Host"))
	}
}
