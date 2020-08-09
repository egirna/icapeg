package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"icapeg/dtos"
	"icapeg/transformers"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/spf13/viper"
)

type (

	// Vmray represents the vmray service
	Vmray struct {
		BaseURL                  string
		Timeout                  time.Duration
		APIKey                   string
		SubmitEndpoint           string
		GetSampleEndpoint        string
		SubmissionStatusEndpoint string
		statusCheckInterval      time.Duration
		statusCheckTimeout       time.Duration
		badFileStatus            []string
		okFileStatus             []string
		statusEndPointExists     bool
		respSupported            bool
		reqSupported             bool
	}
)

// NewVmrayService populates a new vmray instance as a service
func NewVmrayService() Service {
	return &Vmray{
		BaseURL:                  viper.GetString("vmray.base_url"),
		Timeout:                  viper.GetDuration("vmray.timeout") * time.Second,
		APIKey:                   viper.GetString("vmray.api_key"),
		SubmitEndpoint:           viper.GetString("vmray.submit_endpoint"),
		GetSampleEndpoint:        viper.GetString("vmray.get_sample_endpoint"),
		SubmissionStatusEndpoint: viper.GetString("vmray.submission_status_endpoint"),
		statusCheckInterval:      viper.GetDuration("vmray.status_check_interval") * time.Second,
		statusCheckTimeout:       viper.GetDuration("vmray.status_check_timeout") * time.Second,
		badFileStatus:            viper.GetStringSlice("vmray.bad_file_status"),
		okFileStatus:             viper.GetStringSlice("vmray.ok_file_status"),
		statusEndPointExists:     true,
		respSupported:            true,
		reqSupported:             true,
	}
}

// SubmitFile calls the submission api for vmray
func (v *Vmray) SubmitFile(f *bytes.Buffer, filename string) (*dtos.SubmitResponse, error) {

	urlStr := v.BaseURL + v.SubmitEndpoint

	bodyBuf := &bytes.Buffer{}

	bodyWriter := multipart.NewWriter(bodyBuf)

	part, err := bodyWriter.CreateFormFile("sample_file", filename)

	if err != nil {
		return nil, err
	}

	if f == nil {
		return nil, errors.New("Invalid file")
	}

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
	req.Header.Add("Authorization", fmt.Sprintf("api_key %s", v.APIKey))

	resp, err := client.Do(req)
	if err != nil {
		errorLogger.LogToFile("service: vmray: failed to do request:", err.Error())
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		bdy, _ := ioutil.ReadAll(resp.Body)
		return nil, errors.New(string(bdy))
	}

	sresp := dtos.VmraySubmitResponse{}

	if err := json.NewDecoder(resp.Body).Decode(&sresp); err != nil {
		return nil, err
	}

	if len(sresp.Data.Errors) > 0 {
		errByte, _ := json.Marshal(sresp.Data.Errors)
		return nil, errors.New(string(errByte))
	}

	return transformers.TransformVmrayToSubmitResponse(&sresp), nil
}

// GetSampleFileInfo returns the submitted sample file's info
func (v *Vmray) GetSampleFileInfo(sampleID string, filemetas ...dtos.FileMetaInfo) (*dtos.SampleInfo, error) {

	urlStr := fmt.Sprintf("%s%s/%s", v.BaseURL, v.GetSampleEndpoint, sampleID)

	req, err := http.NewRequest(http.MethodGet, urlStr, nil)

	if err != nil {
		return nil, err
	}

	client := http.Client{}
	ctx, cancel := context.WithTimeout(context.Background(), v.Timeout)
	defer cancel()
	req = req.WithContext(ctx)

	req.Header.Add("Authorization", fmt.Sprintf("api_key %s", v.APIKey))

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		bdy, _ := ioutil.ReadAll(resp.Body)
		return nil, errors.New(string(bdy))
	}

	sampleResp := dtos.GetVmraySampleResponse{}

	if err := json.NewDecoder(resp.Body).Decode(&sampleResp); err != nil {
		return nil, err
	}

	return transformers.TransformVmrayToSampleInfo(&sampleResp), nil
}

