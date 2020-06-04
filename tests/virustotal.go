package tests

import (
	"encoding/json"
	"icapeg/dtos"
	"log"
	"net"
	"net/http"
	"net/http/httptest"

	"github.com/spf13/viper"
)

const (
	virustotalFileScan   = "virustotal_file_scan_endpoint"
	virustotalFileReport = "virustotal_file_report_endpoint"
	virustotalURLScan    = "virustotal_url_scan_endpoint"
	virustotalURReport   = "virustotal_url_report_endpoint"
	serviceVirustotal    = "virustotal"
	virustotalURL        = "127.0.0.1:8001"
)

func getVirusTotalMockServer() *httptest.Server {
	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		virustotalEndpointMap := map[string]string{
			viper.GetString("virustotal.file_scan_endpoint"):   virustotalFileScan,
			viper.GetString("virustotal.file_report_endpoint"): virustotalFileReport,
			viper.GetString("virustotal.url_scan_endpoint"):    virustotalURLScan,
			viper.GetString("virustotal.url_report_endpoint"):  virustotalURReport,
		}

		w.Header().Add("Content-type", "application/json")

		urlStr := r.URL.EscapedPath()

		if _, exists := virustotalEndpointMap[urlStr]; !exists {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"message":"Route not found"}`))
			return
		}

		var jsonRep []byte
		var err error
		endpoint := virustotalEndpointMap[urlStr]

		if endpoint == virustotalFileScan || endpoint == virustotalURLScan {

			if r.Method != http.MethodPost {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			if err := r.ParseMultipartForm(1024); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"message":"Something went wrong"}`))
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

	lstnr, err := net.Listen("tcp", virustotalURL)

	if err != nil {
		log.Fatal(err.Error())
	}

	ts.Listener = lstnr

	return ts

}
