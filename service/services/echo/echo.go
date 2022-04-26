package echo

import (
	"bytes"
	"icapeg/logger"
	"icapeg/readValues"
	"icapeg/service/services-utilities/general-functions"
	"icapeg/utils"
	"io"
	"net/http"
	"time"
)

// Echo represents the information regarding the Echo service
type Echo struct {
	httpMsg                *utils.HttpMsg
	elapsed                time.Duration
	serviceName            string
	methodName             string
	maxFileSize            int
	bypassExts             []string
	processExts            []string
	BaseURL                string
	Timeout                time.Duration
	APIKey                 string
	ScanEndpoint           string
	FailThreshold          int
	returnOrigIfMaxSizeExc bool
	returnOrigIf400        bool
	generalFunc            *general_functions.GeneralFunc
	logger                 *logger.ZLogger
}

// NewEchoService returns a new populated instance of the Echo service
func NewEchoService(serviceName, methodName string, httpMsg *utils.HttpMsg, elapsed time.Duration, logger *logger.ZLogger) *Echo {
	return &Echo{
		httpMsg:                httpMsg,
		elapsed:                elapsed,
		serviceName:            serviceName,
		methodName:             methodName,
		maxFileSize:            readValues.ReadValuesInt(serviceName + ".max_filesize"),
		bypassExts:             readValues.ReadValuesSlice(serviceName + ".bypass_extensions"),
		processExts:            readValues.ReadValuesSlice(serviceName + ".process_extensions"),
		BaseURL:                readValues.ReadValuesString(serviceName + ".base_url"),
		Timeout:                readValues.ReadValuesDuration(serviceName+".timeout") * time.Second,
		APIKey:                 readValues.ReadValuesString(serviceName + ".api_key"),
		ScanEndpoint:           readValues.ReadValuesString(serviceName + ".scan_endpoint"),
		FailThreshold:          readValues.ReadValuesInt(serviceName + ".fail_threshold"),
		returnOrigIfMaxSizeExc: readValues.ReadValuesBool(serviceName + ".return_original_if_max_file_size_exceeded"),
		generalFunc:            general_functions.NewGeneralFunc(httpMsg, elapsed, logger),
		logger:                 logger,
	}
}

//Processing is a func used for to processing the http message
func (e *Echo) Processing() (int, interface{}, map[string]string) {

	//extracting the file from http message
	file, reqContentType, err := e.generalFunc.CopyingFileToTheBuffer(e.methodName)
	if err != nil {
		return utils.InternalServerErrStatusCodeStr, nil, nil
	}

	//getting the extension of the file
	fileExtension := utils.GetMimeExtension(file.Bytes())

	//check if the file extension is a bypass extension
	//if yes we will not modify the file, and we will return 204 No modifications
	err = e.generalFunc.IfFileExtIsBypass(fileExtension, e.bypassExts)
	if err != nil {
		return utils.NoModificationStatusCodeStr,
			nil, nil
	}

	//check if the file extension is a bypass extension and not a process extension
	//if yes we will not modify the file, and we will return 204 No modifications
	err = e.generalFunc.IfFileExtIsBypassAndNotProcess(fileExtension, e.bypassExts, e.processExts)
	if err != nil {
		return utils.NoModificationStatusCodeStr,
			nil, nil
	}

	//check if the file size is greater than max file size of the service
	//if yes we will return 200 ok or 204 no modification, it depends on the configuration of the service
	if e.maxFileSize != 0 && e.maxFileSize < file.Len() {
		status, file, httpMsg := e.generalFunc.IfMaxFileSeizeExc(e.returnOrigIfMaxSizeExc, file, e.maxFileSize)
		fileAfterPrep, httpMsg := e.generalFunc.IfStatusIs204WithFile(e.methodName, status, file, false, reqContentType, httpMsg)
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

	//returning the scanned file if everything is ok
	scannedFile := e.generalFunc.PreparingFileAfterScanning(file.Bytes(), reqContentType, e.methodName)
	return utils.OkStatusCodeStr, e.generalFunc.ReturningHttpMessageWithFile(e.methodName, scannedFile), nil
}
