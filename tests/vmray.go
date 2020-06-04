package tests

import (
	"encoding/json"
	"fmt"
	"icapeg/dtos"
	"icapeg/utils"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
)

const (
	vmraySubmitFile       = "vmray_submit_file_endpoint"
	vmraySampleFileInfo   = "vmray_sample_file_info_endpoint"
	vmraySubmissionStatus = "vmray_submission_status_endpoint"
	vmrayURL              = "127.0.0.1:8003"
	serviceVmray          = "vmray"
	submissionID          = 5624736
	badSampleID           = 4841630
	goodSampleID          = 4321
)

var (
	vmrayEndpointMap = map[string]string{
		"/sample/submit": vmraySubmitFile,
		fmt.Sprintf("/submission/%d", submissionID): vmraySubmissionStatus,
		fmt.Sprintf("/sample/%d", badSampleID):      vmraySampleFileInfo,
		fmt.Sprintf("/sample/%d", goodSampleID):     vmraySampleFileInfo,
	}
)

func getVmrayMockServer() *httptest.Server {
	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

			resourceName := ""

			if r.MultipartForm.File["sample_file"] != nil {
				resourceName = r.MultipartForm.File["sample_file"][0].Filename
			}

			if url, exists := r.MultipartForm.Value["sample_url"]; exists {
				resourceName = url[0]
			}

			vresp := dtos.VmraySubmitResponse{
				Data: dtos.VmraySubmitData{
					Submissions: []dtos.VmraySubmissions{
						{
							SubmissionID:       submissionID,
							SubmissionSampleID: badSampleID,
						},
					},
				},
			}

			if resourceName != "eicar.com" && resourceName != badFileURL {
				vresp.Data.Submissions[0].SubmissionSampleID = goodSampleID
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
					SampleFilename: "eicar.com",
					SampleType:     "com",
					SampleSeverity: "malicious",
					SampleVtiScore: 20,
					SampleFilesize: 3028,
				},
			}

			if sampleID == strconv.Itoa(goodSampleID) {
				vresp.Data.SampleSeverity = "not_suspicious"
				vresp.Data.SampleVtiScore = 0
				vresp.Data.SampleFilename = "somepdf.pdf"
				vresp.Data.SampleType = "pdf"
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

	lstnr, err := net.Listen("tcp", vmrayURL)

	if err != nil {
		log.Fatal(err.Error())
	}

	ts.Listener = lstnr

	return ts

}
