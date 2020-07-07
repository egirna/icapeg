package service

import (
	"bytes"
	"fmt"
	"icapeg/dtos"
	"icapeg/utils"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"reflect"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/egirna/icap"
)

const (
	goodMockFile = "goodfile12345ghjtkl"
	badMockFile  = "badfile12345ghtjkly"
	goodMockURL  = "http://file.com/good.pdf"
	badMockURL   = "http://file.com/bad.exe"
)

type remoteICAPTester struct {
	port                int
	w                   icap.ResponseWriter
	req                 *icap.Request
	mode                string
	previewBytes        string
	transferPreview     string
	isTag               string
	service             string
	stop                chan os.Signal
	callOPTIONS         bool
	httpResp            *http.Response
	httpReq             *http.Request
	wantedOPTIONSHeader http.Header
	wantedRESPMODHeader http.Header
	wantedREQMODHeader  http.Header
}

func TestRemoteICAP(t *testing.T) {

	// preparing the remote ICAP tester
	testers := []remoteICAPTester{
		{
			port:            1345,
			mode:            utils.ICAPModeResp,
			previewBytes:    "0",
			transferPreview: "*",
			isTag:           "remote_icap_server",
			service:         "test",
			stop:            make(chan os.Signal, 1),
			callOPTIONS:     true,
			httpResp:        makeHTTPResponse(goodMockURL),
			httpReq:         makeHTTPRequest(goodMockURL),
			wantedOPTIONSHeader: http.Header{
				"Methods":          []string{utils.ICAPModeResp},
				"Allow":            []string{"204"},
				"Transfer-Preview": []string{"*"},
			},
		},
	}

	for _, rit := range testers {
		//preparing the remote ICAP service
		svc := &RemoteICAPService{
			url:             fmt.Sprintf("icap://localhost:%d", rit.port),
			respmodEndpoint: "/remote-resp",
			optionsEndpoint: "",
			reqmodEndpoint:  "remote-req",
			timeout:         5 * time.Second,
			requestHeader:   http.Header{},
		}

		// starting the remote icap server
		go rit.startRemoteICAPMockServer()

		if rit.callOPTIONS {
			/* Performing OPTIONS */
			if svc.optionsEndpoint == "" {
				if rit.mode == utils.ICAPModeResp {
					svc.optionsEndpoint = svc.respmodEndpoint
				}
				if rit.mode == utils.ICAPModeReq {
					svc.optionsEndpoint = svc.reqmodEndpoint
				}
			}
			resp, err := svc.DoOptions()

			if err != nil {
				t.Error("RemoteICAP failed: ", err.Error())
			}

			if resp.StatusCode != http.StatusOK {
				t.Errorf("RemoteICAP failed for OPTIONS, wanted status code: %d, got: %d", http.StatusOK, resp.StatusCode)
			}

			for header, value := range rit.wantedOPTIONSHeader {
				if _, exists := resp.Header[header]; !exists {
					t.Errorf("RemoteICAP failed for OPTIONS, expected header: %s", header)
					continue
				}
				if !reflect.DeepEqual(value, resp.Header[header]) {
					t.Errorf("RemoteICAP failed for OPTIONS header(%s) value, wanted: %v , got: %v", header, value, resp.Header[header])
				}
			}

			/* --------------------------------------------------------------------- */
		}

		if rit.mode == utils.ICAPModeReq {
			// TODO: Perform reqmod tests here
		}
		if rit.mode == utils.ICAPModeResp {
			// TODO: Perform respmod tests here
		}

		rit.stopRemoteICAPMockServer()

	}

}

func (rit *remoteICAPTester) startRemoteICAPMockServer() {
	icap.HandleFunc("/remote-resp", func(w icap.ResponseWriter, req *icap.Request) {
		rit.w = w
		rit.req = req
		rit.remoteICAPMockRespmod()
	})

	signal.Notify(rit.stop, syscall.SIGKILL, syscall.SIGINT, syscall.SIGQUIT)

	go func() {
		if err := icap.ListenAndServe(fmt.Sprintf(":%d", rit.port), nil); err != nil {
			log.Println(err.Error())
			rit.stopRemoteICAPMockServer()
		}
	}()

	<-rit.stop

}

