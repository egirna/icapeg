package service

import (
	"bytes"
	"encoding/json"
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
	ic "github.com/egirna/icap-client"
)

const (
	goodMockFile = "goodfile12345ghjtkl"
	badMockFile  = "badfile12345ghtjkly"
	goodMockURL  = "http://file.com/good.pdf"
	badMockURL   = "http://file.com/bad.exe"
)

type remoteICAPTester struct {
	port                int
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
	wantedOPTIONSStatus int
	wantedREQMODStatus  int
	wantedRESPMODStatus int
}

func TestRemoteICAP(t *testing.T) {

	// preparing the remote ICAP tester
	testers := getRemoteICAPTesters()

	for _, rit := range testers {
		//preparing the remote ICAP service
		svc := &RemoteICAPService{
			url:             fmt.Sprintf("icap://localhost:%d", rit.port),
			respmodEndpoint: "/remote-resp",
			optionsEndpoint: "",
			reqmodEndpoint:  "/remote-req",
			timeout:         5 * time.Second,
			requestHeader:   http.Header{},
		}

		// starting the remote icap server
		go rit.startRemoteICAPMockServer()

		/* Performing OPTIONS */
		var oResp *ic.Response
		if rit.callOPTIONS {

			if svc.optionsEndpoint == "" {
				if rit.mode == utils.ICAPModeResp {
					svc.optionsEndpoint = svc.respmodEndpoint
				}
				if rit.mode == utils.ICAPModeReq {
					svc.optionsEndpoint = svc.reqmodEndpoint
				}
			}
			var err error
			oResp, err = svc.DoOptions()

			if err != nil {
				t.Error("RemoteICAP failed: ", err.Error())
				rit.stopRemoteICAPMockServer()
				continue
			}

			if oResp.StatusCode != rit.wantedOPTIONSStatus {
				t.Errorf("RemoteICAP failed for OPTIONS, wanted status code: %d, got: %d", rit.wantedOPTIONSStatus, oResp.StatusCode)
			}

			for header, value := range rit.wantedOPTIONSHeader {
				if _, exists := oResp.Header[header]; !exists {
					t.Errorf("RemoteICAP failed for OPTIONS, expected header: %s", header)
					continue
				}
				if !reflect.DeepEqual(value, oResp.Header[header]) {
					t.Errorf("RemoteICAP failed for OPTIONS header(%s) value, wanted: %v , got: %v", header, value, oResp.Header[header])
				}
			}

			/* --------------------------------------------------------------------- */
		}

		if oResp != nil {
			svc.SetHeader(oResp.Header) // setting the headers received from OPTIONS for client's next call
		}

		/* Performing REQMOD */
		if rit.mode == utils.ICAPModeReq {
			svc.SetHTTPRequest(rit.httpReq)
			resp, err := svc.DoReqmod()

			if err != nil {
				t.Error("RemoteICAP failed for REQMOD: ", err.Error())
				rit.stopRemoteICAPMockServer()
				continue
			}

			if resp.StatusCode != rit.wantedREQMODStatus {
				t.Errorf("RemoteICAP failed for REQMOD, wanted status code: %d ,got: %d", rit.wantedREQMODStatus, resp.StatusCode)
			}

			for header, value := range rit.wantedREQMODHeader {
				if _, exists := resp.Header[header]; !exists {
					t.Errorf("RemoteICAP failed for REQMOD, expected header: %s", header)
					continue
				}
				if !reflect.DeepEqual(value, resp.Header[header]) {
					t.Errorf("RemoteICAP failed for REQMOD header(%s) value, wanted: %v , got: %v", header, value, resp.Header[header])
				}
			}

		}

		/* ------------------------------------------------------------------------ */

		/* Performing RESPMOD */
		if rit.mode == utils.ICAPModeResp {
			svc.SetHTTPRequest(rit.httpReq)
			svc.SetHTTPResponse(rit.httpResp)
			resp, err := svc.DoRespmod()

			if err != nil {
				t.Error("RemoteICAP failed for RESPMOD: ", err.Error())
				rit.stopRemoteICAPMockServer()
				continue
			}

			if resp.StatusCode != rit.wantedRESPMODStatus {
				t.Errorf("RemoteICAP failed for RESPMOD, wanted status code: %d ,got: %d", rit.wantedRESPMODStatus, resp.StatusCode)
			}

			for header, value := range rit.wantedRESPMODHeader {
				if _, exists := resp.Header[header]; !exists {
					t.Errorf("RemoteICAP failed for RESPMOD, expected header: %s", header)
					continue
				}
				if !reflect.DeepEqual(value, resp.Header[header]) {
					t.Errorf("RemoteICAP failed for RESPMOD header(%s) value, wanted: %v , got: %v", header, value, resp.Header[header])
				}
			}
		}

		/* ------------------------------------------------------------------------ */

		rit.stopRemoteICAPMockServer()

	}

}

