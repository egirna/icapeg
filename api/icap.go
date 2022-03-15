package api

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"net/textproto"
	"strconv"
	"strings"
	"time"

	"icapeg/api/ContentTypes"
	"icapeg/config"
	"icapeg/dtos"
	"icapeg/icap"
	"icapeg/logger"
	"icapeg/readValues"
	"icapeg/service"
	"icapeg/utils"

	zLog "github.com/rs/zerolog/log"
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

	serviceMAxFileSize := readValues.ReadValuesInt(serviceName + ".max_filesize")

	// checking if request method is allowed or not
	methodName := getMethodName(req.Method)
	if methodName != "OPTIONS" {
		isMethodEnabled := readValues.ReadValuesBool(serviceName + "." + methodName)
		if !isMethodEnabled {
			zLog.Debug().Dur("duration", elapsed).Str("value", methodName+" is not enabled").
				Msgf("this_method_is_not_enabled_in_GO_ICAP_configuration")
			w.WriteHeader(http.StatusMethodNotAllowed, nil, false)
			return
		}
	}

	vendor := getVendorName(serviceName)

	h := w.Header()
	h.Set("ISTag", readValues.ReadValuesString(serviceName+".service_tag"))
	h.Set("Service", readValues.ReadValuesString(serviceName+".service_caption"))
	zLog.Info().Dur("duration", elapsed).Str("value", fmt.Sprintf("with method:%s url:%s", req.Method, req.RawURL)).Msgf("request_received_on_icap")

	appCfg := config.App()

	Is204Allowed := is204Allowed(req.Header)

	isShadowServiceEnabled := readValues.ReadValuesBool(serviceName + ".shadow_service")

	if isShadowServiceEnabled {
		shadowService(elapsed, Is204Allowed, req, w, zlogger)
	}

	switch req.Method {
	case utils.ICAPModeOptions:
		optionsMode(h, serviceName, appCfg, vendor, req, w, zlogger)

	case utils.ICAPModeResp:
		defer req.Response.Body.Close()
		// misunderstanding of RFC, to be fixed later
		// if val, exist := req.Header["Allow"]; !exist || (len(val) > 0 && val[0] != utils.NoModificationStatusCodeStr) { // following RFC3507, if the request has Allow: 204 header, it is to be checked and if it doesn't exists, return the request as it is to the ICAP client, https://tools.ietf.org/html/rfc3507#section-4.6
		//	debugLogger.LogToFile("Processing not required for this request")
		//	w.WriteHeader(http.StatusNoContent, nil, false)
		//	return
		// }

		// change body to service name
		/* If any remote icap is enabled, the work flow is controlled by the remote icap */
		/*if strings.HasPrefix(vendor, utils.ICAPPrefix) {
			doRemoteRESPMOD(req, w, vendor, appCfg.RespScannerVendorShadow)
			return
		}*/

		/* If the shadow icap wants to run independently */
		/*if vendor == utils.NoVendor && strings.HasPrefix(appCfg.RespScannerVendorShadow, utils.ICAPPrefix) {
			siSvc := service.GetICAPService(appCfg.RespScannerVendorShadow)
			siSvc.SetHeader(req.Header)
			shadowHTTPResp := utils.GetHTTPResponseCopy(req.Response)
			go doShadowRESPMOD(siSvc, *req.Request, shadowHTTPResp, zLog)
			w.WriteHeader(http.StatusNoContent, nil, false)
			return
		}*/

		/*if vendor == utils.NoVendor && appCfg.RespScannerVendorShadow == utils.NoVendor {  // if no scanner name provided, then bypass everything
			debugLogger.LogToFile("No respmod scanner provided...bypassing everything")
			w.WriteHeader(http.StatusNoContent, nil, false)
			return
		 }*/

		// getting the content type to determine if the response is for a file, and if so, if its allowed to be processed
		// according to the configuration
		if req.Request == nil {
			req.Request = &http.Request{}
		}
		ct := utils.GetMimeExtension(req.Preview)
		processExts := appCfg.ProcessExtensions
		bypassExts := appCfg.BypassExtensions

		if utils.InStringSlice(ct, bypassExts) {
			elapsed = time.Since(zlogger.LogStartTime)
			zLog.Debug().Dur("duration", elapsed).Str("value", fmt.Sprintf("processing not required for file type- %s", ct)).Msgf("belongs_bypassable_extensions")
			w.WriteHeader(http.StatusNoContent, nil, false)
			return
		}

		if utils.InStringSlice(utils.Any, bypassExts) && !utils.InStringSlice(ct, processExts) { // if extension does not belong to "All bypassable except the processable ones" group
			elapsed = time.Since(zlogger.LogStartTime)
			zLog.Debug().Dur("duration", elapsed).Str("value", fmt.Sprintf("processing not required for file type- %s", ct)).Msgf("dont_belong_to_processable_extensions")
			w.WriteHeader(http.StatusNoContent, nil, false)
			return
		}

		// copying the file to a buffer for scanner vendor processing as the file is allowed according the co figuration
		buf := &bytes.Buffer{}

		if _, err := io.Copy(buf, req.Response.Body); err != nil {
			elapsed = time.Since(zlogger.LogStartTime)
			zLog.Error().Dur("duration", elapsed).Err(err).Str("value", "Failed to copy the response body to buffer").Msgf("read_request_body_error")
			w.WriteHeader(http.StatusNoContent, nil, false)
			return
		}
		isGzip := false
		if req.Response.Header.Get("Content-Encoding") == "gzip" {
			isGzip = true
			reader, _ := gzip.NewReader(buf)
			// Empty byte slice.
			var result []byte
			result, err := ioutil.ReadAll(reader)
			if err != nil {
				elapsed = time.Since(zlogger.LogStartTime)
				zLog.Error().Dur("duration", elapsed).Err(err).Str("value", "failed to decompress input file").Msgf("decompress_gz_file_failed")
				w.WriteHeader(http.StatusBadRequest, nil, false)
				return
			}
			buf = bytes.NewBuffer(result)
		}
		if serviceMAxFileSize != 0 && buf.Len() > serviceMAxFileSize {
			elapsed = time.Since(zlogger.LogStartTime)
			zLog.Debug().Dur("duration", elapsed).Str("value", fmt.Sprintf("file size exceeds max filesize limit %d", serviceMAxFileSize)).Msgf("large_file_size")
			if readValues.ReadValuesBool(serviceName + ".return_original_if_max_file_size_exceeded") {
				if Is204Allowed {
					w.WriteHeader(utils.NoModificationStatusCodeStr, nil, false)
				} else {
					req.Response.Body = io.NopCloser(buf)
					w.WriteHeader(utils.OkStatusCodeStr, req.Response, true)
					w.Write(buf.Bytes())
				}
			} else {
				tmpl, _ := template.ParseFiles("service/unprocessable-file.html")
				htmlBuf := &bytes.Buffer{}
				tmpl.Execute(htmlBuf, &errorPage{
					Reason:       "The Max file size is exceeded",
					RequestedURL: req.Request.RequestURI,
				})
				newResp := &http.Response{
					StatusCode: http.StatusForbidden,
					Status:     strconv.Itoa(http.StatusForbidden) + " " + http.StatusText(http.StatusForbidden),
					Header: http.Header{
						utils.ContentType:   []string{utils.HTMLContentType},
						utils.ContentLength: []string{strconv.Itoa(htmlBuf.Len())},
					},
				}
				w.WriteHeader(utils.OkStatusCodeStr, newResp, true)
				w.Write(htmlBuf.Bytes())
			}
			return
		}
		// preparing the file meta information
		filename := utils.GetFileName(req.Request)
		fileExt := utils.GetFileExtension(req.Request)
		fmi := dtos.FileMetaInfo{
			FileName: filename,
			FileType: fileExt,
			FileSize: float64(buf.Len()),
		}
		/* If the shadow virus scanner wants to run independently */
		if vendor == utils.NoVendor && appCfg.RespScannerVendorShadow != utils.NoVendor {
			go doShadowScan(vendor, serviceName, filename, fmi, buf, "", zlogger)
			w.WriteHeader(http.StatusNoContent, nil, false)
			return
		}
		// Gw rebuild service req api , resp ICAP client
		if vendor == "glasswall" {
			filename = "test"
			resp, statusCode, html, x_adaption_id, err := DoCDR(vendor, serviceName, buf, filename, req.Request.RequestURI, zlogger)
			if err != nil {
				zLog.Debug().Dur("duration", elapsed).Str("value", fmt.Sprintf("file wasn't processed")).
					Msgf("forbidden")
				newResp := &http.Response{
					StatusCode: http.StatusForbidden,
					Status:     http.StatusText(http.StatusForbidden),
				}
				w.WriteHeader(http.StatusForbidden, newResp, true)

			} else {
				defer resp.Body.Close()
				bodybyte, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					elapsed = time.Since(zlogger.LogStartTime)
					zLog.Error().Dur("duration", elapsed).Err(err).Str("value", "failed to read the response body from GW engine response").Msgf("read_response_body_from_glasswall_error")
					w.WriteHeader(http.StatusInternalServerError, nil, false)
					return
				}
				if isShadowServiceEnabled {
					// add logs and reports here
					return
				} else {
					if isGzip {
						var newBuf bytes.Buffer
						gz := gzip.NewWriter(&newBuf)
						if _, err := gz.Write(bodybyte); err != nil {
							elapsed = time.Since(zlogger.LogStartTime)
							zLog.Error().Dur("duration", elapsed).Err(err).Str("value", "failed to decompress input file").Msgf("decompress_gz_file_failed")
							w.WriteHeader(http.StatusInternalServerError, nil, false)
							return
						}
						gz.Close()
						bodybyte = newBuf.Bytes()
					}
					if html {
						zLog.Error().Dur("duration", elapsed).Err(err).Str("value", "file wasn't processed because of cloud api failure").
							Msgf("cloud_api_failure")
						newResp := &http.Response{
							StatusCode: http.StatusForbidden,
							Status:     strconv.Itoa(http.StatusForbidden) + " " + http.StatusText(http.StatusForbidden),
							Header: http.Header{
								utils.ContentType:   []string{utils.HTMLContentType},
								utils.ContentLength: []string{strconv.Itoa(len(string(bodybyte)))},
							},
						}
						w.Header().Set("x-adaptation-file-id", x_adaption_id)
						w.WriteHeader(utils.OkStatusCodeStr, newResp, true)
						w.Write(bodybyte)
						return
					}
					if statusCode == 204 && Is204Allowed {
						w.WriteHeader(utils.NoModificationStatusCodeStr, nil, false)
						return
					}
					zLog.Info().Dur("duration", elapsed).Err(err).Str("value", "file was processed").
						Msgf("file_processed_successfully")
					newResp := req.Response
					newResp.Header.Set(utils.ContentLength, strconv.Itoa(len(string(bodybyte))))
					w.Header().Set("x-adaptation-file-id", x_adaption_id)
					w.WriteHeader(utils.OkStatusCodeStr, newResp, true)
					w.Write(bodybyte)
				}
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
					utils.ContentLength: []string{strconv.Itoa(len(string(bodybyte)))},
				},
			}
			w.WriteHeader(http.StatusOK, newResp, true)
			w.Write(bodybyte)

			return

		}
		status, sampleInfo := doScan(vendor, serviceName, filename, fmi, buf, "", zlogger) // scan the file for any anomalies

		if status == http.StatusOK && sampleInfo != nil {
			elapsed = time.Since(zlogger.LogStartTime)
			zLog.Info().Dur("duration", elapsed).Str("value", fmt.Sprintf("The file:%s is %s", filename, sampleInfo.SampleSeverity)).Msgf("scanned_files_for_any_anomalies")
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
			elapsed = time.Since(zlogger.LogStartTime)
			zLog.Info().Dur("duration", elapsed).Str("value", fmt.Sprintf("The file %s is good to go", filename)).Msgf("good_to_go")
		}
		w.WriteHeader(status, nil, false) // \

	case utils.ICAPModeReq:
		defer req.Request.Body.Close()
		// misunderstanding of RFC, to be fixed later
		// if val, exist := req.Header["Allow"]; !exist || (len(val) > 0 && val[0] != utils.NoModificationStatusCodeStr) { // following RFC3507, if the request has Allow: 204 header, it is to be checked and if it doesn't exists, return the request as it is to the ICAP client, https://tools.ietf.org/html/rfc3507#section-4.6
		//	debugLogger.LogToFile("Processing not required for this request")
		//	w.WriteHeader(http.StatusNoContent, nil, false)
		//	return
		// }

		// bypass CONNECT method scanning as a quick fix
		if req.Request != nil {
			if req.Request.Method == "CONNECT" {
				w.WriteHeader(utils.NoModificationStatusCodeStr, nil, false)
				return
			}
		}

		/* If any remote icap is enabled, the work flow is controlled by the remote icap */
		if strings.HasPrefix(vendor, utils.ICAPPrefix) {
			doRemoteRESPMOD(req, w, vendor, appCfg.RespScannerVendorShadow, zlogger)
			return
		}

		/* If the shadow icap wants to run independently */
		if vendor == utils.NoVendor && strings.HasPrefix(appCfg.RespScannerVendorShadow, utils.ICAPPrefix) {
			siSvc := service.GetICAPService(appCfg.RespScannerVendorShadow)
			siSvc.SetHeader(req.Header)
			shadowHTTPResp := utils.GetHTTPResponseCopy(req.Response)
			go doShadowRESPMOD(siSvc, *req.Request, shadowHTTPResp, zlogger)
			w.WriteHeader(http.StatusNoContent, nil, false)
			return
		}

		if vendor == utils.NoVendor && appCfg.RespScannerVendorShadow == utils.NoVendor { // if no scanner name provided, then bypass everything
			elapsed = time.Since(zlogger.LogStartTime)
			zLog.Debug().Dur("duration", elapsed).Str("value", "no respmod scanner provided...bypassing everything").Msgf("no_response_mode_scanner_provided")
			w.WriteHeader(http.StatusNoContent, nil, false)
			return
		}

		// getting the content type to determine if the response is for a file, and if so, if its allowed to be processed
		// according to the configuration

		ct := utils.GetMimeExtension(req.Preview)

		processExts := appCfg.ProcessExtensions
		bypassExts := appCfg.BypassExtensions

		if utils.InStringSlice(ct, bypassExts) { // if the extension is bypassable
			elapsed = time.Since(zlogger.LogStartTime)
			zLog.Debug().Dur("duration", elapsed).Str("value", fmt.Sprintf("processing not required for file type- %s", ct)).Msgf("belongs_bypassable_extensions")
			w.WriteHeader(http.StatusNoContent, nil, false)
			return
		}
		if utils.InStringSlice(utils.Any, bypassExts) && !utils.InStringSlice(ct, processExts) { // if extension does not belong to "All bypassable except the processable ones" group
			elapsed = time.Since(zlogger.LogStartTime)
			zLog.Debug().Dur("duration", elapsed).Str("value", fmt.Sprintf("processing not required for file type- %s", ct)).Msgf("dont_belong_to_processable_extensions")
			w.WriteHeader(http.StatusNoContent, nil, false)
			return
		}

		// get an instance from the struct which fits with content-type in the request
		reqContentType := ContentTypes.GetContentType(req.Request)
		// getting the file from request and store it in buf as a type of bytes.Buffer
		buf := reqContentType.GetFileFromRequest()

		if serviceMAxFileSize != 0 && buf.Len() > serviceMAxFileSize {
			elapsed = time.Since(zlogger.LogStartTime)
			zLog.Debug().Dur("duration", elapsed).Str("value", fmt.Sprintf("file size exceeds max filesize limit %d", serviceMAxFileSize)).Msgf("large_file_size")
			if readValues.ReadValuesBool(serviceName + ".return_original_if_max_file_size_exceeded") {
				if Is204Allowed {
					w.WriteHeader(utils.NoModificationStatusCodeStr, nil, false)
				} else {
					req.Response.Body = io.NopCloser(buf)
					w.WriteHeader(utils.OkStatusCodeStr, req.Request, true)
					w.Write(buf.Bytes())
				}
			} else {
				tmpl, _ := template.ParseFiles("service/unprocessable-file.html")
				htmlBuf := &bytes.Buffer{}
				tmpl.Execute(htmlBuf, &errorPage{
					Reason:       "The Max file size is exceeded",
					RequestedURL: req.Request.RequestURI,
				})
				newResp := &http.Response{
					StatusCode: http.StatusForbidden,
					Status:     strconv.Itoa(http.StatusForbidden) + " " + http.StatusText(http.StatusForbidden),
					Header: http.Header{
						utils.ContentType:   []string{utils.HTMLContentType},
						utils.ContentLength: []string{strconv.Itoa(htmlBuf.Len())},
					},
				}
				w.WriteHeader(utils.OkStatusCodeStr, newResp, true)
				w.Write(htmlBuf.Bytes())
			}
			return
		}
		// preparing the file meta information
		filename := utils.GetFileName(req.Request)
		fileExt := utils.GetFileExtension(req.Request)
		fmi := dtos.FileMetaInfo{
			FileName: filename,
			FileType: fileExt,
			FileSize: float64(buf.Len()),
		}
		/* If the shadow virus scanner wants to run independently */
		if vendor == utils.NoVendor && appCfg.RespScannerVendorShadow != utils.NoVendor {
			go doShadowScan(vendor, serviceName, filename, fmi, buf, "", zlogger)
			w.WriteHeader(http.StatusNoContent, nil, false)
			return
		}

		// Gw rebuid servise req api , resp icap client
		if vendor == "glasswall" {
			if req.Request == nil {
				req.Request = &http.Request{}
			}
			filename = "test"
			resp, _, _, x_adaption_id, err := DoCDR(vendor, serviceName, buf, filename, req.Request.RequestURI, zlogger)
			if err != nil {
				zLog.Error().Dur("duration", elapsed).Err(err).Str("value", "file wasn't processed").
					Msgf("file_wasn't_processed")
				fmt.Println(err)
				newReq := &http.Request{}
				w.WriteHeader(http.StatusForbidden, newReq, true)
			} else {
				defer resp.Body.Close()
				bodyByte, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					fmt.Println(err)
				}
				if isShadowServiceEnabled {
					// add logs and reports here
					return
				} else {
					newReq := req.Request
					// adding the file after scanning in the request again
					finalBody := reqContentType.BodyAfterScanning(bodyByte)
					newReq.ContentLength = int64(len(finalBody))
					newReq.Header.Set(utils.ContentLength, strconv.Itoa(len(finalBody)))
					zLog.Info().Dur("duration", elapsed).Err(err).Str("value", "file was processed").
						Msgf("file_processed_successfully")
					w.Header().Set("x-adaptation-file-id", x_adaption_id)
					w.WriteHeader(utils.OkStatusCodeStr, newReq, true)
					w.Write([]byte(finalBody))
				}
			}
			return

		}
		// echo service
		if vendor == "echo" {
			bodybyte, err := ioutil.ReadAll(buf)
			if err != nil {
				fmt.Println(err)
			}

			newResp := &http.Response{
				StatusCode: http.StatusOK,
				Status:     http.StatusText(http.StatusOK),
				Header: http.Header{
					utils.ContentLength: []string{strconv.Itoa(len(string(bodybyte)))},
				},
			}
			w.WriteHeader(http.StatusOK, newResp, true)
			w.Write(bodybyte)

			return

		}
		status, sampleInfo := doScan(vendor, serviceName, filename, fmi, buf, "", zlogger) // scan the file for any anomalies

		if status == http.StatusOK && sampleInfo != nil {
			elapsed = time.Since(zlogger.LogStartTime)
			zLog.Info().Dur("duration", elapsed).Str("value", fmt.Sprintf("The file:%s is %s", filename, sampleInfo.SampleSeverity)).Msgf("scanned_files_for_any_anomalies")
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
			elapsed = time.Since(zlogger.LogStartTime)
			zLog.Info().Dur("duration", elapsed).Str("value", fmt.Sprintf("The file %s is good to go", filename)).Msgf("good_to_go")
		}
		w.WriteHeader(status, nil, false)

	case "ERRECHO":
		w.WriteHeader(http.StatusBadRequest, nil, false)
		elapsed = time.Since(zlogger.LogStartTime)
		zLog.Debug().Dur("duration", elapsed).Str("value", "Malformed request").Msgf("request_received_malformed")
	default:
		w.WriteHeader(http.StatusMethodNotAllowed, nil, false)
		elapsed = time.Since(zlogger.LogStartTime)
		zLog.Debug().Dur("duration", elapsed).Str("value", fmt.Sprintf("Invalid request method %s- respmod", req.Method)).Msgf("invalid_request_method")
	}
}