func (rit *remoteICAPTester) stopRemoteICAPMockServer() {
	rit.stop <- syscall.SIGKILL
}

func (rit *remoteICAPTester) remoteICAPMockOptions() {

	h := rit.w.Header()
	h.Set("Methods", rit.mode)
	h.Set("Allow", "204")
	if pb, _ := strconv.Atoi(rit.previewBytes); pb > 0 {
		h.Set("Preview", rit.previewBytes)
	}
	h.Set("Transfer-Preview", rit.transferPreview)
	rit.w.WriteHeader(http.StatusOK, nil, false)
}

func (rit *remoteICAPTester) remoteICAPMockRespmod() {
	h := rit.w.Header()
	h.Set("ISTag", rit.isTag)
	h.Set("Service", rit.service)

	switch rit.req.Method {
	case utils.ICAPModeOptions:
		rit.remoteICAPMockOptions()
	case utils.ICAPModeResp:

		defer rit.req.Response.Body.Close()

		if val, exist := rit.req.Header["Allow"]; !exist || (len(val) > 0 && val[0] != utils.NoModificationStatusCodeStr) {
			rit.w.WriteHeader(http.StatusNoContent, nil, false)
			return
		}

		ct := utils.GetMimeExtension(rit.req.Preview)

		buf := &bytes.Buffer{}

		if _, err := io.Copy(buf, rit.req.Response.Body); err != nil {
			rit.w.WriteHeader(http.StatusNoContent, nil, false)
			return
		}

		var sampleInfo *dtos.SampleInfo
		var status int

		filename := utils.GetFileName(rit.req.Request)

		if buf.String() == badMockFile {
			sampleInfo = &dtos.SampleInfo{
				SampleSeverity: "malicious",
				FileName:       filename,
				SampleType:     ct,
				FileSizeStr:    "2.2mb",
				VTIScore:       "10/100",
			}
			status = http.StatusOK
		}

		if buf.String() == goodMockFile {
			status = http.StatusNoContent
		}

		if status == http.StatusOK && sampleInfo != nil {
			htmlBuf, newResp := utils.GetTemplateBufferAndResponse(utils.BadFileTemplate, &dtos.TemplateData{
				FileName:     sampleInfo.FileName,
				FileType:     sampleInfo.SampleType,
				FileSizeStr:  sampleInfo.FileSizeStr,
				RequestedURL: utils.BreakHTTPURL(rit.req.Request.RequestURI),
				Severity:     sampleInfo.SampleSeverity,
				Score:        sampleInfo.VTIScore,
				ResultsBy:    "some_vendor",
			})
			rit.w.WriteHeader(http.StatusOK, newResp, true)
			rit.w.Write(htmlBuf.Bytes())
			return
		}

		rit.w.WriteHeader(status, nil, false)

	}

}

func remoteICAPMockReqmod(rit remoteICAPTester) {

}

func makeHTTPResponse(urlStr string) *http.Response {
	var resp *http.Response
	if urlStr == goodMockURL {
		resp = &http.Response{
			Body:          ioutil.NopCloser(strings.NewReader(goodMockFile)),
			ContentLength: int64(len(goodMockFile)),
		}
	}

	if urlStr == badMockURL {
		resp = &http.Response{
			Body:          ioutil.NopCloser(strings.NewReader(badMockFile)),
			ContentLength: int64(len(badMockFile)),
		}
	}

	return resp
}

func makeHTTPRequest(urlStr string) *http.Request {
	u, _ := url.Parse(urlStr)
	req := &http.Request{
		URL:  u,
		Host: "file.com",
	}

	return req
}
