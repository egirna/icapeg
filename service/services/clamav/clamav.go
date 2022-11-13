package clamav

import (
	"bytes"
	"github.com/dutchcoders/go-clamd"
	"icapeg/consts"
	"icapeg/logging"
	"io"
	"net/http"
	"strconv"
	"time"
)

func (c *Clamav) Processing(partial bool) (int, interface{}, map[string]string, map[string]interface{},
	map[string]interface{}, map[string]interface{}) {
	logging.Logger.Info(utils.PrepareLogMsg(c.xICAPMetadata, c.serviceName+" service has started processing"))
	serviceHeaders := make(map[string]string)
	serviceHeaders["X-ICAP-Metadata"] = c.xICAPMetadata
	msgHeadersBeforeProcessing := c.generalFunc.LogHTTPMsgHeaders(c.methodName)
	msgHeadersAfterProcessing := make(map[string]interface{})
	vendorMsgs := make(map[string]interface{})
	// no need to scan part of the file, this service needs all the file at ine time
	if partial {
		logging.Logger.Info(utils.PrepareLogMsg(c.xICAPMetadata,
			c.serviceName+" service has stopped processing partially"))
		return utils.Continue, nil, nil, msgHeadersBeforeProcessing, msgHeadersAfterProcessing, vendorMsgs
	}

	isGzip := false
	c.generalFunc.LogHTTPMsgHeaders(c.methodName)
	logging.Logger.Info(utils.PrepareLogMsg(c.xICAPMetadata, c.serviceName+" service has started processing"))

	//extracting the file from http message
	file, reqContentType, err := c.generalFunc.CopyingFileToTheBuffer(c.methodName)
	if err != nil {
		logging.Logger.Error(utils.PrepareLogMsg(c.xICAPMetadata, c.serviceName+" error: "+err.Error()))
		logging.Logger.Info(utils.PrepareLogMsg(c.xICAPMetadata, c.serviceName+" service has stopped processing"))
		return utils.InternalServerErrStatusCodeStr, nil, serviceHeaders, msgHeadersBeforeProcessing,
			msgHeadersAfterProcessing, vendorMsgs
	}

	//if the http method is Connect, return the request as it is because it has no body
	if c.httpMsg.Request.Method == http.MethodConnect {
		return utils.OkStatusCodeStr, c.generalFunc.ReturningHttpMessageWithFile(c.methodName, file.Bytes()),
			serviceHeaders, msgHeadersBeforeProcessing, msgHeadersAfterProcessing, vendorMsgs
	}

	//getting the extension of the file
	var contentType []string
	if len(contentType) == 0 {
		contentType = append(contentType, "")
	}
	var fileName string
	if c.methodName == utils.ICAPModeReq {
		contentType = c.httpMsg.Request.Header["Content-Type"]
		fileName = c.generalFunc.GetFileName()
	} else {
		contentType = c.httpMsg.Response.Header["Content-Type"]
		fileName = c.generalFunc.GetFileName()
	}
	if len(contentType) == 0 {
		contentType = append(contentType, "")
	}
	fileExtension := c.generalFunc.GetMimeExtension(file.Bytes(), contentType[0], fileName)

	//check if the file extension is a bypass extension
	//if yes we will not modify the file, and we will return 204 No modifications
	isProcess, icapStatus, httpMsg := c.generalFunc.CheckTheExtension(fileExtension, c.extArrs,
		c.processExts, c.rejectExts, c.bypassExts, c.return400IfFileExtRejected, isGzip,
		c.serviceName, c.methodName, ClamavIdentifier, c.httpMsg.Request.RequestURI, reqContentType, file)
	if !isProcess {
		logging.Logger.Info(utils.PrepareLogMsg(c.xICAPMetadata, c.serviceName+" service has stopped processing"))
		msgHeadersAfterProcessing = c.generalFunc.LogHTTPMsgHeaders(c.methodName)
		return icapStatus, httpMsg, serviceHeaders, msgHeadersBeforeProcessing,
			msgHeadersAfterProcessing, vendorMsgs
	}

	//check if the file size is greater than max file size of the service
	//if yes we will return 200 ok or 204 no modification, it depends on the configuration of the service
	if c.maxFileSize != 0 && c.maxFileSize < file.Len() {
		status, file, httpMsg := c.generalFunc.IfMaxFileSizeExc(c.returnOrigIfMaxSizeExc, c.serviceName, c.methodName, file, c.maxFileSize)
		fileAfterPrep, httpMsg := c.generalFunc.IfStatusIs204WithFile(c.methodName, status, file, isGzip, reqContentType, httpMsg, true)
		if fileAfterPrep == nil && httpMsg == nil {
			logging.Logger.Info(utils.PrepareLogMsg(c.xICAPMetadata, c.serviceName+" service has stopped processing"))
			return utils.InternalServerErrStatusCodeStr, nil, serviceHeaders, msgHeadersBeforeProcessing,
				msgHeadersAfterProcessing, vendorMsgs
		}
		switch msg := httpMsg.(type) {
		case *http.Request:
			msg.Body = io.NopCloser(bytes.NewBuffer(fileAfterPrep))
			logging.Logger.Info(utils.PrepareLogMsg(c.xICAPMetadata, c.serviceName+" service has stopped processing"))
			msgHeadersAfterProcessing = c.generalFunc.LogHTTPMsgHeaders(c.methodName)
			return status, msg, nil, msgHeadersBeforeProcessing,
				msgHeadersAfterProcessing, vendorMsgs
		case *http.Response:
			msg.Body = io.NopCloser(bytes.NewBuffer(fileAfterPrep))
			logging.Logger.Info(utils.PrepareLogMsg(c.xICAPMetadata,
				c.serviceName+" service has stopped processing"))
			msgHeadersAfterProcessing = c.generalFunc.LogHTTPMsgHeaders(c.methodName)
			return status, msg, nil, msgHeadersBeforeProcessing,
				msgHeadersAfterProcessing, vendorMsgs
		}
		msgHeadersAfterProcessing = c.generalFunc.LogHTTPMsgHeaders(c.methodName)
		return status, nil, nil, msgHeadersBeforeProcessing,
			msgHeadersAfterProcessing, vendorMsgs
	}
	clmd := clamd.NewClamd(c.SocketPath)
	logging.Logger.Debug(utils.PrepareLogMsg(c.xICAPMetadata,
		"sending the HTTP msg body to the ClamAV through antivirus socket"))
	response, err := clmd.ScanStream(bytes.NewReader(file.Bytes()), make(chan bool))
	if err != nil {
		logging.Logger.Error(utils.PrepareLogMsg(c.xICAPMetadata, c.serviceName+" error: "+err.Error()))
		logging.Logger.Info(utils.PrepareLogMsg(c.xICAPMetadata, c.serviceName+" service has stopped processing"))
		return utils.InternalServerErrStatusCodeStr, nil, serviceHeaders, msgHeadersBeforeProcessing,
			msgHeadersAfterProcessing, vendorMsgs
	}

	result := &clamd.ScanResult{}
	scanFinished := false

	go func() {
		for s := range response {
			result = s
		}
		scanFinished = true
	}()

	time.Sleep(5 * time.Second)

	if !scanFinished {
		logging.Logger.Error(utils.PrepareLogMsg(c.xICAPMetadata, c.serviceName+" error: scanning timeout"))
		logging.Logger.Info(utils.PrepareLogMsg(c.xICAPMetadata, c.serviceName+" service has stopped processing"))
		return utils.RequestTimeOutStatusCodeStr, nil, serviceHeaders, msgHeadersBeforeProcessing,
			msgHeadersAfterProcessing, vendorMsgs
	}

	if result.Status == ClamavMalStatus {
		logging.Logger.Debug(utils.PrepareLogMsg(c.xICAPMetadata, c.serviceName+"File is not safe"))
		if c.methodName == utils.ICAPModeResp {
			errPage := c.generalFunc.GenHtmlPage(utils.BlockPagePath, utils.ErrPageReasonFileIsNotSafe, c.serviceName, "CLAMAV ID", c.httpMsg.Request.RequestURI)
			c.httpMsg.Response = c.generalFunc.ErrPageResp(http.StatusForbidden, errPage.Len())
			c.httpMsg.Response.Body = io.NopCloser(bytes.NewBuffer(errPage.Bytes()))
			logging.Logger.Info(utils.PrepareLogMsg(c.xICAPMetadata, c.serviceName+" service has stopped processing"))
			msgHeadersAfterProcessing = c.generalFunc.LogHTTPMsgHeaders(c.methodName)
			return utils.OkStatusCodeStr, c.httpMsg.Response, serviceHeaders, msgHeadersBeforeProcessing,
				msgHeadersAfterProcessing, vendorMsgs
		} else {
			htmlPage, req, err := c.generalFunc.ReqModErrPage(utils.ErrPageReasonFileIsNotSafe, c.serviceName, ClamavIdentifier)
			if err != nil {
				return utils.InternalServerErrStatusCodeStr, nil, nil, msgHeadersBeforeProcessing,
					msgHeadersAfterProcessing, vendorMsgs
			}
			req.Body = io.NopCloser(htmlPage)
			msgHeadersAfterProcessing = c.generalFunc.LogHTTPMsgHeaders(c.methodName)
			return utils.OkStatusCodeStr, req, serviceHeaders, msgHeadersBeforeProcessing,
				msgHeadersAfterProcessing, vendorMsgs
		}
	}
	logging.Logger.Debug(utils.PrepareLogMsg(c.xICAPMetadata, c.serviceName+"File is safe"))
	//returning the scanned file if everything is ok
	fileAfterPrep, httpMsg := c.generalFunc.IfICAPStatusIs204(c.methodName, utils.NoModificationStatusCodeStr,
		file, false, reqContentType, c.httpMsg)
	if fileAfterPrep == nil && httpMsg == nil {
		logging.Logger.Info(utils.PrepareLogMsg(c.xICAPMetadata, c.serviceName+" service has stopped processing"))
		return utils.InternalServerErrStatusCodeStr, nil, nil, msgHeadersBeforeProcessing,
			msgHeadersAfterProcessing, vendorMsgs
	}

	//returning the http message and the ICAP status code
	switch msg := httpMsg.(type) {
	case *http.Request:
		msg.Body = io.NopCloser(bytes.NewBuffer(fileAfterPrep))
		logging.Logger.Info(utils.PrepareLogMsg(c.xICAPMetadata, c.serviceName+" service has stopped processing"))
		msgHeadersAfterProcessing = c.generalFunc.LogHTTPMsgHeaders(c.methodName)
		return utils.NoModificationStatusCodeStr, msg, serviceHeaders, msgHeadersBeforeProcessing,
			msgHeadersAfterProcessing, vendorMsgs
	case *http.Response:
		msg.Body = io.NopCloser(bytes.NewBuffer(fileAfterPrep))
		logging.Logger.Info(utils.PrepareLogMsg(c.xICAPMetadata, c.serviceName+" service has stopped processing"))
		msgHeadersAfterProcessing = c.generalFunc.LogHTTPMsgHeaders(c.methodName)
		return utils.NoModificationStatusCodeStr, msg, serviceHeaders, msgHeadersBeforeProcessing,
			msgHeadersAfterProcessing, vendorMsgs
	}
	c.generalFunc.LogHTTPMsgHeaders(c.methodName)
	logging.Logger.Info(utils.PrepareLogMsg(c.xICAPMetadata, c.serviceName+" service has stopped processing"))
	msgHeadersAfterProcessing = c.generalFunc.LogHTTPMsgHeaders(c.methodName)
	return utils.NoModificationStatusCodeStr, nil, serviceHeaders, msgHeadersBeforeProcessing,
		msgHeadersAfterProcessing, vendorMsgs
}

func (c *Clamav) ISTagValue() string {
	epochTime := strconv.FormatInt(time.Now().Unix(), 10)
	return "epoch-" + epochTime
}
