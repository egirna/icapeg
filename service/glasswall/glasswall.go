package glasswall

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"time"

	"icapeg/dtos"
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
	returnOrigIfUnprocessableFileType bool
	returnOrigIf400                   bool
	authID                            string
	logger                            *logger.ZLogger
}

// NewGlasswallService returns a new populated instance of the Glasswall service
func NewGlasswallService(serviceName string, logger *logger.ZLogger) *Glasswall {
	gw := &Glasswall{
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

func (g *Glasswall) SubmitFile(f *bytes.Buffer, filename string) (*dtos.SubmitResponse, error) {

	urlStr := g.BaseURL + g.ScanEndpoint

	bodyBuf := &bytes.Buffer{}

	bodyWriter := multipart.NewWriter(bodyBuf)

	part, err := bodyWriter.CreateFormFile("file", filename)

	if err != nil {
		return nil, err
	}

	io.Copy(part, bytes.NewReader(f.Bytes()))
	if err = bodyWriter.Close(); err != nil {
		elapsed := time.Since(g.logger.LogStartTime)
		zLog.Error().Dur("duration", elapsed).Err(err).Str("value", "failed to close writer").Msgf("cant_close_writer_while_submitting_files_gw")
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, urlStr, bodyBuf)
	if err != nil {
		return nil, err
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: utils.InitSecure()},
	}
	client := &http.Client{Transport: tr}
	ctx, cancel := context.WithTimeout(context.Background(), g.Timeout)
	defer cancel()
	req = req.WithContext(ctx)

	req.Header.Add("Content-Type", bodyWriter.FormDataContentType())
	if g.authID != "" {
		req.Header.Add("authorization", g.authID)
	}

	resp, err := client.Do(req)
	if err != nil {
		elapsed := time.Since(g.logger.LogStartTime)
		zLog.Error().Dur("duration", elapsed).Err(err).Str("value", "service: Glasswall: failed to do request").Msgf("gw_service_fail_to_serve")
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Println("resp")

		bdy, _ := ioutil.ReadAll(resp.Body)
		bdyStr := ""
		if string(bdy) == "" {
			bdyStr = http.StatusText(resp.StatusCode)
		} else {
			bdyStr = string(bdy)

		}
		fmt.Println(bdyStr)

		return nil, errors.New(bdyStr)
	}

	scanResp := dtos.GlasswallScanFileResponse{}
	if err := json.NewDecoder(resp.Body).Decode(&scanResp); err != nil {
		//	stop until know the returned response

		// return nil, err
	}
	fmt.Println("5")
	scanResp.DataID = "15"
	return toSubmitResponse(&scanResp), nil
}

// GetSampleFileInfo returns the submitted sample file's info
func (g *Glasswall) GetSampleFileInfo(sampleID string, filemetas ...dtos.FileMetaInfo) (*dtos.SampleInfo, error) {

	urlStr := g.BaseURL + fmt.Sprintf(g.ReportEndpoint+"/"+sampleID)
	// urlStr := v.BaseURL + fmt.Sprintf(viper.GetString("glasswall.report_endpoint"), viper.GetString("glasswall.api_key"), sampleID)

	req, err := http.NewRequest(http.MethodGet, urlStr, nil)

	if err != nil {
		return nil, err
	}
	if g.authID != "" {
		req.Header.Add("authorization", g.authID)
	}

	client := http.Client{}
	ctx, cancel := context.WithTimeout(context.Background(), g.Timeout)
	defer cancel()
	req = req.WithContext(ctx)

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		bdy, _ := ioutil.ReadAll(resp.Body)
		bdyStr := ""
		if string(bdy) == "" {
			bdyStr = fmt.Sprintf("Status code received:%d with no body", resp.StatusCode)
		} else {
			bdyStr = string(bdy)
		}
		return nil, errors.New(bdyStr)
	}

	sampleResp := dtos.GlasswallReportResponse{}

	if err := json.NewDecoder(resp.Body).Decode(&sampleResp); err != nil {
		return nil, err
	}

	fm := dtos.FileMetaInfo{}

	if len(filemetas) > 0 {
		fm = filemetas[0]
	}

	return toSampleInfo(&sampleResp, fm, g.FailThreshold), nil

}

// GetSubmissionStatus returns the submission status of a submitted sample
func (g *Glasswall) GetSubmissionStatus(submissionID string) (*dtos.SubmissionStatusResponse, error) {

	urlStr := g.BaseURL + fmt.Sprintf(g.ReportEndpoint+"/"+submissionID)

	req, err := http.NewRequest(http.MethodGet, urlStr, nil)

	if err != nil {
		return nil, err
	}
	if g.authID != "" {
		req.Header.Add("authorization", g.authID)
	}
	client := http.Client{}
	ctx, cancel := context.WithTimeout(context.Background(), g.Timeout)
	defer cancel()
	req = req.WithContext(ctx)

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNoContent {
			return nil, errors.New("No content receive from Glasswall on status check, maybe request quota expired")
		}
		bdy, _ := ioutil.ReadAll(resp.Body)
		return nil, errors.New(string(bdy))
	}

	sampleResp := dtos.GlasswallReportResponse{}

	if err := json.NewDecoder(resp.Body).Decode(&sampleResp); err != nil {
		return nil, err
	}

	return toSubmissionStatusResponse(&sampleResp), nil
}

