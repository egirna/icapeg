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
	"log"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/spf13/viper"
)

// VirusTotal represents the informations regarding the virustotal service
type VirusTotal struct {
	BaseURL string
	Timeout time.Duration
	APIKey  string
}

// NewVirusTotalService returns a new populated instance of the virustotal service
func NewVirusTotalService() Service {
	return &VirusTotal{
		BaseURL: viper.GetString("virustotal.base_url"),
		Timeout: viper.GetDuration("virustotal.timeout") * time.Second,
		APIKey:  viper.GetString("virustotal.api_key"),
	}
}

// SubmitFile calls the submission api for virustotal
func (v *VirusTotal) SubmitFile(f *bytes.Buffer, filename string) (*dtos.SubmitResponse, error) {

	urlStr := v.BaseURL + viper.GetString("virustotal.scan_endpoint")

	bodyBuf := &bytes.Buffer{}

	bodyWriter := multipart.NewWriter(bodyBuf)

	bodyWriter.WriteField("apikey", v.APIKey)

	part, err := bodyWriter.CreateFormFile("file", filename)

	if err != nil {
		return nil, err
	}

	io.Copy(part, bytes.NewReader(f.Bytes()))
	if err := bodyWriter.Close(); err != nil {
		log.Println("failed to close writer", err.Error())
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
		log.Println("service: virustotal: failed to do request:", err.Error())
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		bdy, _ := ioutil.ReadAll(resp.Body)
		return nil, errors.New(string(bdy))
	}

	scanResp := dtos.VirusTotalScanFileResponse{}

	if err := json.NewDecoder(resp.Body).Decode(&scanResp); err != nil {
		return nil, err
	}

	return transformers.TransformVirusTotalToSubmitResponse(&scanResp), nil
}

// GetSampleFileInfo returns the submitted sample file's info
func (v *VirusTotal) GetSampleFileInfo(sampleID string, filemetas ...dtos.FileMetaInfo) (*dtos.SampleInfo, error) {

	urlStr := v.BaseURL + fmt.Sprintf(viper.GetString("virustotal.report_endpoint"), viper.GetString("virustotal.api_key"), sampleID)

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
		bdy, _ := ioutil.ReadAll(resp.Body)
		bdyStr := ""
		if string(bdy) == "" {
			bdyStr = fmt.Sprintf("Status code received:%d with no body", resp.StatusCode)
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

	return transformers.TransformVirusTotalToSampleInfo(&sampleResp, fm), nil

}

// GetSubmissionStatus returns the submission status of a submitted sample
func (v *VirusTotal) GetSubmissionStatus(submissionID string) (*dtos.SubmissionStatusResponse, error) {

	urlStr := v.BaseURL + fmt.Sprintf(viper.GetString("virustotal.report_endpoint"), viper.GetString("virustotal.api_key"), submissionID)

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
