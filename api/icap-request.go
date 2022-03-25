package api

import (
	"bytes"
	"errors"
	"fmt"
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
	"strconv"
	"time"
)

type ICAPRequest struct {
	w                      icap.ResponseWriter
	req                    *icap.Request
	h                      http.Header
	Is204Allowed           bool
	isShadowServiceEnabled bool
	appCfg                 *config.AppConfig
	logger                 *logger.ZLogger
	elapsed                time.Duration
	serviceName            string
	methodName             string
	vendor                 string
}

func NewICAPRequest(w icap.ResponseWriter, req *icap.Request, logger *logger.ZLogger) *ICAPRequest {
	ICAPRequest := &ICAPRequest{
		w:       w,
		req:     req,
		h:       w.Header(),
		logger:  logger,
		elapsed: time.Since(logger.LogStartTime),
	}
	return ICAPRequest
}

func (i *ICAPRequest) RequestInitialization() error {

	//adding headers to the log
	i.addHeadersToLogs()

	// checking if the service doesn't exist in toml file
	// if it does not exist, the response will be 404 ICAP Service Not Found
	i.serviceName = i.req.URL.Path[1:len(i.req.URL.Path)]
	if !i.isServiceExists() {
		i.w.WriteHeader(http.StatusNotFound, nil, false)
		err := errors.New("service doesn't exist")
		return err
	}

	// checking if request method is allowed or not
	i.methodName = i.req.Method
	i.methodName = i.getMethodName()
	if !i.isMethodAllowed() {
		i.w.WriteHeader(http.StatusMethodNotAllowed, nil, false)
		err := errors.New("method is not allowed")
		return err
	}
	i.methodName = i.req.Method

	i.vendor = i.getVendorName()

	i.addingISTAGServiceHeaders()

	ct := utils.GetMimeExtension(i.req.Preview)

	bypassExts := readValues.ReadValuesSlice(i.serviceName + ".bypass_extensions")
	if utils.InStringSlice(ct, bypassExts) {
		i.elapsed = time.Since(i.logger.LogStartTime)
		zLog.Debug().Dur("duration", i.elapsed).Str("value", fmt.Sprintf("processing not required for file type- %s", ct)).Msgf("belongs_bypassable_extensions")
		i.w.WriteHeader(http.StatusNoContent, nil, false)
		return errors.New("processing not required for file type")
	}
	processExts := readValues.ReadValuesSlice(i.serviceName + ".process_extensions")
	if utils.InStringSlice(utils.Any, bypassExts) && !utils.InStringSlice(ct, processExts) {
		// if extension does not belong to "All bypassable except the processable ones" group
		i.elapsed = time.Since(i.logger.LogStartTime)
		zLog.Debug().Dur("duration", i.elapsed).Str("value", fmt.Sprintf("processing not required for file type- %s", ct)).Msgf("dont_belong_to_processable_extensions")
		i.w.WriteHeader(http.StatusNoContent, nil, false)
		return errors.New("processing not required for file type")
	}

	i.Is204Allowed = i.is204Allowed()

	i.isShadowServiceEnabled = readValues.ReadValuesBool(i.serviceName + ".shadow_service")

	i.appCfg = config.App()

	if i.isShadowServiceEnabled && i.methodName != "OPTIONS" {
		i.shadowService()
		go i.RequestProcessing()
		return errors.New("shadow service")
	}

	return nil
}

func (i *ICAPRequest) RequestProcessing() {
	switch i.methodName {
	case utils.ICAPModeOptions:
		i.optionsMode()
		break

	case utils.ICAPModeResp:
		defer i.req.Response.Body.Close()
		if i.req.Request == nil {
			i.req.Request = &http.Request{}
		}
		gw := service.GetService(i.vendor, i.serviceName, i.methodName, i.req.Request, i.req.Response, i.elapsed, i.logger)
		IcapStatusCode, file, httpResponse, serviceHeaders := gw.Processing()
		if serviceHeaders != nil {
			for key, value := range serviceHeaders {
				i.w.Header().Set(key, value)
			}
		}
		if i.isShadowServiceEnabled {
			//add logs here
			return
		}
		switch IcapStatusCode {
		case utils.InternalServerErrStatusCodeStr:
			i.w.WriteHeader(IcapStatusCode, nil, false)
			break
		case utils.NoModificationStatusCodeStr:
			if i.Is204Allowed {
				i.w.WriteHeader(utils.NoModificationStatusCodeStr, nil, false)
			} else {
				i.w.WriteHeader(utils.OkStatusCodeStr, i.req.Response, true)
				i.w.Write(file)
			}
		case utils.OkStatusCodeStr:
			i.w.WriteHeader(utils.OkStatusCodeStr, httpResponse, true)
			i.w.Write(file)
		}
	case utils.ICAPModeReq:
	}

}

