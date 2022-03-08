package virustotal

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

// VirusTotal represents the informations regarding the virustotal service
type VirusTotal struct {
	BaseURL              string
	Timeout              time.Duration
	APIKey               string
	FileScanEndpoint     string
	URLScanEndpoint      string
	FileReportEndpoint   string
	URLReportEndpoint    string
	FailThreshold        int
	statusCheckInterval  time.Duration
	statusCheckTimeout   time.Duration
	badFileStatus        []string
	okFileStatus         []string
	statusEndPointExists bool
	respSupported        bool
	reqSupported         bool
	logger               *logger.ZLogger
}

// NewVirusTotalService returns a new populated instance of the virustotal service
func NewVirusTotalService(serviceName string, logger *logger.ZLogger) *VirusTotal {
	return &VirusTotal{
		BaseURL:              readValues.ReadValuesString(serviceName + ".base_url"),
		Timeout:              readValues.ReadValuesDuration(serviceName+".timeout") * time.Second,
		APIKey:               readValues.ReadValuesString(serviceName + ".api_key"),
		FileScanEndpoint:     readValues.ReadValuesString(serviceName + ".file_scan_endpoint"),
		URLScanEndpoint:      readValues.ReadValuesString(serviceName + ".url_scan_endpoint"),
		FileReportEndpoint:   readValues.ReadValuesString(serviceName + ".file_report_endpoint"),
		URLReportEndpoint:    readValues.ReadValuesString(serviceName + ".url_report_endpoint"),
		FailThreshold:        readValues.ReadValuesInt(serviceName + ".fail_threshold"),
		statusCheckInterval:  readValues.ReadValuesDuration(serviceName+".status_check_interval") * time.Second,
		statusCheckTimeout:   readValues.ReadValuesDuration(serviceName+".status_check_timeout") * time.Second,
		badFileStatus:        readValues.ReadValuesSlice(serviceName + ".bad_file_status"),
		okFileStatus:         readValues.ReadValuesSlice(serviceName + ".ok_file_status"),
		statusEndPointExists: false,
		respSupported:        readValues.ReadValuesBool(serviceName + ".resp_mode"),
		reqSupported:         readValues.ReadValuesBool(serviceName + ".req_mode"),
		logger:               logger,
	}
}

// SubmitFile calls the submission api for virustotal
func (v *VirusTotal) SubmitFile(f *bytes.Buffer, filename string) (*dtos.SubmitResponse, error) {

	urlStr := v.BaseURL + v.FileScanEndpoint

	bodyBuf := &bytes.Buffer{}

	bodyWriter := multipart.NewWriter(bodyBuf)

	bodyWriter.WriteField("apikey", v.APIKey)

	part, err := bodyWriter.CreateFormFile("file", filename)

	if err != nil {
		return nil, err
	}

	io.Copy(part, bytes.NewReader(f.Bytes()))
	if err := bodyWriter.Close(); err != nil {
		zLog.Error().Msgf("failed to close writer %s", err.Error())
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, urlStr, bodyBuf)

	if err != nil {
		return nil, err
	}

	client := http.Client{}
	ctx, cancel := context.WithTimeout(context.Background(), v.Timeout)
	defer cancel()
	req = req.WithContext(ctx)

	req.Header.Add("Content-Type", bodyWriter.FormDataContentType())

	resp, err := client.Do(req)
	if err != nil {
		zLog.Logger.Error().Msgf("service: virustotal: failed to do request: %s", err.Error())
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		bdy, _ := ioutil.ReadAll(resp.Body)
		bdyStr := ""
		if string(bdy) == "" {
			bdyStr = http.StatusText(resp.StatusCode)
		} else {
			bdyStr = string(bdy)

		}
		return nil, errors.New(bdyStr)
	}

	scanResp := dtos.VirusTotalScanFileResponse{}

	if err := json.NewDecoder(resp.Body).Decode(&scanResp); err != nil {
		return nil, err
	}

	return toSubmitResponse(&scanResp), nil
}

// SubmitURL calls the submission api for virustotal
func (v *VirusTotal) SubmitURL(fileURL, filename string) (*dtos.SubmitResponse, error) {

	urlStr := v.BaseURL + v.URLScanEndpoint

	bodyBuf := &bytes.Buffer{}

	bodyWriter := multipart.NewWriter(bodyBuf)

	bodyWriter.WriteField("apikey", v.APIKey)
	bodyWriter.WriteField("url", fileURL)

	if err := bodyWriter.Close(); err != nil {
		zLog.Logger.Error().Msgf("failed to close writer %s", err.Error())
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, urlStr, bodyBuf)

	if err != nil {
		return nil, err
	}

	client := http.Client{}
	ctx, cancel := context.WithTimeout(context.Background(), v.Timeout)
	defer cancel()
	req = req.WithContext(ctx)

	req.Header.Add("Content-Type", bodyWriter.FormDataContentType())

	resp, err := client.Do(req)
	if err != nil {
		zLog.Logger.Error().Msgf("service: virustotal: failed to do request: %s", err.Error())
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		bdy, _ := ioutil.ReadAll(resp.Body)
		bdyStr := ""
		if string(bdy) == "" {
			bdyStr = http.StatusText(resp.StatusCode)
		} else {
			bdyStr = string(bdy)

		}
		return nil, errors.New(bdyStr)
	}

	scanResp := dtos.VirusTotalScanFileResponse{}

	if err := json.NewDecoder(resp.Body).Decode(&scanResp); err != nil {
		return nil, err
	}

	return toSubmitResponse(&scanResp), nil
}

