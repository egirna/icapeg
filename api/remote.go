package api

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"icapeg/icap"
	"icapeg/logger"
	"icapeg/service"
	"icapeg/utils"

	zLog "github.com/rs/zerolog/log"
)

func doRemoteOPTIONS(req *icap.Request, w icap.ResponseWriter, vendor, shadowVendor, mode string, logger *logger.ZLogger) {

	riSvc := service.GetICAPService(vendor)
	// riSvc.SetHeader(req.Header)

	if shadowVendor != utils.NoVendor && strings.HasPrefix(shadowVendor, utils.ICAPPrefix) {
		siSvc := service.GetICAPService(shadowVendor)
		siSvc.SetHeader(req.Header)
		updateEmptyOptionsEndpoint(siSvc, mode)
		go doShadowOPTIONS(siSvc, logger)
	}

	updateEmptyOptionsEndpoint(riSvc, mode)
	elapsed := time.Since(logger.LogStartTime)
	zLog.Info().Dur("duration", elapsed).Str("value", "passing request to the remote ICAP server").Msgf("send_request_to_remote_ICAP")
	resp, err := riSvc.DoOptions()

	if err != nil {
		zLog.Logger.Error().Msgf("Failed to make OPTIONS call of remote icap server: %s", err.Error())
		w.WriteHeader(utils.IfPropagateError(http.StatusFailedDependency, http.StatusNoContent), nil, false)
		return
	}
	elapsed = time.Since(logger.LogStartTime)
	zLog.Info().Dur("duration", elapsed).Str("value", fmt.Sprintf("Received response from the remote ICAP server wwith status code: %d", resp.StatusCode)).Msgf("receive_response_from_remote_ICAP")
	utils.CopyHeaders(resp.Header, w.Header(), utils.HeaderEncapsulated)
	w.WriteHeader(resp.StatusCode, nil, false)
}

func doRemoteRESPMOD(req *icap.Request, w icap.ResponseWriter, vendor, shadowVendor string, logger *logger.ZLogger) {

	riSvc := service.GetICAPService(vendor)
	riSvc.SetHeader(req.Header)

	if shadowVendor != utils.NoVendor && strings.HasPrefix(shadowVendor, utils.ICAPPrefix) {
		siSvc := service.GetICAPService(shadowVendor)
		siSvc.SetHeader(req.Header)
		shadowHTTPResp := utils.GetHTTPResponseCopy(req.Response)
		go doShadowRESPMOD(siSvc, *req.Request, shadowHTTPResp, logger)
	}

	riSvc.SetHTTPRequest(req.Request)
	riSvc.SetHTTPResponse(req.Response)

	if req.Request.URL.Scheme == "" {
		elapsed := time.Since(logger.LogStartTime)
		zLog.Debug().Dur("duration", elapsed).Str("value", "scheme not found, changing the url").Msgf("scheme_not_found")
		req.Request.URL = utils.GetNewURL(req.Request)
	}

	b, err := ioutil.ReadAll(req.Response.Body)

	if err != nil {
		elapsed := time.Since(logger.LogStartTime)
		zLog.Error().Dur("duration", elapsed).Err(err).Str("value", fmt.Sprintf("Error reading the body: %s", err.Error())).Msgf("read_response_body_error")
	}

	bdyStr := string(b)
	if len(b) > int(req.Response.ContentLength) {
		if strings.HasSuffix(bdyStr, "\n\n") {
			bdyStr = strings.TrimSuffix(bdyStr, "\n\n")
		}
	}

	req.Response.Body = ioutil.NopCloser(bytes.NewBuffer([]byte(bdyStr)))
	elapsed := time.Since(logger.LogStartTime)
	zLog.Info().Dur("duration", elapsed).Str("value", "passing request to the remote ICAP server").Msgf("send_request_to_remote_ICAP")
	resp, err := riSvc.DoRespmod()

	if err != nil {
		zLog.Logger.Error().Msgf("Failed to make RESPMOD call to remote icap server: %s", err.Error())
		w.WriteHeader(utils.IfPropagateError(http.StatusFailedDependency, http.StatusNoContent), nil, false)
		return
	}

	utils.CopyHeaders(resp.Header, w.Header(), "")
	elapsed = time.Since(logger.LogStartTime)
	zLog.Info().Dur("duration", elapsed).Str("value", fmt.Sprintf("Received response from the remote ICAP server wwith status code: %d", resp.StatusCode)).Msgf("receive_response_from_remote_ICAP")

	if resp.StatusCode == http.StatusOK { // NOTE: this is done to render the error html page, not sure this is the proper way

		if resp.ContentResponse != nil && resp.ContentResponse.Body != nil {

			bdyByte, err := ioutil.ReadAll(resp.ContentResponse.Body)
			if err != nil && err != io.ErrUnexpectedEOF {
				elapsed = time.Since(logger.LogStartTime)
				zLog.Error().Dur("duration", elapsed).Err(err).Str("value", fmt.Sprintf("Failed to read body from the remote icap response: %s", err.Error())).Msgf("read_remote_icap_response_body_error")
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

func doRemoteREQMOD(req *icap.Request, w icap.ResponseWriter, vendor, shadowVendor string, zLog *logger.ZLogger) {

	riSvc := service.GetICAPService(vendor)
	riSvc.SetHeader(req.Header)

	if shadowVendor != utils.NoVendor && strings.HasPrefix(shadowVendor, utils.ICAPPrefix) {
		siSvc := service.GetICAPService(shadowVendor)
		siSvc.SetHeader(req.Header)
		go doShadowREQMOD(siSvc, *req.Request, zLog)
	}

	riSvc.SetHTTPRequest(req.Request)

	if req.Request.URL.Scheme == "" {
		zLog.Logger.Debug().Msg("Scheme not found, changing the url")
		req.Request.URL = utils.GetNewURL(req.Request)
	}

	ext := utils.GetFileExtension(req.Request)

	if ext == "" {
		zLog.Logger.Debug().Msg("Processing not required...")
		w.WriteHeader(http.StatusNoContent, nil, false)
		return
	}

	zLog.Logger.Info().Msg("Passing request to the remote ICAP server...")
	resp, err := riSvc.DoReqmod()

	if err != nil {
		zLog.Logger.Error().Msgf("Failed to make REQMOD call to remote icap server: %s", err.Error())
		w.WriteHeader(utils.IfPropagateError(http.StatusFailedDependency, http.StatusNoContent), nil, false)
		return
	}

	utils.CopyHeaders(resp.Header, w.Header(), "")

	zLog.Logger.Info().Msgf("Received response from the remote ICAP server with status code: %d...", resp.StatusCode)

	if resp.StatusCode == http.StatusOK {

		if resp.ContentResponse != nil && resp.ContentResponse.Body != nil {

			bdyByte, err := ioutil.ReadAll(resp.ContentResponse.Body)
			if err != nil && err != io.ErrUnexpectedEOF {
				zLog.Logger.Error().Msgf("Failed to read body from the remote icap response: %s", err.Error())
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