// SubmitURL calls the submission api for Glasswall
func (g *Glasswall) SubmitURL(fileURL, filename string) (*dtos.SubmitResponse, error) {
	return nil, nil
}

// GetSampleURLInfo returns the submitted sample url's info
func (g *Glasswall) GetSampleURLInfo(sampleID string, filemetas ...dtos.FileMetaInfo) (*dtos.SampleInfo, error) {
	return nil, nil
}

// GetStatusCheckInterval returns the status_check_interval duration of the service
func (g *Glasswall) GetStatusCheckInterval() time.Duration {
	return g.statusCheckInterval
}

// GetStatusCheckTimeout returns the status_check_timeout duraion of the service
func (g *Glasswall) GetStatusCheckTimeout() time.Duration {
	return g.statusCheckTimeout
}

// GetBadFileStatus returns the bad_file_status slice of the service
func (g *Glasswall) GetBadFileStatus() []string {
	return g.badFileStatus
}

// GetOkFileStatus returns the ok_file_status slice of the service
func (g *Glasswall) GetOkFileStatus() []string {
	return g.okFileStatus
}

// StatusEndpointExists returns the status_endpoint_exists boolean value of the service
func (g *Glasswall) StatusEndpointExists() bool {
	return g.statusEndPointExists
}

// RespSupported returns the respSupported field of the service
func (g *Glasswall) RespSupported() bool {
	return g.respSupported
}

// ReqSupported returns the reqSupported field of the service
func (g *Glasswall) ReqSupported() bool {
	return g.reqSupported
}

// SendFileApi sends file to api GW rebuild services
func (g *Glasswall) SendFileApi(f *bytes.Buffer, filename string, reqURL string) (*http.Response, int, bool, string, error) {

	urlStr := g.BaseURL + g.ScanEndpoint

	bodyBuf := &bytes.Buffer{}

	bodyWriter := multipart.NewWriter(bodyBuf)

	// adding policy in the request
	bodyWriter.WriteField("contentManagementFlagJson", g.policy)

	part, err := bodyWriter.CreateFormFile("file", filename)

	if err != nil {
		return nil, utils.BadRequestStatusCodeStr, false, "", err
	}

	io.Copy(part, bytes.NewReader(f.Bytes()))
	if err := bodyWriter.Close(); err != nil {
		elapsed := time.Since(g.logger.LogStartTime)
		zLog.Error().Dur("duration", elapsed).Err(err).Str("value", "failed to close writer").Msgf("cant_close_writer_while_sending_api_files_gw")
		return nil, utils.BadRequestStatusCodeStr, false, "", err
	}

	req, err := http.NewRequest(http.MethodPost, urlStr, bodyBuf)
	if err != nil {
		return nil, utils.BadRequestStatusCodeStr, false, "", err
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
		return nil, utils.BadRequestStatusCodeStr, false, "", err
	}
	IsUnprocessableFileType := g.IsUnprocessableFileType(resp, f)
	if !IsUnprocessableFileType {
		if resp.StatusCode == http.StatusBadRequest {
			if g.returnOrigIf400 {
				resp.Body = io.NopCloser(f)
				return resp, utils.NoModificationStatusCodeStr, false, resp.Header.Get("x-adaptation-file-id"), err
			}
			resp.Body = io.NopCloser(getErrorPage("service/unprocessable-file.html",
				&errorPage{
					XAdaptationFileId:         resp.Header.Get("x-adaptation-file-id"),
					XSdkEngineVersion:         resp.Header.Get("x-sdk-engine-version"),
					RequestedURL:              reqURL,
					XGlasswallCloudApiVersion: resp.Header.Get("x-glasswall-cloud-api-version"),
				}))
			return resp, utils.OkStatusCodeStr, true, resp.Header.Get("x-adaptation-file-id"), err
		}
		return resp, utils.OkStatusCodeStr, false, resp.Header.Get("x-adaptation-file-id"), err
	}
	if g.returnOrigIfUnprocessableFileType {
		resp.Body = io.NopCloser(f)
		return resp, utils.NoModificationStatusCodeStr, false, resp.Header.Get("x-adaptation-file-id"), err
	}
	resp.Body = io.NopCloser(getErrorPage("service/unprocessable-file.html",
		&errorPage{
			Reason:                    "File can't be processed",
			XAdaptationFileId:         resp.Header.Get("x-adaptation-file-id"),
			XSdkEngineVersion:         resp.Header.Get("x-sdk-engine-version"),
			RequestedURL:              reqURL,
			XGlasswallCloudApiVersion: resp.Header.Get("x-glasswall-cloud-api-version"),
		}))
	return resp, utils.OkStatusCodeStr, true, resp.Header.Get("x-adaptation-file-id"), err
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

type (
	errorPage struct {
		Reason                    string `json:"reason"`
		XAdaptationFileId         string `json:"x-adaptation-file-id"`
		XSdkEngineVersion         string `json:"x-sdk-engine-version"`
		RequestedURL              string `json:"requested_url"`
		XGlasswallCloudApiVersion string `json:"x-glasswall-cloud-api-version"`
	}
)

func getErrorPage(templateName string, data *errorPage) *bytes.Buffer {
	tmpl, _ := template.ParseFiles(templateName)
	htmlBuf := &bytes.Buffer{}
	tmpl.Execute(htmlBuf, data)
	return htmlBuf
}
