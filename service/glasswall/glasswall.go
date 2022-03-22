package glasswall

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"icapeg/icap"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strconv"
	"time"

	"icapeg/logger"
	"icapeg/readValues"
	"icapeg/utils"

	zLog "github.com/rs/zerolog/log"
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
	w                                 icap.ResponseWriter
	req                               *http.Request
	resp                              *http.Response
	elapsed                           time.Duration
	zlogger                           *logger.ZLogger
	serviceName                       string
	maxFileSize                       int
	is204Allowed                      bool
	shadowService                     bool
	methodName                        string
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
	logger                            *logger.ZLogger
}

// NewGlasswallService returns a new populated instance of the Glasswall service
func NewGlasswallService(w icap.ResponseWriter, req *http.Request, resp *http.Response, elapsed time.Duration, Is204Allowed bool,
	serviceName string, methodName string, logger *logger.ZLogger) *Glasswall {
	gw := &Glasswall{
		w:                                 w,
		req:                               req,
		resp:                              resp,
		elapsed:                           elapsed,
		serviceName:                       serviceName,
		maxFileSize:                       readValues.ReadValuesInt(serviceName + ".max_filesize"),
		is204Allowed:                      Is204Allowed,
		shadowService:                     readValues.ReadValuesBool(serviceName + ".shadow_service"),
		methodName:                        methodName,
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

func (g *Glasswall) SendReqToAPI(f *bytes.Buffer, filename string) *http.Response {

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

func (g *Glasswall) writingIcapResp(resp, serviceResp *http.Response, scannedFile []byte) {
	newResp := resp
	newResp.Header.Set(utils.ContentLength, strconv.Itoa(len(string(scannedFile))))
	g.w.Header().Set("x-adaptation-file-id", serviceResp.Header.Get("x-adaptation-file-id"))
	g.w.WriteHeader(utils.OkStatusCodeStr, newResp, true)
	g.w.Write(scannedFile)
}

func (g *Glasswall) RespMode(req *http.Request, resp *http.Response) {
	if req == nil {
		req = &http.Request{}
	}
	if g.shadowService {
		go g.utilRespMode(req, resp)
		return
	}
	g.utilRespMode(req, resp)
}

func (g *Glasswall) resp400(serviceResp *http.Response, returnOrig bool, reason, requestedUrl string, file *bytes.Buffer) {
	if returnOrig {
		if g.is204Allowed {
			g.w.Header().Set("x-adaptation-file-id", serviceResp.Header.Get("x-adaptation-file-id"))
			g.w.WriteHeader(utils.NoModificationStatusCodeStr, nil, false)
		} else {
			if g.methodName == "RESPMOD" {
				g.resp.Body = io.NopCloser(file)
				g.w.WriteHeader(utils.OkStatusCodeStr, g.resp, true)
			} else {
				g.req.Body = io.NopCloser(file)
				g.w.WriteHeader(utils.OkStatusCodeStr, g.req, true)
			}
			g.w.Write(file.Bytes())
		}
	}
	errPage := GenHtmlPage("service/unprocessable-file.html", reason, requestedUrl)
	g.w.Header().Set("x-adaptation-file-id", serviceResp.Header.Get("x-adaptation-file-id"))
	g.w.WriteHeader(utils.BadRequestStatusCodeStr, ErrPageResp(http.StatusBadRequest, errPage.Len()), true)
	g.w.Write(errPage.Bytes())
}

func (g *Glasswall) utilRespMode(req *http.Request, resp *http.Response) {
	// preparing the file meta information
	filename := utils.GetFileName(req)
	//fileExt := utils.GetFileExtension(req)
	//fmi := dtos.FileMetaInfo{
	//	FileName: filename,
	//	FileType: fileExt,
	//	FileSize: float64(file.Len()),
	//}
	/* If the shadow virus scanner wants to run independently */
	//if vendor == utils.NoVendor && appCfg.RespScannerVendorShadow != utils.NoVendor {
	//	go doShadowScan(vendor, serviceName, filename, fmi, buf, "", zlogger)
	//	w.WriteHeader(http.StatusNoContent, nil, false)
	//	return
	//}

	file, err := CopyingFileToTheBuffer(resp, g.w, g.elapsed, g.zlogger)
	if err != nil {
		return
	}

	isGzip := resp.Header.Get("Content-Encoding") == "gzip"
	if isGzip {
		if file, err = DecodeGzip(file, g.w, g.elapsed, g.zlogger); err != nil {
			return
		}
	}

	if g.maxFileSize != 0 && g.maxFileSize < file.Len() {
		MaxFileSeizeExc(g.returnOrigIfMaxSizeExc, g.is204Allowed, g.w,
			req, resp, file, g.maxFileSize, g.elapsed, g.zlogger)
		return
	}

	serviceResp := g.SendReqToAPI(file, filename)
	scannedFile, err := ApiRespAnalysis(serviceResp, g.w, isGzip, g.elapsed, g.zlogger)
	if err != nil {
		return
	}
	if serviceResp.StatusCode == 400 {
		reason := "File can't be processed by Glasswall engine"
		returnOrig := g.returnOrigIfUnprocessableFileType
		if g.IsUnprocessableFileType(resp, file) {
			reason = "The file type is unsupported by Glasswall engine"
			returnOrig = g.returnOrigIf400
		}
		g.resp400(serviceResp, returnOrig, reason, req.RequestURI, file)
		return
	}

	if isGzip {
		scannedFile, err = compressGzip(scannedFile, g.w, g.elapsed, g.zlogger)
		if err != nil {
			return
		}
	}
	if g.shadowService {
		//add logs here
	}
	g.writingIcapResp(resp, serviceResp, scannedFile)
}
