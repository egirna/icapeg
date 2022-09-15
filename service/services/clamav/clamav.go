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

func (c *Clamav) Processing(partial bool) (int, interface{}, map[string]string) {
	logging.Logger.Info(c.serviceName + " service has started processing")
	serviceHeaders := make(map[string]string)
	// no need to scan part of the file, this service needs all the file at ine time
	if partial {
		logging.Logger.Info(c.serviceName + " service has stopped processing partially")
		return utils.Continue, nil, nil
	}

	isGzip := false

	//extracting the file from http message
	file, reqContentType, err := c.generalFunc.CopyingFileToTheBuffer(c.methodName)
	if err != nil {
		logging.Logger.Error(c.serviceName + " error: " + err.Error())
		logging.Logger.Info(c.serviceName + " service has stopped processing")
		return utils.InternalServerErrStatusCodeStr, nil, serviceHeaders
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
		logging.Logger.Info(c.serviceName + " service has stopped processing")
		return icapStatus, httpMsg, serviceHeaders
	}

	//check if the file size is greater than max file size of the service
	//if yes we will return 200 ok or 204 no modification, it depends on the configuration of the service
	if c.maxFileSize != 0 && c.maxFileSize < file.Len() {
		status, file, httpMsg := c.generalFunc.IfMaxFileSizeExc(c.returnOrigIfMaxSizeExc, c.serviceName, c.methodName, file, c.maxFileSize)
		fileAfterPrep, httpMsg := c.generalFunc.IfStatusIs204WithFile(c.methodName, status, file, isGzip, reqContentType, httpMsg)
		if fileAfterPrep == nil && httpMsg == nil {
			logging.Logger.Info(c.serviceName + " service has stopped processing")
			return utils.InternalServerErrStatusCodeStr, nil, serviceHeaders
		}
		switch msg := httpMsg.(type) {
		case *http.Request:
			msg.Body = io.NopCloser(bytes.NewBuffer(fileAfterPrep))
			logging.Logger.Info(c.serviceName + " service has stopped processing")
			return status, msg, nil
		case *http.Response:
			msg.Body = io.NopCloser(bytes.NewBuffer(fileAfterPrep))
			logging.Logger.Info(c.serviceName + " service has stopped processing")
			return status, msg, nil
		}
		return status, nil, nil
	}

	clmd := clamd.NewClamd(c.SocketPath)
	logging.Logger.Debug("sending the HTTP msg body to the ClamAV through antivirus socket")
	response, err := clmd.ScanStream(bytes.NewReader(file.Bytes()), make(chan bool))
	if err != nil {
		logging.Logger.Error(c.serviceName + " error: " + err.Error())
		logging.Logger.Info(c.serviceName + " service has stopped processing")
		return utils.InternalServerErrStatusCodeStr, nil, serviceHeaders
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
		logging.Logger.Error(c.serviceName + " error: scanning timeout")
		logging.Logger.Info(c.serviceName + " service has stopped processing")
		return utils.RequestTimeOutStatusCodeStr, nil, serviceHeaders
	}

	if result.Status == ClamavMalStatus {
		logging.Logger.Debug(c.serviceName + "File is not safe")
		reason := "File is not safe"
		errPage := c.generalFunc.GenHtmlPage(utils.BlockPagePath, reason, c.serviceName, "CLAMAV ID", c.httpMsg.Request.RequestURI)
		c.httpMsg.Response = c.generalFunc.ErrPageResp(http.StatusForbidden, errPage.Len())
		c.httpMsg.Response.Body = io.NopCloser(bytes.NewBuffer(errPage.Bytes()))
		logging.Logger.Info(c.serviceName + " service has stopped processing")
		return utils.OkStatusCodeStr, c.httpMsg.Response, serviceHeaders
	}
	logging.Logger.Debug(c.serviceName + "File is safe")
	//returning the scanned file if everything is ok
	fileAfterPrep, httpMsg := c.generalFunc.IfICAPStatusIs204(c.methodName, utils.NoModificationStatusCodeStr,
		file, false, reqContentType, c.httpMsg)
	if fileAfterPrep == nil && httpMsg == nil {
		logging.Logger.Info(c.serviceName + " service has stopped processing")
		return utils.InternalServerErrStatusCodeStr, nil, nil
	}

	//returning the http message and the ICAP status code
	switch msg := httpMsg.(type) {
	case *http.Request:
		msg.Body = io.NopCloser(bytes.NewBuffer(fileAfterPrep))
		logging.Logger.Info(c.serviceName + " service has stopped processing")
		return utils.NoModificationStatusCodeStr, msg, serviceHeaders
	case *http.Response:
		msg.Body = io.NopCloser(bytes.NewBuffer(fileAfterPrep))
		logging.Logger.Info(c.serviceName + " service has stopped processing")
		return utils.NoModificationStatusCodeStr, msg, serviceHeaders
	}
	logging.Logger.Info(c.serviceName + " service has stopped processing")
	return utils.NoModificationStatusCodeStr, nil, serviceHeaders

}

func (c *Clamav) ISTagValue() string {
	epochTime := strconv.FormatInt(time.Now().Unix(), 10)
	return "epoch-" + epochTime
}
