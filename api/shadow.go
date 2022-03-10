package api

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"icapeg/config"
	"icapeg/dtos"
	"icapeg/logger"
	"icapeg/service"
	"icapeg/utils"

	zLog "github.com/rs/zerolog/log"
)

func doShadowOPTIONS(svc service.ICAPService, logger *logger.ZLogger) {

	elapsed := time.Since(logger.LogStartTime)
	zLog.Info().Dur("duration", elapsed).Str("value", "Passing request to the shadow ICAP server...").Msgf("request_to_shadow_icap_server")
	resp, err := svc.DoOptions()
	if err != nil {
		elapsed = time.Since(logger.LogStartTime)
		zLog.Error().Dur("duration", elapsed).Err(err).Str("value", "Failed to make OPTIONS call of shadow icap server").Msgf("error_making_options_call")
		return
	}

	elapsed = time.Since(logger.LogStartTime)
	zLog.Info().Dur("duration", elapsed).
		Str("value", fmt.Sprintf("Received response from the shadow ICAP server with status Code: %d ", resp.StatusCode)).
		Msgf("response_from_shadow_icap_server")

	respHeaderString := ""
	for header, values := range resp.Header {
		respHeaderString += fmt.Sprintf("%s: %v ", header, values)
	}
	elapsed = time.Since(logger.LogStartTime)
	zLog.Info().Dur("duration", elapsed).
		Str("value", fmt.Sprintf("headers of response %s", respHeaderString)).
		Msgf("header_of_response")
}

func doShadowRESPMOD(svc service.ICAPService, httpReq http.Request, httpResp http.Response, logger *logger.ZLogger) {
	svc.SetHTTPRequest(&httpReq)
	svc.SetHTTPResponse(&httpResp)

	if httpReq.URL.Scheme == "" {
		elapsed := time.Since(logger.LogStartTime)
		zLog.Debug().Dur("duration", elapsed).Str("value", "Scheme not found, changing the url").Msgf("scheme_not_found_shadow_respmode")
		httpReq.URL = utils.GetNewURL(&httpReq)
	}

	b, err := ioutil.ReadAll(httpResp.Body)

	if err != nil {
		elapsed := time.Since(logger.LogStartTime)
		zLog.Error().Dur("duration", elapsed).Err(err).Str("value", "Error reading the response body from shadow server").Msgf("error_reading_body_from_shadow_server")
	}

	bdyStr := string(b)
	if len(b) > int(httpResp.ContentLength) {
		if strings.HasSuffix(bdyStr, "\n\n") {
			bdyStr = strings.TrimSuffix(bdyStr, "\n\n")
		}
	}

	httpResp.Body = ioutil.NopCloser(bytes.NewBuffer([]byte(bdyStr)))

	elapsed := time.Since(logger.LogStartTime)
	zLog.Info().Dur("duration", elapsed).Str("value", "Passing request to the RESPMOD shadow ICAP server...").Msgf("request_to_shadow_icap_server")

	resp, err := svc.DoRespmod()

	if err != nil {
		elapsed = time.Since(logger.LogStartTime)
		zLog.Error().Dur("duration", elapsed).Err(err).Str("value", "Failed to make RESPMOD call of shadow icap server").Msgf("error_making_shadow_respmode_call")
		return
	}

	elapsed = time.Since(logger.LogStartTime)
	zLog.Info().Dur("duration", elapsed).
		Str("value", fmt.Sprintf("Received response from the shadow ICAP server with status Code: %d ", resp.StatusCode)).
		Msgf("response_from_shadow_icap_server")

	respHeaderString := ""
	for header, values := range resp.Header {
		respHeaderString += fmt.Sprintf("%s: %v ", header, values)
	}
	elapsed = time.Since(logger.LogStartTime)
	zLog.Info().Dur("duration", elapsed).
		Str("value", fmt.Sprintf("headers of response %s", respHeaderString)).
		Msgf("header_of_response")
	ContentResponseHeaderString := ""
	if resp.ContentResponse != nil {
		for header, values := range resp.ContentResponse.Header {
			ContentResponseHeaderString += fmt.Sprintf("%s: %v ", header, values)
		}
		elapsed = time.Since(logger.LogStartTime)
		zLog.Info().Dur("duration", elapsed).
			Str("value", fmt.Sprintf("headers of content response %s", ContentResponseHeaderString)).
			Msgf("header_of_content_response")
	}
}

