package service

import (
	"bytes"
	"encoding/json"
	"icapeg/dtos"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestVirusTotalSubmitFile(t *testing.T) {

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Add("Content-type", "application/json")

		if r.URL.String() != "/file/scan" {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"message":"Route not found"}`))
			return
		}

		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if err := r.ParseMultipartForm(1024); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"Somethign went wrong"}`))
			return
		}

		if apikey, exists := r.MultipartForm.Value["apikey"]; !exists || (len(apikey) > 0 && apikey[0] == "") {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		vresp := dtos.VirusTotalScanFileResponse{
			ScanID:       "275a021bbfb6489e54d471899f7db9d1663fc695ec2fe2a2c4538aabf651fd0f-1588501687",
			Sha1:         "3395856ce81f2b7382dee72602f798b642f14140",
			Resource:     "275a021bbfb6489e54d471899f7db9d1663fc695ec2fe2a2c4538aabf651fd0f",
			ResponseCode: 1,
			Sha256:       "275a021bbfb6489e54d471899f7db9d1663fc695ec2fe2a2c4538aabf651fd0f",
			Permalink:    "https://www.virustotal.com/file/275a021bbfb6489e54d471899f7db9d1663fc695ec2fe2a2c4538aabf651fd0f/analysis/1588501687/",
			Md5:          "44d88612fea8a8f36de82e1278abb02f",
			VerboseMsg:   "Scan request successfully queued, come back later for the report",
		}

		jsonRep, err := json.Marshal(vresp)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"Somethign went wrong"}`))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(jsonRep)
		return
	}))

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

		if (sample.sresp == nil && resp != nil) || (resp == nil && sample.sresp != nil) {
			t.Errorf("Unexpected result for virustotal submit file  , wanted: %v got: %v",
				sample.sresp, resp)
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

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Add("Content-type", "application/json")

		if r.URL.String() != "/url/scan" {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"message":"Route not found"}`))
			return
		}

		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if err := r.ParseMultipartForm(1024); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"Somethign went wrong"}`))
			return
		}

		if apikey, exists := r.MultipartForm.Value["apikey"]; !exists || (len(apikey) > 0 && apikey[0] == "") {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		vresp := dtos.VirusTotalScanFileResponse{
			ScanID:       "275a021bbfb6489e54d471899f7db9d1663fc695ec2fe2a2c4538aabf651fd0f-1588501687",
			Resource:     "https://www.eicar.org/download/eicar.com",
			ResponseCode: 1,
			Permalink:    "https://www.virustotal.com/url/b0088072c305c3ded6bedd90a4cfdd0e6a414116f7e4934622d7493ee0063d58/analysis/1588518740/",
			VerboseMsg:   "Scan request successfully queued, come back later for the report",
		}

		jsonRep, err := json.Marshal(vresp)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"Somethign went wrong"}`))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(jsonRep)
		return
	}))

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

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Add("Content-type", "application/json")

		if r.URL.EscapedPath() != "/file/report" {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"message":"Route not found"}`))
			return
		}

		if r.Method != http.MethodGet && r.Method != http.MethodPost {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		apikey := r.URL.Query().Get("apikey")

		if apikey == "" {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		resource := r.URL.Query().Get("resource")

		vresp := dtos.VirusTotalReportResponse{
			Scans: dtos.Scans{
				Bkav: dtos.Scanner{
					Detected: true,
					Version:  "1.3.0.9899",
					Result:   "DOS.EiracA.Trojan",
					Update:   "20200429",
				},
				DrWeb: dtos.Scanner{
					Detected: true,
					Version:  "7.0.46.3050",
					Result:   "EICAR Test File (NOT a Virus!)",
					Update:   "20200503",
				},
				MicroWorldEScan: dtos.Scanner{
					Detected: true,
					Version:  "14.0.409.0",
					Result:   "EICAR-Test-File",
					Update:   "20200503",
				},
				VBA32: dtos.Scanner{
					Detected: true,
					Version:  "4.3.0",
					Result:   "EICAR-Test-File",
					Update:   "20200430",
				},
				FireEye: dtos.Scanner{
					Detected: true,
					Version:  "32.31.0.0",
					Result:   "EICAR-Test-File (not a virus)",
					Update:   "20200316",
				},
			},
			ScanID:       "275a021bbfb6489e54d471899f7db9d1663fc695ec2fe2a2c4538aabf651fd0f-1588522969",
			Sha1:         "3395856ce81f2b7382dee72602f798b642f14140",
			Resource:     "275a021bbfb6489e54d471899f7db9d1663fc695ec2fe2a2c4538aabf651fd0f",
			ResponseCode: 1,
			ScanDate:     "2020-05-03 16:22:49",
			Permalink:    "https://www.virustotal.com/file/275a021bbfb6489e54d471899f7db9d1663fc695ec2fe2a2c4538aabf651fd0f/analysis/1588522969/",
			VerboseMsg:   "Scan finished, information embedded",
			Total:        5,
			Positives:    5,
			Sha256:       "275a021bbfb6489e54d471899f7db9d1663fc695ec2fe2a2c4538aabf651fd0f",
			Md5:          "44d88612fea8a8f36de82e1278abb02f",
		}

		if resource == "abcd12345" {
			vresp.Positives = 1
		}

		jsonRep, err := json.Marshal(vresp)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"Somethign went wrong"}`))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(jsonRep)
		return
	}))

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

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Add("Content-type", "application/json")

		if r.URL.EscapedPath() != "/url/report" {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"message":"Route not found"}`))
			return
		}

		if r.Method != http.MethodGet && r.Method != http.MethodPost {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		apikey := r.URL.Query().Get("apikey")

		if apikey == "" {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		resource := r.URL.Query().Get("resource")

		vresp := dtos.VirusTotalReportResponse{
			Scans: dtos.Scans{
				Bkav: dtos.Scanner{
					Detected: true,
					Version:  "1.3.0.9899",
					Result:   "DOS.EiracA.Trojan",
					Update:   "20200429",
				},
				DrWeb: dtos.Scanner{
					Detected: true,
					Version:  "7.0.46.3050",
					Result:   "EICAR Test File (NOT a Virus!)",
					Update:   "20200503",
				},
				MicroWorldEScan: dtos.Scanner{
					Detected: true,
					Version:  "14.0.409.0",
					Result:   "EICAR-Test-File",
					Update:   "20200503",
				},
				VBA32: dtos.Scanner{
					Detected: true,
					Version:  "4.3.0",
					Result:   "EICAR-Test-File",
					Update:   "20200430",
				},
				FireEye: dtos.Scanner{
					Detected: true,
					Version:  "32.31.0.0",
					Result:   "EICAR-Test-File (not a virus)",
					Update:   "20200316",
				},
			},
			ScanID:       "275a021bbfb6489e54d471899f7db9d1663fc695ec2fe2a2c4538aabf651fd0f-1588522969",
			Resource:     "275a021bbfb6489e54d471899f7db9d1663fc695ec2fe2a2c4538aabf651fd0f",
			ResponseCode: 1,
			ScanDate:     "2020-05-03 16:22:49",
			Permalink:    "https://www.virustotal.com/file/275a021bbfb6489e54d471899f7db9d1663fc695ec2fe2a2c4538aabf651fd0f/analysis/1588522969/",
			VerboseMsg:   "Scan finished, information embedded",
			Total:        5,
			Positives:    5,
		}

		if resource == "abcd12345" {
			vresp.Positives = 1
		}

		jsonRep, err := json.Marshal(vresp)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"Somethign went wrong"}`))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(jsonRep)
		return
	}))

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
