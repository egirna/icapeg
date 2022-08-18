package echo

import (
	"bytes"
	"icapeg/utils"
	"io"
	"net/http"
	"strconv"
	"time"
)

// Processing is a func used for to processing the http message
func (e *Echo) Processing(partial bool) (int, interface{}, map[string]string) {
	serviceHeaders := make(map[string]string)
	// no need to scan part of the file, this service needs all the file at ine time
	if partial {
		return utils.Continue, nil, nil
	}

	isGzip := false

	//extracting the file from http message
	file, reqContentType, err := e.generalFunc.CopyingFileToTheBuffer(e.methodName)
	if err != nil {
		return utils.InternalServerErrStatusCodeStr, nil, serviceHeaders
	}
	//getting the extension of the file
	var contentType []string
	if len(contentType) == 0 {
		contentType = append(contentType, "")
	}
	var fileName string
	if e.methodName == utils.ICAPModeReq {
		contentType = e.httpMsg.Request.Header["Content-Type"]
		fileName = utils.GetFileName(e.httpMsg.Request)
	} else {
		contentType = e.httpMsg.Response.Header["Content-Type"]
		fileName = utils.GetFileName(e.httpMsg.Response)
	}
	if len(contentType) == 0 {
		contentType = append(contentType, "")
	}
	fileExtension := utils.GetMimeExtension(file.Bytes(), contentType[0], fileName)

	//check if the file extension is a bypass extension
	//if yes we will not modify the file, and we will return 204 No modifications
	isProcess, icapStatus, httpMsg := e.generalFunc.CheckTheExtension(fileExtension, e.extArrs,
		e.processExts, e.rejectExts, e.bypassExts, e.return400IfFileExtRejected, isGzip,
		e.serviceName, e.methodName, EchoIdentifier, e.httpMsg.Request.RequestURI, reqContentType, file)
	if !isProcess {
		return icapStatus, httpMsg, serviceHeaders
	}

	//check if the file size is greater than max file size of the service
	//if yes we will return 200 ok or 204 no modification, it depends on the configuration of the service
	if e.maxFileSize != 0 && e.maxFileSize < file.Len() {
		status, file, httpMsg := e.generalFunc.IfMaxFileSeizeExc(e.returnOrigIfMaxSizeExc, e.serviceName, file, e.maxFileSize)
		fileAfterPrep, httpMsg := e.generalFunc.IfStatusIs204WithFile(e.methodName, status, file, isGzip, reqContentType, httpMsg)
		if fileAfterPrep == nil && httpMsg == nil {
			return utils.InternalServerErrStatusCodeStr, nil, serviceHeaders
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

	scannedFile := file.Bytes()

	//returning the scanned file if everything is ok
	scannedFile = e.generalFunc.PreparingFileAfterScanning(scannedFile, reqContentType, e.methodName)
	return utils.OkStatusCodeStr, e.generalFunc.ReturningHttpMessageWithFile(e.methodName, scannedFile), serviceHeaders
}

func (e *Echo) ISTagValue() string {
	epochTime := strconv.FormatInt(time.Now().Unix(), 10)
	return "epoch-" + epochTime
}
