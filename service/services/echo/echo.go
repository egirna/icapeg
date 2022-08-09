package echo

import (
	"bytes"
	"fmt"
	"icapeg/utils"
	"io"
	"net/http"
	"strconv"
	"time"
)

// Processing is a func used for to processing the http message
func (e *Echo) Processing(partial bool) (int, interface{}, map[string]string) {
	fmt.Println("here echo")
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
	for i := 0; i < 3; i++ {
		if e.extArrs[i].Name == "process" {
			if e.generalFunc.IfFileExtIsX(fileExtension, e.processExts) {
				break
			}
		} else if e.extArrs[i].Name == "reject" {
			if e.generalFunc.IfFileExtIsX(fileExtension, e.rejectExts) {
				reason := "File rejected"
				if e.return400IfFileExtRejected {
					return utils.BadRequestStatusCodeStr, nil, serviceHeaders
				}
				errPage := e.generalFunc.GenHtmlPage("service/unprocessable-file.html", reason, e.httpMsg.Request.RequestURI)
				e.httpMsg.Response = e.generalFunc.ErrPageResp(http.StatusForbidden, errPage.Len())
				e.httpMsg.Response.Body = io.NopCloser(bytes.NewBuffer(errPage.Bytes()))
				return utils.OkStatusCodeStr, e.httpMsg.Response, serviceHeaders
			}
		} else if e.extArrs[i].Name == "bypass" {
			if e.generalFunc.IfFileExtIsX(fileExtension, e.bypassExts) {
				fileAfterPrep, httpMsg := e.generalFunc.IfICAPStatusIs204(e.methodName, utils.NoModificationStatusCodeStr,
					file, false, reqContentType, e.httpMsg)
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
	if e.maxFileSize != 0 && e.maxFileSize < file.Len() {
		status, file, httpMsg := e.generalFunc.IfMaxFileSeizeExc(e.returnOrigIfMaxSizeExc, file, e.maxFileSize)
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

	//check if the body of the http message is compressed in Gzip or not
	//isGzip = e.generalFunc.IsBodyGzipCompressed(e.methodName)
	////if it's compressed, we decompress it to send it to Glasswall service
	//if isGzip {
	//	if file, err = e.generalFunc.DecompressGzipBody(file); err != nil {
	//		fmt.Println("here")
	//		return utils.InternalServerErrStatusCodeStr, nil, nil
	//	}
	//}

	scannedFile := file.Bytes()

	//if the original file was compressed in GZIP, we will compress the scanned file in GZIP also
	//if isGzip {
	//	scannedFile, err = e.generalFunc.CompressFileGzip(scannedFile)
	//	if err != nil {
	//		return utils.InternalServerErrStatusCodeStr, nil, nil
	//	}
	//}

	//returning the scanned file if everything is ok
	scannedFile = e.generalFunc.PreparingFileAfterScanning(scannedFile, reqContentType, e.methodName)
	return utils.OkStatusCodeStr, e.generalFunc.ReturningHttpMessageWithFile(e.methodName, scannedFile), serviceHeaders
}

func (e *Echo) ISTagValue() string {
	epochTime := strconv.FormatInt(time.Now().Unix(), 10)
	return "epoch-" + epochTime
}
