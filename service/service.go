package service

import (
	"bytes"
	"icapeg/dtos"
)

// The service names
const (
	SVCVirusTotal   = "virustotal"
	SVCMetaDefender = "metadefender"
)

type (
	// Service holds the info to distinguish a service
	Service interface {
		SubmitFile(*bytes.Buffer, string) (*dtos.SubmitResponse, error)
		GetSubmissionStatus(string) (*dtos.SubmissionStatusResponse, error)
		GetSampleFileInfo(string, ...dtos.FileMetaInfo) (*dtos.SampleInfo, error)
	}
)

// GetService returns a service based on the service name
func GetService(name string) Service {
	switch name {
	case SVCVirusTotal:
		return NewVirusTotalService()
	case SVCMetaDefender:
		return NewMetaDefenderService()
	}

	return nil
}
