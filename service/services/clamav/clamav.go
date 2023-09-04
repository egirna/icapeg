package clamav

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/textproto"
	"strconv"
	"strings"
	"time"

	"github.com/egirna/icapeg/logging"
	utils "github.com/egirna/icapeg/utils"
	"go.uber.org/zap"

	"github.com/dutchcoders/go-clamd"
)

func (c *Clamav) Processing(partial bool, IcapHeader textproto.MIMEHeader) (int, interface{}, map[string]string, map[string]interface{},
	map[string]interface{}, map[string]interface{}) {
	logging.Logger.Info(c.serviceName+" service has started processing", zap.String("X-ICAP-Metadata", c.xICAPMetadata))
	serviceHeaders := make(map[string]string)
	serviceHeaders["X-ICAP-Metadata"] = c.xICAPMetadata
	msgHeadersBeforeProcessing := c.generalFunc.LogHTTPMsgHeaders(c.methodName)
	msgHeadersAfterProcessing := make(map[string]interface{})
	vendorMsgs := make(map[string]interface{})
	c.IcapHeaders = IcapHeader
	c.IcapHeaders.Add("X-ICAP-Metadata", c.xICAPMetadata)
	// no need to scan part of the file, this service needs all the file at one time
	if partial {
		logging.Logger.Info(c.serviceName+" service has stopped processing partially", zap.String("X-ICAP-Metadata", c.xICAPMetadata))
		return utils.Continue, nil, nil,
			msgHeadersBeforeProcessing, msgHeadersAfterProcessing, vendorMsgs
	}
	if c.methodName == utils.ICAPModeResp {
		if c.httpMsg.Response != nil {
			if c.httpMsg.Response.StatusCode == 206 {
				logging.Logger.Info(c.serviceName+" service has stopped processing byte range received", zap.String("X-ICAP-Metadata", c.xICAPMetadata))
				return utils.NoModificationStatusCodeStr, c.httpMsg, serviceHeaders,
					msgHeadersBeforeProcessing, msgHeadersAfterProcessing, vendorMsgs
			}
		}
	}
	isGzip := false
	ExceptionPagePath := utils.BlockPagePath

	if c.ExceptionPage != "" {
		ExceptionPagePath = c.ExceptionPage
	}

	//extracting the file from http message
	file, reqContentType, err := c.generalFunc.CopyingFileToTheBuffer(c.methodName)
	if err != nil {
		logging.Logger.Error(c.serviceName+" error: "+err.Error(), zap.String("X-ICAP-Metadata", c.xICAPMetadata))
		logging.Logger.Info(c.serviceName+" service has stopped processing", zap.String("X-ICAP-Metadata", c.xICAPMetadata))
		return utils.InternalServerErrStatusCodeStr, nil, serviceHeaders,
			msgHeadersBeforeProcessing, msgHeadersAfterProcessing, vendorMsgs
	}

	//if the http method is Connect, return the request as it is because it has no body
	if c.methodName == utils.ICAPModeReq {
		if c.httpMsg.Request.Method == http.MethodConnect {
			return utils.OkStatusCodeStr, c.generalFunc.ReturningHttpMessageWithFile(c.methodName, file.Bytes()),
				serviceHeaders, msgHeadersBeforeProcessing, msgHeadersAfterProcessing, vendorMsgs
		}
	}

	//getting the extension of the file
	var contentType []string

	var fileName string
	if c.methodName == utils.ICAPModeReq {
		contentType = c.httpMsg.Request.Header["Content-Type"]
		fileName = c.generalFunc.GetFileName(c.serviceName, c.xICAPMetadata)
	} else {
		contentType = c.httpMsg.Response.Header["Content-Type"]
		fileName = c.generalFunc.GetFileName(c.serviceName, c.xICAPMetadata)
	}
	if len(contentType) == 0 {
		contentType = append(contentType, "")
	}

	logging.Logger.Info(c.serviceName+" file name : "+fileName, zap.String("X-ICAP-Metadata", c.xICAPMetadata))

	fileExtension := c.generalFunc.GetMimeExtension(file.Bytes(), contentType[0], fileName)

	//check if the file extension is a bypass extension
	//if yes we will not modify the file, and we will return 204 No modifications
	hash := sha256.New()
	f := file
	_, err = hash.Write(f.Bytes())
	if err != nil {
		fmt.Println(err.Error())
	}
	fileSize := fmt.Sprintf("%v", file.Len())
	fileHash := hex.EncodeToString(hash.Sum([]byte(nil)))
	logging.Logger.Info(c.serviceName+" file hash : "+fileHash, zap.String("X-ICAP-Metadata", c.xICAPMetadata))
	isProcess, icapStatus, httpMsg := c.generalFunc.CheckTheExtension(fileExtension, c.extArrs,
		c.processExts, c.rejectExts, c.bypassExts, c.return400IfFileExtRejected, isGzip,
		c.serviceName, c.methodName, fileHash, c.httpMsg.Request.RequestURI, reqContentType, file, ExceptionPagePath, fileSize)
	if !isProcess {
		logging.Logger.Info(c.serviceName+" service has stopped processing", zap.String("X-ICAP-Metadata", c.xICAPMetadata))
		msgHeadersAfterProcessing = c.generalFunc.LogHTTPMsgHeaders(c.methodName)
		return icapStatus, httpMsg, serviceHeaders,
			msgHeadersBeforeProcessing, msgHeadersAfterProcessing, vendorMsgs
	}

	//check if the file size is greater than max file size of the service
	//if yes we will return 200 ok or 204 no modification, it depends on the configuration of the service
	if c.maxFileSize != 0 && c.maxFileSize < file.Len() {
		status, file, httpMsg := c.generalFunc.IfMaxFileSizeExc(c.returnOrigIfMaxSizeExc, c.serviceName, c.methodName, file, c.maxFileSize, ExceptionPagePath, fileSize)
		fileAfterPrep, httpMsg := c.generalFunc.IfStatusIs204WithFile(c.methodName, status, file, isGzip, reqContentType, httpMsg, true)
		if fileAfterPrep == nil && httpMsg == nil {
			logging.Logger.Info(c.serviceName+" service has stopped processing", zap.String("X-ICAP-Metadata", c.xICAPMetadata))
			return utils.InternalServerErrStatusCodeStr, nil, serviceHeaders,
				msgHeadersBeforeProcessing, msgHeadersAfterProcessing, vendorMsgs
		}
		switch msg := httpMsg.(type) {
		case *http.Request:
			msg.Body = io.NopCloser(bytes.NewBuffer(fileAfterPrep))
			logging.Logger.Info(c.serviceName+" service has stopped processing", zap.String("X-ICAP-Metadata", c.xICAPMetadata))
			msgHeadersAfterProcessing = c.generalFunc.LogHTTPMsgHeaders(c.methodName)
			return status, msg, nil,
				msgHeadersBeforeProcessing, msgHeadersAfterProcessing, vendorMsgs
		case *http.Response:
			msg.Body = io.NopCloser(bytes.NewBuffer(fileAfterPrep))
			msgHeadersAfterProcessing = c.generalFunc.LogHTTPMsgHeaders(c.methodName)
			logging.Logger.Info(c.serviceName+" service has stopped processing", zap.String("X-ICAP-Metadata", c.xICAPMetadata))
			return status, msg, nil,
				msgHeadersBeforeProcessing, msgHeadersAfterProcessing, vendorMsgs
		}
		msgHeadersAfterProcessing = c.generalFunc.LogHTTPMsgHeaders(c.methodName)
		return status, nil, nil,
			msgHeadersBeforeProcessing, msgHeadersAfterProcessing, vendorMsgs
	}

	clmd := clamd.NewClamd(c.SocketPath)
	logging.Logger.Debug("sending the HTTP msg body to the ClamAV through antivirus socket", zap.String("X-ICAP-Metadata", c.xICAPMetadata))
	response, err := clmd.ScanStream(bytes.NewReader(file.Bytes()), make(chan bool))
	if err != nil {
		logging.Logger.Error(c.serviceName+" error: "+err.Error(), zap.String("X-ICAP-Metadata", c.xICAPMetadata))
		logging.Logger.Info(c.serviceName+" service has stopped processing", zap.String("X-ICAP-Metadata", c.xICAPMetadata))
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
		logging.Logger.Error(c.serviceName+" error: "+err.Error(), zap.String("X-ICAP-Metadata", c.xICAPMetadata))
		if strings.Contains(err.Error(), "context deadline exceeded") {
			logging.Logger.Info(c.serviceName+" service has stopped processing", zap.String("X-ICAP-Metadata", c.xICAPMetadata))
			return utils.RequestTimeOutStatusCodeStr, nil, nil,
				msgHeadersBeforeProcessing, msgHeadersAfterProcessing, vendorMsgs
		}
		logging.Logger.Info(c.serviceName+" service has stopped processing", zap.String("X-ICAP-Metadata", c.xICAPMetadata))
		return utils.BadRequestStatusCodeStr, nil, nil,
			msgHeadersBeforeProcessing, msgHeadersAfterProcessing, vendorMsgs
	}
	if result.Status == ClamavMalStatus {
		logging.Logger.Debug(c.serviceName+"File is not safe", zap.String("X-ICAP-Metadata", c.xICAPMetadata))
		if c.methodName == utils.ICAPModeResp {
			errPage := c.generalFunc.GenHtmlPage(ExceptionPagePath, utils.ErrPageReasonFileIsNotSafe, c.serviceName, c.FileHash, c.httpMsg.Request.RequestURI, fileSize, c.xICAPMetadata)

			c.httpMsg.Response = c.generalFunc.ErrPageResp(c.CaseBlockHttpResponseCode, errPage.Len())
			if c.CaseBlockHttpBody {
				c.httpMsg.Response.Body = io.NopCloser(bytes.NewBuffer(errPage.Bytes()))
			} else {
				var r []byte
				c.httpMsg.Response.Body = io.NopCloser(bytes.NewBuffer(r))
				delete(c.httpMsg.Response.Header, "Content-Type")
				delete(c.httpMsg.Response.Header, "Content-Length")
			}
			logging.Logger.Info(c.serviceName+" service has stopped processing", zap.String("X-ICAP-Metadata", c.xICAPMetadata))
			msgHeadersAfterProcessing = c.generalFunc.LogHTTPMsgHeaders(c.methodName)
			return utils.OkStatusCodeStr, c.httpMsg.Response, serviceHeaders,
				msgHeadersBeforeProcessing, msgHeadersAfterProcessing, vendorMsgs
		} else {
			htmlPage, req, err := c.generalFunc.ReqModErrPage(utils.ErrPageReasonFileIsNotSafe, c.serviceName, c.FileHash, fileSize)
			if err != nil {
				logging.Logger.Error(c.serviceName+" error: "+err.Error(), zap.String("X-ICAP-Metadata", c.xICAPMetadata))

				return utils.InternalServerErrStatusCodeStr, nil, nil,
					msgHeadersBeforeProcessing, msgHeadersAfterProcessing, vendorMsgs
			}
			req.Body = io.NopCloser(htmlPage)
			serviceHeaders["X-Virus-ID"] = result.Description
			msgHeadersAfterProcessing = c.generalFunc.LogHTTPMsgHeaders(c.methodName)
			return utils.OkStatusCodeStr, req, serviceHeaders,
				msgHeadersBeforeProcessing, msgHeadersAfterProcessing, vendorMsgs
		}
	}
	//returning the scanned file if everything is ok
	fileAfterPrep, httpMsg := c.generalFunc.IfICAPStatusIs204(c.methodName, utils.NoModificationStatusCodeStr,
		file, false, reqContentType, c.httpMsg)
	if fileAfterPrep == nil && httpMsg == nil {
		logging.Logger.Info(c.serviceName+" service has stopped processing", zap.String("X-ICAP-Metadata", c.xICAPMetadata))
		return utils.InternalServerErrStatusCodeStr, nil, nil, msgHeadersBeforeProcessing,
			msgHeadersAfterProcessing, vendorMsgs
	}

	//returning the http message and the ICAP status code
	switch msg := httpMsg.(type) {
	case *http.Request:
		msg.Body = io.NopCloser(bytes.NewBuffer(fileAfterPrep))
		logging.Logger.Info(c.serviceName+" service has stopped processing", zap.String("X-ICAP-Metadata", c.xICAPMetadata))
		msgHeadersAfterProcessing = c.generalFunc.LogHTTPMsgHeaders(c.methodName)
		return utils.NoModificationStatusCodeStr, msg, serviceHeaders, msgHeadersBeforeProcessing,
			msgHeadersAfterProcessing, vendorMsgs
	case *http.Response:
		msg.Body = io.NopCloser(bytes.NewBuffer(fileAfterPrep))
		logging.Logger.Info(c.serviceName+" service has stopped processing", zap.String("X-ICAP-Metadata", c.xICAPMetadata))
		msgHeadersAfterProcessing = c.generalFunc.LogHTTPMsgHeaders(c.methodName)
		return utils.NoModificationStatusCodeStr, msg, serviceHeaders, msgHeadersBeforeProcessing,
			msgHeadersAfterProcessing, vendorMsgs
	}
	c.generalFunc.LogHTTPMsgHeaders(c.methodName)
	logging.Logger.Info(c.serviceName+" service has stopped processing", zap.String("X-ICAP-Metadata", c.xICAPMetadata))
	msgHeadersAfterProcessing = c.generalFunc.LogHTTPMsgHeaders(c.methodName)
	return utils.NoModificationStatusCodeStr, nil, serviceHeaders, msgHeadersBeforeProcessing,
		msgHeadersAfterProcessing, vendorMsgs
}

func (c *Clamav) ISTagValue() string {
	epochTime := strconv.FormatInt(time.Now().Unix(), 10)
	return "epoch-" + epochTime
}
