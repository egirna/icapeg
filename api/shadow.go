package api

import (
	"bytes"
	"icapeg/config"
	"icapeg/service"
	"icapeg/utils"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

func performShadowOPTIONS(svc service.RemoteICAPService) {
	siCfg := config.ShadowICAP()
	svc.Endpoint = siCfg.RespmodEndpoint
	if siCfg.OptionsEndpoint != "" {
		svc.Endpoint = siCfg.OptionsEndpoint
	}

	resp, err := service.RemoteICAPOptions(svc)

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

func performShadowRESPMOD(svc service.RemoteICAPService, httpReq http.Request, httpResp http.Response) {
	siCfg := config.ShadowICAP()
	svc.Endpoint = siCfg.RespmodEndpoint
	svc.HTTPRequest = &httpReq
	svc.HTTPResponse = &httpResp

	if httpReq.URL.Scheme == "" {
		debugLogger.LogToFile("Scheme not found, changing the url")
		u, _ := url.Parse("http://" + httpReq.Host + httpReq.URL.Path)
		httpReq.URL = u
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

	resp, err := service.RemoteICAPRespmod(svc)

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

func performShadowREQMOD(svc service.RemoteICAPService, httpReq http.Request) {
	siCfg := config.ShadowICAP()
	svc.Endpoint = siCfg.ReqmodEndpoint
	svc.HTTPRequest = &httpReq

	if httpReq.URL.Scheme == "" {
		debugLogger.LogToFile("Scheme not found, changing the url")
		u, _ := url.Parse("http://" + httpReq.Host + httpReq.URL.Path)
		httpReq.URL = u
	}

	ext := utils.GetFileExtension(&httpReq)

	if ext == "" {
		debugLogger.LogToFile("Processing not required...")
		return
	}

	resp, err := service.RemoteICAPReqmod(svc)

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
