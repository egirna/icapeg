package api

import (
	"bytes"
	"fmt"
	"icapeg/config"
	"icapeg/dtos"
	"icapeg/icap"
	"icapeg/logger"
	"icapeg/readValues"
	"icapeg/service"
	"icapeg/utils"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

var (
	infoLogger  = logger.NewLogger(logger.LogLevelInfo, logger.LogLevelDebug, logger.LogLevelError)
	debugLogger = logger.NewLogger(logger.LogLevelDebug)
	errorLogger = logger.NewLogger(logger.LogLevelError, logger.LogLevelDebug)
)

//isServiceExists is a func which get the array of service which is the config file
//and return true if the service which included in the request exists, rather than that it returns false
func isServiceExists(path string) bool {
	services := readValues.ReadValuesSlice("app.services")
	for i := 0; i < len(services); i++ {
		if path == services[i] {
			return true
		}
	}
	return false
}

// ToICAPEGServe is the ICAP Request Handler for all modes and services:
func ToICAPEGServe(w icap.ResponseWriter, req *icap.Request) {
	//checking if the service doesn't exist in toml file
	//if it does not exist, the response will be 404 ICAP Service Not Found
	serviceName := req.URL.Path[1:len(req.URL.Path)]
	if !isServiceExists(serviceName) {
		w.WriteHeader(http.StatusNotFound, nil, false)
		return
	}

	//this variable is used to check if the link(route) exists
	fullLink := readValues.ReadValuesString(serviceName+".base_url") +
		readValues.ReadValuesString(serviceName+".scan_endpoint")
	//making http request to check if it exists
	resp, err := http.Get(fullLink)
	if err != nil {
		print(err.Error())
	} else {
		if resp.StatusCode == 404 {
			w.WriteHeader(http.StatusNotFound, nil, false)
			return
		}
	}
	//checking if request method is allowed or not
	methodName := req.Method
	if methodName == "REQMOD" {
		methodName = "req_mode"
	} else if methodName == "RESPMOD" {
		methodName = "resp_mode"
	}
	if methodName != "OPTIONS" {
		isMethodEnabled := readValues.ReadValuesBool(serviceName + "." + methodName)
		if !isMethodEnabled {
			w.WriteHeader(http.StatusMethodNotAllowed, nil, false)
			return
		}
	}
	vendor := serviceName + ".vendor"
	vendor = readValues.ReadValuesString(vendor)
	h := w.Header()
	h.Set("ISTag", utils.ISTag)
	h.Set("Service", "Egirna ICAP-EG")
	infoLogger.LogfToFile("Request received---> METHOD:%s URL:%s\n", req.Method, req.RawURL)
	//println("Request received---> METHOD:%s URL:%s\n", req.Method, req.RawURL)
	appCfg := config.App()
	switch req.Method {
	case utils.ICAPModeOptions:

		/* If any remote icap is enabled, the work flow is controlled by the remote icap */
		if strings.HasPrefix(vendor, utils.ICAPPrefix) {
			doRemoteOPTIONS(req, w, vendor, appCfg.RespScannerVendorShadow, utils.ICAPModeResp)
			return
		} else if strings.HasPrefix(appCfg.RespScannerVendorShadow, utils.ICAPPrefix) { // if the shadow wants to run independently
			siSvc := service.GetICAPService(appCfg.RespScannerVendorShadow)
			siSvc.SetHeader(req.Header)
			updateEmptyOptionsEndpoint(siSvc, utils.ICAPModeResp)
			go doShadowOPTIONS(siSvc)
		}
		//getting Methods which enabled in the service
		var allMethods []string
		if readValues.ReadValuesBool(serviceName + ".resp_mode") {
			allMethods = append(allMethods, "RESPMOD")
		}
		if readValues.ReadValuesBool(serviceName + ".req_mode") {
			allMethods = append(allMethods, "REQMOD")
		}
		if len(allMethods) == 1 {
			h.Set("Methods", allMethods[0])
		} else {
			h.Set("Methods", allMethods[0]+", "+allMethods[1])
		}

		h.Set("Allow", "204")
		// Add preview if preview_enabled is true in config
		if appCfg.PreviewEnabled == true {
			if pb, _ := strconv.Atoi(appCfg.PreviewBytes); pb >= 0 {
				h.Set("Preview", appCfg.PreviewBytes)
			}
		}

		h.Set("Transfer-Preview", utils.Any)

		w.WriteHeader(http.StatusOK, nil, false)

	case utils.ICAPModeResp:
		defer req.Response.Body.Close()
		//misunderstanding of RFC, to be fixed later
		//if val, exist := req.Header["Allow"]; !exist || (len(val) > 0 && val[0] != utils.NoModificationStatusCodeStr) { // following RFC3507, if the request has Allow: 204 header, it is to be checked and if it doesn't exists, return the request as it is to the ICAP client, https://tools.ietf.org/html/rfc3507#section-4.6
		//	debugLogger.LogToFile("Processing not required for this request")
		//	w.WriteHeader(http.StatusNoContent, nil, false)
		//	return
		//}

		//change body to service name
		/* If any remote icap is enabled, the work flow is controlled by the remote icap */
		if strings.HasPrefix(vendor, utils.ICAPPrefix) {
			doRemoteRESPMOD(req, w, vendor, appCfg.RespScannerVendorShadow)
			return
		}

		/* If the shadow icap wants to run independently */
		if vendor == utils.NoVendor && strings.HasPrefix(appCfg.RespScannerVendorShadow, utils.ICAPPrefix) {
			siSvc := service.GetICAPService(appCfg.RespScannerVendorShadow)
			siSvc.SetHeader(req.Header)
			shadowHTTPResp := utils.GetHTTPResponseCopy(req.Response)
			go doShadowRESPMOD(siSvc, *req.Request, shadowHTTPResp)
			w.WriteHeader(http.StatusNoContent, nil, false)
			return
		}

		if vendor == utils.NoVendor && appCfg.RespScannerVendorShadow == utils.NoVendor { // if no scanner name provided, then bypass everything
			debugLogger.LogToFile("No respmod scanner provided...bypassing everything")
			w.WriteHeader(http.StatusNoContent, nil, false)
			return
		}

		// getting the content type to determine if the response is for a file, and if so, if its allowed to be processed
		// according to the configuration

		ct := utils.GetMimeExtension(req.Preview)
		processExts := appCfg.ProcessExtensions
		bypassExts := appCfg.BypassExtensions

		if utils.InStringSlice(ct, bypassExts) { // if the extension is bypassable
			debugLogger.LogToFile("Processing not required for file type-", ct)
			debugLogger.LogToFile("Reason: Belongs bypassable extensions")
			w.WriteHeader(http.StatusNoContent, nil, false)
			return
		}

		if utils.InStringSlice(utils.Any, bypassExts) && !utils.InStringSlice(ct, processExts) { // if extension does not belong to "All bypassable except the processable ones" group
			debugLogger.LogToFile("Processing not required for file type-", ct)
			debugLogger.LogToFile("Reason: Doesn't belong to processable extensions")
			w.WriteHeader(http.StatusNoContent, nil, false)
			return
		}

		// copying the file to a buffer for scanner vendor processing as the file is allowed according the co figuration

		buf := &bytes.Buffer{}

		if _, err := io.Copy(buf, req.Response.Body); err != nil {
			errorLogger.LogToFile("Failed to copy the response body to buffer: ", err.Error())
			w.WriteHeader(http.StatusNoContent, nil, false)
			return
		}

		if buf.Len() > appCfg.MaxFileSize {
			debugLogger.LogToFile("File size too large")
			w.WriteHeader(http.StatusNoContent, nil, false)
			return
		}
		// preparing the file meta informations
		filename := utils.GetFileName(req.Request)
		fileExt := utils.GetFileExtension(req.Request)
		fmi := dtos.FileMetaInfo{
			FileName: filename,
			FileType: fileExt,
			FileSize: float64(buf.Len()),
		}
		/* If the shadow virus scanner wants to run independently */
		if vendor == utils.NoVendor && appCfg.RespScannerVendorShadow != utils.NoVendor {
			go doShadowScan(vendor, serviceName, filename, fmi, buf, "")
			w.WriteHeader(http.StatusNoContent, nil, false)
			return
		}
		// Gw rebuild service req api , resp ICAP client
		if vendor == "glasswall" {

			filename = "test"
			resp, err := DoCDR(vendor, serviceName, buf, filename)
			if err != nil {
				fmt.Println(err)
				newResp := &http.Response{
					StatusCode: http.StatusForbidden,
					Status:     http.StatusText(http.StatusForbidden),
				}
				w.WriteHeader(http.StatusForbidden, newResp, true)

			} else {
				defer resp.Body.Close()
				bodybyte, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					fmt.Println(err)
				}

				newResp := req.Response
				newResp.Header.Set("Content-Length", strconv.Itoa(len(string(bodybyte))))
				w.WriteHeader(http.StatusOK, newResp, true)
				w.Write(bodybyte)

			}
			return

		}
		// echo servise
		if vendor == "echo" {
			bodybyte, err := ioutil.ReadAll(buf)
			if err != nil {
				fmt.Println(err)
			}

			newResp := &http.Response{
				StatusCode: http.StatusOK,
				Status:     http.StatusText(http.StatusOK),
				Header: http.Header{
					"Content-Length": []string{strconv.Itoa(len(string(bodybyte)))},
				},
			}
			w.WriteHeader(http.StatusOK, newResp, true)
			w.Write(bodybyte)

			return

		}
		status, sampleInfo := doScan(vendor, serviceName, filename, fmi, buf, "") // scan the file for any anomalies

		if status == http.StatusOK && sampleInfo != nil {
			infoLogger.LogfToFile("The file:%s is %s\n", filename, sampleInfo.SampleSeverity)
			htmlBuf, newResp := utils.GetTemplateBufferAndResponse(utils.BadFileTemplate, &dtos.TemplateData{
				FileName:     sampleInfo.FileName,
				FileType:     sampleInfo.SampleType,
				FileSizeStr:  sampleInfo.FileSizeStr,
				RequestedURL: utils.BreakHTTPURL(req.Request.RequestURI),
				Severity:     sampleInfo.SampleSeverity,
				Score:        sampleInfo.VTIScore,
				ResultsBy:    vendor,
			})
			w.WriteHeader(http.StatusOK, newResp, true)
			w.Write(htmlBuf.Bytes())
			return
		}

		if status == http.StatusNoContent {
			infoLogger.LogfToFile("The file %s is good to go\n", filename)
		}
		w.WriteHeader(status, nil, false) // \

	case utils.ICAPModeReq:
		defer req.Request.Body.Close()
		//misunderstanding of RFC, to be fixed later
		//if val, exist := req.Header["Allow"]; !exist || (len(val) > 0 && val[0] != utils.NoModificationStatusCodeStr) { // following RFC3507, if the request has Allow: 204 header, it is to be checked and if it doesn't exists, return the request as it is to the ICAP client, https://tools.ietf.org/html/rfc3507#section-4.6
		//	debugLogger.LogToFile("Processing not required for this request")
		//	w.WriteHeader(http.StatusNoContent, nil, false)
		//	return
		//}

		/* If any remote icap is enabled, the work flow is controlled by the remote icap */
		if strings.HasPrefix(vendor, utils.ICAPPrefix) {
			doRemoteRESPMOD(req, w, vendor, appCfg.RespScannerVendorShadow)
			return
		}

		/* If the shadow icap wants to run independently */
		if vendor == utils.NoVendor && strings.HasPrefix(appCfg.RespScannerVendorShadow, utils.ICAPPrefix) {
			siSvc := service.GetICAPService(appCfg.RespScannerVendorShadow)
			siSvc.SetHeader(req.Header)
			shadowHTTPResp := utils.GetHTTPResponseCopy(req.Response)
			go doShadowRESPMOD(siSvc, *req.Request, shadowHTTPResp)
			w.WriteHeader(http.StatusNoContent, nil, false)
			return
		}

		if vendor == utils.NoVendor && appCfg.RespScannerVendorShadow == utils.NoVendor { // if no scanner name provided, then bypass everything
			debugLogger.LogToFile("No respmod scanner provided...bypassing everything")
			w.WriteHeader(http.StatusNoContent, nil, false)
			return
		}

		// getting the content type to determine if the response is for a file, and if so, if its allowed to be processed
		// according to the configuration

		ct := utils.GetMimeExtension(req.Preview)

		processExts := appCfg.ProcessExtensions
		bypassExts := appCfg.BypassExtensions

		if utils.InStringSlice(ct, bypassExts) { // if the extension is bypassable
			debugLogger.LogToFile("Processing not required for file type-", ct)
			debugLogger.LogToFile("Reason: Belongs bypassable extensions")
			w.WriteHeader(http.StatusNoContent, nil, false)
			return
		}
		if utils.InStringSlice(utils.Any, bypassExts) && !utils.InStringSlice(ct, processExts) { // if extension does not belong to "All bypassable except the processable ones" group
			debugLogger.LogToFile("Processing not required for file type-", ct)
			debugLogger.LogToFile("Reason: Doesn't belong to processable extensions")
			w.WriteHeader(http.StatusNoContent, nil, false)
			return
		}

		// copying the file to a buffer for scanner vendor processing as the file is allowed according the co figuration

		buf := &bytes.Buffer{}

		if _, err := io.Copy(buf, req.Request.Body); err != nil {
			errorLogger.LogToFile("Failed to copy the response body to buffer: ", err.Error())
			w.WriteHeader(http.StatusNoContent, nil, false)
			return
		}
		if buf.Len() > appCfg.MaxFileSize {
			debugLogger.LogToFile("File size too large")
			w.WriteHeader(http.StatusNoContent, nil, false)
			return
		}
		// preparing the file meta informations
		filename := utils.GetFileName(req.Request)
		fileExt := utils.GetFileExtension(req.Request)
		fmi := dtos.FileMetaInfo{
			FileName: filename,
			FileType: fileExt,
			FileSize: float64(buf.Len()),
		}
		/* If the shadow virus scanner wants to run independently */
		if vendor == utils.NoVendor && appCfg.RespScannerVendorShadow != utils.NoVendor {
			go doShadowScan(vendor, serviceName, filename, fmi, buf, "")
			w.WriteHeader(http.StatusNoContent, nil, false)
			return
		}
		//fmt.Println("I am here now")

		// Gw rebuid servise req api , resp icap client
		if vendor == "glasswall" {

			filename = "test"
			resp, err := DoCDR(vendor, serviceName, buf, filename)
			if err != nil {
				fmt.Println(err)
				newReq := &http.Request{}
				w.WriteHeader(http.StatusForbidden, newReq, true)
			} else {
				defer resp.Body.Close()
				bodyByte, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					fmt.Println(err)
				}

				newReq := req.Request
				newReq.ContentLength = int64(len(string(bodyByte)))
				newReq.Header.Set("Content-Length", strconv.Itoa(len(string(bodyByte))))
				w.WriteHeader(http.StatusOK, newReq, true)
				w.Write(bodyByte)
			}
			return

		}
		// echo servise
		if vendor == "echo" {
			bodybyte, err := ioutil.ReadAll(buf)
			if err != nil {
				fmt.Println(err)
			}

			newResp := &http.Response{
				StatusCode: http.StatusOK,
				Status:     http.StatusText(http.StatusOK),
				Header: http.Header{
					"Content-Length": []string{strconv.Itoa(len(string(bodybyte)))},
				},
			}
			w.WriteHeader(http.StatusOK, newResp, true)
			w.Write(bodybyte)

			return

		}
		status, sampleInfo := doScan(vendor, serviceName, filename, fmi, buf, "") // scan the file for any anomalies

		if status == http.StatusOK && sampleInfo != nil {
			infoLogger.LogfToFile("The file:%s is %s\n", filename, sampleInfo.SampleSeverity)
			htmlBuf, newResp := utils.GetTemplateBufferAndResponse(utils.BadFileTemplate, &dtos.TemplateData{
				FileName:     sampleInfo.FileName,
				FileType:     sampleInfo.SampleType,
				FileSizeStr:  sampleInfo.FileSizeStr,
				RequestedURL: utils.BreakHTTPURL(req.Request.RequestURI),
				Severity:     sampleInfo.SampleSeverity,
				Score:        sampleInfo.VTIScore,
				ResultsBy:    vendor,
			})
			w.WriteHeader(http.StatusOK, newResp, true)
			w.Write(htmlBuf.Bytes())
			return
		}

		if status == http.StatusNoContent {
			infoLogger.LogfToFile("The file %s is good to go\n", filename)
		}
		w.WriteHeader(status, nil, false)

	case "ERRECHO":
		fmt.Println("ERRECHO")
		w.WriteHeader(http.StatusBadRequest, nil, false)
		debugLogger.LogToFile("Malformed request")
	default:
		fmt.Println("default")

		w.WriteHeader(http.StatusMethodNotAllowed, nil, false)
		debugLogger.LogfToFile("Invalid request method %s- respmod\n", req.Method)
	}
}
