package api

import (
	"bytes"
	"icapeg/config"
	"icapeg/dtos"
	"icapeg/service"
	"icapeg/utils"
	"io/ioutil"
	"net/http"
	"strings"
)

func doShadowOPTIONS(svc service.ICAPService) {

	resp, err := svc.DoOptions()

	if err != nil {
		errorLogger.LogfToFile("Failed to make OPTIONS call of shadow icap server: %s\n", err.Error())
		return
	}

	infoLogger.LogToFile("Received response from the shadow ICAP server with the following info:")
	infoLogger.LogToFile("Status Code: ", resp.StatusCode)
	infoLogger.LogToFile("Headers:")
	infoLogger.LogToFile("---------")
	for header, values := range resp.Header {
		infoLogger.LogfToFile("%s: %v\n", header, values)
	}
}

func doShadowRESPMOD(svc service.ICAPService, httpReq http.Request, httpResp http.Response) {
	svc.SetHTTPRequest(&httpReq)
	svc.SetHTTPResponse(&httpResp)

	if httpReq.URL.Scheme == "" {
		debugLogger.LogToFile("Scheme not found, changing the url")
		httpReq.URL = utils.GetNewURL(&httpReq)
	}

	b, err := ioutil.ReadAll(httpResp.Body)

	if err != nil {
		errorLogger.LogToFile("Error reading the body: ", err.Error())
	}

	bdyStr := string(b)
	if len(b) > int(httpResp.ContentLength) {
		if strings.HasSuffix(bdyStr, "\n\n") {
			bdyStr = strings.TrimSuffix(bdyStr, "\n\n")
		}
	}

	httpResp.Body = ioutil.NopCloser(bytes.NewBuffer([]byte(bdyStr)))

	resp, err := svc.DoRespmod()

	if err != nil {
		errorLogger.LogfToFile("Failed to make RESPMOD call to shadow icap server: %s\n", err.Error())
		return
	}

	infoLogger.LogToFile("Received response from the shadow ICAP server with the following info:")
	infoLogger.LogToFile("Status Code: ", resp.StatusCode)
	infoLogger.LogToFile("Headers:")
	infoLogger.LogToFile("---------")
	for header, values := range resp.Header {
		infoLogger.LogfToFile("%s: %v\n", header, values)
	}
	if resp.ContentResponse != nil {
		infoLogger.LogToFile("HTTP Response Headers:")
		infoLogger.LogToFile("----------------------")
		for header, values := range resp.ContentResponse.Header {
			infoLogger.LogfToFile("%s: %v\n", header, values)
		}
	}
}

func doShadowREQMOD(svc service.ICAPService, httpReq http.Request) {
	svc.SetHTTPRequest(&httpReq)

	if httpReq.URL.Scheme == "" {
		debugLogger.LogToFile("Scheme not found, changing the url")
		httpReq.URL = utils.GetNewURL(&httpReq)
	}

	ext := utils.GetFileExtension(&httpReq)

	if ext == "" {
		debugLogger.LogToFile("Processing not required...")
		return
	}

	resp, err := svc.DoReqmod()

	if err != nil {
		errorLogger.LogfToFile("Failed to make REQMOD call to shadow icap server: %s\n", err.Error())
		return
	}

	infoLogger.LogToFile("Received response from the shadow ICAP server with the following info:")
	infoLogger.LogToFile("Status Code: ", resp.StatusCode)
	infoLogger.LogToFile("Headers:")
	infoLogger.LogToFile("---------")
	for header, values := range resp.Header {
		infoLogger.LogfToFile("%s: %v\n", header, values)
	}
	if resp.ContentResponse != nil {
		infoLogger.LogToFile("HTTP Response Headers:")
		infoLogger.LogToFile("----------------------")
		for header, values := range resp.ContentResponse.Header {
			infoLogger.LogfToFile("%s: %v\n", header, values)
		}
	}

}

func doShadowScan(filename string, fmi dtos.FileMetaInfo, buf *bytes.Buffer, fileURL string) {

	scannerName := ""
	if buf == nil && fileURL != "" {
		scannerName = config.App().ReqScannerVendorShadow
	}
	if buf != nil && fileURL == "" {
		scannerName = config.App().RespScannerVendorShadow
	}

	var sts int
	var si *dtos.SampleInfo

	localService := service.IsServiceLocal(scannerName)

	if localService && buf != nil { // if the scanner is installed locally
		sts, si = doLocalScan(scannerName, fmi, buf)
	}

	if !localService { // if the scanner is an external service requiring API calls.

		if buf == nil && fileURL != "" { // indicates this is a URL scan request
			sts, si = doRemoteURLScan(scannerName, filename, fmi, fileURL)
		}

		if buf != nil && fileURL == "" { // indicates this is a File scan request
			sts, si = doRemoteFileScan(scannerName, filename, fmi, buf)
		}

	}

	infoLogger.LogToFile("The Shadow Result:")
	infoLogger.LogToFile("------------------------------")

	infoLogger.LogfToFile("Response Status from the shadow scanner(%s): %d", scannerName, sts)
	if sts == http.StatusNoContent {
		infoLogger.LogfToFile("The file:%s is good to go\n", filename)
	}
	if sts == http.StatusOK {
		infoLogger.LogToFile("File Name: ", si.FileName)
		infoLogger.LogToFile("File Type: ", si.SampleType)
		infoLogger.LogToFile("File Size: ", si.FileSizeStr)
		infoLogger.LogToFile("Requested URL: ", utils.BreakHTTPURL(fileURL))
		infoLogger.LogToFile("Severity", si.SampleSeverity)
		infoLogger.LogToFile("Positive Score: ", si.VTIScore)
		infoLogger.LogToFile("Results By: ", scannerName)
	}

	infoLogger.LogToFile("------------------------------")

}