// GetSampleFileInfo returns the submitted sample file's info
func (v *VirusTotal) GetSampleFileInfo(sampleID string, filemetas ...dtos.FileMetaInfo) (*dtos.SampleInfo, error) {

	urlStr := v.BaseURL + fmt.Sprintf(v.FileReportEndpoint, v.APIKey, sampleID)

	req, err := http.NewRequest(http.MethodGet, urlStr, nil)

	if err != nil {
		return nil, err
	}

	client := http.Client{}
	ctx, cancel := context.WithTimeout(context.Background(), v.Timeout)
	defer cancel()
	req = req.WithContext(ctx)

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {

		if resp.StatusCode == http.StatusNoContent {
			return nil, errors.New("Virustotal returned no content, maybe request quota expired")
		}
		bdy, _ := ioutil.ReadAll(resp.Body)
		bdyStr := ""
		if string(bdy) == "" {
			bdyStr = http.StatusText(resp.StatusCode)
		} else {
			bdyStr = string(bdy)

		}
		return nil, errors.New(bdyStr)
	}

	sampleResp := dtos.VirusTotalReportResponse{}

	if err := json.NewDecoder(resp.Body).Decode(&sampleResp); err != nil {
		return nil, err
	}

	fm := dtos.FileMetaInfo{}

	if len(filemetas) > 0 {
		fm = filemetas[0]
	}

	return toSampleInfo(&sampleResp, fm, v.FailThreshold), nil

}

// GetSampleURLInfo returns the submitted sample url's info
func (v *VirusTotal) GetSampleURLInfo(sampleID string, filemetas ...dtos.FileMetaInfo) (*dtos.SampleInfo, error) {

	urlStr := v.BaseURL + fmt.Sprintf(v.URLReportEndpoint, v.APIKey, sampleID)

	req, err := http.NewRequest(http.MethodGet, urlStr, nil)

	if err != nil {
		return nil, err
	}

	client := http.Client{}
	ctx, cancel := context.WithTimeout(context.Background(), v.Timeout)
	defer cancel()
	req = req.WithContext(ctx)

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {

		if resp.StatusCode == http.StatusNoContent {
			return nil, errors.New("Virustotal returned no content, maybe request quota expired")
		}
		bdy, _ := ioutil.ReadAll(resp.Body)
		bdyStr := ""
		if string(bdy) == "" {
			bdyStr = http.StatusText(resp.StatusCode)
		} else {
			bdyStr = string(bdy)

		}
		return nil, errors.New(bdyStr)
	}

	sampleResp := dtos.VirusTotalReportResponse{}

	if err := json.NewDecoder(resp.Body).Decode(&sampleResp); err != nil {
		return nil, err
	}

	fm := dtos.FileMetaInfo{}

	if len(filemetas) > 0 {
		fm = filemetas[0]
	}

	return toSampleInfo(&sampleResp, fm, v.FailThreshold), nil

}

// GetSubmissionStatus returns the submission status of a submitted sample
func (v *VirusTotal) GetSubmissionStatus(submissionID string) (*dtos.SubmissionStatusResponse, error) {

	urlStr := v.BaseURL + fmt.Sprintf(readValues.ReadValuesString("virustotal.report_endpoint"), readValues.ReadValuesString("virustotal.api_key"), submissionID)

	req, err := http.NewRequest(http.MethodGet, urlStr, nil)

	if err != nil {
		return nil, err
	}

	client := http.Client{}
	ctx, cancel := context.WithTimeout(context.Background(), v.Timeout)
	defer cancel()
	req = req.WithContext(ctx)

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNoContent {
			return nil, errors.New("No content receive from virustotal on status check, maybe request quota expired")
		}
		bdy, _ := ioutil.ReadAll(resp.Body)
		return nil, errors.New(string(bdy))
	}

	sampleResp := dtos.VirusTotalReportResponse{}

	if err := json.NewDecoder(resp.Body).Decode(&sampleResp); err != nil {
		return nil, err
	}

	return toSubmissionStatusResponse(&sampleResp), nil
}

// GetStatusCheckInterval returns the status_check_interval duration of the service
func (v *VirusTotal) GetStatusCheckInterval() time.Duration {
	return v.statusCheckInterval
}

// GetStatusCheckTimeout returns the status_check_timeout duraion of the service
func (v *VirusTotal) GetStatusCheckTimeout() time.Duration {
	return v.statusCheckTimeout
}

// GetBadFileStatus returns the bad_file_status slice of the service
func (v *VirusTotal) GetBadFileStatus() []string {
	return v.badFileStatus
}

// GetOkFileStatus returns the ok_file_status slice of the service
func (v *VirusTotal) GetOkFileStatus() []string {
	return v.okFileStatus
}

// StatusEndpointExists returns the status_endpoint_exists boolean value of the service
func (v *VirusTotal) StatusEndpointExists() bool {
	return v.statusEndPointExists
}

// RespSupported returns the respSupported field of the service
func (v *VirusTotal) RespSupported() bool {
	return v.respSupported
}

// ReqSupported returns the reqSupported field of the service
func (v *VirusTotal) ReqSupported() bool {
	return v.reqSupported
}
func (v *VirusTotal) SendFileApi(f *bytes.Buffer, filename string, reqURL string) (*http.Response, int, bool, string, error) {

	urlStr := v.BaseURL

	bodyBuf := &bytes.Buffer{}

	req, err := http.NewRequest(http.MethodPost, urlStr, bodyBuf)
	// fmt.Println(req)
	if err != nil {
		return nil, utils.BadRequestStatusCodeStr, false, "", err
	}

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		zLog.Logger.Error().Msgf("service: Glasswall: failed to do request: %s", err.Error())
		return nil, utils.BadRequestStatusCodeStr, false, "", err
	}
	return resp, utils.OkStatusCodeStr, false, "", err

}
