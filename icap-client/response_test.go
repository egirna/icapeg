package icapclient

import (
	"bufio"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

func TestResponse(t *testing.T) {

	t.Run("ReadResponse REQMOD", func(t *testing.T) { // FIXME: headers and content request aren't being tested properly

		type testSample struct {
			headers      http.Header
			status       string
			statusCode   int
			previewBytes int
			respStr      string
			httpReqStr   string
		}

		sampleTable := []testSample{
			{
				headers: http.Header{
					"Date":         []string{"Mon, 10 Jan 2000  09:55:21 GMT"},
					"Server":       []string{"ICAP-Server-Software/1.0"},
					"Istag":        []string{"\"W3E4R7U9-L2E4-2\""},
					"Encapsulated": []string{"req-hdr=0, null-body=231"},
				},
				status:       "OK",
				statusCode:   200,
				previewBytes: 0,
				respStr: "ICAP/1.0 200 OK\r\n" +
					"Date: Mon, 10 Jan 2000  09:55:21 GMT\r\n" +
					"Server: ICAP-Server-Software/1.0\r\n" +
					"Connection: close\r\n" +
					"Istag: \"W3E4R7U9-L2E4-2\"\r\n" +
					"Encapsulated: req-hdr=0, null-body=231\r\n\r\n",
				httpReqStr: "GET /modified-path HTTP/1.1\r\n" +
					"Host: www.origin-server.com\r\n" +
					"Via: 1.0 icap-server.net (ICAP Example ReqMod Service 1.1)\r\n" +
					"Accept: text/html, text/plain, image/gif\r\n" +
					"Accept-Encoding: gzip, compress\r\n" +
					"If-None-Match: \"xyzzy\", \"r2d2xxxx\"\r\n\r\n",
			},
			{
				headers: http.Header{
					"Date":         []string{"Mon, 10 Jan 2000  09:55:21 GMT"},
					"Server":       []string{"ICAP-Server-Software/1.0"},
					"Istag":        []string{"\"W3E4R7U9-L2E4-2\""},
					"Encapsulated": []string{"req-hdr=0, req-body=244"},
				},
				status:       "OK",
				statusCode:   200,
				previewBytes: 0,
				respStr: "ICAP/1.0 200 OK\r\n" +
					"Date: Mon, 10 Jan 2000  09:55:21 GMT\r\n" +
					"Server: ICAP-Server-Software/1.0\r\n" +
					"Connection: close\r\n" +
					"Istag: \"W3E4R7U9-L2E4-2\"\r\n" +
					"Encapsulated: req-hdr=0, req-body=244\r\n\r\n",
				httpReqStr: "POST /origin-resource/form.pl HTTP/1.1\r\n" +
					"Host: www.origin-server.com\r\n" +
					"Via: 1.0 icap-server.net (ICAP Example ReqMod Service 1.1)\r\n" +
					"Accept: text/html, text/plain, image/gif\r\n" +
					"Accept-Encoding: gzip, compress\r\n" +
					"Pragma: no-cache\r\n" +
					"Content-Length: 45\r\n\r\n" +
					"2d\r\n" +
					"I am posting this information.  ICAP powered!\r\n" +
					"0\r\n\r\n",
			},
		}

		for _, sample := range sampleTable {
			resp, err := ReadResponse(bufio.NewReader(strings.NewReader(sample.respStr + sample.httpReqStr)))
			if err != nil {
				t.Fatal(err.Error())
			}

			if resp.StatusCode != sample.statusCode {
				t.Logf("Wanted ICAP status code: %d , got: %d", sample.statusCode, resp.StatusCode)
				t.Fail()
			}
			if resp.Status != sample.status {
				t.Logf("Wanted ICAP status: %s , got: %s", sample.status, resp.Status)
				t.Fail()
			}
			if resp.PreviewBytes != sample.previewBytes {
				t.Logf("Wanted preview bytes: %d, got: %d", sample.previewBytes, resp.PreviewBytes)
				t.Fail()
			}

			for k, v := range sample.headers {
				if val, exists := resp.Header[k]; !exists || !reflect.DeepEqual(val, v) {
					t.Logf("Wanted Header: %s with value: %v, got: %v", k, v, val)
					t.Fail()
					break
				}
			}
			if resp.ContentRequest == nil {
				t.Log("ContentRequest is nil")
				t.Fail()
			}

			wantedHTTPReq, err := http.ReadRequest(bufio.NewReader(strings.NewReader(sample.httpReqStr)))
			if err != nil {
				t.Fatal(err.Error())
			}

			if !reflect.DeepEqual(resp.ContentRequest, wantedHTTPReq) {
				t.Logf("Wanted http request: %v, got: %v", wantedHTTPReq, resp.ContentRequest)
				t.Fail()
			}

		}

	})

	t.Run("ReadResponse RESPMOD", func(t *testing.T) {
		type testSample struct {
			headers      http.Header
			status       string
			statusCode   int
			previewBytes int
			respStr      string
			httpRespStr  string
			httpReqStr   string
		}

		sampleTable := []testSample{
			{
				headers: http.Header{
					"Date":         []string{"Mon, 10 Jan 2000  09:55:21 GMT"},
					"Server":       []string{"ICAP-Server-Software/1.0"},
					"Istag":        []string{"\"W3E4R7U9-L2E4-2\""},
					"Encapsulated": []string{"req-hdr=0, res-body=222"},
				},
				status:       "OK",
				statusCode:   200,
				previewBytes: 0,
				respStr: "ICAP/1.0 200 OK\r\n" +
					"Date: Mon, 10 Jan 2000  09:55:21 GMT\r\n" +
					"Server: ICAP-Server-Software/1.0\r\n" +
					"Connection: close\r\n" +
					"ISTag: \"W3E4R7U9-L2E4-2\"\r\n" +
					"Encapsulated: req-hdr=0, res-body=222\r\n\r\n",
				httpRespStr: "HTTP/1.1 200 OK\r\n" +
					"Date: Mon, 10 Jan 2000  09:55:21 GMT\r\n" +
					"Via: 1.0 icap.example.org (ICAP Example RespMod Service 1.1)\r\n" +
					"Server: Apache/1.3.6 (Unix)\r\n" +
					"ETag: \"63840-1ab7-378d415b\"\r\n" +
					"Content-Type: text/plain\r\n" +
					"Content-Length: 92\r\n\r\n" +
					"5c\r\n" +
					"This is data that was returned by an origin server, but with value added by an ICAP server.\r\n" +
					"0\r\n\r\n",
			},
		}

		for _, sample := range sampleTable {
			resp, err := ReadResponse(bufio.NewReader(strings.NewReader(sample.respStr + sample.httpRespStr)))
			if err != nil {
				t.Fatal(err.Error())
			}

			if resp.StatusCode != sample.statusCode {
				t.Logf("Wanted ICAP status code: %d , got: %d", sample.statusCode, resp.StatusCode)
				t.Fail()
			}
			if resp.Status != sample.status {
				t.Logf("Wanted ICAP status: %s , got: %s", sample.status, resp.Status)
				t.Fail()
			}
			if resp.PreviewBytes != sample.previewBytes {
				t.Logf("Wanted preview bytes: %d, got: %d", sample.previewBytes, resp.PreviewBytes)
				t.Fail()
			}

			for k, v := range sample.headers {
				if val, exists := resp.Header[k]; !exists || !reflect.DeepEqual(val, v) {
					t.Logf("Wanted Header: %s with value: %v, got: %v", k, v, val)
					t.Fail()
					break
				}
			}
			if resp.ContentResponse == nil {
				t.Log("ContentResponse is nil")
				t.Fail()
			}

			wantedHTTPResp, err := http.ReadResponse(bufio.NewReader(strings.NewReader(sample.httpRespStr)), nil)
			if err != nil {
				t.Fatal(err.Error())
			}

			if !reflect.DeepEqual(resp.ContentResponse, wantedHTTPResp) {
				t.Logf("Wanted http response: %v, got: %v", wantedHTTPResp, resp.ContentResponse)
				t.Fail()
			}

		}

	})

}
