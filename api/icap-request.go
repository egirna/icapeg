package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/egirna/icapeg/config"
	http_message "github.com/egirna/icapeg/http-message"
	"github.com/egirna/icapeg/icap"
	"github.com/egirna/icapeg/logging"
	"github.com/egirna/icapeg/service"
	utils "github.com/egirna/icapeg/utils"
	"go.uber.org/zap"
)

// ICAPRequest struct is used to encapsulate important information of the ICAP request like method name, etc
type ICAPRequest struct {
	w                      icap.ResponseWriter
	req                    *icap.Request
	h                      http.Header
	Is204Allowed           bool
	isShadowServiceEnabled bool
	appCfg                 *config.AppConfig
	serviceName            string
	methodName             string
	vendor                 string
	optionsReqHeaders      map[string]interface{}
	optionsRespHeaders     map[string]interface{}
	generalReqHeaders      map[string]interface{}
	generalRespHeaders     map[string]interface{}
}

// NewICAPRequest is a func to create a new instance from struct IcapRequest yo handle upcoming ICAP requests
func NewICAPRequest(w icap.ResponseWriter, req *icap.Request) *ICAPRequest {
	ICAPRequest := &ICAPRequest{
		w:      w,
		req:    req,
		h:      w.Header(),
		appCfg: config.App(),
	}
	for serviceName, serviceInstance := range ICAPRequest.appCfg.ServicesInstances {
		service.InitServiceConfig(serviceInstance.Vendor, serviceName)
	}
	return ICAPRequest
}

// RequestInitialization is a fun to retrieve the important information from the ICAP request
// and initialize the ICAP response
func (i *ICAPRequest) RequestInitialization() (string, error) {
	xICAPMetadata := i.generateICAPReqMetaData(utils.ICAPRequestIdLen)
	logging.Logger.Info("Validating the received ICAP request", zap.String("X-ICAP-Metadata", xICAPMetadata))
	logging.Logger.Debug("Creating an instance from ICAPeg configuration", zap.String("X-ICAP-Metadata", xICAPMetadata))
	i.appCfg = config.App()

	// checking if the service doesn't exist in toml file
	// if it does not exist, the response will be 404 ICAP Service Not Found
	logging.Logger.Debug("checking if the service doesn't exist in toml file", zap.String("X-ICAP-Metadata", xICAPMetadata))
	i.serviceName = i.req.URL.Path[1:len(i.req.URL.Path)]
	if !i.isServiceExists(xICAPMetadata) {
		i.w.WriteHeader(utils.ICAPServiceNotFoundCodeStr, nil, false)
		err := errors.New("service doesn't exist")
		logging.Logger.Error(err.Error())
		return xICAPMetadata, err
	}

	// checking if request method is allowed or not
	logging.Logger.Debug("checking if request method is allowed or not", zap.String("X-ICAP-Metadata", xICAPMetadata))
	i.methodName = i.req.Method
	if i.methodName != "options" {
		if !i.isMethodAllowed(xICAPMetadata) {
			i.w.WriteHeader(utils.MethodNotAllowedForServiceCodeStr, nil, false)
			err := errors.New("method is not allowed")
			logging.Logger.Error(err.Error())
			return xICAPMetadata, err
		}
		i.methodName = i.req.Method
	}

	//getting vendor name which depends on the name of the service
	i.vendor = i.getVendorName(xICAPMetadata)

	//adding important headers to options ICAP response
	requiredService := service.GetService(i.vendor, i.serviceName, i.methodName,
		&http_message.HttpMsg{Request: i.req.Request, Response: i.req.Response}, xICAPMetadata)
	logging.Logger.Debug("adding ISTAG Service Headers", zap.String("X-ICAP-Metadata", xICAPMetadata))
	i.addingISTAGServiceHeaders(requiredService.ISTagValue())

	logging.Logger.Debug("checking if returning 24 to ICAP client is allowed or not", zap.String("X-ICAP-Metadata", xICAPMetadata))
	i.Is204Allowed = i.is204Allowed(xICAPMetadata)

	i.isShadowServiceEnabled = config.AppCfg.ServicesInstances[i.serviceName].ShadowService

	//checking if the shadow service is enabled or not to apply shadow service mode
	logging.Logger.Debug("checking if the shadow service is enabled or not to apply shadow service mode", zap.String("X-ICAP-Metadata", xICAPMetadata))
	if i.isShadowServiceEnabled && i.methodName != "OPTIONS" {
		logging.Logger.Debug("shadow service mode i on", zap.String("X-ICAP-Metadata", xICAPMetadata))
		i.shadowService(xICAPMetadata)
		go i.RequestProcessing(xICAPMetadata)
		return xICAPMetadata, errors.New("shadow service")
	} else {
		if i.appCfg.DebuggingHeaders {
			logging.Logger.Debug(
				"adding header to ICAP response in OPTIONS mode indicates that shadow service is off", zap.String("X-ICAP-Metadata", xICAPMetadata))
			i.h["X-ICAPeg-Shadow-Service"] = []string{"false"}
		}
	}

	return xICAPMetadata, nil
}

