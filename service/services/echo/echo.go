package echo

import (
	"bytes"
	"fmt"
	utils "icapeg/consts"
	"icapeg/logging"
	"icapeg/service/services-utilities/ContentTypes"
	"io"
	"net/http"
	"net/textproto"
	"strconv"
	"time"
)

// Processing is a func used for to processing the http message
func (e *Echo) Processing(partial bool, IcapHeader textproto.MIMEHeader) (int, interface{}, map[string]string, map[string]interface{},
	map[string]interface{}, map[string]interface{}) {
	serviceHeaders := make(map[string]string)
	serviceHeaders["X-ICAP-Metadata"] = e.xICAPMetadata
	msgHeadersBeforeProcessing := e.generalFunc.LogHTTPMsgHeaders(e.methodName)
	msgHeadersAfterProcessing := make(map[string]interface{})
	vendorMsgs := make(map[string]interface{})
	logging.Logger.Info(utils.PrepareLogMsg(e.xICAPMetadata, e.serviceName+" service has started processing"))

	// no need to scan part of the file, this service needs all the file at ine time
	if partial {
		logging.Logger.Info(utils.PrepareLogMsg(e.xICAPMetadata,
			e.serviceName+" service has stopped processing partially"))
		return utils.Continue, nil, nil, msgHeadersBeforeProcessing, msgHeadersAfterProcessing, vendorMsgs
	}
	isGzip := false

	//extracting the file from http message
	file, reqContentType, err := e.CopyingFileToTheBuffer(e.methodName)
	if err != nil {
		logging.Logger.Error(utils.PrepareLogMsg(e.xICAPMetadata, e.serviceName+" error: "+err.Error()))
		logging.Logger.Info(utils.PrepareLogMsg(e.xICAPMetadata, e.serviceName+" service has stopped processing"))
		return utils.InternalServerErrStatusCodeStr, nil, serviceHeaders,
			msgHeadersBeforeProcessing, msgHeadersAfterProcessing, vendorMsgs
	}

	//if the http method is Connect, return the request as it is because it has no body
	if e.httpMsg.Request.Method == http.MethodConnect {
		return utils.OkStatusCodeStr, e.generalFunc.ReturningHttpMessageWithFile(e.methodName, file),
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
		fileName = e.generalFunc.GetFileName(e.serviceName, e.xICAPMetadata)
	} else {
		contentType = e.httpMsg.Response.Header["Content-Type"]
		fileName = e.generalFunc.GetFileName(e.serviceName, e.xICAPMetadata)
	}
	if len(contentType) == 0 {
		contentType = append(contentType, "")
	}
	fileExtension := e.generalFunc.GetMimeExtension(file, contentType[0], fileName, true)
	fileSize := fmt.Sprintf("%v kb", len(file)/1000)

	//check if the file extension is a bypass extension
	//if yes we will not modify the file, and we will return 204 No modifications
	isProcess, icapStatus, httpMsg := e.generalFunc.CheckTheExtension(fileExtension, e.extArrs,
		e.processExts, e.rejectExts, e.bypassExts, e.return400IfFileExtRejected, isGzip,
		e.serviceName, e.methodName, EchoIdentifier, e.httpMsg.Request.RequestURI, reqContentType, bytes.NewBuffer(file), utils.BlockPagePath, fileSize)
	if !isProcess && e.methodName == utils.ICAPModeReq {
		logging.Logger.Info(utils.PrepareLogMsg(e.xICAPMetadata, e.serviceName+" service has stopped processing"))
		msgHeadersAfterProcessing = e.generalFunc.LogHTTPMsgHeaders(e.methodName)
		return icapStatus, httpMsg, serviceHeaders, msgHeadersBeforeProcessing,
			msgHeadersAfterProcessing, vendorMsgs
	}

	//check if the file size is greater than max file size of the service
	//if yes we will return 200 ok or 204 no modification, it depends on the configuration of the service
	if e.maxFileSize != 0 && e.maxFileSize < len(file) && isProcess {
		status, file, httpMsgAfter := e.generalFunc.IfMaxFileSizeExc(e.returnOrigIfMaxSizeExc, e.serviceName, e.methodName, bytes.NewBuffer(file), e.maxFileSize, utils.BlockPagePath, fileSize)
		fileAfterPrep, httpMsgAfter := e.generalFunc.IfStatusIs204WithFile(e.methodName, status, file, isGzip, reqContentType, httpMsgAfter, true)
		if fileAfterPrep == nil && httpMsgAfter == nil {
			logging.Logger.Info(utils.PrepareLogMsg(e.xICAPMetadata, e.serviceName+" service has stopped processing"))
			return utils.InternalServerErrStatusCodeStr, nil, serviceHeaders,
				msgHeadersBeforeProcessing, msgHeadersAfterProcessing, vendorMsgs
		}
		switch msg := httpMsgAfter.(type) {

		case *http.Request:
			msg.Body = io.NopCloser(bytes.NewBuffer(fileAfterPrep))
			logging.Logger.Info(utils.PrepareLogMsg(e.xICAPMetadata, e.serviceName+" service has stopped processing"))
			msgHeadersAfterProcessing = e.generalFunc.LogHTTPMsgHeaders(e.methodName)
			return status, msg, nil, msgHeadersBeforeProcessing, msgHeadersAfterProcessing, vendorMsgs
		case *http.Response:
			msg.Body = io.NopCloser(bytes.NewBuffer(fileAfterPrep))
			logging.Logger.Info(utils.PrepareLogMsg(e.xICAPMetadata, e.serviceName+" service has stopped processing"))
			msgHeadersAfterProcessing = e.generalFunc.LogHTTPMsgHeaders(e.methodName)
			return status, msg, nil, msgHeadersBeforeProcessing, msgHeadersAfterProcessing, vendorMsgs
		}
		msgHeadersAfterProcessing = e.generalFunc.LogHTTPMsgHeaders(e.methodName)
		return status, nil, nil, msgHeadersBeforeProcessing, msgHeadersAfterProcessing, vendorMsgs
	}

	scannedFile := file

	//returning the scanned file if everything is ok
	scannedFile = e.generalFunc.PreparingFileAfterScanning(scannedFile, reqContentType, e.methodName)
	msgHeadersAfterProcessing = e.generalFunc.LogHTTPMsgHeaders(e.methodName)
	logging.Logger.Info(utils.PrepareLogMsg(e.xICAPMetadata, e.serviceName+" service has stopped processing"))
	return utils.OkStatusCodeStr, e.generalFunc.ReturningHttpMessageWithFile(e.methodName, scannedFile),
		serviceHeaders, msgHeadersBeforeProcessing, msgHeadersAfterProcessing, vendorMsgs
}

