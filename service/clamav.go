package service

import (
	"errors"
	"fmt"
	"icapeg/dtos"
	"icapeg/helpers"
	"io"
	"time"

	"github.com/dutchcoders/go-clamd"
	"github.com/spf13/viper"
)

// the clamav constants
const (
	ClamavMalStatus = "FOUND"
)

// Clamav represents the informations regarding the clamav service
type Clamav struct {
	SocketPath  string
	WaitTimeOut time.Duration
}

// NewClamavService returns a new populated instance of the clamav service
func NewClamavService() LocalService {
	return &Clamav{
		SocketPath:  viper.GetString("clamav.socket_path"),
		WaitTimeOut: viper.GetDuration("clamav.wait_timeout") * time.Second,
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

	fileSizeStr := fmt.Sprintf("%.2fmb", helpers.ByteToMegaBytes(int(fileMetaInfo.FileSize)))

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
