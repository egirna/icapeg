package service

import (
	"encoding/json"
	"icapeg/dtos"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestGetService(t *testing.T) {
	type testSample struct {
		svcName string
		svc     Service
	}

	sampleTable := []testSample{
		{
			svcName: "virustotal",
			svc:     NewVirusTotalService(),
		},
		{
			svcName: "vmray",
			svc:     NewVmrayService(),
		},
		{
			svcName: "metadefender",
			svc:     NewMetaDefenderService(),
		},
		{
			svcName: "somename",
			svc:     nil,
		},
	}

	for _, sample := range sampleTable {
		svc := GetService(sample.svcName)

		gotType := reflect.TypeOf(svc)
		wantType := reflect.TypeOf(sample.svc)
		if gotType != wantType {
			t.Errorf("GetService failed for %s , wanted: %v , got: %v", sample.svcName, wantType, gotType)
		}
	}

}

func TestLocalService(t *testing.T) {
	type testSample struct {
		svcName string
		svc     LocalService
	}

	sampleTable := []testSample{
		{
			svcName: "clamav",
			svc:     NewClamavService(),
		},
		{
			svcName: "somename",
			svc:     nil,
		},
	}

	for _, sample := range sampleTable {
		svc := GetLocalService(sample.svcName)

		gotType := reflect.TypeOf(svc)
		wantType := reflect.TypeOf(sample.svc)
		if gotType != wantType {
			t.Errorf("GetLocalService failed for %s , wanted: %v , got: %v", sample.svcName, wantType, gotType)
		}
	}
}

func TestIsServiceLocal(t *testing.T) {
	type testSample struct {
		svcName string
		isLocal bool
	}

	sampleTable := []testSample{
		{
			svcName: "clamav",
			isLocal: true,
		},
		{
			svcName: "virustotal",
			isLocal: false,
		},
		{
			svcName: "somename",
			isLocal: false,
		},
	}

	for _, sample := range sampleTable {
		got := IsServiceLocal(sample.svcName)
		want := sample.isLocal

		if got != want {
			t.Errorf("IsServiceLocal failed for %s, wanted: %v , got: %v", sample.svcName, want, got)
		}

	}
}

func getVirusTotalMockServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-type", "application/json")

		urlStr := r.URL.EscapedPath()

		if _, exists := virustalEndpointMap[urlStr]; !exists {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"message":"Route not found"}`))
			return
		}

		var jsonRep []byte
		var err error
		endpoint := virustalEndpointMap[urlStr]

		if endpoint == virustotalFileScan || endpoint == virustotalURLScan {

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

			if endpoint == virustotalURLScan {
				vresp.Resource = "https://www.eicar.org/download/eicar.com"
				vresp.Permalink = "https://www.virustotal.com/url/b0088072c305c3ded6bedd90a4cfdd0e6a414116f7e4934622d7493ee0063d58/analysis/1588518740/"
			}

			jsonRep, err = json.Marshal(vresp)

			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"message":"Somethign went wrong"}`))
				return
			}

		}

		if endpoint == virustotalFileReport || endpoint == virustotalURReport {

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

			if endpoint == virustotalURReport {
				vresp.Resource = resource
				vresp.Permalink = "https://www.virustotal.com/url/b0088072c305c3ded6bedd90a4cfdd0e6a414116f7e4934622d7493ee0063d58/analysis/1588518740/"
			}

			if resource == "abcd12345" {
				vresp.Positives = 1
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
