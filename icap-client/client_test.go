package icapclient

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestClient(t *testing.T) {
	if !testServerRunning() {
		go startTestServer()
	}

	t.Run("Client Do RESPMOD", func(t *testing.T) {

		httpReq, err := http.NewRequest(http.MethodGet, "http://someurl.com", nil)
		if err != nil {
			t.Log(err.Error())
			t.Fail()
			return
		}

		type testSample struct {
			httpResp            *http.Response
			wantedStatusCode    int
			wantedStatus        string
			wantedTimeout       time.Duration
			wantedDialerTimeout time.Duration
			wantedReadTimeout   time.Duration
			wantedWriteTimeout  time.Duration
		}

		sampleTable := []testSample{
			{
				httpResp: &http.Response{
					Status:     "200 OK",
					StatusCode: http.StatusOK,
					Proto:      "HTTP/1.0",
					ProtoMajor: 1,
					ProtoMinor: 0,
					Header: http.Header{
						"Content-Type":   []string{"plain/text"},
						"Content-Length": []string{"19"},
					},
					ContentLength: 19,
					Body:          ioutil.NopCloser(strings.NewReader("This is a GOOD FILE")),
				},
				wantedStatusCode:    http.StatusNoContent,
				wantedStatus:        "No Modifications",
				wantedTimeout:       defaultTimeout,
				wantedDialerTimeout: defaultTimeout,
				wantedReadTimeout:   defaultTimeout,
				wantedWriteTimeout:  defaultTimeout,
			},
			{
				httpResp: &http.Response{
					Status:     "200 OK",
					StatusCode: http.StatusOK,
					Proto:      "HTTP/1.0",
					ProtoMajor: 1,
					ProtoMinor: 0,
					Header: http.Header{
						"Content-Type":   []string{"plain/text"},
						"Content-Length": []string{"18"},
					},
					ContentLength: 18,
					Body:          ioutil.NopCloser(strings.NewReader("This is a BAD FILE")),
				},
				wantedStatusCode:    http.StatusOK,
				wantedStatus:        "OK",
				wantedTimeout:       defaultTimeout,
				wantedDialerTimeout: defaultTimeout,
				wantedReadTimeout:   defaultTimeout,
				wantedWriteTimeout:  defaultTimeout,
			},
		}

		for _, sample := range sampleTable {
			req, err := NewRequest(MethodRESPMOD, fmt.Sprintf("icap://localhost:%d/respmod", port), httpReq, sample.httpResp)
			if err != nil {
				t.Log(err.Error())
				t.Fail()
				return
			}

			client := Client{}
			resp, err := client.Do(req)
			if err != nil {
				t.Log(err.Error())
				t.Fail()
				return
			}

			if resp.StatusCode != sample.wantedStatusCode {
				t.Logf("Wanted status code:%d, got:%d", sample.wantedStatusCode, resp.StatusCode)
				t.Fail()
			}

			if resp.Status != sample.wantedStatus {
				t.Logf("Wanted status:%s, got:%s", sample.wantedStatus, resp.Status)
				t.Fail()
			}

			if client.Timeout != sample.wantedTimeout {
				t.Logf("Wanted timeout to be:%v, got:%v", sample.wantedTimeout, client.Timeout)
				t.Fail()
			}
			if client.scktDriver.DialerTimeout != sample.wantedDialerTimeout {
				t.Logf("Wanted DialierTimeout to be:%v , got:%v", sample.wantedDialerTimeout, client.scktDriver.DialerTimeout)
				t.Fail()
			}
			if client.scktDriver.ReadTimeout != sample.wantedReadTimeout {
				t.Logf("Wanted ReadTimeout to be:%v , got:%v", sample.wantedReadTimeout, client.scktDriver.ReadTimeout)
				t.Fail()
			}
			if client.scktDriver.WriteTimeout != sample.wantedWriteTimeout {
				t.Logf("Wanted WriteTimeout to be:%v , got:%v", sample.wantedWriteTimeout, client.scktDriver.WriteTimeout)
				t.Fail()
			}
		}

	})

	t.Run("Client Do REQMOD", func(t *testing.T) {

		type testSample struct {
			urlStr              string
			wantedStatusCode    int
			wantedStatus        string
			wantedTimeout       time.Duration
			wantedDialerTimeout time.Duration
			wantedReadTimeout   time.Duration
			wantedWriteTimeout  time.Duration
		}

		sampleTable := []testSample{
			{
				urlStr:              "http://goodifle.com",
				wantedStatusCode:    http.StatusNoContent,
				wantedStatus:        "No Modifications",
				wantedTimeout:       defaultTimeout,
				wantedDialerTimeout: defaultTimeout,
				wantedReadTimeout:   defaultTimeout,
				wantedWriteTimeout:  defaultTimeout,
			},
			{
				urlStr:              "http://badfile.com",
				wantedStatusCode:    http.StatusOK,
				wantedStatus:        "OK",
				wantedTimeout:       defaultTimeout,
				wantedDialerTimeout: defaultTimeout,
				wantedReadTimeout:   defaultTimeout,
				wantedWriteTimeout:  defaultTimeout,
			},
		}

		for _, sample := range sampleTable {
			httpReq, err := http.NewRequest(http.MethodGet, sample.urlStr, nil)
			if err != nil {
				t.Log(err.Error())
				t.Fail()
				return
			}

			req, err := NewRequest(MethodREQMOD, fmt.Sprintf("icap://localhost:%d/reqmod", port), httpReq, nil)
			if err != nil {
				t.Log(err.Error())
				t.Fail()
				return
			}

			client := Client{}
			resp, err := client.Do(req)
			if err != nil {
				t.Log(err.Error())
				t.Fail()
				return
			}

			if resp.StatusCode != sample.wantedStatusCode {
				t.Logf("Wanted status code:%d, got:%d", sample.wantedStatusCode, resp.StatusCode)
				t.Fail()
			}

			if resp.Status != sample.wantedStatus {
				t.Logf("Wanted status:%s, got:%s", sample.wantedStatus, resp.Status)
				t.Fail()
			}

			if client.Timeout != sample.wantedTimeout {
				t.Logf("Wanted timeout to be:%v, got:%v", sample.wantedTimeout, client.Timeout)
				t.Fail()
			}
			if client.scktDriver.DialerTimeout != sample.wantedDialerTimeout {
				t.Logf("Wanted DialierTimeout to be:%v , got:%v", sample.wantedDialerTimeout, client.scktDriver.DialerTimeout)
				t.Fail()
			}
			if client.scktDriver.ReadTimeout != sample.wantedReadTimeout {
				t.Logf("Wanted ReadTimeout to be:%v , got:%v", sample.wantedReadTimeout, client.scktDriver.ReadTimeout)
				t.Fail()
			}
			if client.scktDriver.WriteTimeout != sample.wantedWriteTimeout {
				t.Logf("Wanted WriteTimeout to be:%v , got:%v", sample.wantedWriteTimeout, client.scktDriver.WriteTimeout)
				t.Fail()
			}

		}
	})

	t.Run("Clien Do RESPMOD with OPTIONS", func(t *testing.T) {

		httpReq, err := http.NewRequest(http.MethodGet, "http://someurl.com", nil)
		if err != nil {
			t.Log(err.Error())
			t.Fail()
			return
		}

		type testSample struct {
			httpResp               *http.Response
			wantedStatusCode       int
			wantedStatus           string
			wantedPreviewBytes     int
			wantedOptionStatusCode int
			wantedOptionStatus     string
			wantedOptionHeader     http.Header
		}

		sampleTable := []testSample{
			{
				httpResp: &http.Response{
					Status:     "200 OK",
					StatusCode: http.StatusOK,
					Proto:      "HTTP/1.0",
					ProtoMajor: 1,
					ProtoMinor: 0,
					Header: http.Header{
						"Content-Type":   []string{"plain/text"},
						"Content-Length": []string{"41"},
					},
					ContentLength: 41,
					Body:          ioutil.NopCloser(strings.NewReader("Hello World!This is a GOOD FILE! bye bye!")),
				},
				wantedStatusCode:       http.StatusNoContent,
				wantedStatus:           "No Modifications",
				wantedPreviewBytes:     previewBytes,
				wantedOptionStatusCode: http.StatusOK,
				wantedOptionStatus:     "OK",
				wantedOptionHeader: http.Header{
					"Methods":          []string{"RESPMOD"},
					"Allow":            []string{"204"},
					"Preview":          []string{strconv.Itoa(previewBytes)},
					"Transfer-Preview": []string{"*"},
				},
			},
			{
				httpResp: &http.Response{
					Status:     "200 OK",
					StatusCode: http.StatusOK,
					Proto:      "HTTP/1.0",
					ProtoMajor: 1,
					ProtoMinor: 0,
					Header: http.Header{
						"Content-Type":   []string{"plain/text"},
						"Content-Length": []string{"18"},
					},
					ContentLength: 18,
					Body:          ioutil.NopCloser(strings.NewReader("This is a BAD FILE")),
				},
				wantedStatusCode:       http.StatusOK,
				wantedStatus:           "OK",
				wantedPreviewBytes:     previewBytes,
				wantedOptionStatusCode: http.StatusOK,
				wantedOptionStatus:     "OK",
				wantedOptionHeader: http.Header{
					"Methods":          []string{"RESPMOD"},
					"Allow":            []string{"204"},
					"Preview":          []string{strconv.Itoa(previewBytes)},
					"Transfer-Preview": []string{"*"},
				},
			},
		}

		for _, sample := range sampleTable {

			urlStr := fmt.Sprintf("icap://localhost:%d/respmod", port)

			optReq, err := NewRequest(MethodOPTIONS, urlStr, nil, nil)
			if err != nil {
				t.Log(err.Error())
				t.Fail()
				return
			}

			client := Client{}
			optResp, err := client.Do(optReq)
			if err != nil {
				t.Log(err.Error())
				t.Fail()
				return
			}

			if optResp.Status != sample.wantedOptionStatus {
				t.Logf("Wanted status:%s, got:%s", sample.wantedOptionStatus, optResp.Status)
				t.Fail()
			}

			if optResp.StatusCode != sample.wantedOptionStatusCode {
				t.Logf("Wanted status code:%d, got:%d", sample.wantedOptionStatusCode, optResp.StatusCode)
				t.Fail()
			}

			if optResp.PreviewBytes != sample.wantedPreviewBytes {
				t.Logf("Wanted preview bytes:%d , got:%d", sample.wantedPreviewBytes, optResp.PreviewBytes)
				t.Fail()
			}

			for k, v := range sample.wantedOptionHeader {
				if val, exists := optResp.Header[k]; exists {
					if !reflect.DeepEqual(val, v) {
						t.Logf("Wanted value for header:%s to be:%v, got:%v", k, v, val)
						t.Fail()
					}
					continue
				}

				t.Logf("Expected header:%s but not found", k)
				t.Fail()

			}

			req, err := NewRequest(MethodRESPMOD, urlStr, httpReq, sample.httpResp)
			if err != nil {
				t.Log(err.Error())
				t.Fail()
				return
			}

			if err := req.ExtendHeader(optResp.Header); err != nil {
				t.Log(err.Error())
				t.Fail()
				return
			}

			resp, err := client.Do(req)
			if err != nil {
				t.Log(err.Error())
				t.Fail()
				return
			}

			if resp.StatusCode != sample.wantedStatusCode {
				t.Logf("Wanted status code:%d, got:%d", sample.wantedStatusCode, resp.StatusCode)
				t.Fail()
			}

			if resp.Status != sample.wantedStatus {
				t.Logf("Wanted status:%s, got:%s", sample.wantedStatus, resp.Status)
				t.Fail()
			}

		}
	})

	t.Run("Client Do REQMOD with OPTIONS", func(t *testing.T) {
		type testSample struct {
			urlStr                 string
			wantedStatusCode       int
			wantedStatus           string
			wantedOptionStatus     string
			wantedOptionStatusCode int
			wantedOptionHeader     http.Header
		}

		sampleTable := []testSample{
			{
				urlStr:                 "http://goodifle.com",
				wantedStatusCode:       http.StatusNoContent,
				wantedStatus:           "No Modifications",
				wantedOptionStatus:     "OK",
				wantedOptionStatusCode: http.StatusOK,
				wantedOptionHeader: http.Header{
					"Methods":          []string{"REQMOD"},
					"Allow":            []string{"204"},
					"Preview":          []string{strconv.Itoa(previewBytes)},
					"Transfer-Preview": []string{"*"},
				},
			},
			{
				urlStr:                 "http://badfile.com",
				wantedStatusCode:       http.StatusOK,
				wantedStatus:           "OK",
				wantedOptionStatus:     "OK",
				wantedOptionStatusCode: http.StatusOK,
				wantedOptionHeader: http.Header{
					"Methods":          []string{"REQMOD"},
					"Allow":            []string{"204"},
					"Preview":          []string{strconv.Itoa(previewBytes)},
					"Transfer-Preview": []string{"*"},
				},
			},
		}

		for _, sample := range sampleTable {

			client := Client{}
			urlStr := fmt.Sprintf("icap://localhost:%d/reqmod", port)

			optReq, err := NewRequest(MethodOPTIONS, urlStr, nil, nil)
			if err != nil {
				t.Log(err.Error())
				t.Fail()
				return
			}

			optResp, err := client.Do(optReq)
			if err != nil {
				t.Log(err.Error())
				t.Fail()
				return
			}

			if optResp.Status != sample.wantedOptionStatus {
				t.Logf("Wanted status:%s , got:%s", sample.wantedOptionStatus, optResp.Status)
				t.Fail()
			}
			if optResp.StatusCode != sample.wantedOptionStatusCode {
				t.Logf("Wanted status code:%d , got:%d", sample.wantedOptionStatusCode, optResp.StatusCode)
				t.Fail()
			}
			for k, v := range sample.wantedOptionHeader {
				if val, exists := optResp.Header[k]; exists {
					if !reflect.DeepEqual(val, v) {
						t.Logf("Wanted header:%s to have value:%v, got:%v", k, v, val)
						t.Fail()
					}
					continue
				}

				t.Logf("Expected header:%s but not found", k)
				t.Fail()
			}

			httpReq, err := http.NewRequest(http.MethodGet, sample.urlStr, nil)
			if err != nil {
				t.Log(err.Error())
				t.Fail()
				return
			}

			req, err := NewRequest(MethodREQMOD, urlStr, httpReq, nil)
			if err != nil {
				t.Log(err.Error())
				t.Fail()
				return
			}

			if err := req.ExtendHeader(optResp.Header); err != nil {
				t.Log(err.Error())
				t.Fail()
				return
			}

			resp, err := client.Do(req)
			if err != nil {
				t.Log(err.Error())
				t.Fail()
				return
			}

			if resp.StatusCode != sample.wantedStatusCode {
				t.Logf("Wanted status code:%d, got:%d", sample.wantedStatusCode, resp.StatusCode)
				t.Fail()
			}

			if resp.Status != sample.wantedStatus {
				t.Logf("Wanted status:%s, got:%s", sample.wantedStatus, resp.Status)
				t.Fail()
			}

		}
	})

	t.Run("Client Do REQMOD with Custom Driver", func(t *testing.T) {

		type testSample struct {
			urlStr              string
			wantedStatusCode    int
			wantedStatus        string
			wantedTimeout       time.Duration
			wantedDialerTimeout time.Duration
			wantedReadTimeout   time.Duration
			wantedWriteTimeout  time.Duration
		}

		sampleTable := []testSample{
			{
				urlStr:              "http://goodifle.com",
				wantedStatusCode:    http.StatusNoContent,
				wantedStatus:        "No Modifications",
				wantedTimeout:       defaultTimeout,
				wantedDialerTimeout: 2 * time.Second,
				wantedReadTimeout:   2 * time.Second,
				wantedWriteTimeout:  2 * time.Second,
			},
			{
				urlStr:              "http://badfile.com",
				wantedStatusCode:    http.StatusOK,
				wantedStatus:        "OK",
				wantedTimeout:       defaultTimeout,
				wantedDialerTimeout: 2 * time.Second,
				wantedReadTimeout:   2 * time.Second,
				wantedWriteTimeout:  2 * time.Second,
			},
		}

		for _, sample := range sampleTable {
			httpReq, err := http.NewRequest(http.MethodGet, sample.urlStr, nil)
			if err != nil {
				t.Log(err.Error())
				t.Fail()
				return
			}

			req, err := NewRequest(MethodREQMOD, fmt.Sprintf("icap://localhost:%d/reqmod", port), httpReq, nil)
			if err != nil {
				t.Log(err.Error())
				t.Fail()
				return
			}

			client := Client{}
			client.SetDriver(&Driver{
				Host:          "127.0.0.1",
				Port:          1344,
				DialerTimeout: 2 * time.Second,
				ReadTimeout:   2 * time.Second,
				WriteTimeout:  2 * time.Second,
			})
			resp, err := client.Do(req)
			if err != nil {
				t.Log(err.Error())
				t.Fail()
				return
			}

			if resp.StatusCode != sample.wantedStatusCode {
				t.Logf("Wanted status code:%d, got:%d", sample.wantedStatusCode, resp.StatusCode)
				t.Fail()
			}

			if resp.Status != sample.wantedStatus {
				t.Logf("Wanted status:%s, got:%s", sample.wantedStatus, resp.Status)
				t.Fail()
			}

			if client.Timeout != sample.wantedTimeout {
				t.Logf("Wanted timeout to be:%v, got:%v", sample.wantedTimeout, client.Timeout)
				t.Fail()
			}
			if client.scktDriver.DialerTimeout != sample.wantedDialerTimeout {
				t.Logf("Wanted DialierTimeout to be:%v , got:%v", sample.wantedDialerTimeout, client.scktDriver.DialerTimeout)
				t.Fail()
			}
			if client.scktDriver.ReadTimeout != sample.wantedReadTimeout {
				t.Logf("Wanted ReadTimeout to be:%v , got:%v", sample.wantedReadTimeout, client.scktDriver.ReadTimeout)
				t.Fail()
			}
			if client.scktDriver.WriteTimeout != sample.wantedWriteTimeout {
				t.Logf("Wanted WriteTimeout to be:%v , got:%v", sample.wantedWriteTimeout, client.scktDriver.WriteTimeout)
				t.Fail()
			}

		}
	})

	if testServerRunning() {
		defer stopTestServer()
	}
}
