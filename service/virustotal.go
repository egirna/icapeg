package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"icapeg/dtos"
	"icapeg/readValues"
	"icapeg/transformers"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"time"
)

// VirusTotal represents the informations regarding the virustotal service
type VirusTotal struct {
	BaseURL              string
	Timeout              time.Duration
	APIKey               string
	FileScanEndpoint     string
	FileAnalysisEndpoint string
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
}

// NewVirusTotalService returns a new populated instance of the virustotal service
func NewVirusTotalService(serviceName string) Service {
	return &VirusTotal{
		BaseURL:              readValues.ReadValuesString("virustotal.base_url"),
		Timeout:              readValues.ReadValuesDuration("virustotal.timeout") * time.Second,
		APIKey:               readValues.ReadValuesString("virustotal.api_key"),
		FileScanEndpoint:     readValues.ReadValuesString("virustotal.file_scan_endpoint"),
		URLScanEndpoint:      readValues.ReadValuesString("virustotal.url_scan_endpoint"),
		FileAnalysisEndpoint: readValues.ReadValuesString("virustotal.file_analysis_endpoint"),
		FileReportEndpoint:   readValues.ReadValuesString("virustotal.file_report_endpoint"),
		URLReportEndpoint:    readValues.ReadValuesString("virustotal.url_report_endpoint"),
		FailThreshold:        readValues.ReadValuesInt("virustotal.fail_threshold"),
		statusCheckInterval:  readValues.ReadValuesDuration("virustotal.status_check_interval") * time.Second,
		statusCheckTimeout:   readValues.ReadValuesDuration("virustotal.status_check_timeout") * time.Second,
		badFileStatus:        readValues.ReadValuesSlice("virustotal.bad_file_status"),
		okFileStatus:         readValues.ReadValuesSlice("virustotal.ok_file_status"),
		statusEndPointExists: false,
		respSupported:        true,
		reqSupported:         true,
	}
}

