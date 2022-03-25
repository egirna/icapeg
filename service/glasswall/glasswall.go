package glasswall

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
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
	"icapeg/service/general-functions"
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
	req                               *http.Request
	resp                              *http.Response
	elapsed                           time.Duration
	serviceName                       string
	methodName                        string
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
func NewGlasswallService(serviceName, methodName string, req *http.Request, resp *http.Response, elapsed time.Duration, logger *logger.ZLogger) *Glasswall {
	gw := &Glasswall{
		req:                               req,
		resp:                              resp,
		elapsed:                           elapsed,
		serviceName:                       serviceName,
		methodName:                        methodName,
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
		generalFunc:                       general_functions.NewGeneralFunc(req, resp, elapsed, logger),
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

func (g *Glasswall) Processing() (int, []byte, *http.Response, map[string]string) {

	file, err := g.generalFunc.CopyingFileToTheBuffer(g.methodName)
	if err != nil {
		return utils.InternalServerErrStatusCodeStr, nil, nil, nil
	}

	isGzip := g.generalFunc.IsBodyGzipCompressed(g.methodName)
	if isGzip {
		if file, err = g.generalFunc.DecompressGzipBody(file); err != nil {
			return utils.InternalServerErrStatusCodeStr, nil, nil, nil
		}
	}

	if g.maxFileSize != 0 && g.maxFileSize < file.Len() {
		status, file, httpResponse := g.generalFunc.IfMaxFileSeizeExc(g.returnOrigIfMaxSizeExc, file, g.maxFileSize)
		return status, file.Bytes(), httpResponse, nil
	}

	filename := g.generalFunc.GetFileName()
	serviceResp := g.SendFileToAPI(file, filename)
	serviceHeaders := make(map[string]string)
	serviceHeaders["X-Adaptation-File-Id"] = serviceResp.Header.Get("x-adaptation-file-id")

	if serviceResp.StatusCode == 400 {
		reason := "File can't be processed by Glasswall engine"
		returnOrig := g.returnOrigIf400
		if g.IsUnprocessableFileType(serviceResp, file) {
			reason = "The file type is unsupported by Glasswall engine"
			returnOrig = g.returnOrigIfUnprocessableFileType
		}
		status, file, httpResponse := g.resp400(returnOrig, reason, file)
		return status, file.Bytes(), httpResponse, serviceHeaders
	}

	scannedFile, err := g.generalFunc.ExtractFileFromServiceResp(serviceResp)
	if err != nil {
		return utils.InternalServerErrStatusCodeStr, nil, nil, serviceHeaders
	}

	if isGzip {
		scannedFile, err = g.generalFunc.CompressFileGzip(scannedFile)
		if err != nil {
			return utils.InternalServerErrStatusCodeStr, nil, nil, serviceHeaders
		}
	}
	g.resp.Header.Set(utils.ContentLength, strconv.Itoa(len(string(scannedFile))))
	return utils.OkStatusCodeStr, scannedFile, g.resp, serviceHeaders
}

func (g *Glasswall) resp400(returnOrig bool, reason string, file *bytes.Buffer) (int, *bytes.Buffer, *http.Response) {
	if returnOrig {
		return utils.NoModificationStatusCodeStr, file, nil
	}
	errPage := g.generalFunc.GenHtmlPage("service/unprocessable-file.html", reason, g.req.RequestURI)
	g.resp = g.generalFunc.ErrPageResp(http.StatusForbidden, errPage.Len())
	return utils.OkStatusCodeStr, errPage, g.resp
}

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
