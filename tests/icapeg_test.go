package tests

import (
	"fmt"
	"icapeg/config"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	ic "github.com/egirna/icap-client"
)

func TestICAPeg(t *testing.T) {
	// initializing the test configurations
	config.InitTestConfig()

	// making the stop channel to control the stoppage of the test sever
	stop := make(chan os.Signal, 1)

	//starting the test ICAP server
	go startTestServer(stop)

	//preparing the third-party mock servers
	tss := getThirdPartyServers()
	startThirdPartyServers(tss)

	startTesting(t)

	//stopping the third-party mock servers & test ICAP server
	stopThirdPartyServers(tss)
	stopTestServer(stop)
}

func startTesting(t *testing.T) {

	t.Run("Testing With A Bad File", func(t *testing.T) {
		httpReq, err := makeDownloadFileHTTPRequest(badFileURL)

		if err != nil {
			t.Errorf("Failed to make download file request: %s", err.Error())
			return
		}

		t.Log("Performing REQMOD...")

		resp, err := performReqmod(fmt.Sprintf("icap://localhost:%d/reqmod-icapeg", config.App().Port), httpReq)

		if err != nil {
			t.Errorf("Failed to perform reqmod request: %s", err.Error())
			return
		}

		wantedStatusCode := http.StatusOK
		gotStatusCode := resp.StatusCode
		if wantedStatusCode != gotStatusCode {
			t.Errorf("Wanted status code: %d got: %d", wantedStatusCode, gotStatusCode)
			return
		}

		httpResp, err := makeDownloadFileHTTPResponse(httpReq)

		if err != nil {
			t.Errorf("Failed to get download file response: %s", err.Error())
			return
		}

		t.Log("Performing RESPMOD...")

		resp, err = performRespmod(fmt.Sprintf("icap://localhost:%d/respmod-icapeg", config.App().Port), httpReq, httpResp)

		if err != nil {
			t.Errorf("Failed to perform respmod request: %s", err.Error())
			return
		}

		wantedStatusCode = http.StatusOK
		gotStatusCode = resp.StatusCode
		if wantedStatusCode != gotStatusCode {
			t.Errorf("Wanted status code: %d got: %d", wantedStatusCode, gotStatusCode)
			return
		}

		wantedContentType := "text/html"
		gotContentType := resp.ContentResponse.Header.Get("Content-Type")
		if wantedContentType != gotContentType {
			t.Errorf("Wanted content-type: %s got: %s", wantedContentType, gotContentType)
			return
		}
	})

	t.Run("Testing With A Good File", func(t *testing.T) {
		httpReq, err := makeDownloadFileHTTPRequest(goodFileURL)

		if err != nil {
			t.Errorf("Failed to make download file request: %s", err.Error())
			return
		}

		t.Log("Performing REQMOD...")

		resp, err := performReqmod(fmt.Sprintf("icap://localhost:%d/reqmod-icapeg", config.App().Port), httpReq)

		if err != nil {
			t.Errorf("Failed to perform reqmod request: %s", err.Error())
			return
		}

		wantedStatusCode := http.StatusNoContent
		gotStatusCode := resp.StatusCode
		if wantedStatusCode != gotStatusCode {
			t.Errorf("Wanted status code: %d got: %d", wantedStatusCode, gotStatusCode)
			return
		}

		httpResp, err := makeDownloadFileHTTPResponse(httpReq)

		if err != nil {
			t.Errorf("Failed to get download file response: %s", err.Error())
			return
		}

		t.Log("Performing RESPMOD...")

		resp, err = performRespmod(fmt.Sprintf("icap://localhost:%d/respmod-icapeg", config.App().Port), httpReq, httpResp)

		if err != nil {
			t.Errorf("Failed to perform respmod request: %s", err.Error())
			return
		}

		wantedStatusCode = http.StatusNoContent
		gotStatusCode = resp.StatusCode
		if wantedStatusCode != gotStatusCode {
			t.Errorf("Wanted status code: %d got: %d", wantedStatusCode, gotStatusCode)
			return
		}

		wantedEncapsulated := "null-body=0"
		gotEncapsulated := resp.Header.Get("Encapsulated")
		if wantedEncapsulated != gotEncapsulated {
			t.Errorf("Wanted encapsulated: %s got: %s", wantedEncapsulated, gotEncapsulated)
			return
		}
	})

}

func performReqmod(url string, httpReq *http.Request) (*ic.Response, error) {

	optReq, err := ic.NewRequest(ic.MethodOPTIONS, url, nil, nil)

	if err != nil {
		return nil, err
	}

	client := &ic.Client{
		Timeout: 10 * time.Second,
	}

	optResp, err := client.Do(optReq)

	if err != nil {
		return nil, err
	}

	req, err := ic.NewRequest(ic.MethodREQMOD, url, httpReq, nil)

	if err != nil {
		return nil, err
	}

	if err := req.SetPreview(optResp.PreviewBytes); err != nil {
		return nil, err
	}

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}
	return resp, nil
}

func performRespmod(url string, httpReq *http.Request, httpResp *http.Response) (*ic.Response, error) {
	optReq, err := ic.NewRequest(ic.MethodOPTIONS, url, nil, nil)

	if err != nil {
		return nil, err
	}

	client := &ic.Client{
		Timeout: 10 * time.Second,
	}

	optResp, err := client.Do(optReq)

	if err != nil {
		return nil, err
	}

	req, err := ic.NewRequest(ic.MethodRESPMOD, url, httpReq, httpResp)

	if err != nil {
		return nil, err
	}

	if err := req.SetPreview(optResp.PreviewBytes); err != nil {
		return nil, err
	}

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}
	return resp, nil
}

func makeDownloadFileHTTPRequest(url string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)

	if err != nil {
		return nil, err
	}

	return req, nil
}

func makeDownloadFileHTTPResponse(req *http.Request) (*http.Response, error) {
	client := &http.Client{}

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	return resp, nil
}

func getThirdPartyServers() []*httptest.Server {
	tss := []*httptest.Server{}

	respmodService := config.App().RespScannerVendor
	reqmodService := config.App().ReqScannerVendor

	if respmodService == serviceVirustotal || reqmodService == serviceVirustotal {
		tss = append(tss, getVirusTotalMockServer())
	}

	if respmodService == serviceVmray || reqmodService == serviceVmray {
		tss = append(tss, getVmrayMockServer())
	}

	return tss

}

func startThirdPartyServers(tss []*httptest.Server) {
	for _, ts := range tss {
		ts.Start()
	}
}

func stopThirdPartyServers(tss []*httptest.Server) {
	for _, ts := range tss {
		ts.Close()
	}
}
