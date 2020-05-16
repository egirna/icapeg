package service

import (
	"bytes"
	"icapeg/dtos"
	"testing"
	"time"
)

const (
	virustotalFileScan   = "virustotal_file_scan_endpoint"
	virustotalFileReport = "virustotal_file_report_endpoint"
	virustotalURLScan    = "virustotal_url_scan_endpoint"
	virustotalURReport   = "virustotal_url_report_endpoint"
)

var (
	virustalEndpointMap = map[string]string{
		"/file/scan":   virustotalFileScan,
		"/file/report": virustotalFileReport,
		"/url/scan":    virustotalURLScan,
		"/url/report":  virustotalURReport,
	}
)

func TestVirusTotalSubmitFile(t *testing.T) {

	testServer := getVirusTotalMockServer()

	defer testServer.Close()

	type testSample struct {
		vt    *VirusTotal
		sresp *dtos.SubmitResponse
	}

	sampleTable := []testSample{
		{
			vt: &VirusTotal{
				BaseURL:          testServer.URL,
				Timeout:          10 * time.Second,
				APIKey:           "someapikey",
				FileScanEndpoint: "/file/scan",
			},
			sresp: &dtos.SubmitResponse{
				SubmissionExists:   true,
				SubmissionID:       "275a021bbfb6489e54d471899f7db9d1663fc695ec2fe2a2c4538aabf651fd0f",
				SubmissionSampleID: "275a021bbfb6489e54d471899f7db9d1663fc695ec2fe2a2c4538aabf651fd0f",
			},
		},
		{
			vt: &VirusTotal{
				BaseURL:          testServer.URL,
				Timeout:          10 * time.Second,
				APIKey:           "",
				FileScanEndpoint: "/file/scan",
			},
			sresp: nil,
		},
	}

	for _, sample := range sampleTable {
		resp, err := sample.vt.SubmitFile(&bytes.Buffer{}, "somefile.exe")

		if resp != nil && err != nil {
			t.Error("Unexpected response from virustotal submit file: ", err.Error())
			return
		}

		if (sample.sresp == nil && resp != nil) || (err != nil && sample.sresp != nil) {
			t.Error("Unexpected response from virustotal submit file: ", err.Error())
			return
		}

		if resp != nil && !resp.SubmissionExists {
			t.Errorf("Unexpected result for virustotal submit file SubmissionExists , wanted: %v got: %v",
				sample.sresp.SubmissionExists, resp.SubmissionExists)
		}

		if resp != nil && resp.SubmissionID != sample.sresp.SubmissionID {
			t.Errorf("Unexpected result for virustotal submit file SubmissionID, wanted: %v got: %v",
				sample.sresp.SubmissionID, resp.SubmissionID)
		}

		if resp != nil && resp.SubmissionSampleID != sample.sresp.SubmissionSampleID {
			t.Errorf("Unexpected result for virustotal submit file SubmissionSampleID, wanted: %v got: %v",
				sample.sresp.SubmissionSampleID, resp.SubmissionSampleID)
		}
	}

}

