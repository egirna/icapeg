package tests

import (
	"fmt"
	"icapeg/config"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

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

	defer stopThirdPartyServers(tss)

	//stopping the test ICAP sercer
	defer stopTestServer(stop)
}

func startTesting(t *testing.T) {

	httpReq, err := http.NewRequest(http.MethodGet, "http://www.eicar.org/download/eicar.com", nil)

	if err != nil {
		t.Error(err.Error())
		return
	}

	urlStr := fmt.Sprintf("icap://localhost:%d/reqmod-icapeg", config.App().Port)
	req, err := ic.NewRequest(ic.MethodREQMOD, urlStr, httpReq, nil)

	if err != nil {
		t.Error(err.Error())
		return
	}

	client := &ic.Client{}

	resp, err := client.Do(req)

	if err != nil {
		t.Error(err.Error())
		return
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Wanted status code: %d got: %d", http.StatusOK, resp.StatusCode)
	}
}

func getThirdPartyServers() []*httptest.Server {
	tss := []*httptest.Server{}

	respmodService := config.App().RespScannerVendor
	reqmodService := config.App().ReqScannerVendor

	if respmodService == serviceVirustotal || reqmodService == serviceVirustotal {
		tss = append(tss, getVirusTotalMockServer())
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