//adding headers to the log
func (i *ICAPRequest) addHeadersToLogs() {
	for key, element := range i.req.Header {
		res := key + " : "
		for i := 0; i < len(element); i++ {
			res += element[i]
			if i != len(element)-1 {
				res += ", "
			}
		}
		zLog.Debug().Dur("duration", i.elapsed).Str("value", "ICAP request header").
			Msgf(res)
	}
}

func (i *ICAPRequest) isServiceExists() bool {
	services := readValues.ReadValuesSlice("app.services")
	for r := 0; r < len(services); r++ {
		if i.serviceName == services[r] {
			return true
		}
	}
	return false

}

func (i *ICAPRequest) getMethodName() string {
	if i.methodName == "REQMOD" {
		i.methodName = "req_mode"
	} else if i.methodName == "RESPMOD" {
		i.methodName = "resp_mode"
	}
	return i.methodName
}

func (i *ICAPRequest) isMethodAllowed() bool {
	if i.methodName != "OPTIONS" {
		isMethodEnabled := readValues.ReadValuesBool(i.serviceName + "." + i.methodName)
		if !isMethodEnabled {
			zLog.Debug().Dur("duration", i.elapsed).Str("value", i.methodName+" is not enabled").
				Msgf("this_method_is_not_enabled_in_GO_ICAP_configuration")
			return false
		}
	}
	return true
}

func (i *ICAPRequest) getVendorName() string {
	vendor := i.serviceName + ".vendor"
	vendor = readValues.ReadValuesString(vendor)
	return vendor
}

func (i *ICAPRequest) addingISTAGServiceHeaders() {
	i.h.Set("ISTag", readValues.ReadValuesString(i.serviceName+".service_tag"))
	i.h.Set("Service", readValues.ReadValuesString(i.serviceName+".service_caption"))
	zLog.Info().Dur("duration", i.elapsed).Str("value", fmt.Sprintf("with method:%s url:%s", i.methodName, i.req.RawURL)).
		Msgf("request_received_on_icap")
}

func (i *ICAPRequest) is204Allowed() bool {
	Is204Allowed := false
	if _, exist := i.req.Header["Allow"]; exist &&
		i.req.Header.Get("Allow") == strconv.Itoa(utils.NoModificationStatusCodeStr) {
		Is204Allowed = true
	}
	return Is204Allowed
}

func (i *ICAPRequest) shadowService() {
	zLog.Debug().Dur("duration", i.elapsed).Str("value", "processing not required for this request").
		Msgf("shadow_service_is_enabled")
	if i.Is204Allowed { // following RFC3507, if the request has Allow: 204 header, it is to be checked and if it doesn't exists, return the request as it is to the ICAP client, https://tools.ietf.org/html/rfc3507#section-4.6
		i.elapsed = time.Since(i.logger.LogStartTime)
		zLog.Debug().Dur("duration", i.elapsed).Str("value", "the file won't be modified").
			Msgf("request_received_on_icap_with_header_204")
		i.w.WriteHeader(utils.NoModificationStatusCodeStr, nil, false)
	} else {
		if i.req.Method == "REQMOD" {
			i.w.WriteHeader(utils.OkStatusCodeStr, i.req.Request, true)
			tempBody, _ := ioutil.ReadAll(i.req.Request.Body)
			i.w.Write(tempBody)
			i.req.Request.Body = io.NopCloser(bytes.NewBuffer(tempBody))
		} else if i.req.Method == "RESPMOD" {
			i.w.WriteHeader(utils.OkStatusCodeStr, i.req.Response, true)
			tempBody, _ := ioutil.ReadAll(i.req.Response.Body)
			i.w.Write(tempBody)
			i.req.Response.Body = io.NopCloser(bytes.NewBuffer(tempBody))
		}
	}
}

func (i *ICAPRequest) getEnabledMethods() string {
	var allMethods []string
	if readValues.ReadValuesBool(i.serviceName + ".resp_mode") {
		allMethods = append(allMethods, "RESPMOD")
	}
	if readValues.ReadValuesBool(i.serviceName + ".req_mode") {
		allMethods = append(allMethods, "REQMOD")
	}
	if len(allMethods) == 1 {
		return allMethods[0]
	}
	return allMethods[0] + ", " + allMethods[1]
}

func (i *ICAPRequest) optionsMode() {
	//optionsModeRemote(vendor, req, w, appCfg, zlogger)
	i.h.Set("Methods", i.getEnabledMethods())
	i.h.Set("Allow", "204")
	// Add preview if preview_enabled is true in config
	if i.appCfg.PreviewEnabled == true {
		if pb, _ := strconv.Atoi(i.appCfg.PreviewBytes); pb >= 0 {
			i.h.Set("Preview", i.appCfg.PreviewBytes)
		}
	}
	i.h.Set("Transfer-Preview", utils.Any)
	i.w.WriteHeader(http.StatusOK, nil, false)
}
