package vmray

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"icapeg/dtos"
	"icapeg/utils"
)

const (
	vmraySubmitFile       = "vmray_submit_file_endpoint"
	vmraySampleFileInfo   = "vmray_sample_file_info_endpoint"
	vmraySubmissionStatus = "vmray_submission_status_endpoint"
)

var (
	vmrayEndpointMap = map[string]string{
		"/sample/submit": vmraySubmitFile,
	}
)

func getVmrayMockServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-type", "application/json")

		urlStr := r.URL.EscapedPath()

		if _, exists := vmrayEndpointMap[urlStr]; !exists {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"message":"Route not found"}`))
			return
		}

		apikey := r.Header.Get("Authorization")

		if apikey == "" || !utils.InStringSlice("api_key", strings.Split(apikey, " ")) ||
			len(strings.SplitAfter(apikey, " ")) <= 1 {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error_msg": "Authentication required","result": "error"}`))
			return
		}

		var jsonRep []byte
		var err error
		endpoint := vmrayEndpointMap[urlStr]

		if endpoint == vmraySubmitFile {

			if r.Method != http.MethodPost {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			if err := r.ParseMultipartForm(1024); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"message":"Somethign went wrong"}`))
				return
			}

			if r.MultipartForm.File["sample_file"] != nil {
				if r.MultipartForm.File["sample_file"][0].Filename == "" || r.MultipartForm.File["sample_file"][0].Size < 1 {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(`{
													"error_msg": "Invalid file parameter \"sample_file\". It seems the way you pass the data is incorrect",
  												"result": "error"
						}`))
					return
				}

			} else if sampleURL, exists := r.MultipartForm.Value["sample_url"]; !exists || (len(sampleURL) > 0 && sampleURL[0] == "") {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{
													"error_msg": "Missing parameter: Either \"sample_file\" or \"sample_url\" must be specified.",
  												"result": "error"
												}`))
				return
			}

			vresp := dtos.VmraySubmitResponse{
				Data: dtos.VmraySubmitData{
					Submissions: []dtos.VmraySubmissions{
						{
							SubmissionID:       5624736,
							SubmissionSampleID: 4841630,
						},
					},
				},
			}

			jsonRep, err = json.Marshal(vresp)

			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"message":"Somethign went wrong"}`))
				return
			}

		}

		if endpoint == vmraySubmissionStatus {

			if r.Method != http.MethodGet {
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte(`{
  					"error_msg": "Forbidden internal API function \"submission_update\". You must enable the internal API to use this parameter",
  					"result": "error"
							}`))
				return
			}

			vresp := dtos.VmraySubmissionStatusResponse{
				Data: dtos.VmraySubmissionData{
					SubmissionFinished: true,
				},
			}

			jsonRep, err = json.Marshal(vresp)

			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"message":"Somethign went wrong"}`))
				return
			}

		}

		if endpoint == vmraySampleFileInfo {

			if r.Method != http.MethodGet {
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte(`{
  					"error_msg": "Forbidden internal API function \"sample_update\". You must enable the internal API to use this parameter",
  					"result": "error"
								}`))
				return
			}

			words := strings.Split(r.URL.EscapedPath(), "/")

			sampleID := words[len(words)-1]

			vresp := dtos.GetVmraySampleResponse{
				Data: dtos.VmraySampleData{
					SampleFilename: "somefile.exe",
					SampleType:     "exe",
					SampleSeverity: "not_suspicious",
					SampleVtiScore: 0,
					SampleFilesize: 3028,
				},
			}

			if sampleID == "4321" {
				vresp.Data.SampleSeverity = "malicious"
				vresp.Data.SampleVtiScore = 20
			}

			jsonRep, err = json.Marshal(vresp)

			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"message":"Somethign went wrong"}`))
				return
			}
		}

		w.WriteHeader(http.StatusOK)
		w.Write(jsonRep)
		return
	}))

}

func TestVmraySubmitFile(t *testing.T) {
	testServer := getVmrayMockServer()

	defer testServer.Close()

	type testSample struct {
		vr         *Vmray
		fileBuffer *bytes.Buffer
		filename   string
		sresp      *dtos.SubmitResponse
	}

	sampleTable := []testSample{
		{
			vr: &Vmray{
				BaseURL:        testServer.URL,
				Timeout:        10 * time.Second,
				APIKey:         "",
				SubmitEndpoint: "/sample/submit",
			},
			fileBuffer: &bytes.Buffer{},
			filename:   "somefile.exe",
			sresp:      nil,
		},
		{
			vr: &Vmray{
				BaseURL:        testServer.URL,
				Timeout:        10 * time.Second,
				APIKey:         "someapikey",
				SubmitEndpoint: "/sample/submit",
			},
			fileBuffer: bytes.NewBuffer([]byte{0x52, 0x61, 0x72, 0x21, 0x1A, 0x07, 0x00}),
			filename:   "somefile.exe",
			sresp: &dtos.SubmitResponse{
				SubmissionExists:   true,
				SubmissionID:       "5624736",
				SubmissionSampleID: "4841630",
			},
		},
	}

	for _, sample := range sampleTable {
		resp, err := sample.vr.SubmitFile(sample.fileBuffer, sample.filename)

		if resp != nil && err != nil {
			t.Error("Failed to submit file for vmray: ", err.Error())
			return
		}

		if (sample.sresp == nil && resp != nil) || (sample.sresp != nil && err != nil) {
			t.Error("Failed to submit file for vmray: ", err.Error())
			return
		}

		if resp == nil {
			continue
		}

		if sample.sresp.SubmissionExists != resp.SubmissionExists {
			t.Errorf("Unexpected result for vmray submit file SubmissionExists , wanted: %v got: %v",
				sample.sresp.SubmissionExists, resp.SubmissionExists)
		}

		if sample.sresp.SubmissionID != resp.SubmissionID {
			t.Errorf("Unexpected result for vmray submit file SubmissionID , wanted: %s got: %s",
				sample.sresp.SubmissionID, resp.SubmissionID)
		}

		if sample.sresp.SubmissionSampleID != resp.SubmissionSampleID {
			t.Errorf("Unexpected result for vmray submit file SubmissionSampleID , wanted: %s got: %s",
				sample.sresp.SubmissionSampleID, resp.SubmissionSampleID)
		}

	}

}

func TestVmraySubmitURL(t *testing.T) {
	testServer := getVmrayMockServer()

	defer testServer.Close()

	type testSample struct {
		vr       *Vmray
		fileURL  string
		filename string
		sresp    *dtos.SubmitResponse
	}

	sampleTable := []testSample{
		{
			vr: &Vmray{
				BaseURL:        testServer.URL,
				Timeout:        10 * time.Second,
				APIKey:         "",
				SubmitEndpoint: "/sample/submit",
			},
			fileURL:  "http://somehost.com/somefile.exe",
			filename: "somefile.exe",
			sresp:    nil,
		},
		{
			vr: &Vmray{
				BaseURL:        testServer.URL,
				Timeout:        10 * time.Second,
				APIKey:         "someapikey",
				SubmitEndpoint: "/sample/submit",
			},
			fileURL:  "http://somehost.com/somefile.exe?someparam=somevalue",
			filename: "somefile.exe",
			sresp: &dtos.SubmitResponse{
				SubmissionExists:   true,
				SubmissionID:       "5624736",
				SubmissionSampleID: "4841630",
			},
		},
	}

	for _, sample := range sampleTable {
		resp, err := sample.vr.SubmitURL(sample.fileURL, sample.filename)

		if resp != nil && err != nil {
			t.Error("Failed to submit url for vmray: ", err.Error())
			return
		}

		if sample.sresp == nil && resp != nil {
			t.Errorf("Unexpected result for vmray submit url  , wanted: %v got: %v",
				sample.sresp, resp)
			return
		}

		if resp == nil {
			continue
		}

		if sample.sresp.SubmissionExists != resp.SubmissionExists {
			t.Errorf("Unexpected result for vmray submit url SubmissionExists , wanted: %v got: %v",
				sample.sresp.SubmissionExists, resp.SubmissionExists)
		}

		if sample.sresp.SubmissionID != resp.SubmissionID {
			t.Errorf("Unexpected result for vmray submit url SubmissionID , wanted: %s got: %s",
				sample.sresp.SubmissionID, resp.SubmissionID)
		}

		if sample.sresp.SubmissionSampleID != resp.SubmissionSampleID {
			t.Errorf("Unexpected result for vmray submit url SubmissionSampleID , wanted: %s got: %s",
				sample.sresp.SubmissionSampleID, resp.SubmissionSampleID)
		}

	}
}

func TestVmraySampleFileInfo(t *testing.T) {

	testServer := getVmrayMockServer()

	defer testServer.Close()

	type testSample struct {
		vr       *Vmray
		sampleID string
		sfresp   *dtos.SampleInfo
	}

	sampleTable := []testSample{
		{
			vr: &Vmray{
				BaseURL:           testServer.URL,
				Timeout:           10 * time.Second,
				APIKey:            "someapikey",
				GetSampleEndpoint: "/sample",
			},
			sampleID: "1234",
			sfresp: &dtos.SampleInfo{
				FileName:       "somefile.exe",
				SampleType:     "exe",
				SampleSeverity: "not_suspicious",
				VTIScore:       fmt.Sprintf("%v/100", 0),
				FileSizeStr:    fmt.Sprintf("%.2fmb", utils.ByteToMegaBytes(3028)),
			},
		},
		{
			vr: &Vmray{
				BaseURL:           testServer.URL,
				Timeout:           10 * time.Second,
				APIKey:            "someapikey",
				GetSampleEndpoint: "/sample",
			},
			sampleID: "4321",
			sfresp: &dtos.SampleInfo{
				FileName:       "somefile.exe",
				SampleType:     "exe",
				SampleSeverity: "malicious",
				VTIScore:       fmt.Sprintf("%v/100", 20),
				FileSizeStr:    fmt.Sprintf("%.2fmb", utils.ByteToMegaBytes(3028)),
			},
		},
	}

	for _, sample := range sampleTable {
		vmrayEndpointMap["/sample/"+sample.sampleID] = vmraySampleFileInfo
		resp, err := sample.vr.GetSampleFileInfo(sample.sampleID)

		if err != nil {
			t.Error("Failed to get sample file info for vmray: ", err.Error())
			return
		}

		if sample.sfresp.FileName != resp.FileName {
			t.Errorf("Unexpected result for vmray sample info Filename , wanted: %s got: %s",
				sample.sfresp.FileName, resp.FileName)
		}

		if sample.sfresp.SampleType != resp.SampleType {
			t.Errorf("Unexpected result for vmray sample info SampleType , wanted: %s got: %s",
				sample.sfresp.SampleType, resp.SampleType)
		}

		if sample.sfresp.SampleSeverity != resp.SampleSeverity {
			t.Errorf("Unexpected result for vmray sample info SampleSeverity , wanted: %s got: %s",
				sample.sfresp.SampleSeverity, resp.SampleSeverity)
		}

		if sample.sfresp.FileSizeStr != resp.FileSizeStr {
			t.Errorf("Unexpected result for vmray sample info FileSizeStr , wanted: %s got: %s",
				sample.sfresp.FileSizeStr, resp.FileSizeStr)
		}

		if sample.sfresp.VTIScore != resp.VTIScore {
			t.Errorf("Unexpected result for vmray sample info VTIScore , wanted: %s got: %s",
				sample.sfresp.VTIScore, resp.VTIScore)
		}
	}

}

func TestVmraySubmissionStatus(t *testing.T) {
	testServer := getVmrayMockServer()

	defer testServer.Close()

	type testSample struct {
		vr           *Vmray
		submissionID string
		ssresp       *dtos.SubmissionStatusResponse
	}

	sampleTable := []testSample{
		{
			vr: &Vmray{
				BaseURL:                  testServer.URL,
				Timeout:                  10 * time.Second,
				APIKey:                   "someapikey",
				SubmissionStatusEndpoint: "/submission",
			},
			submissionID: "1234",
			ssresp: &dtos.SubmissionStatusResponse{
				SubmissionFinished: true,
			},
		},
	}

	for _, sample := range sampleTable {
		vmrayEndpointMap["/submission/"+sample.submissionID] = vmraySubmissionStatus
		resp, err := sample.vr.GetSubmissionStatus(sample.submissionID)

		if err != nil {
			t.Error("Failed to get submission status for vmray: ", err.Error())
			return
		}

		if sample.ssresp.SubmissionFinished != resp.SubmissionFinished {
			t.Errorf("Unexpected result for vmray submission status SubmissionFinished , wanted: %v got: %v",
				sample.ssresp.SubmissionFinished, resp.SubmissionFinished)
		}
	}
}