func TestVirusTotalSubmitURL(t *testing.T) {

	testServer := getVirusTotalMockServer()

	defer testServer.Close()

	type testSample struct {
		vt    *VirusTotal
		sresp *dtos.SubmitResponse
	}

	sampleTable := []testSample{
		{
			vt: &VirusTotal{
				BaseURL:         testServer.URL,
				Timeout:         10 * time.Second,
				APIKey:          "someapikey",
				URLScanEndpoint: "/url/scan",
			},
			sresp: &dtos.SubmitResponse{
				SubmissionExists:   true,
				SubmissionID:       "https://www.eicar.org/download/eicar.com",
				SubmissionSampleID: "https://www.eicar.org/download/eicar.com",
			},
		},
		{
			vt: &VirusTotal{
				BaseURL:         testServer.URL,
				Timeout:         10 * time.Second,
				APIKey:          "",
				URLScanEndpoint: "/url/scan",
			},
			sresp: nil,
		},
	}

	for _, sample := range sampleTable {
		resp, err := sample.vt.SubmitURL("http://somehost.com/somefile.exe", "somefile.exe")

		if resp != nil && err != nil {
			t.Error("Unexpected response from virustotal submit url: ", err.Error())
			return
		}

		if sample.sresp == nil && resp != nil {
			t.Errorf("Unexpected result for virustotal submit url  , wanted: %v got: %v",
				sample.sresp, resp)
			return
		}

		if resp == nil {
			continue
		}

		if !resp.SubmissionExists {
			t.Errorf("Unexpected result for virustotal submit url SubmissionExists , wanted: %v got: %v",
				sample.sresp.SubmissionExists, resp.SubmissionExists)
		}

		if resp.SubmissionID != sample.sresp.SubmissionID {
			t.Errorf("Unexpected result for virustotal submit url SubmissionID, wanted: %v got: %v",
				sample.sresp.SubmissionID, resp.SubmissionID)
		}

		if resp.SubmissionSampleID != sample.sresp.SubmissionSampleID {
			t.Errorf("Unexpected result for virustotal submit url SubmissionSampleID, wanted: %v got: %v",
				sample.sresp.SubmissionSampleID, resp.SubmissionSampleID)
		}
	}
}

func TestVirusTotalSampleFileInfo(t *testing.T) {

	testServer := getVirusTotalMockServer()

	defer testServer.Close()

	type testSample struct {
		vt       *VirusTotal
		fi       dtos.FileMetaInfo
		siResp   *dtos.SampleInfo
		sampleID string
	}

	sampleTable := []testSample{
		{
			vt: &VirusTotal{
				BaseURL:            testServer.URL,
				Timeout:            10 * time.Second,
				APIKey:             "someapikey",
				FileReportEndpoint: "/file/report?apikey=%s&resource=%s",
				FailThreshold:      3,
			},
			fi: dtos.FileMetaInfo{
				FileName: "somefile.exe",
				FileSize: 3556000.0,
				FileType: "exe",
			},
			siResp: &dtos.SampleInfo{
				FileName:           "somefile.exe",
				SampleType:         "exe",
				SampleSeverity:     "malicious",
				VTIScore:           "5/5",
				FileSizeStr:        "3.56mb",
				SubmissionFinished: true,
			},
			sampleID: "12345abcd",
		},
		{
			vt: &VirusTotal{
				BaseURL:            testServer.URL,
				Timeout:            10 * time.Second,
				APIKey:             "someapikey",
				FileReportEndpoint: "/file/report?apikey=%s&resource=%s",
				FailThreshold:      3,
			},
			fi: dtos.FileMetaInfo{
				FileName: "somefile.pdf",
				FileSize: 220000.0,
				FileType: "pdf",
			},
			siResp: &dtos.SampleInfo{
				FileName:           "somefile.pdf",
				SampleType:         "pdf",
				SampleSeverity:     "ok",
				VTIScore:           "1/5",
				FileSizeStr:        "0.22mb",
				SubmissionFinished: true,
			},
			sampleID: "abcd12345",
		},
	}

	for _, sample := range sampleTable {
		resp, err := sample.vt.GetSampleFileInfo(sample.sampleID, sample.fi)

		if err != nil {
			t.Error("Failed to make get sample file info request for virustotal: ", err.Error())
			return
		}

		if sample.siResp.FileName != resp.FileName {
			t.Errorf("Unexpected result for virustotal sample file info FileName, wanted: %s got: %s",
				sample.siResp.FileName, resp.FileName)
		}

		if sample.siResp.SampleType != resp.SampleType {
			t.Errorf("Unexpected result for virustotal sample file info SampleType, wanted: %s got: %s",
				sample.siResp.SampleType, resp.SampleType)
		}

		if sample.siResp.SampleSeverity != resp.SampleSeverity {
			t.Errorf("Unexpected result for virustotal sample file info SampleSeverity, wanted: %s got: %s",
				sample.siResp.SampleSeverity, resp.SampleSeverity)
		}

		if sample.siResp.VTIScore != resp.VTIScore {
			t.Errorf("Unexpected result for virustotal sample file info VTIScore, wanted: %s got: %s",
				sample.siResp.VTIScore, resp.VTIScore)
		}

		if sample.siResp.FileSizeStr != resp.FileSizeStr {
			t.Errorf("Unexpected result for virustotal sample file info FileSizeStr, wanted: %s got: %s",
				sample.siResp.FileSizeStr, resp.FileSizeStr)
		}

		if sample.siResp.SubmissionFinished != resp.SubmissionFinished {
			t.Errorf("Unexpected result for virustotal sample file info SubmissionFinished, wanted: %v got: %v",
				sample.siResp.SubmissionFinished, resp.SubmissionFinished)
		}
	}

}

