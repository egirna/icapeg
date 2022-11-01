package echo

import (
	"bytes"
	"icapeg/consts"
	"icapeg/logging"
	"io"
	"net/http"
	"strconv"
	"time"
)

// Processing is a func used for to processing the http message
func (e *Echo) Processing(partial bool) (int, interface{}, map[string]string, map[string]interface{},
	map[string]interface{}, map[string]interface{}) {
	serviceHeaders := make(map[string]string)
	msgHeadersBeforeProcessing := e.generalFunc.LogHTTPMsgHeaders(e.methodName)
	msgHeadersAfterProcessing := make(map[string]interface{})
	vendorMsgs := make(map[string]interface{})
	logging.Logger.Info(e.serviceName + " service has started processing")

	// no need to scan part of the file, this service needs all the file at ine time
	if partial {
		logging.Logger.Info(e.serviceName + " service has stopped processing partially")
		return utils.Continue, nil, nil, msgHeadersBeforeProcessing, msgHeadersAfterProcessing, vendorMsgs
	}
	isGzip := false

	//extracting the file from http message
	file, reqContentType, err := e.generalFunc.CopyingFileToTheBuffer(e.methodName)
	if err != nil {
		logging.Logger.Error(e.serviceName + " error: " + err.Error())
		logging.Logger.Info(e.serviceName + " service has stopped processing")
		return utils.InternalServerErrStatusCodeStr, nil, serviceHeaders,
			msgHeadersBeforeProcessing, msgHeadersAfterProcessing, vendorMsgs
	}

	//if the http method is Connect, return the request as it is because it has no body
	if e.httpMsg.Request.Method == http.MethodConnect {
		return utils.OkStatusCodeStr, e.generalFunc.ReturningHttpMessageWithFile(e.methodName, file.Bytes()),
			serviceHeaders, msgHeadersBeforeProcessing, msgHeadersAfterProcessing, vendorMsgs
	}

	//getting the extension of the file
	var contentType []string
	if len(contentType) == 0 {
		contentType = append(contentType, "")
	}
	var fileName string
	if e.methodName == utils.ICAPModeReq {
		contentType = e.httpMsg.Request.Header["Content-Type"]
		fileName = e.generalFunc.GetFileName()
	} else {
		contentType = e.httpMsg.Response.Header["Content-Type"]
		fileName = e.generalFunc.GetFileName()
	}
	if len(contentType) == 0 {
		contentType = append(contentType, "")
	}
	fileExtension := e.generalFunc.GetMimeExtension(file.Bytes(), contentType[0], fileName)

	//check if the file extension is a bypass extension
	//if yes we will not modify the file, and we will return 204 No modifications
	isProcess, icapStatus, httpMsg := e.generalFunc.CheckTheExtension(fileExtension, e.extArrs,
		e.processExts, e.rejectExts, e.bypassExts, e.return400IfFileExtRejected, isGzip,
		e.serviceName, e.methodName, EchoIdentifier, e.httpMsg.Request.RequestURI, reqContentType, file)
	if !isProcess {
		logging.Logger.Info(e.serviceName + " service has stopped processing")
		msgHeadersAfterProcessing = e.generalFunc.LogHTTPMsgHeaders(e.methodName)
		return icapStatus, httpMsg, serviceHeaders, msgHeadersBeforeProcessing,
			msgHeadersAfterProcessing, vendorMsgs
	}

	//check if the file size is greater than max file size of the service
	//if yes we will return 200 ok or 204 no modification, it depends on the configuration of the service
	if e.maxFileSize != 0 && e.maxFileSize < file.Len() {
		status, file, httpMsgAfter := e.generalFunc.IfMaxFileSizeExc(e.returnOrigIfMaxSizeExc, e.serviceName, e.methodName, file, e.maxFileSize)
		fileAfterPrep, httpMsgAfter := e.generalFunc.IfStatusIs204WithFile(e.methodName, status, file, isGzip, reqContentType, httpMsgAfter, true)
		if fileAfterPrep == nil && httpMsgAfter == nil {
			logging.Logger.Info(e.serviceName + " service has stopped processing")
			return utils.InternalServerErrStatusCodeStr, nil, serviceHeaders,
				msgHeadersBeforeProcessing, msgHeadersAfterProcessing, vendorMsgs
		}
		switch msg := httpMsgAfter.(type) {

		case *http.Request:
			msg.Body = io.NopCloser(bytes.NewBuffer(fileAfterPrep))
			logging.Logger.Info(e.serviceName + " service has stopped processing")
			msgHeadersAfterProcessing = e.generalFunc.LogHTTPMsgHeaders(e.methodName)
			return status, msg, nil, msgHeadersBeforeProcessing, msgHeadersAfterProcessing, vendorMsgs
		case *http.Response:
			msg.Body = io.NopCloser(bytes.NewBuffer(fileAfterPrep))
			logging.Logger.Info(e.serviceName + " service has stopped processing")
			msgHeadersAfterProcessing = e.generalFunc.LogHTTPMsgHeaders(e.methodName)
			return status, msg, nil, msgHeadersBeforeProcessing, msgHeadersAfterProcessing, vendorMsgs
		}
		msgHeadersAfterProcessing = e.generalFunc.LogHTTPMsgHeaders(e.methodName)
		return status, nil, nil, msgHeadersBeforeProcessing, msgHeadersAfterProcessing, vendorMsgs
	}

	scannedFile := file.Bytes()

	//returning the scanned file if everything is ok
	scannedFile = e.generalFunc.PreparingFileAfterScanning(scannedFile, reqContentType, e.methodName)
	msgHeadersAfterProcessing = e.generalFunc.LogHTTPMsgHeaders(e.methodName)
	logging.Logger.Info(e.serviceName + " service has stopped processing")
	return utils.OkStatusCodeStr, e.generalFunc.ReturningHttpMessageWithFile(e.methodName, scannedFile),
		serviceHeaders, msgHeadersBeforeProcessing, msgHeadersAfterProcessing, vendorMsgs
}

func (e *Echo) ISTagValue() string {
	epochTime := strconv.FormatInt(time.Now().Unix(), 10)
	return "epoch-" + epochTime
}
