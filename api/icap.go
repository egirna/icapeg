package api

import (
	_ "bytes"
	_ "compress/gzip"
	"fmt"
	_ "html/template"
	"icapeg/service"
	_ "io"
	_ "io/ioutil"
	"net/http"
	_ "strconv"
	_ "strings"
	"time"

	_ "icapeg/api/ContentTypes"
	"icapeg/config"
	"icapeg/icap"
	"icapeg/logger"
	"icapeg/readValues"
	_ "icapeg/service"
	"icapeg/utils"

	zLog "github.com/rs/zerolog/log"
)

// ToICAPEGServe is the ICAP Request Handler for all modes and services:
func ToICAPEGServe(w icap.ResponseWriter, req *icap.Request, zlogger *logger.ZLogger) {

	// setting up logging for each request
	elapsed := time.Since(zlogger.LogStartTime)

	//adding headers to the log
	addHeadersToLogs(req.Header, elapsed)

	// checking if the service doesn't exist in toml file
	// if it does not exist, the response will be 404 ICAP Service Not Found
	serviceName := req.URL.Path[1:len(req.URL.Path)]
	if !isServiceExists(serviceName) {
		w.WriteHeader(http.StatusNotFound, nil, false)
		return
	}

	// checking if request method is allowed or not
	methodName := getMethodName(req.Method)
	if !isMethodAllowed(serviceName, methodName, elapsed) {
		w.WriteHeader(http.StatusMethodNotAllowed, nil, false)
		return
	}
	methodName = req.Method

	vendor := getVendorName(serviceName)

	h := w.Header()
	addingISTAGServiceHeaders(h, serviceName, methodName, req.RawURL, elapsed)

	appCfg := config.App()

	ct := utils.GetMimeExtension(req.Preview)

	bypassExts := readValues.ReadValuesSlice(serviceName + ".bypass_extensions")
	if utils.InStringSlice(ct, bypassExts) {
		elapsed = time.Since(zlogger.LogStartTime)
		zLog.Debug().Dur("duration", elapsed).Str("value", fmt.Sprintf("processing not required for file type- %s", ct)).Msgf("belongs_bypassable_extensions")
		w.WriteHeader(http.StatusNoContent, nil, false)
		return
	}
	processExts := readValues.ReadValuesSlice(serviceName + ".process_extensions")
	if utils.InStringSlice(utils.Any, bypassExts) && !utils.InStringSlice(ct, processExts) {
		// if extension does not belong to "All bypassable except the processable ones" group
		elapsed = time.Since(zlogger.LogStartTime)
		zLog.Debug().Dur("duration", elapsed).Str("value", fmt.Sprintf("processing not required for file type- %s", ct)).Msgf("dont_belong_to_processable_extensions")
		w.WriteHeader(http.StatusNoContent, nil, false)
		return
	}
	Is204Allowed := is204Allowed(req.Header)

	isShadowServiceEnabled := readValues.ReadValuesBool(serviceName + ".shadow_service")

	if isShadowServiceEnabled && methodName != "OPTIONS" {
		shadowService(elapsed, Is204Allowed, req, w, zlogger)
		go func() {
			time.Sleep(10 * time.Second)
			fmt.Println("finish")
		}()
		return
	}

	switch methodName {
	case utils.ICAPModeOptions:
		optionsMode(h, serviceName, appCfg, vendor, req, w, zlogger)

	case utils.ICAPModeResp:
		defer req.Response.Body.Close()
		gw := service.GetService("glasswall", "gw_rebuild", w, req.Request, req.Response, elapsed, Is204Allowed, methodName, zlogger)
		gw.RespMode(req.Request, req.Response)
		break

	case utils.ICAPModeReq:
	}
}
