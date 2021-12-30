package service

import (
	"errors"
	"fmt"
	"icapeg/dtos"
	"icapeg/readValues"
	"icapeg/utils"
	"io"
	"time"

	"github.com/dutchcoders/go-clamd"
)

// the clamav constants
const (
	ClamavMalStatus = "FOUND"
)

// Clamav represents the informations regarding the clamav service
type Clamav struct {
	SocketPath    string
	WaitTimeOut   time.Duration
	badFileStatus []string
	okFileStatus  []string
	respSupported bool
	reqSupported  bool
}

// NewClamavService returns a new populated instance of the clamav service
func NewClamavService(serviceName string) LocalService {
	return &Clamav{
		SocketPath:    readValues.ReadValuesString(serviceName + "socket_path"),
		WaitTimeOut:   readValues.ReadValuesDuration(serviceName+"wait_timeout") * time.Second,
		badFileStatus: readValues.ReadValuesSlice(serviceName + "bad_file_status"),
		okFileStatus:  readValues.ReadValuesSlice(serviceName + "ok_file_status"),
		respSupported: true,
		reqSupported:  false,
	}
}

// ScanFileStream scans a file stream using clamav
func (c *Clamav) ScanFileStream(file io.Reader, fileMetaInfo dtos.FileMetaInfo) (*dtos.SampleInfo, error) {

	clmd := clamd.NewClamd(c.SocketPath)

	response, err := clmd.ScanStream(file, make(chan bool))

	if err != nil {
		return nil, err
	}

	result := &clamd.ScanResult{}
	scanFinished := false

	go func() {
		for s := range response {
			result = s
		}
		scanFinished = true
	}()

	time.Sleep(c.WaitTimeOut)

	if !scanFinished {
		return nil, errors.New("Scanning time out")
	}

	severity := "ok"

	if result.Status == ClamavMalStatus {
		severity = "malicious"
	}

	fileSizeStr := fmt.Sprintf("%.2fmb", utils.ByteToMegaBytes(int(fileMetaInfo.FileSize)))

	si := &dtos.SampleInfo{
		FileName:           fileMetaInfo.FileName,
		SampleType:         fileMetaInfo.FileType,
		SampleSeverity:     severity,
		FileSizeStr:        fileSizeStr,
		VTIScore:           "N/A",
		SubmissionFinished: scanFinished,
	}

	return si, nil
}

// GetBadFileStatus returns the bad_file_status slice of the service
func (c *Clamav) GetBadFileStatus() []string {
	return c.badFileStatus
}

// GetOkFileStatus returns the ok_file_status slice of the service
func (c *Clamav) GetOkFileStatus() []string {
	return c.okFileStatus
}

// RespSupported returns the respSupported field of the service
func (c *Clamav) RespSupported() bool {
	return c.respSupported
}

// ReqSupported returns the reqSupported field of the service
func (c *Clamav) ReqSupported() bool {
	return c.reqSupported
}