func (rit *remoteICAPTester) startRemoteICAPMockServer() {
	icap.HandleFunc("/remote-resp", func(w icap.ResponseWriter, req *icap.Request) {
		rit.remoteICAPMockRespmod(w, req)
	})

	icap.HandleFunc("/remote-req", func(w icap.ResponseWriter, req *icap.Request) {
		rit.remoteICAPMockReqmod(w, req)
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

func (rit *remoteICAPTester) remoteICAPMockOptions(w icap.ResponseWriter, req *icap.Request) {

	h := w.Header()
	h.Set("Methods", rit.mode)
	h.Set("Allow", "204")
	if pb, _ := strconv.Atoi(rit.previewBytes); pb > 0 {
		h.Set("Preview", rit.previewBytes)
	}
	h.Set("Transfer-Preview", rit.transferPreview)
	w.WriteHeader(http.StatusOK, nil, false)
}

func (rit *remoteICAPTester) remoteICAPMockRespmod(w icap.ResponseWriter, req *icap.Request) {
	h := w.Header()
	h.Set("ISTag", rit.isTag)
	h.Set("Service", rit.service)

	switch req.Method {
	case utils.ICAPModeOptions:
		rit.remoteICAPMockOptions(w, req)
	case utils.ICAPModeResp:

		defer req.Response.Body.Close()

		if val, exist := req.Header["Allow"]; !exist || (len(val) > 0 && val[0] != utils.NoModificationStatusCodeStr) {
			w.WriteHeader(http.StatusNoContent, nil, false)
			return
		}

		ct := utils.GetMimeExtension(req.Preview)

		buf := &bytes.Buffer{}

		if _, err := io.Copy(buf, req.Response.Body); err != nil {
			w.WriteHeader(http.StatusNoContent, nil, false)
			return
		}

		var sampleInfo *dtos.SampleInfo
		var status int

		filename := utils.GetFileName(req.Request)

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
			td := &dtos.TemplateData{
				FileName:     sampleInfo.FileName,
				FileType:     sampleInfo.SampleType,
				FileSizeStr:  sampleInfo.FileSizeStr,
				RequestedURL: utils.BreakHTTPURL(req.Request.RequestURI),
				Severity:     sampleInfo.SampleSeverity,
				Score:        sampleInfo.VTIScore,
				ResultsBy:    "some_vendor",
			}
			htmlStr := fmt.Sprintf("<html><body>%v</body></htm>", *td)
			htmlBuf := bytes.NewBuffer([]byte(htmlStr))
			newResp := &http.Response{
				StatusCode: http.StatusForbidden,
				Status:     http.StatusText(http.StatusForbidden),
				Header: http.Header{
					"Content-Type":   []string{"text/html"},
					"Content-Length": []string{strconv.Itoa(htmlBuf.Len())},
				},
			}
			w.WriteHeader(http.StatusOK, newResp, true)
			w.Write(htmlBuf.Bytes())
			return
		}

		w.WriteHeader(status, nil, false)

	}

}

func (rit *remoteICAPTester) remoteICAPMockReqmod(w icap.ResponseWriter, req *icap.Request) {
	h := w.Header()
	h.Set("ISTag", rit.isTag)
	h.Set("Service", rit.service)

	switch req.Method {
	case utils.ICAPModeOptions:
		rit.remoteICAPMockOptions(w, req)
	case utils.ICAPModeReq:
		if val, exist := req.Header["Allow"]; !exist || (len(val) > 0 && val[0] != utils.NoModificationStatusCodeStr) {
			w.WriteHeader(http.StatusNoContent, nil, false)
			return
		}

		fileURL := req.Request.RequestURI

		filename := utils.GetFileName(req.Request)
		fileExt := utils.GetFileExtension(req.Request)

		var sampleInfo *dtos.SampleInfo
		var status int

		if fileURL == badMockURL {
			sampleInfo = &dtos.SampleInfo{
				SampleSeverity: "malicious",
				FileName:       filename,
				SampleType:     fileExt,
				FileSizeStr:    "2.2mb",
				VTIScore:       "10/100",
			}
			status = http.StatusOK
		}

		if fileURL == goodMockURL {
			status = http.StatusNoContent
		}

		if status == http.StatusOK && sampleInfo != nil {
			data := &dtos.TemplateData{
				FileName:     sampleInfo.FileName,
				FileType:     sampleInfo.SampleType,
				FileSizeStr:  sampleInfo.FileSizeStr,
				RequestedURL: utils.BreakHTTPURL(req.Request.RequestURI),
				Severity:     sampleInfo.SampleSeverity,
				Score:        sampleInfo.VTIScore,
			}

			dataByte, err := json.Marshal(data)

			if err != nil {
				w.WriteHeader(utils.IfPropagateError(http.StatusInternalServerError, http.StatusNoContent), nil, false)
				return
			}

			req.Request.Body = ioutil.NopCloser(bytes.NewReader(dataByte))

			icap.ServeLocally(w, req)

			return
		}

		w.WriteHeader(status, nil, false)

	}

}

func getRemoteICAPTesters() []remoteICAPTester {
	return []remoteICAPTester{
		{
			port:            1345,
			mode:            utils.ICAPModeResp,
			previewBytes:    "0",
			transferPreview: "*",
			isTag:           "remote_icap_server",
			service:         "test",
			stop:            make(chan os.Signal, 1),
			callOPTIONS:     false,
			httpResp:        makeHTTPResponse(goodMockURL),
			httpReq:         makeHTTPRequest(goodMockURL),
			wantedRESPMODHeader: http.Header{
				"Istag":   []string{"remote_icap_server"},
				"Service": []string{"test"},
			},
			wantedOPTIONSStatus: http.StatusOK,
			wantedRESPMODStatus: http.StatusNoContent,
		},
		{
			port:            1346,
			mode:            utils.ICAPModeResp,
			previewBytes:    "4096",
			transferPreview: "*",
			isTag:           "remote_icap_server",
			service:         "test",
			stop:            make(chan os.Signal, 1),
			callOPTIONS:     true,
			httpResp:        makeHTTPResponse(badMockURL),
			httpReq:         makeHTTPRequest(badMockURL),
			wantedOPTIONSHeader: http.Header{
				"Methods":          []string{utils.ICAPModeResp},
				"Allow":            []string{"204"},
				"Transfer-Preview": []string{"*"},
			},
			wantedRESPMODHeader: http.Header{
				"Istag":   []string{"remote_icap_server"},
				"Service": []string{"test"},
			},
			wantedOPTIONSStatus: http.StatusOK,
			wantedRESPMODStatus: http.StatusOK,
		},
		{
			port:            1347,
			mode:            utils.ICAPModeReq,
			previewBytes:    "0",
			transferPreview: "*",
			isTag:           "remote_icap_server",
			service:         "test",
			stop:            make(chan os.Signal, 1),
			callOPTIONS:     true,
			httpReq:         makeHTTPRequest(goodMockURL),
			wantedOPTIONSHeader: http.Header{
				"Methods":          []string{utils.ICAPModeReq},
				"Allow":            []string{"204"},
				"Transfer-Preview": []string{"*"},
			},
			wantedREQMODHeader: http.Header{
				"Istag":   []string{"remote_icap_server"},
				"Service": []string{"test"},
			},
			wantedOPTIONSStatus: http.StatusOK,
			wantedREQMODStatus:  http.StatusNoContent,
		},
		{
			port:            1348,
			mode:            utils.ICAPModeReq,
			previewBytes:    "4096",
			transferPreview: "*",
			isTag:           "remote_icap_server",
			service:         "test",
			stop:            make(chan os.Signal, 1),
			callOPTIONS:     true,
			httpReq:         makeHTTPRequest(badMockURL),
			wantedOPTIONSHeader: http.Header{
				"Methods":          []string{utils.ICAPModeReq},
				"Allow":            []string{"204"},
				"Transfer-Preview": []string{"*"},
			},
			wantedREQMODHeader: http.Header{
				"Istag":   []string{"remote_icap_server"},
				"Service": []string{"test"},
			},
			wantedOPTIONSStatus: http.StatusOK,
			wantedREQMODStatus:  http.StatusOK,
		},
	}
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

	resp.Header = http.Header{}

	return resp
}

func makeHTTPRequest(urlStr string) *http.Request {
	u, _ := url.Parse(urlStr)
	req := &http.Request{
		URL:    u,
		Host:   "file.com",
		Header: http.Header{},
	}

	return req
}