// SubmitFile calls the submission api for virustotal
func (v *VirusTotal) SubmitFile(f *bytes.Buffer, filename string) (*dtos.SubmitResponse, error) {

	urlStr := v.BaseURL + v.FileScanEndpoint

	bodyBuf := &bytes.Buffer{}

	bodyWriter := multipart.NewWriter(bodyBuf)

	//bodyWriter.WriteField("apikey", v.APIKey)

	part, err := bodyWriter.CreateFormFile("file", filename)

	if err != nil {
		return nil, err
	}
	//("Before or after?")
	io.Copy(part, bytes.NewReader(f.Bytes()))
	if err := bodyWriter.Close(); err != nil {
		errorLogger.LogToFile("failed to close writer", err.Error())
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
	req.Header.Add("x-apikey", v.APIKey)
	resp, err := client.Do(req)
	if err != nil {
		errorLogger.LogToFile("service: virustotal: failed to do request:", err.Error())
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

	scanResp := dtos.VirusTotalUploadResponse{}

	if err := json.NewDecoder(resp.Body).Decode(&scanResp); err != nil {
		return nil, err
	}

	//v.FileAnalysis(f, scanResp.Data.Id)

	return transformers.TransformVirusTotalToSubmitResponse(&scanResp.Data), nil
	//return nil, nil
}

// SubmitURL calls the submission api for virustotal
func (v *VirusTotal) SubmitURL(fileURL, filename string) (*dtos.SubmitResponse, error) {

	urlStr := v.BaseURL + v.URLScanEndpoint

	bodyBuf := &bytes.Buffer{}

	bodyWriter := multipart.NewWriter(bodyBuf)

	bodyWriter.WriteField("apikey", v.APIKey)
	bodyWriter.WriteField("url", fileURL)

	if err := bodyWriter.Close(); err != nil {
		errorLogger.LogToFile("failed to close writer", err.Error())
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
		errorLogger.LogToFile("service: virustotal: failed to do request:", err.Error())
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

	scanResp := dtos.VirusTotalUploadResponse{}

	if err := json.NewDecoder(resp.Body).Decode(&scanResp); err != nil {
		fmt.Println("Have I reach here")
		return nil, err
	}

	return transformers.TransformVirusTotalToSubmitResponse(&scanResp.Data), nil
}

// GetSampleFileInfo returns the submitted sample file's info
func (v *VirusTotal) GetSampleFileInfo(sampleID string, filemetas ...dtos.FileMetaInfo) (*dtos.SampleInfo, error) {

	urlStr := v.BaseURL + v.FileAnalysisEndpoint + sampleID
	//fmt.Println(urlStr, "Report is hereeee")
	req, err := http.NewRequest(http.MethodGet, urlStr, nil)

	if err != nil {
		return nil, err
	}

	client := http.Client{}
	ctx, cancel := context.WithTimeout(context.Background(), v.Timeout)
	defer cancel()
	req = req.WithContext(ctx)
	req.Header.Add("x-apikey", v.APIKey)
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

	sampleResp := dtos.VirusTotalReportResponseV3{}
	/*bdy, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(bdy))*/
	//fmt.Println(resp.Body)

	if err := json.NewDecoder(resp.Body).Decode(&sampleResp); err != nil {
		fmt.Println("Error here")
		return nil, err
	}

	fm := dtos.FileMetaInfo{}
	if len(filemetas) > 0 {
		fm = filemetas[0]
	}

	return transformers.TransformVirusTotalToSampleInfo(&sampleResp, fm, v.FailThreshold), nil
	//return nil, nil

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

	sampleResp := dtos.VirusTotalReportResponseV3{}

	if err := json.NewDecoder(resp.Body).Decode(&sampleResp); err != nil {
		return nil, err
	}

	fm := dtos.FileMetaInfo{}

	if len(filemetas) > 0 {
		fm = filemetas[0]
	}

	return transformers.TransformVirusTotalToSampleInfo(&sampleResp, fm, v.FailThreshold), nil
	//return nil, nil

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

	return transformers.TransformVirusTotalToSubmissionStatusResponse(&sampleResp), nil
}

// SubmitFile calls the submission api for virustotal
func (v *VirusTotal) FileAnalysis(f *bytes.Buffer, id string) (*dtos.SubmitResponse, error) {

	urlStr := v.BaseURL + v.FileAnalysisEndpoint + id

	bodyBuf := &bytes.Buffer{}

	bodyWriter := multipart.NewWriter(bodyBuf)

	//bodyWriter.WriteField("apikey", v.APIKey)

	req, err := http.NewRequest(http.MethodGet, urlStr, bodyBuf)

	if err != nil {
		return nil, err
	}

	client := http.Client{}
	ctx, cancel := context.WithTimeout(context.Background(), v.Timeout)
	defer cancel()
	req = req.WithContext(ctx)
	fmt.Println(req)
	req.Header.Add("Content-Type", bodyWriter.FormDataContentType())
	req.Header.Add("x-apikey", v.APIKey)
	fmt.Println(req)

	fmt.Println("Am I here?")
	resp, err := client.Do(req)
	fmt.Println("The response is ", resp)
	if err != nil {
		errorLogger.LogToFile("service: virustotal: failed to do request:", err.Error())
		return nil, err
	}
	fmt.Println(resp.StatusCode)
	//fmt.Println(resp.Body)

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
	bdy, _ := ioutil.ReadAll(resp.Body)

	fmt.Println("The response body is ", string(bdy))
	//scanResp := dtos.VirusTotalUploadResponse{}
	/*
		if err := json.NewDecoder(resp.Body).Decode(&scanResp); err != nil {
			fmt.Println("did i reach here?")
			fmt.Println(scanResp)
			return nil, err
		}*/
	//fmt.Println(scanResp.Data.Type)
	fmt.Println("ME HERE But who?")
	//return transformers.TransformVirusTotalToSubmitResponse(&scanResp), nil
	return nil, nil
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
func (g *VirusTotal) SendFileApi(f *bytes.Buffer, filename string) (*http.Response, error) {

	urlStr := g.BaseURL

	bodyBuf := &bytes.Buffer{}

	req, err := http.NewRequest(http.MethodPost, urlStr, bodyBuf)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}

	resp, err := client.Do(req)
	fmt.Println(req)
	if err != nil {
		errorLogger.LogToFile("service: Glasswall: failed to do request:", err.Error())
		return nil, err
	}
	return resp, err

}
