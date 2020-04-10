package service

import (
	"bytes"
	"icapeg/dtos"
	"io"
	"time"
)

// The service names
const (
	SVCVirusTotal   = "virustotal"
	SVCMetaDefender = "metadefender"
	SVCVmray        = "vmray"
	SVCClamav       = "clamav"
)

type (
	// Service holds the info to distinguish a service
	Service interface {
		SubmitFile(*bytes.Buffer, string) (*dtos.SubmitResponse, error)
		GetSubmissionStatus(string) (*dtos.SubmissionStatusResponse, error)
		GetSampleFileInfo(string, ...dtos.FileMetaInfo) (*dtos.SampleInfo, error)
		GetSampleURLInfo(string, ...dtos.FileMetaInfo) (*dtos.SampleInfo, error)
		SubmitURL(string, string) (*dtos.SubmitResponse, error)
		GetStatusCheckInterval() time.Duration
		GetStatusCheckTimeout() time.Duration
		GetBadFileStatus() []string
		GetOkFileStatus() []string
		StatusEndpointExists() bool
		IsLocalScanner() bool
	}

	// LocalService holds the blueprint of a local service
	LocalService interface {
		ScanFileStream(io.Reader, dtos.FileMetaInfo) (*dtos.SampleInfo, error)
		GetBadFileStatus() []string
		GetOkFileStatus() []string
		IsLocalScanner() bool
	}
)

// GetService returns a service based on the service name
func GetService(name string) Service {
	switch name {
	case SVCVirusTotal:
		return NewVirusTotalService()
	case SVCMetaDefender:
		return NewMetaDefenderService()
	case SVCVmray:
		return NewVmrayService()
	}

	return nil
}

// GetLocalService returns a local service based on the name
func GetLocalService(name string) LocalService {
	switch name {
	case SVCClamav:
		return NewClamavService()
	}

	return nil
}