// RequestProcessing is a func to process the ICAP request upon the service and method required
func (i *ICAPRequest) RequestProcessing(xICAPMetadata string) {
	logging.Logger.Info("processing ICAP request upon the service and method required", zap.String("X-ICAP-Metadata", xICAPMetadata))
	partial := false
	if i.methodName != utils.ICAPModeOptions {
		file := &bytes.Buffer{}
		fileLen := 0

		if i.methodName == utils.ICAPModeResp {
			io.Copy(file, i.req.Response.Body)
			fileLen = file.Len()
			i.req.Response.Header.Set(utils.ContentLength, strconv.Itoa(len(file.Bytes())))
			i.req.Response.Body = io.NopCloser(bytes.NewBuffer(file.Bytes()))

		} else {
			if i.req.Method == utils.ICAPModeReq {

				new, err := http.NewRequest(i.req.Request.Method, i.req.Request.URL.String(), nil)
				if err != nil {
					i.req.OrgRequest = i.req.Request

				} else {
					i.req.OrgRequest = new
				}
				body, _ := ioutil.ReadAll(i.req.Request.Body)
				i.req.OrgRequest.Body = io.NopCloser(bytes.NewBuffer(body))
				i.req.OrgRequest.Header = i.req.Request.Header
				i.req.OrgRequest.Header.Set(utils.ContentLength, strconv.Itoa(len(body)))
				i.req.Request.Body = io.NopCloser(bytes.NewBuffer(body))

			}

		}
		if fileLen == 0 {
			partial = false

		} else {
			if i.req.Header.Get("Preview") != "" && i.req.EndIndicator != "0; ieof" && i.req.EndIndicator != "" {
				partial = true
			}
		}
		if i.req.Header.Get("Preview") != "" && i.req.EndIndicator != "0; ieof" && i.req.EndIndicator != "" {
			partial = true
		}
	}

	i.HostHeader()

	// check the method name
	switch i.methodName {
	// for options mode
	case utils.ICAPModeOptions:
		logging.Logger.Debug("OPTIONS mode", zap.String("X-ICAP-Metadata", xICAPMetadata))
		i.optionsReqHeaders = i.LogICAPReqHeaders()
		i.optionsMode(i.serviceName, xICAPMetadata)
		optionsReqResp := make(map[string]interface{})
		optionsReqResp["ICAP-OPTIONS-Request"] = i.optionsReqHeaders
		optionsReqResp["ICAP-OPTIONS-Response"] = i.optionsRespHeaders
		jsonHeaders, _ := json.Marshal(optionsReqResp)
		final := string(jsonHeaders)
		final = strings.ReplaceAll(final, `\`, "")
		logging.Logger.Debug(final, zap.String("X-ICAP-Metadata", xICAPMetadata))
		break

	//for reqmod and respmod
	default:
		logging.Logger.Debug("Response or Request mode", zap.String("X-ICAP-Metadata", xICAPMetadata))
		i.generalReqHeaders = i.LogICAPReqHeaders()
		i.RespAndReqMods(partial, xICAPMetadata)
	}

}

func (i *ICAPRequest) HostHeader() {
	if i.methodName == "REQMOD" {
		i.req.Request.Header.Set("Host", i.req.Request.Host)
	}
}

func (i *ICAPRequest) RespAndReqMods(partial bool, xICAPMetadata string) {

	if i.methodName == utils.ICAPModeReq {
		defer i.req.Request.Body.Close()
		defer i.req.OrgRequest.Body.Close()

	} else {
		defer i.req.Response.Body.Close()
		//someString := "hello world nand hello go and more"
		//r := strings.NewReader(someString)

		//defer Original_rsp.Body.Close()

	}
	if i.req.Request == nil {
		i.req.Request = &http.Request{}
	}
	//initialize the service by creating instance from the required service
	logging.Logger.Debug("initialize the service by creating instance from the required service", zap.String("X-ICAP-Metadata", xICAPMetadata))
	requiredService := service.GetService(i.vendor, i.serviceName, i.methodName,
		&http_message.HttpMsg{Request: i.req.Request, Response: i.req.Response}, xICAPMetadata)

	logging.Logger.Debug("calling Processing func to process the http message which encapsulated inside the ICAP request", zap.String("X-ICAP-Metadata", xICAPMetadata))
	//calling Processing func to process the http message which encapsulated inside the ICAP request
	// send request to services
	///////////////// start service ////////////////////////////////////////////////////////////////////

	//icap.Request.Response
	IcapStatusCode, httpMsg, serviceHeaders, httpMshHeadersBeforeProcessing, httpMshHeadersAfterProcessing,
		vendorMsgs := requiredService.Processing(partial, i.req.Header)

	// adding the headers which the service wants to add them in the ICAP response
	logging.Logger.Debug("adding the headers which the service wants to add them in the ICAP response", zap.String("X-ICAP-Metadata", xICAPMetadata))
	if serviceHeaders != nil {
		for key, value := range serviceHeaders {
			i.h[key] = []string{value}
		}
	}

	//checking if shadow service mode is enabled to add logs instead of returning another
	//ICAP response beside the one who was sent to the client in line 88
	logging.Logger.Debug("checking if shadow service mode is enabled to add logs instead of returning another", zap.String("X-ICAP-Metadata", xICAPMetadata))
	if i.isShadowServiceEnabled {
		//add logs here
		return
	}

	//check the ICAP status code which returned from the service to decide
	//how should be the ICAP response
	switch IcapStatusCode {
	case utils.InternalServerErrStatusCodeStr:
		logging.Logger.Debug(i.serviceName+" returned ICAP response with status code "+strconv.Itoa(utils.InternalServerErrStatusCodeStr), zap.String("X-ICAP-Metadata", xICAPMetadata))
		i.w.WriteHeader(IcapStatusCode, nil, false)
	case utils.Continue:
		logging.Logger.Debug(i.serviceName+" returned ICAP response with status code "+strconv.Itoa(utils.Continue), zap.String("X-ICAP-Metadata", xICAPMetadata))
		//in case the service returned 100 continue
		//we will get the rest of the body from the client
		httpMsgBody := i.preview(xICAPMetadata)
		i.methodName = i.req.Method
		if i.req.Method == utils.ICAPModeReq {
			i.req.Request.Body = io.NopCloser(bytes.NewBuffer(httpMsgBody.Bytes()))
			i.req.OrgRequest.Body = io.NopCloser(bytes.NewBuffer(httpMsgBody.Bytes()))
		} else {
			i.req.Response.Body = io.NopCloser(bytes.NewBuffer(httpMsgBody.Bytes()))
		}
		i.allHeaders(IcapStatusCode, httpMshHeadersBeforeProcessing, httpMshHeadersAfterProcessing, vendorMsgs,
			xICAPMetadata)
		i.RespAndReqMods(false, xICAPMetadata)
	case utils.RequestTimeOutStatusCodeStr:
		logging.Logger.Debug(i.serviceName+" returned ICAP response with status code "+strconv.Itoa(utils.RequestTimeOutStatusCodeStr), zap.String("X-ICAP-Metadata", xICAPMetadata))
		i.w.WriteHeader(IcapStatusCode, nil, false)
	case utils.NoModificationStatusCodeStr:
		logging.Logger.Debug(i.serviceName+" returned ICAP response with status code "+strconv.Itoa(utils.NoModificationStatusCodeStr), zap.String("X-ICAP-Metadata", xICAPMetadata))
		if i.Is204Allowed {
			i.w.WriteHeader(utils.NoModificationStatusCodeStr, nil, false)
		} else {

			IcapStatusCode = utils.OkStatusCodeStr
			if i.methodName == utils.ICAPModeReq {
				IcapStatusCode = utils.OkStatusCodeStr
				body, _ := ioutil.ReadAll(i.req.OrgRequest.Body)
				i.req.Request.Body = io.NopCloser(bytes.NewBuffer(body))
				i.req.Request.Header.Set(utils.ContentLength, strconv.Itoa(len(body)))
				defer i.req.Request.Body.Close()
				i.w.WriteHeader(utils.OkStatusCodeStr, i.req.Request, true)
			} else {
				IcapStatusCode = utils.OkStatusCodeStr
				i.w.WriteHeader(utils.OkStatusCodeStr, httpMsg, true)
			}

			//i.w.WriteHeader(utils.OkStatusCodeStr, httpMsg, true)
		}
	case utils.OkStatusCodeStr:
		logging.Logger.Debug(i.serviceName+" returned ICAP response with status code "+strconv.Itoa(utils.OkStatusCodeStr), zap.String("X-ICAP-Metadata", xICAPMetadata))
		i.w.WriteHeader(utils.OkStatusCodeStr, httpMsg, true)
	case utils.BadRequestStatusCodeStr:
		logging.Logger.Debug(i.serviceName+" returned ICAP response with status code "+strconv.Itoa(utils.BadRequestStatusCodeStr), zap.String("X-ICAP-Metadata", xICAPMetadata))
		i.w.WriteHeader(IcapStatusCode, httpMsg, true)
	}
	i.allHeaders(IcapStatusCode, httpMshHeadersBeforeProcessing, httpMshHeadersAfterProcessing, vendorMsgs, xICAPMetadata)
}

func (i *ICAPRequest) allHeaders(IcapStatusCode int, httpMshHeadersBeforeProcessing map[string]interface{},
	httpMshHeadersAfterProcessing map[string]interface{}, vendorMsgs map[string]interface{}, xICAPMetadata string) {
	i.generalRespHeaders = i.LogICAPResHeaders(IcapStatusCode)
	generalReqResp := make(map[string]interface{})
	generalReqResp["Vendor-Messages"] = vendorMsgs
	i.generalReqHeaders["HTTP-Message"] = httpMshHeadersBeforeProcessing
	if IcapStatusCode == utils.OkStatusCodeStr {
		i.generalRespHeaders["HTTP-Message"] = httpMshHeadersAfterProcessing
	}
	if i.methodName == utils.ICAPModeReq {
		generalReqResp["ICAP-REQMOD-Request"] = i.generalReqHeaders
		generalReqResp["ICAP-REQMOD-Response"] = i.generalRespHeaders

	} else {
		generalReqResp["ICAP-RESPMOD-Request"] = i.generalReqHeaders
		generalReqResp["ICAP-RESPMOD-Response"] = i.generalRespHeaders
	}
	jsonHeaders, _ := json.Marshal(generalReqResp)
	final := string(jsonHeaders)
	final = strings.ReplaceAll(final, `\`, "")
	logging.Logger.Debug(final, zap.String("X-ICAP-Metadata", xICAPMetadata))
}

// adding headers to the logging
func (i *ICAPRequest) addHeadersToLogs(xICAPMetadata string) {
	logging.Logger.Debug("printing ICAP request headers in logs", zap.String("X-ICAP-Metadata", xICAPMetadata))
	for key, element := range i.req.Header {
		res := key + " : "
		innerRes := ""
		for i := 0; i < len(element); i++ {
			innerRes += element[i]
			if i != len(element)-1 {
				innerRes += ", "
			}
		}
		res += innerRes
		logging.Logger.Debug("An ICAP request header -> "+res, zap.String("X-ICAP-Metadata", xICAPMetadata))
		res = ""
	}
}

// isServiceExists is a func to make sure that service which required in ICAP
// request is existing in the config.go file
func (i *ICAPRequest) isServiceExists(xICAPMetadata string) bool {
	services := i.appCfg.Services
	logging.Logger.Debug("looping over services exist in config.toml file to checking if the service doesn't exist or exist", zap.String("X-ICAP-Metadata", xICAPMetadata))
	for r := 0; r < len(services); r++ {
		if i.serviceName == services[r] {
			return true
		}
	}
	return false

}

// getMethodName is a func to get the name of the method of the ICAP request
func (i *ICAPRequest) getMethodName(xICAPMetadata string) string {
	logging.Logger.Debug("getting the method name", zap.String("X-ICAP-Metadata", xICAPMetadata))
	if i.methodName == "REQMOD" {
		i.methodName = "req_mode"
	} else if i.methodName == "RESPMOD" {
		i.methodName = "resp_mode"
	}
	return i.methodName
}

// isMethodAllowed is a func to check if the method in the ICAP request is allowed in config.go file or not
func (i *ICAPRequest) isMethodAllowed(xICAPMetadata string) bool {
	logging.Logger.Debug("checking if the method in the ICAP request is allowed in config.go file or not", zap.String("X-ICAP-Metadata", xICAPMetadata))
	if i.methodName == "RESPMOD" {
		return i.appCfg.ServicesInstances[i.serviceName].RespMode
	} else if i.methodName == "REQMOD" {
		return i.appCfg.ServicesInstances[i.serviceName].ReqMode

	}
	if i.methodName == "OPTIONS" {
		return true
	}
	return false
}

// getVendorName is a func to get the vendor of the service which in the ICAP request
func (i *ICAPRequest) getVendorName(xICAPMetadata string) string {
	logging.Logger.Debug("getting the vendor of the service which in the ICAP request", zap.String("X-ICAP-Metadata", xICAPMetadata))
	return i.appCfg.ServicesInstances[i.serviceName].Vendor
}

// addingISTAGServiceHeaders is a func to add the important header to ICAP response
func (i *ICAPRequest) addingISTAGServiceHeaders(ISTgValue string) {
	i.h["ISTag"] = []string{ISTgValue}
	i.h["Service"] = []string{i.appCfg.ServicesInstances[i.serviceName].ServiceCaption}
}

// is204Allowed is a func to check if ICAP request has the header "204 : Allowed" or not
func (i *ICAPRequest) is204Allowed(xICAPMetadata string) bool {
	logging.Logger.Debug("checking if (Allow : 204) header exists in ICAP request", zap.String("X-ICAP-Metadata", xICAPMetadata))
	Is204Allowed := false
	if _, exist := i.req.Header["Allow"]; exist &&
		i.req.Header.Get("Allow") == strconv.Itoa(utils.NoModificationStatusCodeStr) {
		logging.Logger.Debug("Allow : 204 header exists in ICAP request", zap.String("X-ICAP-Metadata", xICAPMetadata))
		Is204Allowed = true
	} else if _, exist := i.req.Header["Allow"]; exist &&
		strings.Contains(i.req.Header.Get("Allow"), strconv.Itoa(utils.NoModificationStatusCodeStr)) {
		logging.Logger.Debug("Allow : 204 header exists in ICAP request", zap.String("X-ICAP-Metadata", xICAPMetadata))
		Is204Allowed = true

	}
	return Is204Allowed
}

// shadowService is a func to apply the shadow service
func (i *ICAPRequest) shadowService(xICAPMetadata string) {
	logging.Logger.Debug("applying shadow service", zap.String("X-ICAP-Metadata", xICAPMetadata))
	if i.appCfg.DebuggingHeaders {
		logging.Logger.Debug(
			"adding (X-ICAPeg-Shadow-Service : true) to ICAP response because this"+
				" configuration is enabled in config.toml file", zap.String("X-ICAP-Metadata", xICAPMetadata))
		i.h["X-ICAPeg-Shadow-Service"] = []string{"true"}
	}
	if i.Is204Allowed { // following RFC3507, if the request has Allow: 204 header, it is to be checked and if it doesn't exists, return the request as it is to the ICAP client, https://tools.ietf.org/html/rfc3507#section-4.6
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

// getEnabledMethods is a func get all enable method of a specific service
func (i *ICAPRequest) getEnabledMethods(xICAPMetadata string) string {
	logging.Logger.Debug("getting all enable method of a specific service)", zap.String("X-ICAP-Metadata", xICAPMetadata))
	var allMethods []string
	if i.appCfg.ServicesInstances[i.serviceName].RespMode {
		allMethods = append(allMethods, "RESPMOD")
	}
	if i.appCfg.ServicesInstances[i.serviceName].ReqMode {
		allMethods = append(allMethods, "REQMOD")
	}
	if len(allMethods) == 1 {
		return allMethods[0]
	}
	return allMethods[0] + ", " + allMethods[1]
}

func (i *ICAPRequest) servicePreview() (bool, string) {
	return i.appCfg.ServicesInstances[i.serviceName].PreviewEnabled,
		i.appCfg.ServicesInstances[i.serviceName].PreviewBytes
}

// optionsMode is a func to return an ICAP response in OPTIONS mode
func (i *ICAPRequest) optionsMode(serviceName, xICAPMetadata string) {
	logging.Logger.Debug("preparing headers in OPTIONS mode response", zap.String("X-ICAP-Metadata", xICAPMetadata))
	i.h.Set("Methods", i.getEnabledMethods(xICAPMetadata))
	i.h.Set("Allow", "204")
	// Add preview if preview_enabled is true in config.go
	previewEnabled, previewBytes := i.servicePreview()
	if previewEnabled == true {
		if pb, _ := strconv.Atoi(previewBytes); pb >= 0 {
			i.h.Set("Preview", previewBytes)
		}
	}
	i.h.Set("Transfer-Preview", utils.Any)
	i.w.WriteHeader(http.StatusOK, nil, false)
	i.optionsRespHeaders = i.LogICAPResHeaders(http.StatusOK)
}

// preview function is used to get the rest of the http message from the client after sending
// a preview about the body first
func (i *ICAPRequest) preview(xICAPMetadata string) *bytes.Buffer {
	logging.Logger.Debug(
		"getting the rest of the body from client after the service returned ICAP "+
			"response with status code"+strconv.Itoa(utils.Continue), zap.String("X-ICAP-Metadata", xICAPMetadata))
	r := icap.GetTheRest()
	c := io.NopCloser(r)
	buf := new(bytes.Buffer)
	buf.ReadFrom(c)
	return buf
}

func (i *ICAPRequest) LogICAPReqHeaders() map[string]interface{} {
	reqHeaders := make(map[string]interface{})
	reqHeaders["ICAP-Requested-URL"] = "icap://" + i.req.URL.Host + "/" + i.serviceName
	for key, value := range i.req.Header {
		values := ""
		for i := 0; i < len(value); i++ {
			values += value[0]
			if i != len(value)-1 {
				values += ", "
			}
		}
		reqHeaders[key] = value
	}
	return reqHeaders
}

func (i *ICAPRequest) LogICAPResHeaders(statusCode int) map[string]interface{} {
	respHeaders := make(map[string]interface{})
	respHeaders["ICAP-Response-Status-Code"] = statusCode
	for key, value := range i.h {
		values := ""
		for i := 0; i < len(value); i++ {
			values += value[0]
			if i != len(value)-1 {
				values += ", "
			}
		}
		respHeaders[key] = value
	}
	return respHeaders
}

func (i *ICAPRequest) generateICAPReqMetaData(size int) string {

	name := make([]rune, size)
	var letterRunes = []rune(utils.IdentifierString)
	rand.Seed(time.Now().UnixNano())
	for i := range name {
		name[i] += letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(name)
}
