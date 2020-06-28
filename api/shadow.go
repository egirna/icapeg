package api

import (
	"bytes"
	"fmt"
	"icapeg/config"
	"icapeg/service"
	"icapeg/utils"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/viper"
)

func getShadowConfig(name string) config.RemoteICAPConfig {
	return config.RemoteICAPConfig{
		BaseURL:         viper.GetString(fmt.Sprintf("%s.base_url", name)),
		ReqmodEndpoint:  viper.GetString(fmt.Sprintf("%s.reqmod_endpoint", name)),
		RespmodEndpoint: viper.GetString(fmt.Sprintf("%s.respmod_endpoint", name)),
		OptionsEndpoint: viper.GetString(fmt.Sprintf("%s.options_endpoint", name)),
		Timeout:         viper.GetDuration(fmt.Sprintf("%s.timeout", name)) * time.Second,
	}
}

func doShadowOPTIONS(svc service.RemoteICAPService, alternativeEndpoint string) {
	siCfg := getShadowConfig(config.Shadow().RemoteICAP)
	svc.Endpoint = alternativeEndpoint
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

func doShadowRESPMOD(svc service.RemoteICAPService, httpReq http.Request, httpResp http.Response) {
	siCfg := getShadowConfig(config.Shadow().RemoteICAP)
	svc.Endpoint = siCfg.RespmodEndpoint
	svc.HTTPRequest = &httpReq
	svc.HTTPResponse = &httpResp

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

func doShadowREQMOD(svc service.RemoteICAPService, httpReq http.Request) {
	siCfg := getShadowConfig(config.Shadow().RemoteICAP)
	svc.Endpoint = siCfg.ReqmodEndpoint
	svc.HTTPRequest = &httpReq

	if httpReq.URL.Scheme == "" {
		debugLogger.LogToFile("Scheme not found, changing the url")
		httpReq.URL = utils.GetNewURL(&httpReq)
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
