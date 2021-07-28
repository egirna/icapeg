package service

import (
	"bytes"
	"context"
	"crypto/tls"
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

// scan_endpoint = "/file"
// report_endpoint = "/file"

// Glasswall represents the informations regarding the Glasswall service
type Glasswall struct {
	BaseURL              string
	Timeout              time.Duration
	APIKey               string
	ScanEndpoint         string
	ReportEndpoint       string
	FailThreshold        int
	statusCheckInterval  time.Duration
	statusCheckTimeout   time.Duration
	badFileStatus        []string
	okFileStatus         []string
	statusEndPointExists bool
	respSupported        bool
	reqSupported         bool
}

// NewGlasswallService returns a new populated instance of the Glasswall service
func NewGlasswallService() Service {
	return &Glasswall{
		BaseURL:              viper.GetString("Glasswall.base_url"),
		Timeout:              viper.GetDuration("Glasswall.timeout") * time.Second,
		APIKey:               "888",
		ScanEndpoint:         viper.GetString("Glasswall.scan_endpoint"),
		ReportEndpoint:       "/",
		FailThreshold:        viper.GetInt("Glasswall.fail_threshold"),
		statusCheckInterval:  2 * time.Second,
		statusCheckTimeout:   viper.GetDuration("Glasswall.status_check_timeout") * time.Second,
		badFileStatus:        viper.GetStringSlice("Glasswall.bad_file_status"),
		okFileStatus:         viper.GetStringSlice("Glasswall.ok_file_status"),
		statusEndPointExists: false,
		respSupported:        true,
		reqSupported:         true,
	}
}

// SubmitFile calls the submission api for Glasswall

func (g *Glasswall) SubmitFile(f *bytes.Buffer, filename string) (*dtos.SubmitResponse, error) {

	urlStr := g.BaseURL + g.ScanEndpoint

	bodyBuf := &bytes.Buffer{}

	bodyWriter := multipart.NewWriter(bodyBuf)

	bodyWriter.WriteField("apikey", g.APIKey)

	part, err := bodyWriter.CreateFormFile("file", filename)

	if err != nil {
		return nil, err
	}

	io.Copy(part, bytes.NewReader(f.Bytes()))
	if err := bodyWriter.Close(); err != nil {
		errorLogger.LogToFile("failed to close writer", err.Error())
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, urlStr, bodyBuf)
	//fmt.Println(req)
	if err != nil {
		return nil, err
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	ctx, cancel := context.WithTimeout(context.Background(), g.Timeout)
	defer cancel()
	req = req.WithContext(ctx)

	req.Header.Add("Content-Type", bodyWriter.FormDataContentType())

	resp, err := client.Do(req)
	if err != nil {
		errorLogger.LogToFile("service: Glasswall: failed to do request:", err.Error())
		return nil, err
	}
	fmt.Println(http.StatusOK)

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

		//return nil, err
	}
	fmt.Println("5")
	scanResp.DataID = "15"
	return transformers.TransformGlasswallToSubmitResponse(&scanResp), nil
}

// GetSampleFileInfo returns the submitted sample file's info
func (m *Glasswall) GetSampleFileInfo(sampleID string, filemetas ...dtos.FileMetaInfo) (*dtos.SampleInfo, error) {

	urlStr := m.BaseURL + fmt.Sprintf(m.ReportEndpoint+"/"+sampleID)
	//urlStr := v.BaseURL + fmt.Sprintf(viper.GetString("Glasswall.report_endpoint"), viper.GetString("Glasswall.api_key"), sampleID)

	req, err := http.NewRequest(http.MethodGet, urlStr, nil)

	if err != nil {
		return nil, err
	}
	req.Header.Add("apikey", m.APIKey)
	client := http.Client{}
	ctx, cancel := context.WithTimeout(context.Background(), m.Timeout)
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

	return transformers.TransformGlasswallToSampleInfo(&sampleResp, fm, m.FailThreshold), nil

}

// GetSubmissionStatus returns the submission status of a submitted sample
func (m *Glasswall) GetSubmissionStatus(submissionID string) (*dtos.SubmissionStatusResponse, error) {

	urlStr := m.BaseURL + fmt.Sprintf(m.ReportEndpoint+"/"+submissionID)

	req, err := http.NewRequest(http.MethodGet, urlStr, nil)

	if err != nil {
		return nil, err
	}
	req.Header.Add("apikey", viper.GetString("Glasswall.api_key"))
	client := http.Client{}
	ctx, cancel := context.WithTimeout(context.Background(), m.Timeout)
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

	return transformers.TransformGlasswallToSubmissionStatusResponse(&sampleResp), nil
}

// SubmitURL calls the submission api for Glasswall
func (m *Glasswall) SubmitURL(fileURL, filename string) (*dtos.SubmitResponse, error) {
	return nil, nil
}

// GetSampleURLInfo returns the submitted sample url's info
func (m *Glasswall) GetSampleURLInfo(sampleID string, filemetas ...dtos.FileMetaInfo) (*dtos.SampleInfo, error) {
	return nil, nil
}

// GetStatusCheckInterval returns the status_check_interval duration of the service
func (m *Glasswall) GetStatusCheckInterval() time.Duration {
	return m.statusCheckInterval
}

// GetStatusCheckTimeout returns the status_check_timeout duraion of the service
func (m *Glasswall) GetStatusCheckTimeout() time.Duration {
	return m.statusCheckTimeout
}

// GetBadFileStatus returns the bad_file_status slice of the service
func (m *Glasswall) GetBadFileStatus() []string {
	return m.badFileStatus
}

// GetOkFileStatus returns the ok_file_status slice of the service
func (m *Glasswall) GetOkFileStatus() []string {
	return m.okFileStatus
}

// StatusEndpointExists returns the status_endpoint_exists boolean value of the service
func (m *Glasswall) StatusEndpointExists() bool {
	return m.statusEndPointExists
}

// RespSupported returns the respSupported field of the service
func (m *Glasswall) RespSupported() bool {
	return m.respSupported
}

// ReqSupported returns the reqSupported field of the service
func (m *Glasswall) ReqSupported() bool {
	return m.reqSupported
}

//send file to api GW rebuild services
func (g *Glasswall) SendFileApi(f *bytes.Buffer, filename string) (*http.Response, error) {

	urlStr := g.BaseURL + g.ScanEndpoint

	bodyBuf := &bytes.Buffer{}

	bodyWriter := multipart.NewWriter(bodyBuf)

	bodyWriter.WriteField("apikey", g.APIKey)

	part, err := bodyWriter.CreateFormFile("file", filename)

	if err != nil {
		return nil, err
	}

	io.Copy(part, bytes.NewReader(f.Bytes()))
	if err := bodyWriter.Close(); err != nil {
		errorLogger.LogToFile("failed to close writer", err.Error())
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, urlStr, bodyBuf)
	//fmt.Println(req)
	if err != nil {
		return nil, err
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	ctx, cancel := context.WithTimeout(context.Background(), g.Timeout)
	defer cancel()
	req = req.WithContext(ctx)

	req.Header.Add("Content-Type", bodyWriter.FormDataContentType())

	resp, err := client.Do(req)
	if err != nil {
		errorLogger.LogToFile("service: Glasswall: failed to do request:", err.Error())
		return nil, err
	}
	return resp, err

}
