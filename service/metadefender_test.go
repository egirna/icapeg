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
	metadefenderScanFile   = "metadefender_submit_file_endpoint"
	metadefenderReportFile = "metadefender_report_file_endpoint"
)

var (
	metadefenderEndpointMap = map[string]string{
		"/file": metadefenderScanFile,
	}
)

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
