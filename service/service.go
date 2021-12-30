package service

import (
	"bytes"
	"icapeg/config"
	"icapeg/dtos"
	"icapeg/logger"
	"io"
	"net/http"
	"time"

	ic "icapeg/icap-client"
)

// The service names
const (
	SVCVirusTotal   = "virustotal"
	SVCMetaDefender = "metadefender"
	SVCVmray        = "vmray"
	SVCClamav       = "clamav"
	SVCGlasswall    = "glasswall"
)

type (
	// Service holds the info to distinguish a service
	Service interface {
		SubmitFile(*bytes.Buffer, string) (*dtos.SubmitResponse, error)
		SendFileApi(*bytes.Buffer, string) (*http.Response, error)
		GetSubmissionStatus(string) (*dtos.SubmissionStatusResponse, error)
		GetSampleFileInfo(string, ...dtos.FileMetaInfo) (*dtos.SampleInfo, error)
		GetSampleURLInfo(string, ...dtos.FileMetaInfo) (*dtos.SampleInfo, error)
		SubmitURL(string, string) (*dtos.SubmitResponse, error)
		GetStatusCheckInterval() time.Duration
		GetStatusCheckTimeout() time.Duration
		GetBadFileStatus() []string
		GetOkFileStatus() []string
		StatusEndpointExists() bool
		RespSupported() bool
		ReqSupported() bool
	}

	// LocalService holds the blueprint of a local service
	LocalService interface {
		ScanFileStream(io.Reader, dtos.FileMetaInfo) (*dtos.SampleInfo, error)
		GetBadFileStatus() []string
		GetOkFileStatus() []string
		RespSupported() bool
		ReqSupported() bool
	}

	// ICAPService holds the blueprint of a Remote ICAP service
	ICAPService interface {
		DoReqmod() (*ic.Response, error)
		DoRespmod() (*ic.Response, error)
		DoOptions() (*ic.Response, error)
		GetURL() string
		GetRespmodEndpoint() string
		GetReqmodEndpoint() string
		GetOptionsEndpoint() string
		GetTimeout() time.Duration
		SetHTTPRequest(*http.Request)
		SetHTTPResponse(*http.Response)
		SetHeader(map[string][]string)
		ChangeOptionsEndpoint(string)
	}
)

var (
	errorLogger = logger.NewLogger(logger.LogLevelError, logger.LogLevelDebug)
)

// IsServiceLocal determines if a service is local or not
func IsServiceLocal(vendor string, serviceName string) bool {
	svc := GetService(vendor, serviceName)

	if svc != nil {
		return false
	}

	lsvc := GetLocalService(vendor, serviceName)

	if lsvc != nil {
		return true
	}

	return false
}

// GetService returns a service based on the service name
//change name to vendor and add parameter service name
func GetService(vendor string, serviceName string) Service {
	switch vendor {
	case SVCVirusTotal:
		return NewVirusTotalService(serviceName)
	case SVCMetaDefender:
		return NewMetaDefenderService(serviceName)
	case SVCVmray:
		return NewVmrayService(serviceName)
	case SVCGlasswall:
		return NewGlasswallService(serviceName)
	}
	return nil
}

// GetLocalService returns a local service based on the name
func GetLocalService(vendor string, serviceName string) LocalService {
	switch vendor {
	case SVCClamav:
		return NewClamavService(serviceName)
	}

	return nil
}

// GetICAPService returns a remote ICAP service based on the name
func GetICAPService(name string) ICAPService {
	if config.App().LogLevel == logger.LogLevelDebug {
		ic.SetDebugMode(true)
		ic.SetDebugOutput(logger.LogFile())
	}
	return NewRemoteICAPService(name)
}