func doShadowREQMOD(svc service.ICAPService, httpReq http.Request, logger *logger.ZLogger) {
	svc.SetHTTPRequest(&httpReq)

	if httpReq.URL.Scheme == "" {
		elapsed := time.Since(logger.LogStartTime)
		zLog.Debug().Dur("duration", elapsed).Str("value", "Scheme not found, changing the url").Msgf("scheme_not_found_shadow_reqmode")
		httpReq.URL = utils.GetNewURL(&httpReq)
	}

	ext := utils.GetFileExtension(&httpReq)

	if ext == "" {
		elapsed := time.Since(logger.LogStartTime)
		zLog.Debug().Dur("duration", elapsed).Str("value", "Processing not required").Msgf("file_processing_not_required")
		return
	}

	elapsed := time.Since(logger.LogStartTime)
	zLog.Info().Dur("duration", elapsed).Str("value", "Passing request to the RESPMODE shadow ICAP server...").Msgf("request_to_shadow_icap_server")
	resp, err := svc.DoReqmod()

	if err != nil {
		elapsed = time.Since(logger.LogStartTime)
		zLog.Error().Dur("duration", elapsed).Err(err).Str("value", "Failed to make REQMOD call of shadow icap server").Msgf("error_making_shadow_reqmode_call")
		return
	}

	elapsed = time.Since(logger.LogStartTime)
	zLog.Info().Dur("duration", elapsed).
		Str("value", fmt.Sprintf("Received response from the shadow ICAP server with status Code: %d ", resp.StatusCode)).
		Msgf("response_from_shadow_icap_server")

	respHeaderString := ""
	for header, values := range resp.Header {
		respHeaderString += fmt.Sprintf("%s: %v ", header, values)
	}
	elapsed = time.Since(logger.LogStartTime)
	zLog.Info().Dur("duration", elapsed).
		Str("value", fmt.Sprintf("headers of response %s", respHeaderString)).
		Msgf("header_of_response")
	ContentResponseHeaderString := ""
	if resp.ContentResponse != nil {
		for header, values := range resp.ContentResponse.Header {
			ContentResponseHeaderString += fmt.Sprintf("%s: %v ", header, values)
		}
		elapsed = time.Since(logger.LogStartTime)
		zLog.Info().Dur("duration", elapsed).
			Str("value", fmt.Sprintf("headers of content response %s", ContentResponseHeaderString)).
			Msgf("header_of_content_response")
	}
}

func doShadowScan(vendor string, serviceName string, filename string, fmi dtos.FileMetaInfo, buf *bytes.Buffer, fileURL string, logger *logger.ZLogger) {

	newBuf := utils.CopyBuffer(buf)

	scannerName := ""
	if buf == nil && fileURL != "" {
		scannerName = config.App().ReqScannerVendorShadow
	}
	if buf != nil && fileURL == "" {
		scannerName = config.App().RespScannerVendorShadow
	}

	var sts int
	var si *dtos.SampleInfo

	localService := service.IsServiceLocal(vendor, scannerName, logger)

	if localService && buf != nil { // if the scanner is installed locally
		sts, si = doLocalScan(scannerName, serviceName, fmi, newBuf, logger)
	}

	if !localService { // if the scanner is an external service requiring API calls.

		if buf == nil && fileURL != "" { // indicates this is a URL scan request
			sts, si = doRemoteURLScan(scannerName, serviceName, filename, fmi, fileURL, logger)
		}

		if buf != nil && fileURL == "" { // indicates this is a File scan request
			sts, si = doRemoteFileScan(scannerName, serviceName, filename, fmi, newBuf, logger)
		}

	}
	elapsed := time.Since(logger.LogStartTime)
	zLog.Info().Dur("duration", elapsed).
		Str("value", fmt.Sprintf("Response Status from the shadow scanner(%s): %d", scannerName, sts)).
		Msgf("shadow_scanner_response")

	zLog.Logger.Info().Msgf("Response Status from the shadow scanner(%s): %d", scannerName, sts)
	elapsed = time.Since(logger.LogStartTime)
	zLog.Info().Dur("duration", elapsed).Str("value", fmt.Sprintf("The file %s is good to go", filename)).Msgf("good_to_go")
	if sts == http.StatusOK {
		elapsed = time.Since(logger.LogStartTime)
		logString := fmt.Sprintf("File Name: %s\nFile Type: %s\nFile Size: %s\nRequested URL: %s\nSeverity: %s\nPositive Score: %s\nResults By: %s",
			si.FileName, si.SampleType, si.FileSizeStr, utils.BreakHTTPURL(fileURL), si.SampleSeverity, si.VTIScore, scannerName)
		zLog.Info().Dur("duration", elapsed).
			Str("value", fmt.Sprintf("content %s", logString)).
			Msgf("status_ok_information")
	}
}
