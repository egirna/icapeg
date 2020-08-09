package tests

import (
	"encoding/json"
	"fmt"
	"icapeg/dtos"
	"icapeg/utils"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
)

const (
	metadefenderScanFile   = "metadefender_submit_file_endpoint"
	metadefenderReportFile = "metadefender_report_file_endpoint"
	serviceMetadefender    = "metadefender"
	metadefenderURL        = "127.0.0.1:8002"
	badDataID              = "bzIwMDUwNW9Nb2lxOWktcHZkUXhpVVcyS05W"
	goodDataID             = "4321abcd"
)

var (
	metadefenderEndpointMap = map[string]string{
		"/file":                             metadefenderScanFile,
		fmt.Sprintf("/file/%s", badDataID):  metadefenderReportFile,
		fmt.Sprintf("/file/%s", goodDataID): metadefenderReportFile,
	}
)

func getMetaDefenderMockServer() *httptest.Server {
	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
				DataID: badDataID,
			}

			ext := utils.GetMimeExtension(data)

			if ext != "com" {
				mresp.DataID = goodDataID
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
					TotalDetectedAvs:   33,
					ProgressPercentage: 100,
				},
			}

			if dataID == goodDataID {
				mresp.ScanResults.TotalDetectedAvs = 1
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

	lstnr, err := net.Listen("tcp", metadefenderURL)

	if err != nil {
		log.Fatal(err.Error())
	}

	ts.Listener = lstnr

	return ts
}
