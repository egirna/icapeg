package service

import (
	"bytes"
	"fmt"
	"icapeg/dtos"
	"icapeg/utils"
	"testing"
	"time"
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
