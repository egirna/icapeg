package api

import (
	"bytes"
	"icapeg/config"
	"icapeg/service"
	"icapeg/utils"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/egirna/icap"
)

func doRemoteOPTIONS(req *icap.Request, w icap.ResponseWriter, alternativeEndpoint string) {
	riCfg := config.RemoteICAP()
	var riSvc, siSvc service.RemoteICAPService

	riSvc = prepareRemoteSvc(req.Header, riCfg.BaseURL, riCfg.Timeout)
	infoLogger.LogToFile("Passing request to the remote ICAP server...")

	if config.Shadow().RemoteICAP != "" {
		siCfg := getShadowConfig(config.Shadow().RemoteICAP)
		siSvc = prepareRemoteSvc(req.Header, siCfg.BaseURL, siCfg.Timeout)
		infoLogger.LogToFile("Passing request to the shadow ICAP server...")
		go doShadowOPTIONS(siSvc, alternativeEndpoint)
	}

	riSvc.Endpoint = alternativeEndpoint
	if riCfg.OptionsEndpoint != "" {
		riSvc.Endpoint = riCfg.OptionsEndpoint
	}

	resp, err := service.RemoteICAPOptions(riSvc)

	if err != nil {
		errorLogger.LogfToFile("Failed to make OPTIONS call of remote icap server: %s\n", err.Error())
		w.WriteHeader(utils.IfPropagateError(http.StatusFailedDependency, http.StatusNoContent), nil, false)
		return
	}

	infoLogger.LogfToFile("Received response from the remote ICAP server wwith status code: %d...\n", resp.StatusCode)

	utils.CopyHeaders(resp.Header, w.Header(), utils.HeaderEncapsulated)

	w.WriteHeader(resp.StatusCode, nil, false)

}

func doRemoteRESPMOD(req *icap.Request, w icap.ResponseWriter) {

	riCfg := config.RemoteICAP()
	var riSvc, siSvc service.RemoteICAPService

	riSvc = prepareRemoteSvc(req.Header, riCfg.BaseURL, riCfg.Timeout)
	infoLogger.LogToFile("Passing request to the remote ICAP server...")

	if config.Shadow().RemoteICAP != "" {
		siCfg := getShadowConfig(config.Shadow().RemoteICAP)
		siSvc = prepareRemoteSvc(req.Header, siCfg.BaseURL, siCfg.Timeout)
		infoLogger.LogToFile("Passing request to the shadow ICAP server...")
		shadowHTTPResp := utils.GetHTTPResponseCopy(req.Response)
		go doShadowRESPMOD(siSvc, *req.Request, shadowHTTPResp)
	}

	riSvc.Endpoint = riCfg.RespmodEndpoint
	riSvc.HTTPRequest = req.Request
	riSvc.HTTPResponse = req.Response

	if req.Request.URL.Scheme == "" {
		debugLogger.LogToFile("Scheme not found, changing the url")
		req.Request.URL = utils.GetNewURL(req.Request)
	}

	b, err := ioutil.ReadAll(req.Response.Body)

	if err != nil {
		errorLogger.LogToFile("Error reading the body: ", err.Error())
	}

	bdyStr := string(b)
	if len(b) > int(req.Response.ContentLength) {
		if strings.HasSuffix(bdyStr, "\n\n") {
			bdyStr = strings.TrimSuffix(bdyStr, "\n\n")
		}
	}

	req.Response.Body = ioutil.NopCloser(bytes.NewBuffer([]byte(bdyStr)))

	resp, err := service.RemoteICAPRespmod(riSvc)

	if err != nil {
		errorLogger.LogfToFile("Failed to make RESPMOD call to remote icap server: %s\n", err.Error())
		w.WriteHeader(utils.IfPropagateError(http.StatusFailedDependency, http.StatusNoContent), nil, false)
		return
	}

	utils.CopyHeaders(resp.Header, w.Header(), "")

	infoLogger.LogfToFile("Received response from the remote ICAP server with status code: %d...\n", resp.StatusCode)

	if resp.StatusCode == http.StatusOK { // NOTE: this is done to render the error html page, not sure this is the proper way

		if resp.ContentResponse != nil && resp.ContentResponse.Body != nil {

			bdyByte, err := ioutil.ReadAll(resp.ContentResponse.Body)
			if err != nil && err != io.ErrUnexpectedEOF {
				errorLogger.LogToFile("Failed to read body from the remote icap response: ", err.Error())
				w.WriteHeader(utils.IfPropagateError(http.StatusInternalServerError, http.StatusNoContent), nil, false)
				return
			}

			defer resp.ContentResponse.Body.Close()

			w.WriteHeader(resp.StatusCode, resp.ContentResponse, true)

			w.Write(bdyByte)
			return
		}
	}
	w.WriteHeader(resp.StatusCode, nil, false)

}

func doRemoteREQMOD(req *icap.Request, w icap.ResponseWriter) {
	riCfg := config.RemoteICAP()
	var riSvc, siSvc service.RemoteICAPService

	riSvc = prepareRemoteSvc(req.Header, riCfg.BaseURL, riCfg.Timeout)
	infoLogger.LogToFile("Passing request to the remote ICAP server...")

	if config.Shadow().RemoteICAP != "" {
		siCfg := getShadowConfig(config.Shadow().RemoteICAP)
		siSvc = prepareRemoteSvc(req.Header, siCfg.BaseURL, siCfg.Timeout)
		infoLogger.LogToFile("Passing request to the shadow ICAP server...")
		go doShadowREQMOD(siSvc, *req.Request)
	}

	riSvc.Endpoint = riCfg.ReqmodEndpoint
	riSvc.HTTPRequest = req.Request

	if req.Request.URL.Scheme == "" {
		debugLogger.LogToFile("Scheme not found, changing the url")
		req.Request.URL = utils.GetNewURL(req.Request)
	}

	ext := utils.GetFileExtension(req.Request)

	if ext == "" {
		debugLogger.LogToFile("Processing not required...")
		w.WriteHeader(http.StatusNoContent, nil, false)
		return
	}

	resp, err := service.RemoteICAPReqmod(riSvc)

	if err != nil {
		errorLogger.LogfToFile("Failed to make REQMOD call to remote icap server: %s\n", err.Error())
		w.WriteHeader(utils.IfPropagateError(http.StatusFailedDependency, http.StatusNoContent), nil, false)
		return
	}

	utils.CopyHeaders(resp.Header, w.Header(), "")

	infoLogger.LogfToFile("Received response from the remote ICAP server with status code: %d...\n", resp.StatusCode)

	if resp.StatusCode == http.StatusOK {

		if resp.ContentResponse != nil && resp.ContentResponse.Body != nil {

			bdyByte, err := ioutil.ReadAll(resp.ContentResponse.Body)
			if err != nil && err != io.ErrUnexpectedEOF {
				errorLogger.LogToFile("Failed to read body from the remote icap response: ", err.Error())
				w.WriteHeader(utils.IfPropagateError(http.StatusInternalServerError, http.StatusNoContent), nil, false)
				return
			}

			defer resp.ContentResponse.Body.Close()

			w.WriteHeader(resp.StatusCode, resp.ContentResponse, true)

			w.Write(bdyByte)
			return
		}
	}

	w.WriteHeader(resp.StatusCode, nil, false)

}

func prepareRemoteSvc(reqHeaders map[string][]string, url string, timeout time.Duration) service.RemoteICAPService {
	svc := service.RemoteICAPService{
		URL:           url,
		Timeout:       timeout,
		RequestHeader: http.Header{},
	}
	utils.CopyHeaders(reqHeaders, svc.RequestHeader, "")
	return svc
}