// GetSubmissionStatus returns the submission status of a submitted sample
func (v *Vmray) GetSubmissionStatus(submissionID string) (*dtos.SubmissionStatusResponse, error) {
	urlStr := fmt.Sprintf("%s%s/%s", v.BaseURL, v.SubmissionStatusEndpoint, submissionID)

	req, err := http.NewRequest(http.MethodGet, urlStr, nil)

	if err != nil {
		return nil, err
	}

	client := http.Client{}
	ctx, cancel := context.WithTimeout(context.Background(), v.Timeout)
	defer cancel()
	req = req.WithContext(ctx)

	req.Header.Add("Authorization", fmt.Sprintf("api_key %s", v.APIKey))

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		bdy, _ := ioutil.ReadAll(resp.Body)
		return nil, errors.New(string(bdy))
	}

	ssResp := dtos.VmraySubmissionStatusResponse{}

	if err := json.NewDecoder(resp.Body).Decode(&ssResp); err != nil {
		return nil, err
	}

	return transformers.TransformVmrayToSubmissionStatusResponse(&ssResp), nil
}

// SubmitURL calls the submission api for vmray
func (v *Vmray) SubmitURL(fileURL, filename string) (*dtos.SubmitResponse, error) {

	urlStr := v.BaseURL + v.SubmitEndpoint

	bodyBuf := &bytes.Buffer{}

	bodyWriter := multipart.NewWriter(bodyBuf)

	bodyWriter.WriteField("sample_url", fileURL)

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
	req.Header.Add("Authorization", fmt.Sprintf("api_key %s", v.APIKey))

	resp, err := client.Do(req)
	if err != nil {
		errorLogger.LogToFile("service: vmray: failed to do request:", err.Error())
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		bdy, _ := ioutil.ReadAll(resp.Body)
		return nil, errors.New(string(bdy))
	}

	sresp := dtos.VmraySubmitResponse{}

	if err := json.NewDecoder(resp.Body).Decode(&sresp); err != nil {
		return nil, err
	}

	if len(sresp.Data.Errors) > 0 {
		errByte, _ := json.Marshal(sresp.Data.Errors)
		return nil, errors.New(string(errByte))
	}

	return transformers.TransformVmrayToSubmitResponse(&sresp), nil
}

// GetSampleURLInfo returns the submitted sample url's info
func (v *Vmray) GetSampleURLInfo(sampleID string, filemetas ...dtos.FileMetaInfo) (*dtos.SampleInfo, error) {
	return v.GetSampleFileInfo(sampleID, filemetas...)
}

// GetStatusCheckInterval returns the status_check_interval duration of the service
func (v *Vmray) GetStatusCheckInterval() time.Duration {
	return v.statusCheckInterval
}

// GetStatusCheckTimeout returns the status_check_timeout duraion of the service
func (v *Vmray) GetStatusCheckTimeout() time.Duration {
	return v.statusCheckTimeout
}

// GetBadFileStatus returns the bad_file_status slice of the service
func (v *Vmray) GetBadFileStatus() []string {
	return v.badFileStatus
}

// GetOkFileStatus returns the ok_file_status slice of the service
func (v *Vmray) GetOkFileStatus() []string {
	return v.okFileStatus
}

// StatusEndpointExists returns the status_endpoint_exists boolean value of the service
func (v *Vmray) StatusEndpointExists() bool {
	return v.statusEndPointExists
}

// RespSupported returns the respSupported field of the service
func (v *Vmray) RespSupported() bool {
	return v.respSupported
}

// ReqSupported returns the reqSupported field of the service
func (v *Vmray) ReqSupported() bool {
	return v.reqSupported
}
