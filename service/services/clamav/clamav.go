package clamav

import (
	"bytes"
	"fmt"
	"github.com/dutchcoders/go-clamd"
	"github.com/spf13/viper"
	"icapeg/readValues"
	general_functions "icapeg/service/services-utilities/general-functions"
	"net/http"
	"strconv"

	"icapeg/utils"
	"io"
	"time"
)

// the clamav constants
const (
	ClamavMalStatus = "FOUND"
)

// Clamav represents the informations regarding the clamav service
type Clamav struct {
	httpMsg                *utils.HttpMsg
	elapsed                time.Duration
	serviceName            string
	methodName             string
	maxFileSize            int
	bypassExts             []string
	processExts            []string
	SocketPath             string
	WaitTimeOut            time.Duration
	badFileStatus          []string
	okFileStatus           []string
	respSupported          bool
	reqSupported           bool
	returnOrigIfMaxSizeExc bool
	generalFunc            *general_functions.GeneralFunc
}

// NewClamavService returns a new populated instance of the clamav service
func NewClamavService(serviceName, methodName string, httpMsg *utils.HttpMsg) *Clamav {
	return &Clamav{
		httpMsg:                httpMsg,
		serviceName:            serviceName,
		methodName:             methodName,
		maxFileSize:            readValues.ReadValuesInt(serviceName + ".max_filesize"),
		bypassExts:             readValues.ReadValuesSlice(serviceName + ".bypass_extensions"),
		processExts:            readValues.ReadValuesSlice(serviceName + ".process_extensions"),
		SocketPath:             viper.GetString("clamav.socket_path"),
		WaitTimeOut:            viper.GetDuration("clamav.wait_timeout") * time.Second,
		badFileStatus:          viper.GetStringSlice("clamav.bad_file_status"),
		okFileStatus:           viper.GetStringSlice("clamav.ok_file_status"),
		respSupported:          true,
		reqSupported:           false,
		returnOrigIfMaxSizeExc: readValues.ReadValuesBool(serviceName + ".return_original_if_max_file_size_exceeded"),
		generalFunc:            general_functions.NewGeneralFunc(httpMsg),
	}
}

//// ScanFileStream scans a file stream using clamav
//func (c *Clamav) ScanFileStream(file io.Reader, fileMetaInfo dtos.FileMetaInfo) (*dtos.SampleInfo, error) {
//
//	clmd := clamd.NewClamd(c.SocketPath)
//
//	response, err := clmd.ScanStream(file, make(chan bool))
//
//	if err != nil {
//		return nil, err
//	}
//
//	result := &clamd.ScanResult{}
//	scanFinished := false
//
//	go func() {
//		for s := range response {
//			result = s
//		}
//		scanFinished = true
//	}()
//
//	time.Sleep(c.WaitTimeOut)
//
//	if !scanFinished {
//		return nil, errors.New("scanning time out")
//	}
//
//	severity := "ok"
//
//	if result.Status == ClamavMalStatus {
//		severity = "malicious"
//	}
//
//	fileSizeStr := fmt.Sprintf("%.2fmb", utils.ByteToMegaBytes(int(fileMetaInfo.FileSize)))
//
//	si := &dtos.SampleInfo{
//		FileName:           fileMetaInfo.FileName,
//		SampleType:         fileMetaInfo.FileType,
//		SampleSeverity:     severity,
//		FileSizeStr:        fileSizeStr,
//		VTIScore:           "N/A",
//		SubmissionFinished: scanFinished,
//	}
//
//	return si, nil
//}

func (c *Clamav) Processing(partial bool) (int, interface{}, map[string]string) {
	isGzip := false

	//extracting the file from http message
	file, reqContentType, err := c.generalFunc.CopyingFileToTheBuffer(c.methodName)
	if err != nil {
		return utils.InternalServerErrStatusCodeStr, nil, nil
	}

	//getting the extension of the file
	fileExtension := utils.GetMimeExtension(file.Bytes())

	//check if the file extension is a bypass extension
	//if yes we will not modify the file, and we will return 204 No modifications
	err = c.generalFunc.IfFileExtIsBypass(fileExtension, c.bypassExts)
	if err != nil {
		return utils.NoModificationStatusCodeStr,
			nil, nil
	}

	//check if the file extension is a bypass extension and not a process extension
	//if yes we will not modify the file, and we will return 204 No modifications
	err = c.generalFunc.IfFileExtIsBypassAndNotProcess(fileExtension, c.bypassExts, c.processExts)
	if err != nil {
		return utils.NoModificationStatusCodeStr,
			nil, nil
	}

	//check if the file size is greater than max file size of the service
	//if yes we will return 200 ok or 204 no modification, it depends on the configuration of the service
	if c.maxFileSize != 0 && c.maxFileSize < file.Len() {
		status, file, httpMsg := c.generalFunc.IfMaxFileSeizeExc(c.returnOrigIfMaxSizeExc, file, c.maxFileSize)
		fileAfterPrep, httpMsg := c.generalFunc.IfStatusIs204WithFile(c.methodName, status, file, isGzip, reqContentType, httpMsg)
		if fileAfterPrep == nil && httpMsg == nil {
			return utils.InternalServerErrStatusCodeStr, nil, nil
		}
		switch msg := httpMsg.(type) {
		case *http.Request:
			msg.Body = io.NopCloser(bytes.NewBuffer(fileAfterPrep))
			return status, msg, nil
		case *http.Response:
			msg.Body = io.NopCloser(bytes.NewBuffer(fileAfterPrep))
			return status, msg, nil
		}
		return status, nil, nil
	}

	// function to reverse string

	//check if the body of the http message is compressed in Gzip or not
	isGzip = c.generalFunc.IsBodyGzipCompressed(c.methodName)
	//if it's compressed, we decompress it to send it to Glasswall service
	if isGzip {
		if file, err = c.generalFunc.DecompressGzipBody(file); err != nil {
			return utils.InternalServerErrStatusCodeStr, nil, nil
		}
	}

	clamav := clamd.NewClamd("/tmp/clamd.socket")

	reader := bytes.NewReader(clamd.EICAR)
	response, err := clamav.ScanStream(reader, make(chan bool))

	for s := range response {
		fmt.Printf("%v %v\n", s, err)
	}

	return 500, nil, nil
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

func (c *Clamav) ISTagValue() string {
	epochTime := strconv.FormatInt(time.Now().Unix(), 10)
	return "epoch-" + epochTime
}
