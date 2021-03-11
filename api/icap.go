package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"icapeg/config"
	"icapeg/dtos"
	"icapeg/logger"
	"icapeg/service"
	"icapeg/utils"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/egirna/icap"
)

var (
	infoLogger  = logger.NewLogger(logger.LogLevelInfo, logger.LogLevelDebug, logger.LogLevelError)
	debugLogger = logger.NewLogger(logger.LogLevelDebug)
	errorLogger = logger.NewLogger(logger.LogLevelError, logger.LogLevelDebug)
)

// ToICAPEGResp is the ICAP Response Mode Handler:
func ToICAPEGResp(w icap.ResponseWriter, req *icap.Request) {

	h := w.Header()
	h.Set("ISTag", utils.ISTag)
	h.Set("Service", "Egirna ICAP-EG")
	infoLogger.LogfToFile("Request received---> METHOD:%s URL:%s\n", req.Method, req.RawURL)

	appCfg := config.App()

	switch req.Method {
	case utils.ICAPModeOptions:

		/* If any remote icap is enabled, the work flow is controlled by the remote icap */
		if strings.HasPrefix(appCfg.RespScannerVendor, utils.ICAPPrefix) {
			doRemoteOPTIONS(req, w, appCfg.RespScannerVendor, appCfg.RespScannerVendorShadow, utils.ICAPModeResp)
			return
		} else if strings.HasPrefix(appCfg.RespScannerVendorShadow, utils.ICAPPrefix) { // if the shadow wants to run independently
			siSvc := service.GetICAPService(appCfg.RespScannerVendorShadow)
			siSvc.SetHeader(req.Header)
			updateEmptyOptionsEndpoint(siSvc, utils.ICAPModeResp)
			go doShadowOPTIONS(siSvc)
		}

		h.Set("Methods", utils.ICAPModeResp)
		h.Set("Allow", "204")

		if pb, _ := strconv.Atoi(appCfg.PreviewBytes); pb > 0 {
			h.Set("Preview", appCfg.PreviewBytes)
		}

		h.Set("Transfer-Preview", utils.Any)

		w.WriteHeader(http.StatusOK, nil, false)

	case utils.ICAPModeResp:
		defer req.Response.Body.Close()
		if val, exist := req.Header["Allow"]; !exist || (len(val) > 0 && val[0] != utils.NoModificationStatusCodeStr) { // following RFC3507, if the request has Allow: 204 header, it is to be checked and if it doesn't exists, return the request as it is to the ICAP client, https://tools.ietf.org/html/rfc3507#section-4.6
			debugLogger.LogToFile("Processing not required for this request")
			w.WriteHeader(http.StatusNoContent, nil, false)
			return
		}

		/* If any remote icap is enabled, the work flow is controlled by the remote icap */
		if strings.HasPrefix(appCfg.RespScannerVendor, utils.ICAPPrefix) {
			doRemoteRESPMOD(req, w, appCfg.RespScannerVendor, appCfg.RespScannerVendorShadow)
			return
		}

		/* If the shadow icap wants to run independently */
		if appCfg.RespScannerVendor == utils.NoVendor && strings.HasPrefix(appCfg.RespScannerVendorShadow, utils.ICAPPrefix) {
			siSvc := service.GetICAPService(appCfg.RespScannerVendorShadow)
			siSvc.SetHeader(req.Header)
			shadowHTTPResp := utils.GetHTTPResponseCopy(req.Response)
			go doShadowRESPMOD(siSvc, *req.Request, shadowHTTPResp)
			w.WriteHeader(http.StatusNoContent, nil, false)
			return
		}

		if appCfg.RespScannerVendor == utils.NoVendor && appCfg.RespScannerVendorShadow == utils.NoVendor { // if no scanner name provided, then bypass everything
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
		if appCfg.RespScannerVendor == utils.NoVendor && appCfg.RespScannerVendorShadow != utils.NoVendor {
			go doShadowScan(filename, fmi, buf, "")
			w.WriteHeader(http.StatusNoContent, nil, false)
			return
		}
		// Gw rebuid servise req api , resp icap client
		if appCfg.RespScannerVendor == "glasswall" {
			//34.242.219.224:1344
			// "https://52.19.235.59"

			resp, err := DoCDR("glasswall", buf, filename)
			if err != nil {
				fmt.Println(err)
				newResp := &http.Response{
					StatusCode: http.StatusForbidden,
					Status:     http.StatusText(http.StatusForbidden),
				}
				w.WriteHeader(http.StatusForbidden, newResp, true)

			} else {
				bodybyte, _ := ioutil.ReadAll(resp.Body)
				newResp := &http.Response{
					StatusCode: http.StatusOK,
					Status:     http.StatusText(http.StatusOK),
					Header: http.Header{
						"Content-Type":   []string{"application/pdf"},
						"Content-Length": []string{strconv.Itoa(len(string(bodybyte)))},
					},
				}
				w.WriteHeader(http.StatusOK, newResp, true)
				w.Write(bodybyte)

			}
			return

		}
		status, sampleInfo := doScan(appCfg.RespScannerVendor, filename, fmi, buf, "") // scan the file for any anomalies

		if status == http.StatusOK && sampleInfo != nil {
			infoLogger.LogfToFile("The file:%s is %s\n", filename, sampleInfo.SampleSeverity)
			htmlBuf, newResp := utils.GetTemplateBufferAndResponse(utils.BadFileTemplate, &dtos.TemplateData{
				FileName:     sampleInfo.FileName,
				FileType:     sampleInfo.SampleType,
				FileSizeStr:  sampleInfo.FileSizeStr,
				RequestedURL: utils.BreakHTTPURL(req.Request.RequestURI),
				Severity:     sampleInfo.SampleSeverity,
				Score:        sampleInfo.VTIScore,
				ResultsBy:    appCfg.RespScannerVendor,
			})
			w.WriteHeader(http.StatusOK, newResp, true)
			w.Write(htmlBuf.Bytes())
			return
		}

		if status == http.StatusNoContent {
			infoLogger.LogfToFile("The file %s is good to go\n", filename)
		}
		w.WriteHeader(status, nil, false) // \

	case "ERRDUMMY":
		fmt.Println("ERRDUMMY")
		w.WriteHeader(http.StatusBadRequest, nil, false)
		debugLogger.LogToFile("Malformed request")
	default:
		fmt.Println("default")

		w.WriteHeader(http.StatusMethodNotAllowed, nil, false)
		debugLogger.LogfToFile("Invalid request method %s- respmod\n", req.Method)
	}
}

// ToICAPEGReq is the ICAP request Mode Handler:
func ToICAPEGReq(w icap.ResponseWriter, req *icap.Request) {
	h := w.Header()
	h.Set("ISTag", utils.ISTag)
	h.Set("Service", "Egirna ICAP-EG")

	infoLogger.LogfToFile("Request received---> METHOD:%s URL:%s\n", req.Method, req.RawURL)

	appCfg := config.App()

	switch req.Method {
	case utils.ICAPModeOptions:

		/* If any remote icap is enabled, the work flow is controlled by the remote icap */
		if strings.HasPrefix(appCfg.ReqScannerVendor, utils.ICAPPrefix) {
			doRemoteOPTIONS(req, w, appCfg.ReqScannerVendor, appCfg.ReqScannerVendorShadow, utils.ICAPModeReq)
			return
		} else if strings.HasPrefix(appCfg.ReqScannerVendorShadow, utils.ICAPPrefix) { /* If the shadow icap wants to run independently */
			siSvc := service.GetICAPService(appCfg.ReqScannerVendorShadow)
			siSvc.SetHeader(req.Header)
			updateEmptyOptionsEndpoint(siSvc, utils.ICAPModeReq)
			go doShadowOPTIONS(siSvc)
		}

		h.Set("Methods", utils.ICAPModeReq)
		h.Set("Allow", "204")
		h.Set("Preview", "0")
		h.Set("Transfer-Preview", utils.Any)
		w.WriteHeader(http.StatusOK, nil, false)
	case utils.ICAPModeReq:

		if val, exist := req.Header["Allow"]; !exist || (len(val) > 0 && val[0] != utils.NoModificationStatusCodeStr) { // following RFC3507, if the request has Allow: 204 header, it is to be checked and if it doesn't exists, return the request as it is to the ICAP client, https://tools.ietf.org/html/rfc3507#section-4.6
			debugLogger.LogToFile("Processing not required for this request")
			w.WriteHeader(http.StatusNoContent, nil, false)
			return
		}

		// /* If any remote icap is enabled, the work flow is controlled by the remote icap */
		if strings.HasPrefix(appCfg.ReqScannerVendor, utils.ICAPPrefix) {
			doRemoteREQMOD(req, w, appCfg.ReqScannerVendor, appCfg.ReqScannerVendorShadow)
			return
		}

		/* If the shadow icap wants to run independently */
		if appCfg.ReqScannerVendor == utils.NoVendor && strings.HasPrefix(appCfg.ReqScannerVendorShadow, utils.ICAPPrefix) {
			siSvc := service.GetICAPService(appCfg.ReqScannerVendorShadow)
			siSvc.SetHeader(req.Header)
			go doShadowREQMOD(siSvc, *req.Request)
			w.WriteHeader(http.StatusNoContent, nil, false)
			return
		}

		if appCfg.ReqScannerVendor == utils.NoVendor && appCfg.ReqScannerVendorShadow == utils.NoVendor {
			debugLogger.LogToFile("No reqmod scanner provided...bypassing everything")
			w.WriteHeader(http.StatusNoContent, nil, false)
			return
		}

		ext := utils.GetFileExtension(req.Request)

		if ext == "" {
			ext = "html"
		}

		processExts := appCfg.ProcessExtensions
		bypassExts := appCfg.BypassExtensions

		if utils.InStringSlice(ext, bypassExts) { // if the extension is bypassable
			debugLogger.LogToFile("Processing not required for file type-", ext)
			debugLogger.LogToFile("Reason: Belongs bypassable extensions")
			w.WriteHeader(http.StatusNoContent, nil, false)
			return
		}

		if utils.InStringSlice(utils.Any, bypassExts) && !utils.InStringSlice(ext, processExts) { // if extension does not belong to "All bypassable except the processable ones" group
			debugLogger.LogToFile("Processing not required for file type-", ext)
			debugLogger.LogToFile("Reason: Doesn't belong to processable extensions")
			w.WriteHeader(http.StatusNoContent, nil, false)
			return
		}

		fileURL := req.Request.RequestURI

		// preparing the file meta informations
		filename := utils.GetFileName(req.Request)
		fileExt := utils.GetFileExtension(req.Request)
		fmi := dtos.FileMetaInfo{
			FileName: filename,
			FileType: fileExt,
		}

		/* If the shadow virus scanner wants to run independently */
		if appCfg.ReqScannerVendor == utils.NoVendor && appCfg.ReqScannerVendorShadow != utils.NoVendor {
			go doShadowScan(filename, fmi, nil, fileURL)
			w.WriteHeader(http.StatusNoContent, nil, false)
			return
		}

		status, sampleInfo := doScan(appCfg.ReqScannerVendor, filename, fmi, nil, fileURL)

		if status == http.StatusOK && sampleInfo != nil {
			infoLogger.LogfToFile("The url:%s is %s\n", filename, sampleInfo.SampleSeverity)
			data := &dtos.TemplateData{
				FileName:     sampleInfo.FileName,
				FileType:     sampleInfo.SampleType,
				FileSizeStr:  sampleInfo.FileSizeStr,
				RequestedURL: utils.BreakHTTPURL(req.Request.RequestURI),
				Severity:     sampleInfo.SampleSeverity,
				Score:        sampleInfo.VTIScore,
				ResultsBy:    appCfg.ReqScannerVendor,
			}

			dataByte, err := json.Marshal(data)

			if err != nil {
				errorLogger.LogToFile("Failed to marshal template data: ", err.Error())
				w.WriteHeader(utils.IfPropagateError(http.StatusInternalServerError, http.StatusNoContent), nil, false)
				return
			}

			req.Request.Body = ioutil.NopCloser(bytes.NewReader(dataByte))

			icap.ServeLocally(w, req)

			return
		}

		if status == http.StatusNoContent {
			infoLogger.LogfToFile("The url %s is good to go\n", fileURL)
		}

		w.WriteHeader(status, nil, false)

	case "ERRDUMMY":
		w.WriteHeader(http.StatusBadRequest, nil, false)
		debugLogger.LogToFile("Malformed request")
	default:
		w.WriteHeader(http.StatusMethodNotAllowed, nil, false)
		debugLogger.LogfToFile("Invalid request method %s- reqmod\n", req.Method)

	}
}