func (e *Echo) ISTagValue() string {
	epochTime := strconv.FormatInt(time.Now().Unix(), 10)
	return "epoch-" + epochTime
}

// CopyingFileToTheBuffer is a func which used for extracting a file from the body of the http message
func (e *Echo) CopyingFileToTheBuffer(methodName string) ([]byte, ContentTypes.ContentType, error) {
	logging.Logger.Info(utils.PrepareLogMsg(e.xICAPMetadata, "extracting the body of HTTP message"))
	var file []byte
	var err error
	var reqContentType ContentTypes.ContentType
	reqContentType = nil
	switch methodName {
	case utils.ICAPModeReq:
		file, reqContentType, err = e.copyingFileToTheBufferReq()
	case utils.ICAPModeResp:
		file, err = e.copyingFileToTheBufferResp()
	}
	if err != nil {
		return nil, nil, err
	}
	return file, reqContentType, nil
}

// copyingFileToTheBufferResp is a utility function for CopyingFileToTheBuffer func
// it's used for extracting a file from the body of the http response
func (e *Echo) copyingFileToTheBufferResp() ([]byte, error) {
	file, err := e.httpMsg.StorageClient.Load(e.httpMsg.StorageKey)
	if err != nil {
		return file, err
	}
	return file, nil
}

// copyingFileToTheBufferReq is a utility function for CopyingFileToTheBuffer func
// it's used for extracting a file from the body of the http request
func (e *Echo) copyingFileToTheBufferReq() ([]byte, ContentTypes.ContentType, error) {
	reqContentType := ContentTypes.GetContentType(e.httpMsg.Request)
	// getting the file from request and store it in buf as a type of bytes.Buffer
	file := reqContentType.GetFileFromRequest()
	return file.Bytes(), reqContentType, nil

}
