package glasswall

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"icapeg/service/services-utilities/ContentTypes"
	"icapeg/service/services-utilities/general-functions"
	"icapeg/utils"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strconv"
	"time"

	zLog "github.com/rs/zerolog/log"
	"icapeg/logger"
	"icapeg/readValues"
)

type AuthTokens struct {
	Tokens []Tokens `json:"gw-auth-tokens"`
}

type Tokens struct {
	Id           string `json:"id"`
	Role         string `json:"role"`
	Enabled      bool   `json:"enabled"`
	CreationDate int64  `json:"creation_date"`
	ExpiryDate   int64  `json:"expiry_date"`
}

// Glasswall represents the information regarding the Glasswall service
type Glasswall struct {
	httpMsg                           *utils.HttpMsg
	elapsed                           time.Duration
	serviceName                       string
	methodName                        string
	previewEnabled                    bool
	previewBytes                      string
	maxFileSize                       int
	bypassExts                        []string
	processExts                       []string
	BaseURL                           string
	Timeout                           time.Duration
	APIKey                            string
	ScanEndpoint                      string
	ReportEndpoint                    string
	FailThreshold                     int
	statusCheckInterval               time.Duration
	statusCheckTimeout                time.Duration
	badFileStatus                     []string
	okFileStatus                      []string
	statusEndPointExists              bool
	respSupported                     bool
	reqSupported                      bool
	policy                            string
	returnOrigIfMaxSizeExc            bool
	returnOrigIfUnprocessableFileType bool
	returnOrigIf400                   bool
	authID                            string
	generalFunc                       *general_functions.GeneralFunc
	logger                            *logger.ZLogger
}

// NewGlasswallService returns a new populated instance of the Glasswall service
func NewGlasswallService(serviceName, methodName string, httpMsg *utils.HttpMsg, elapsed time.Duration, logger *logger.ZLogger) *Glasswall {
	gw := &Glasswall{
		httpMsg:                           httpMsg,
		elapsed:                           elapsed,
		serviceName:                       serviceName,
		methodName:                        methodName,
		previewEnabled:                    readValues.ReadValuesBool(serviceName + ".preview_enabled"),
		previewBytes:                      readValues.ReadValuesString(serviceName + ".preview_bytes"),
		maxFileSize:                       readValues.ReadValuesInt(serviceName + ".max_filesize"),
		bypassExts:                        readValues.ReadValuesSlice(serviceName + ".bypass_extensions"),
		processExts:                       readValues.ReadValuesSlice(serviceName + ".process_extensions"),
		BaseURL:                           readValues.ReadValuesString(serviceName + ".base_url"),
		Timeout:                           readValues.ReadValuesDuration(serviceName+".timeout") * time.Second,
		APIKey:                            readValues.ReadValuesString(serviceName + ".api_key"),
		ScanEndpoint:                      readValues.ReadValuesString(serviceName + ".scan_endpoint"),
		ReportEndpoint:                    "/",
		FailThreshold:                     readValues.ReadValuesInt(serviceName + ".fail_threshold"),
		statusCheckInterval:               2 * time.Second,
		respSupported:                     readValues.ReadValuesBool(serviceName + ".resp_mode"),
		reqSupported:                      readValues.ReadValuesBool(serviceName + ".req_mode"),
		policy:                            readValues.ReadValuesString(serviceName + ".policy"),
		returnOrigIfMaxSizeExc:            readValues.ReadValuesBool(serviceName + ".return_original_if_max_file_size_exceeded"),
		returnOrigIfUnprocessableFileType: readValues.ReadValuesBool(serviceName + ".return_original_if_unprocessable_file_type"),
		returnOrigIf400:                   readValues.ReadValuesBool(serviceName + ".return_original_if_400_response"),
		generalFunc:                       general_functions.NewGeneralFunc(httpMsg, elapsed, logger),
		logger:                            logger,
	}
	authTokens := new(AuthTokens)
	err := json.Unmarshal([]byte(gw.APIKey), authTokens)
	if err != nil {
		elapsed := time.Since(logger.LogStartTime)
		zLog.Error().Dur("duration", elapsed).Err(err).Str("value", "unable to parse auth token").Msgf("auth_token_read_error")
		gw.authID = ""
		return gw
	}
	for _, token := range authTokens.Tokens {
		if token.Role == "file_operations" {
			gw.authID = token.Id
		}
	}
	return gw
}

