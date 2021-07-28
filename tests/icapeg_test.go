package tests

import (
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

	// making the stop channel to control the stoppage of the test servers
	stop := make(chan os.Signal, 1)
	stopRemote := make(chan os.Signal, 1)
	stopShadow := make(chan os.Signal, 1)

	//starting the test ICAP server
	go startTestServer(stop)

	//preparing the third-party mock virus scanner servers
	tss := getThirdPartyServers()
	startThirdPartyServers(tss)

	//stopping the third-party mock virus scanner servers
	stopThirdPartyServers(tss)

	// Preparing the Remote ICAP servers and its required configurations

	appCfg := config.App()
	appCfg.RespScannerVendor = "icap_something"
	appCfg.ReqScannerVendor = "icap_something"
	appCfg.RespScannerVendorShadow = "icap_somethingelse"
	appCfg.ReqScannerVendorShadow = "icap_somethingelse"

	// stopping the ICAP servers
	stopTestServer(stop)
	stopRemoteICAPMockServer(stopRemote)
	stopRemoteICAPMockServer(stopShadow)
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

	if respmodService == serviceMetadefender || reqmodService == serviceMetadefender {
		tss = append(tss, getMetaDefenderMockServer())
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