func TestVirusTotalSampleURLInfo(t *testing.T) {

	testServer := getVirusTotalMockServer()

	defer testServer.Close()

	type testSample struct {
		vt       *VirusTotal
		fi       dtos.FileMetaInfo
		siResp   *dtos.SampleInfo
		sampleID string
	}

	sampleTable := []testSample{
		{
			vt: &VirusTotal{
				BaseURL:           testServer.URL,
				Timeout:           10 * time.Second,
				APIKey:            "someapikey",
				URLReportEndpoint: "/url/report?apikey=%s&resource=%s",
				FailThreshold:     3,
			},
			fi: dtos.FileMetaInfo{
				FileName: "somefile.exe",
				FileSize: 3556000.0,
				FileType: "exe",
			},
			siResp: &dtos.SampleInfo{
				FileName:           "somefile.exe",
				SampleType:         "exe",
				SampleSeverity:     "malicious",
				VTIScore:           "5/5",
				FileSizeStr:        "3.56mb",
				SubmissionFinished: true,
			},
			sampleID: "12345abcd",
		},
		{
			vt: &VirusTotal{
				BaseURL:           testServer.URL,
				Timeout:           10 * time.Second,
				APIKey:            "",
				URLReportEndpoint: "/url/report?apikey=%s&resource=%s",
				FailThreshold:     3,
			},
			fi: dtos.FileMetaInfo{
				FileName: "somefile.pdf",
				FileSize: 220000.0,
				FileType: "pdf",
			},
			siResp:   nil,
			sampleID: "abcd12345",
		},
	}

	for _, sample := range sampleTable {
		resp, err := sample.vt.GetSampleURLInfo(sample.sampleID, sample.fi)

		if resp != nil && err != nil {
			t.Error("Failed to make get sample url info request for virustotal: ", err.Error())
			return
		}

		if sample.siResp == nil && resp != nil {
			t.Errorf("Unexpected result for virustotal url report , wanted: %v got: %v",
				sample.siResp, resp)
			return
		}

		if resp == nil {
			continue
		}

		if sample.siResp.FileName != resp.FileName {
			t.Errorf("Unexpected result for virustotal sample url info FileName, wanted: %s got: %s",
				sample.siResp.FileName, resp.FileName)
		}

		if sample.siResp.SampleType != resp.SampleType {
			t.Errorf("Unexpected result for virustotal sample url info SampleType, wanted: %s got: %s",
				sample.siResp.SampleType, resp.SampleType)
		}

		if sample.siResp.SampleSeverity != resp.SampleSeverity {
			t.Errorf("Unexpected result for virustotal sample url info SampleSeverity, wanted: %s got: %s",
				sample.siResp.SampleSeverity, resp.SampleSeverity)
		}

		if sample.siResp.VTIScore != resp.VTIScore {
			t.Errorf("Unexpected result for virustotal sample url info VTIScore, wanted: %s got: %s",
				sample.siResp.VTIScore, resp.VTIScore)
		}

		if sample.siResp.FileSizeStr != resp.FileSizeStr {
			t.Errorf("Unexpected result for virustotal sample url info FileSizeStr, wanted: %s got: %s",
				sample.siResp.FileSizeStr, resp.FileSizeStr)
		}

		if sample.siResp.SubmissionFinished != resp.SubmissionFinished {
			t.Errorf("Unexpected result for virustotal sample url info SubmissionFinished, wanted: %v got: %v",
				sample.siResp.SubmissionFinished, resp.SubmissionFinished)
		}
	}

}
