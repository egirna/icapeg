package api

import (
	"bytes"
	"icapeg/service"
	"icapeg/utils"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/egirna/icap"
)

func doRemoteOPTIONS(req *icap.Request, w icap.ResponseWriter, vendor, shadowVendor, mode string) {

	riSvc := service.GetICAPService(vendor)
	// riSvc.SetHeader(req.Header)

	if shadowVendor != utils.NoVendor && strings.HasPrefix(shadowVendor, utils.ICAPPrefix) {
		siSvc := service.GetICAPService(shadowVendor)
		siSvc.SetHeader(req.Header)
		updateEmptyOptionsEndpoint(siSvc, mode)
		go doShadowOPTIONS(siSvc)
	}

	updateEmptyOptionsEndpoint(riSvc, mode)

	infoLogger.LogToFile("Passing request to the remote ICAP server...")
	resp, err := riSvc.DoOptions()

	if err != nil {
		errorLogger.LogfToFile("Failed to make OPTIONS call of remote icap server: %s\n", err.Error())
		w.WriteHeader(utils.IfPropagateError(http.StatusFailedDependency, http.StatusNoContent), nil, false)
		return
	}

	infoLogger.LogfToFile("Received response from the remote ICAP server wwith status code: %d...\n", resp.StatusCode)

	utils.CopyHeaders(resp.Header, w.Header(), utils.HeaderEncapsulated)

	w.WriteHeader(resp.StatusCode, nil, false)

}

func doRemoteRESPMOD(req *icap.Request, w icap.ResponseWriter, vendor, shadowVendor string) {

	riSvc := service.GetICAPService(vendor)
	riSvc.SetHeader(req.Header)

	if shadowVendor != utils.NoVendor && strings.HasPrefix(shadowVendor, utils.ICAPPrefix) {
		siSvc := service.GetICAPService(shadowVendor)
		siSvc.SetHeader(req.Header)
		shadowHTTPResp := utils.GetHTTPResponseCopy(req.Response)
		go doShadowRESPMOD(siSvc, *req.Request, shadowHTTPResp)
	}

	riSvc.SetHTTPRequest(req.Request)
	riSvc.SetHTTPResponse(req.Response)

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

	infoLogger.LogToFile("Passing request to the remote ICAP server...")
	resp, err := riSvc.DoRespmod()

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

func doRemoteREQMOD(req *icap.Request, w icap.ResponseWriter, vendor, shadowVendor string) {

	riSvc := service.GetICAPService(vendor)
	riSvc.SetHeader(req.Header)

	if shadowVendor != utils.NoVendor && strings.HasPrefix(shadowVendor, utils.ICAPPrefix) {
		siSvc := service.GetICAPService(shadowVendor)
		siSvc.SetHeader(req.Header)
		go doShadowREQMOD(siSvc, *req.Request)
	}

	riSvc.SetHTTPRequest(req.Request)

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

	infoLogger.LogToFile("Passing request to the remote ICAP server...")
	resp, err := riSvc.DoReqmod()

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

func updateEmptyOptionsEndpoint(svc service.ICAPService, mode string) {
	if svc.GetOptionsEndpoint() == "" {
		if mode == utils.ICAPModeResp {
			svc.ChangeOptionsEndpoint(svc.GetRespmodEndpoint())
		}
		if mode == utils.ICAPModeReq {
			svc.ChangeOptionsEndpoint(svc.GetReqmodEndpoint())
		}
	}
}
