package api

import (
	"bytes"
	zLog "github.com/rs/zerolog/log"
	"icapeg/config"
	"icapeg/icap"
	"icapeg/logger"
	"icapeg/readValues"
	"icapeg/service"
	"icapeg/utils"
	"io"
	"io/ioutil"
	"net/http"
	"net/textproto"
	"strconv"
	"strings"
	"time"
)

// isServiceExists is a func which get the array of service which is the config file
// and return true if the service which included in the request exists, rather than that it returns false
func isServiceExists(path string) bool {
	services := readValues.ReadValuesSlice("app.services")
	for i := 0; i < len(services); i++ {
		if path == services[i] {
			return true
		}
	}
	return false
}

//adding headers to the logs
func addHeadersToLogs(headers textproto.MIMEHeader, elapsed time.Duration) {
	for key, element := range headers {
		res := key + " : "
		for i := 0; i < len(element); i++ {
			res += element[i]
			if i != len(element)-1 {
				res += ", "
			}
		}
		zLog.Debug().Dur("duration", elapsed).Str("value", "ICAP request header").
			Msgf(res)
	}
}

func getMethodName(methodName string) string {
	if methodName == "REQMOD" {
		methodName = "req_mode"
	} else if methodName == "RESPMOD" {
		methodName = "resp_mode"
	}
	return methodName
}

func is204Allowed(headers textproto.MIMEHeader) bool {
	Is204Allowed := false
	if _, exist := headers["Allow"]; exist &&
		headers.Get("Allow") == strconv.Itoa(utils.NoModificationStatusCodeStr) {
		Is204Allowed = true
	}
	return Is204Allowed
}

func getVendorName(serviceName string) string {
	vendor := serviceName + ".vendor"
	vendor = readValues.ReadValuesString(vendor)
	return vendor
}

func getEnabledMethods(serviceName string) string {
	var allMethods []string
	if readValues.ReadValuesBool(serviceName + ".resp_mode") {
		allMethods = append(allMethods, "RESPMOD")
	}
	if readValues.ReadValuesBool(serviceName + ".req_mode") {
		allMethods = append(allMethods, "REQMOD")
	}
	if len(allMethods) == 1 {
		return allMethods[0]
	}
	return allMethods[0] + ", " + allMethods[1]
}

func shadowService(elapsed time.Duration, Is204Allowed bool, req *icap.Request,
	w icap.ResponseWriter, zlogger *logger.ZLogger) {
	zLog.Debug().Dur("duration", elapsed).Str("value", "processing not required for this request").
		Msgf("shadow_service_is_enabled")
	if Is204Allowed { // following RFC3507, if the request has Allow: 204 header, it is to be checked and if it doesn't exists, return the request as it is to the ICAP client, https://tools.ietf.org/html/rfc3507#section-4.6
		elapsed = time.Since(zlogger.LogStartTime)
		zLog.Debug().Dur("duration", elapsed).Str("value", "the file won't be modified").
			Msgf("request_received_on_icap_with_header_204")
		w.WriteHeader(utils.NoModificationStatusCodeStr, nil, false)
	} else {
		if req.Method == "REQMOD" {
			w.WriteHeader(utils.OkStatusCodeStr, req.Request, true)
			tempBody, _ := ioutil.ReadAll(req.Request.Body)
			w.Write(tempBody)
			req.Request.Body = io.NopCloser(bytes.NewBuffer(tempBody))
		} else if req.Method == "RESPMOD" {
			w.WriteHeader(utils.OkStatusCodeStr, req.Response, true)
			tempBody, _ := ioutil.ReadAll(req.Response.Body)
			w.Write(tempBody)
			req.Response.Body = io.NopCloser(bytes.NewBuffer(tempBody))
		}
	}
}

/* If any remote icap is enabled, the work flow is controlled by the remote icap */
func optionsModeRemote(vendor string, req *icap.Request, w icap.ResponseWriter, appCfg *config.AppConfig, zlogger *logger.ZLogger) {
	if strings.HasPrefix(vendor, utils.ICAPPrefix) {
		doRemoteOPTIONS(req, w, vendor, appCfg.RespScannerVendorShadow, utils.ICAPModeResp, zlogger)
		return
	} else if strings.HasPrefix(appCfg.RespScannerVendorShadow, utils.ICAPPrefix) { // if the shadow wants to run independently
		siSvc := service.GetICAPService(appCfg.RespScannerVendorShadow)
		siSvc.SetHeader(req.Header)
		updateEmptyOptionsEndpoint(siSvc, utils.ICAPModeResp)
		go doShadowOPTIONS(siSvc, zlogger)
	}
}

func optionsMode(headers http.Header, serviceName string, appCfg *config.AppConfig, vendor string, req *icap.Request,
	w icap.ResponseWriter, zlogger *logger.ZLogger) {
	//optionsModeRemote(vendor, req, w, appCfg, zlogger)
	headers.Set("Methods", getEnabledMethods(serviceName))
	headers.Set("Allow", "204")
	// Add preview if preview_enabled is true in config
	if appCfg.PreviewEnabled == true {
		if pb, _ := strconv.Atoi(appCfg.PreviewBytes); pb >= 0 {
			headers.Set("Preview", appCfg.PreviewBytes)
		}
	}
	headers.Set("Transfer-Preview", utils.Any)
	w.WriteHeader(http.StatusOK, nil, false)
}
