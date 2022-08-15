package clamav

import (
	"bytes"
	"github.com/dutchcoders/go-clamd"
	"icapeg/utils"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

func (c *Clamav) Processing(partial bool) (int, interface{}, map[string]string) {
	serviceHeaders := make(map[string]string)
	// no need to scan part of the file, this service needs all the file at ine time
	if partial {
		return utils.Continue, nil, nil
	}

	isGzip := false

	//extracting the file from http message
	file, reqContentType, err := c.generalFunc.CopyingFileToTheBuffer(c.methodName)
	if err != nil {
		return utils.InternalServerErrStatusCodeStr, nil, serviceHeaders
	}

	//getting the extension of the file
	fileExtension := utils.GetMimeExtension(file.Bytes())

	//check if the file extension is a bypass extension
	//if yes we will not modify the file, and we will return 204 No modifications
	for i := 0; i < 3; i++ {
		if c.extArrs[i].Name == "process" {
			if c.generalFunc.IfFileExtIsX(fileExtension, c.processExts) {
				break
			}
		} else if c.extArrs[i].Name == "reject" {
			if c.generalFunc.IfFileExtIsX(fileExtension, c.rejectExts) {
				reason := "File rejected"
				if c.return400IfFileExtRejected {
					return utils.BadRequestStatusCodeStr, nil, serviceHeaders
				}
				errPage := c.generalFunc.GenHtmlPage("service/unprocessable-file.html", reason, c.httpMsg.Request.RequestURI)
				c.httpMsg.Response = c.generalFunc.ErrPageResp(http.StatusForbidden, errPage.Len())
				c.httpMsg.Response.Body = io.NopCloser(bytes.NewBuffer(errPage.Bytes()))
				return utils.OkStatusCodeStr, c.httpMsg.Response, serviceHeaders
			}
		} else if c.extArrs[i].Name == "bypass" {
			if c.generalFunc.IfFileExtIsX(fileExtension, c.bypassExts) {
				fileAfterPrep, httpMsg := c.generalFunc.IfICAPStatusIs204(c.methodName, utils.NoModificationStatusCodeStr,
					file, false, reqContentType, c.httpMsg)
				if fileAfterPrep == nil && httpMsg == nil {
					return utils.InternalServerErrStatusCodeStr, nil, nil
				}

				//returning the http message and the ICAP status code
				switch msg := httpMsg.(type) {
				case *http.Request:
					msg.Body = io.NopCloser(bytes.NewBuffer(fileAfterPrep))
					return utils.NoModificationStatusCodeStr, msg, serviceHeaders
				case *http.Response:
					msg.Body = io.NopCloser(bytes.NewBuffer(fileAfterPrep))
					return utils.NoModificationStatusCodeStr, msg, serviceHeaders
				}
				return utils.NoModificationStatusCodeStr, nil, serviceHeaders
			}
		}
	}

	//check if the file size is greater than max file size of the service
	//if yes we will return 200 ok or 204 no modification, it depends on the configuration of the service
	if c.maxFileSize != 0 && c.maxFileSize < file.Len() {
		status, file, httpMsg := c.generalFunc.IfMaxFileSeizeExc(c.returnOrigIfMaxSizeExc, file, c.maxFileSize)
		fileAfterPrep, httpMsg := c.generalFunc.IfStatusIs204WithFile(c.methodName, status, file, isGzip, reqContentType, httpMsg)
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

	clmd := clamd.NewClamd(c.SocketPath)
	response, err := clmd.ScanStream(bytes.NewReader(file.Bytes()), make(chan bool))
	if err != nil {
		log.Println("error in scanning")
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
		log.Println("Scannng time out")
		return utils.RequestTimeOutStatusCodeStr, nil, serviceHeaders
	}

	if result.Status == ClamavMalStatus {
		reason := "File is not safe"
		errPage := c.generalFunc.GenHtmlPage("service/unprocessable-file.html", reason, c.httpMsg.Request.RequestURI)
		c.httpMsg.Response = c.generalFunc.ErrPageResp(http.StatusForbidden, errPage.Len())
		c.httpMsg.Response.Body = io.NopCloser(bytes.NewBuffer(errPage.Bytes()))
		return utils.OkStatusCodeStr, c.httpMsg.Response, serviceHeaders
	}

	//returning the scanned file if everything is ok
	fileAfterPrep, httpMsg := c.generalFunc.IfICAPStatusIs204(c.methodName, utils.NoModificationStatusCodeStr,
		file, false, reqContentType, c.httpMsg)
	if fileAfterPrep == nil && httpMsg == nil {
		return utils.InternalServerErrStatusCodeStr, nil, nil
	}

	//returning the http message and the ICAP status code
	switch msg := httpMsg.(type) {
	case *http.Request:
		msg.Body = io.NopCloser(bytes.NewBuffer(fileAfterPrep))
		return utils.NoModificationStatusCodeStr, msg, serviceHeaders
	case *http.Response:
		msg.Body = io.NopCloser(bytes.NewBuffer(fileAfterPrep))
		return utils.NoModificationStatusCodeStr, msg, serviceHeaders
	}
	return utils.NoModificationStatusCodeStr, nil, serviceHeaders

}

func (c *Clamav) ISTagValue() string {
	epochTime := strconv.FormatInt(time.Now().Unix(), 10)
	return "epoch-" + epochTime
}
