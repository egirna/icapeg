package tests

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
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/davecgh/go-spew/spew"
	"github.com/egirna/icap"
)

const (
	transferPreview = "*"
	previewBytes    = "4096"
	remoteICAPPort  = 1345
)

func startRemoteICAPMockServer(stop chan os.Signal, port int) error {
	icap.HandleFunc("/remote-resp", func(w icap.ResponseWriter, req *icap.Request) {
		remoteICAPMockRespmod(w, req)
	})

	icap.HandleFunc("/remote-req", func(w icap.ResponseWriter, req *icap.Request) {
		remoteICAPMockReqmod(w, req)
	})

	signal.Notify(stop, syscall.SIGKILL, syscall.SIGINT, syscall.SIGQUIT)

	go func() {
		if err := icap.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
			log.Println(err.Error())
			stopRemoteICAPMockServer(stop)
		}
	}()

	<-stop
	return nil
}

func remoteICAPMockOptions(w icap.ResponseWriter, req *icap.Request, mode string) {

	h := w.Header()
	h.Set("Methods", mode)
	h.Set("Allow", "204")
	if pb, _ := strconv.Atoi(previewBytes); pb > 0 {
		h.Set("Preview", previewBytes)
	}
	h.Set("Transfer-Preview", transferPreview)
	w.WriteHeader(http.StatusOK, nil, false)
}

func remoteICAPMockRespmod(w icap.ResponseWriter, req *icap.Request) {
	h := w.Header()
	h.Set("ISTag", "remote_icap_server")
	h.Set("Service", "remote_service")

	switch req.Method {
	case utils.ICAPModeOptions:
		remoteICAPMockOptions(w, req, utils.ICAPModeResp)
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

		if filename == "eicar.com" {
			sampleInfo = &dtos.SampleInfo{
				SampleSeverity: "malicious",
				FileName:       filename,
				SampleType:     ct,
				FileSizeStr:    "2.2mb",
				VTIScore:       "10/100",
			}
			status = http.StatusOK
		}

		if filename == "file-example_PDF_1MB.pdf" {
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

func remoteICAPMockReqmod(w icap.ResponseWriter, req *icap.Request) {

	spew.Dump("dksfkdsfjksdjfkdsjfkd")

	h := w.Header()
	h.Set("ISTag", "remote_icap_server")
	h.Set("Service", "remote_service")

	switch req.Method {
	case utils.ICAPModeOptions:
		remoteICAPMockOptions(w, req, utils.ICAPModeReq)
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

		if fileURL == badFileURL {
			sampleInfo = &dtos.SampleInfo{
				SampleSeverity: "malicious",
				FileName:       filename,
				SampleType:     fileExt,
				FileSizeStr:    "2.2mb",
				VTIScore:       "10/100",
			}
			status = http.StatusOK
		}

		if fileURL == goodFileURL {
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

func stopRemoteICAPMockServer(stop chan os.Signal) {
	stop <- syscall.SIGKILL
}