//Processing is a func used for to processing the http message
func (g *Glasswall) Processing() (int, interface{}, map[string]string) {

	isGzip := false

	//extracting the file from http message
	file, reqContentType, err := g.generalFunc.CopyingFileToTheBuffer(g.methodName)
	if err != nil {
		return utils.InternalServerErrStatusCodeStr, nil, nil
	}

	//getting the rest of the HTTP msg body in case of the preview is enabled,
	//and it's okay to get the rest of the HTTP msg
	if g.previewEnabled {
		file = g.generalFunc.Preview()
	}

	//getting the extension of the file
	fileExtension := utils.GetMimeExtension(file.Bytes())

	//check if the file extension is a bypass extension
	//if yes we will not modify the file, and we will return 204 No modifications
	err = g.generalFunc.IfFileExtIsBypass(fileExtension, g.bypassExts)
	if err != nil {
		return utils.NoModificationStatusCodeStr,
			nil, nil
	}

	//check if the file extension is a bypass extension and not a process extension
	//if yes we will not modify the file, and we will return 204 No modifications
	err = g.generalFunc.IfFileExtIsBypassAndNotProcess(fileExtension, g.bypassExts, g.processExts)
	if err != nil {
		return utils.NoModificationStatusCodeStr,
			nil, nil
	}

	//check if the file size is greater than max file size of the service
	//if yes we will return 200 ok or 204 no modification, it depends on the configuration of the service
	if g.maxFileSize != 0 && g.maxFileSize < file.Len() {
		status, file, httpMsg := g.generalFunc.IfMaxFileSeizeExc(g.returnOrigIfMaxSizeExc, file, g.maxFileSize)
		fileAfterPrep, httpMsg := g.generalFunc.IfStatusIs204WithFile(g.methodName, status, file, isGzip, reqContentType, httpMsg)
		if fileAfterPrep == nil && httpMsg == nil {
			return utils.InternalServerErrStatusCodeStr, nil, nil
		}
		switch msg := httpMsg.(type) {
		case *http.Request:
			msg.Body = io.NopCloser(bytes.NewBuffer(fileAfterPrep))
			return status, msg, nil
		case *http.Response:
			msg.Body = io.NopCloser(bytes.NewBuffer(fileAfterPrep))
			return status, msg, nil
		}
		return status, nil, nil
	}

	//check if the body of the http message is compressed in Gzip or not
	isGzip = g.generalFunc.IsBodyGzipCompressed(g.methodName)
	//if it's compressed, we decompress it to send it to Glasswall service
	if isGzip {
		if file, err = g.generalFunc.DecompressGzipBody(file); err != nil {
			return utils.InternalServerErrStatusCodeStr, nil, nil
		}
	}

	//getting the name of the file
	filename := g.generalFunc.GetFileName()

	//sending the file to Glasswall service to scan it
	serviceResp := g.SendFileToAPI(file, filename)

	//adding headers that Glasswall service wants to add to the ICAP response
	serviceHeaders := make(map[string]string)
	serviceHeaders["X-Adaptation-File-Id"] = serviceResp.Header.Get("x-adaptation-file-id")

	//checking if the http status code of Glasswall API response is 400
	//if yes, it's because of the type of the file can't be processed of GW API
	//or because of any other reason
	if serviceResp.StatusCode == 400 {
		//initializing the reason of 400 status code and configuration of returning
		//original page or returning an error page
		reason := "File can't be processed by Glasswall engine"
		returnOrig := g.returnOrigIf400

		//check if the reason is the type of the file
		if g.IsUnprocessableFileType(serviceResp, file) {
			//reinitializing the variables if the file type is the reasong
			reason = "The file type is unsupported by Glasswall engine"
			returnOrig = g.returnOrigIfUnprocessableFileType
		}

		//generating the suitable response of this case (200 ok or 204 no modifications)
		status, file, httpMsg := g.resp400(returnOrig, reason, file)

		//preparing the http message if the response should be 204 no modifications
		//to decide if wee should return 200 ok or 204 no modifications
		fileAfterPrep, httpMsg := g.ifICAPStatusIs204(status, file, isGzip, reqContentType, httpMsg)
		if fileAfterPrep == nil && httpMsg == nil {
			return utils.InternalServerErrStatusCodeStr, nil, nil
		}

		//returning the http message and the ICAP status code
		switch msg := httpMsg.(type) {
		case *http.Request:
			msg.Body = io.NopCloser(bytes.NewBuffer(fileAfterPrep))
			return status, msg, serviceHeaders
		case *http.Response:
			msg.Body = io.NopCloser(bytes.NewBuffer(fileAfterPrep))
			return status, msg, serviceHeaders
		}
		return status, nil, serviceHeaders
	}

	//extracting the file from GW API response
	scannedFile, err := g.generalFunc.ExtractFileFromServiceResp(serviceResp)
	if err != nil {
		return utils.InternalServerErrStatusCodeStr, nil, serviceHeaders
	}

	//if the original file was compressed in GZIP, we will compress the scanned file in GZIP also
	if isGzip {
		scannedFile, err = g.generalFunc.CompressFileGzip(scannedFile)
		if err != nil {
			return utils.InternalServerErrStatusCodeStr, nil, serviceHeaders
		}
	}

	//returning the scanned file if everything is ok
	scannedFile = g.generalFunc.PreparingFileAfterScanning(scannedFile, reqContentType, g.methodName)
	return utils.OkStatusCodeStr, g.returningHttpMessage(scannedFile), serviceHeaders
}

