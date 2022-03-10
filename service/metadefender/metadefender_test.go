package metadefender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"icapeg/dtos"
	"icapeg/utils"
)

const (
	metadefenderScanFile   = "metadefender_submit_file_endpoint"
	metadefenderReportFile = "metadefender_report_file_endpoint"
)

var (
	metadefenderEndpointMap = map[string]string{
		"/file": metadefenderScanFile,
	}
)

func getMetaDefenderMockServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-type", "application/json")

		urlStr := r.URL.EscapedPath()

		if _, exists := metadefenderEndpointMap[urlStr]; !exists {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"message":"Route not found"}`))
			return
		}

		apikey := r.Header.Get("Apikey")

		if apikey == "" {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{
											    "error": {
											        "code": 401001,
											        "messages": [
											            "Authentication strategy is invalid"
											        ]
											    }
											}`))
			return
		}

		var jsonRep []byte
		var err error
		endpoint := metadefenderEndpointMap[urlStr]

		if endpoint == metadefenderScanFile {

			if r.Method != http.MethodPost {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte(`{
												    "error": {
												        "code": 404000,
												        "messages": [
												            "Endpoint not found"
												        ]
												    }
												}`))
				return
			}

			data, _ := ioutil.ReadAll(r.Body)
			r.Body.Close()

			if len(data) < 1 {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{
											    "error": {
											        "code": 400145,
											        "messages": [
											            "Request body is empty. Please send a binary file."
											        ]
											    }
											}`))
				return
			}

			mresp := dtos.MetaDefenderScanFileResponse{
				DataID: "bzIwMDUwNW9Nb2lxOWktcHZkUXhpVVcyS05W",
			}

			jsonRep, err = json.Marshal(mresp)

			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"message":"Somethign went wrong"}`))
				return
			}

		}

		if endpoint == metadefenderReportFile {
			if r.Method != http.MethodGet {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte(`{
												    "error": {
												        "code": 404000,
												        "messages": [
												            "Endpoint not found"
												        ]
												    }
												}`))
				return
			}

			words := strings.Split(r.URL.EscapedPath(), "/")

			dataID := words[len(words)-1]

			mresp := dtos.MetaDefenderReportResponse{
				ScanResults: dtos.MetaDefenderScanResults{
					TotalAvs:           38,
					TotalDetectedAvs:   1,
					ProgressPercentage: 100,
				},
			}

			if dataID == "4321abcd" {
				mresp.ScanResults.TotalDetectedAvs = 33
			}

			jsonRep, err = json.Marshal(mresp)

			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"message":"Somethign went wrong"}`))
				return
			}

		}

		w.WriteHeader(http.StatusOK)
		w.Write(jsonRep)
	}))
}

func TestMetaDefenderSubmitFile(t *testing.T) {
	testServer := getMetaDefenderMockServer()

	defer testServer.Close()

	type testSample struct {
		md       *MetaDefender
		fileBfr  *bytes.Buffer
		filename string
		sresp    *dtos.SubmitResponse
	}

	sampleTable := []testSample{
		{
			md: &MetaDefender{
				BaseURL:      testServer.URL,
				Timeout:      10 * time.Second,
				APIKey:       "someapikey",
				ScanEndpoint: "/file",
			},
			fileBfr:  bytes.NewBuffer([]byte{0x52, 0x61, 0x72, 0x21, 0x1A, 0x07, 0x00}),
			filename: "somefile.exe",
			sresp: &dtos.SubmitResponse{
				SubmissionExists:   true,
				SubmissionID:       "bzIwMDUwNW9Nb2lxOWktcHZkUXhpVVcyS05W",
				SubmissionSampleID: "bzIwMDUwNW9Nb2lxOWktcHZkUXhpVVcyS05W",
			},
		},
		{
			md: &MetaDefender{
				BaseURL:      testServer.URL,
				Timeout:      10 * time.Second,
				APIKey:       "",
				ScanEndpoint: "/file",
			},
			fileBfr:  &bytes.Buffer{},
			filename: "somefile.exe",
			sresp:    nil,
		},
	}

	for _, sample := range sampleTable {
		resp, err := sample.md.SubmitFile(sample.fileBfr, sample.filename)

		if resp != nil && err != nil {
			t.Error("Unexpected response from metadefender submit file: ", err.Error())
			return
		}

		if (sample.sresp == nil && resp != nil) || (sample.sresp != nil && err != nil) {
			t.Error("Unexpected response from metadefender submit file: ", err.Error())
			return
		}

		if resp == nil {
			continue
		}

		if sample.sresp.SubmissionExists != resp.SubmissionExists {
			t.Errorf("Unexpected result for metadefender submit file SubmissionExists , wanted: %v got: %v",
				sample.sresp.SubmissionExists, resp.SubmissionExists)
		}

		if sample.sresp.SubmissionID != resp.SubmissionID {
			t.Errorf("Unexpected result for metadefender submit file SubmissionID , wanted: %s got: %s",
				sample.sresp.SubmissionID, resp.SubmissionID)
		}

		if sample.sresp.SubmissionSampleID != resp.SubmissionSampleID {
			t.Errorf("Unexpected result for metadefender submit file SubmissionSampleID , wanted: %s got: %s",
				sample.sresp.SubmissionSampleID, resp.SubmissionSampleID)
		}

	}

}

func TestMetaDefenderSampleInfo(t *testing.T) {
	testServer := getMetaDefenderMockServer()

	defer testServer.Close()

	type testSample struct {
		md       *MetaDefender
		sampleID string
		fm       dtos.FileMetaInfo
		siResp   *dtos.SampleInfo
	}

	sampleTable := []testSample{
		{
			md: &MetaDefender{
				BaseURL:        testServer.URL,
				Timeout:        10 * time.Second,
				APIKey:         "someapikey",
				ReportEndpoint: "/file",
				FailThreshold:  3,
			},
			sampleID: "1234abcd",
			fm: dtos.FileMetaInfo{
				FileName: "somefile.exe",
				FileSize: 2000,
				FileType: "exe",
			},
			siResp: &dtos.SampleInfo{
				FileName:           "somefile.exe",
				SampleType:         "exe",
				SampleSeverity:     "ok",
				VTIScore:           "1/38",
				FileSizeStr:        fmt.Sprintf("%.2fmb", utils.ByteToMegaBytes(2000)),
				SubmissionFinished: true,
			},
		},
		{
			md: &MetaDefender{
				BaseURL:        testServer.URL,
				Timeout:        10 * time.Second,
				APIKey:         "someapikey",
				ReportEndpoint: "/file",
				FailThreshold:  3,
			},
			sampleID: "4321abcd",
			fm: dtos.FileMetaInfo{
				FileName: "somefile.exe",
				FileSize: 2000,
				FileType: "exe",
			},
			siResp: &dtos.SampleInfo{
				FileName:           "somefile.exe",
				SampleType:         "exe",
				SampleSeverity:     "malicious",
				VTIScore:           "33/38",
				FileSizeStr:        fmt.Sprintf("%.2fmb", utils.ByteToMegaBytes(2000)),
				SubmissionFinished: true,
			},
		},
	}

	for _, sample := range sampleTable {
		metadefenderEndpointMap["/file/"+sample.sampleID] = metadefenderReportFile
		resp, err := sample.md.GetSampleFileInfo(sample.sampleID, sample.fm)

		if err != nil {
			t.Error("Unexpected response from metadefender sample info: ", err.Error())
			return
		}

		if sample.siResp.FileName != resp.FileName {
			t.Errorf("Unexpected result for metadefender sample file info FileName, wanted: %s got: %s",
				sample.siResp.FileName, resp.FileName)
		}

		if sample.siResp.SampleType != resp.SampleType {
			t.Errorf("Unexpected result for metadefender sample file info SampleType, wanted: %s got: %s",
				sample.siResp.SampleType, resp.SampleType)
		}

		if sample.siResp.SampleSeverity != resp.SampleSeverity {
			t.Errorf("Unexpected result for metadefender sample file info SampleSeverity, wanted: %s got: %s",
				sample.siResp.SampleSeverity, resp.SampleSeverity)
		}

		if sample.siResp.VTIScore != resp.VTIScore {
			t.Errorf("Unexpected result for metadefender sample file info VTIScore, wanted: %s got: %s",
				sample.siResp.VTIScore, resp.VTIScore)
		}

		if sample.siResp.FileSizeStr != resp.FileSizeStr {
			t.Errorf("Unexpected result for metadefender sample file info FileSizeStr, wanted: %s got: %s",
				sample.siResp.FileSizeStr, resp.FileSizeStr)
		}

		if sample.siResp.SubmissionFinished != resp.SubmissionFinished {
			t.Errorf("Unexpected result for metadefender sample file info SubmissionFinished, wanted: %v got: %v",
				sample.siResp.SubmissionFinished, resp.SubmissionFinished)
		}
	}

}
