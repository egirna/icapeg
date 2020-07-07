package service

import (
	"bytes"
	"fmt"
	"icapeg/dtos"
	"icapeg/utils"
	"io"
	"log"
	"net/http"
	"strconv"
	"testing"

	"github.com/egirna/icap"
)

const (
	goodMockFile = "goodfile12345ghjtkl"
	badMockFile  = "badfile12345ghtjkly"
	goodMockURL  = "http://file.com/good.pdf"
	badMockURL   = "http://file.com/bad.exe"
)

type remoteICAPTester struct {
	port            int
	w               icap.ResponseWriter
	req             *icap.Request
	mode            string
	previewBytes    string
	transferPreview string
	isTag           string
	service         string
}

func TestRemoteICAP(t *testing.T) {

}

func (rit *remoteICAPTester) startRemoteICAPMockServer() {
	icap.HandleFunc("/remote-resp", func(w icap.ResponseWriter, req *icap.Request) {
		rit.w = w
		rit.req = req
		rit.remoteICAPMockRespmod()
	})

	if err := icap.ListenAndServe(fmt.Sprintf(":%d", rit.port), nil); err != nil {
		log.Fatal(err.Error())
	}

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