//function to return the suitable http message (http request, http response)
func (g *Glasswall) returningHttpMessage(file []byte) interface{} {
	switch g.methodName {
	case utils.ICAPModeReq:
		g.httpMsg.Request.Header.Set(utils.ContentLength, strconv.Itoa(len(string(file))))
		g.httpMsg.Request.Body = io.NopCloser(bytes.NewBuffer(file))
		return g.httpMsg.Request
	case utils.ICAPModeResp:
		g.httpMsg.Response.Header.Set(utils.ContentLength, strconv.Itoa(len(string(file))))
		g.httpMsg.Response.Body = io.NopCloser(bytes.NewBuffer(file))
		return g.httpMsg.Response
	}
	return nil
}

//handling the HTTP message if the status should be 204 no modifications
func (g *Glasswall) ifICAPStatusIs204(status int, file *bytes.Buffer, isGzip bool, reqContentType ContentTypes.ContentType, httpMessage interface{}) ([]byte,
	interface{}) {
	var fileAfterPrep []byte
	var err error
	if isGzip {
		fileAfterPrep, err = g.generalFunc.CompressFileGzip(file.Bytes())
		if err != nil {
			return nil, nil
		}
	}

	if g.methodName == utils.ICAPModeReq {
		fileAfterPrep = g.generalFunc.PreparingFileAfterScanning(file.Bytes(), reqContentType, g.methodName)
	} else {
		fileAfterPrep = file.Bytes()
	}
	if status == utils.NoModificationStatusCodeStr {
		return fileAfterPrep, g.returningHttpMessage(fileAfterPrep)
	}
	return fileAfterPrep, httpMessage
}

//generating a suitable http message if the GW API response status code is 400,
//it depends on the configurations of the service. if it's allow returning original file in this case or not
func (g *Glasswall) resp400(returnOrig bool, reason string, file *bytes.Buffer) (int, *bytes.Buffer, interface{}) {
	if returnOrig {
		return utils.NoModificationStatusCodeStr, file, nil
	}
	errPage := g.generalFunc.GenHtmlPage("service/unprocessable-file.html", reason, g.httpMsg.Request.RequestURI)
	g.httpMsg.Response = g.generalFunc.ErrPageResp(http.StatusForbidden, errPage.Len())
	return utils.OkStatusCodeStr, errPage, g.httpMsg.Response
}

//IsUnprocessableFileType is a func to check if the reason of 400 of the GW API response status code
//is because of the type of the fiel or not
func (g *Glasswall) IsUnprocessableFileType(resp *http.Response, f *bytes.Buffer) bool {
	bodyByte, _ := ioutil.ReadAll(resp.Body)
	bodyStr := string(bodyByte)
	var js interface{}
	if json.Unmarshal([]byte(bodyStr), &js) != nil {
		resp.Body = io.NopCloser(bytes.NewBuffer(bodyByte))
		return false
	}
	resp.Body = io.NopCloser(bytes.NewBuffer(bodyByte))
	var data map[string]interface{}
	body, _ := ioutil.ReadAll(resp.Body)
	_ = json.Unmarshal(body, &data)
	if resp.StatusCode == 400 {
		if data["status"] == "GW_UNPROCESSED" && data["rebuildProcessingStatus"] == "FILE_TYPE_UNSUPPORTED" {
			resp.Body = io.NopCloser(f)
			return true
		}
	}
	resp.Body = io.NopCloser(bytes.NewBuffer(bodyByte))
	return false
}

//SendFileToAPI is a function to send the file to GW API
func (g *Glasswall) SendFileToAPI(f *bytes.Buffer, filename string) *http.Response {

	urlStr := g.BaseURL + g.ScanEndpoint

	bodyBuf := &bytes.Buffer{}

	bodyWriter := multipart.NewWriter(bodyBuf)

	// adding policy in the request
	bodyWriter.WriteField("contentManagementFlagJson", g.policy)

	part, err := bodyWriter.CreateFormFile("file", filename)

	if err != nil {
		return nil
	}

	io.Copy(part, bytes.NewReader(f.Bytes()))
	if err := bodyWriter.Close(); err != nil {
		elapsed := time.Since(g.logger.LogStartTime)
		zLog.Error().Dur("duration", elapsed).Err(err).Str("value", "failed to close writer").Msgf("cant_close_writer_while_sending_api_files_gw")
		return nil
	}

	req, err := http.NewRequest(http.MethodPost, urlStr, bodyBuf)
	if err != nil {
		return nil
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: utils.InitSecure()},
	}
	client := &http.Client{Transport: tr}
	ctx, _ := context.WithTimeout(context.Background(), g.Timeout)

	// defer cancel() commit cancel must be on goroutine
	req = req.WithContext(ctx)

	req.Header.Add("Content-Type", bodyWriter.FormDataContentType())
	if g.authID != "" {
		req.Header.Add("authorization", g.authID)
	}

	resp, err := client.Do(req)
	if err != nil {
		elapsed := time.Since(g.logger.LogStartTime)
		zLog.Error().Dur("duration", elapsed).Err(err).Str("value", "service: Glasswall: failed to do request").Msgf("gw_service_fail_to_serve")
		return nil
	}
	return resp
}