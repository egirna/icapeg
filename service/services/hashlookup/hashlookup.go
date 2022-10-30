package hashlookuppackage

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	utils "icapeg/consts"
	"icapeg/logging"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Processing is a func used for to processing the http message
func (h *Hashlookup) Processing(partial bool) (int, interface{}, map[string]string, map[string]interface{},
	map[string]interface{}, map[string]interface{}) {

	logging.Logger.Info(h.serviceName + " service has started processing")
	serviceHeaders := make(map[string]string)
	msgHeadersBeforeProcessing := h.generalFunc.LogHTTPMsgHeaders(h.methodName)
	msgHeadersAfterProcessing := make(map[string]interface{})
	vendorMsgs := make(map[string]interface{})
	// no need to scan part of the file, this service needs all the file at ine time
	if partial {
		logging.Logger.Info(h.serviceName + " service has stopped processing partially")
		return utils.Continue, nil, nil,
			msgHeadersBeforeProcessing, msgHeadersAfterProcessing, vendorMsgs
	}
	isGzip := false

	//extracting the file from http message
	file, reqContentType, err := h.generalFunc.CopyingFileToTheBuffer(h.methodName)
	if err != nil {
		logging.Logger.Error(h.serviceName + " error: " + err.Error())
		logging.Logger.Info(h.serviceName + " service has stopped processing")
		return utils.InternalServerErrStatusCodeStr, nil, serviceHeaders,
			msgHeadersBeforeProcessing, msgHeadersAfterProcessing, vendorMsgs
	}
	//getting the extension of the file
	var contentType []string
	if len(contentType) == 0 {
		contentType = append(contentType, "")
	}
	var fileName string
	if h.methodName == utils.ICAPModeReq {
		contentType = h.httpMsg.Request.Header["Content-Type"]
		fileName = h.generalFunc.GetFileName()
	} else {
		contentType = h.httpMsg.Response.Header["Content-Type"]
		fileName = h.generalFunc.GetFileName()
	}
	if len(contentType) == 0 {
		contentType = append(contentType, "")
	}
	fileExtension := h.generalFunc.GetMimeExtension(file.Bytes(), contentType[0], fileName)

	//check if the file extension is a bypass extension
	//if yes we will not modify the file, and we will return 204 No modifications
	isProcess, icapStatus, httpMsg := h.generalFunc.CheckTheExtension(fileExtension, h.extArrs,
		h.processExts, h.rejectExts, h.bypassExts, h.return400IfFileExtRejected, isGzip,
		h.serviceName, h.methodName, HashLookupIdentifier, h.httpMsg.Request.RequestURI, reqContentType, file)
	if !isProcess {
		logging.Logger.Info(h.serviceName + " service has stopped processing")
		msgHeadersAfterProcessing = h.generalFunc.LogHTTPMsgHeaders(h.methodName)
		return icapStatus, httpMsg, serviceHeaders,
			msgHeadersBeforeProcessing, msgHeadersAfterProcessing, vendorMsgs
	}

	//check if the file size is greater than max file size of the service
	//if yes we will return 200 ok or 204 no modification, it depends on the configuration of the service
	if h.maxFileSize != 0 && h.maxFileSize < file.Len() {
		status, file, httpMsg := h.generalFunc.IfMaxFileSizeExc(h.returnOrigIfMaxSizeExc, h.serviceName, h.methodName, file, h.maxFileSize)
		fileAfterPrep, httpMsg := h.generalFunc.IfStatusIs204WithFile(h.methodName, status, file, isGzip, reqContentType, httpMsg, true)
		if fileAfterPrep == nil && httpMsg == nil {
			logging.Logger.Info(h.serviceName + " service has stopped processing")
			return utils.InternalServerErrStatusCodeStr, nil, serviceHeaders,
				msgHeadersBeforeProcessing, msgHeadersAfterProcessing, vendorMsgs
		}
		switch msg := httpMsg.(type) {
		case *http.Request:
			msg.Body = io.NopCloser(bytes.NewBuffer(fileAfterPrep))
			logging.Logger.Info(h.serviceName + " service has stopped processing")
			msgHeadersAfterProcessing = h.generalFunc.LogHTTPMsgHeaders(h.methodName)
			return status, msg, nil,
				msgHeadersBeforeProcessing, msgHeadersAfterProcessing, vendorMsgs
		case *http.Response:
			msg.Body = io.NopCloser(bytes.NewBuffer(fileAfterPrep))
			logging.Logger.Info(h.serviceName + " service has stopped processing")
			msgHeadersAfterProcessing = h.generalFunc.LogHTTPMsgHeaders(h.methodName)
			return status, msg, nil,
				msgHeadersBeforeProcessing, msgHeadersAfterProcessing, vendorMsgs
		}
		msgHeadersAfterProcessing = h.generalFunc.LogHTTPMsgHeaders(h.methodName)
		return status, nil, nil,
			msgHeadersBeforeProcessing, msgHeadersAfterProcessing, vendorMsgs
	}

	scannedFile := file.Bytes()
	isMal, err := h.sendFileToScan(file)
	if err != nil {
		logging.Logger.Error(h.serviceName + " error: " + err.Error())
		if strings.Contains(err.Error(), "context deadline exceeded") {
			logging.Logger.Info(h.serviceName + " service has stopped processing")
			return utils.RequestTimeOutStatusCodeStr, nil, nil,
				msgHeadersBeforeProcessing, msgHeadersAfterProcessing, vendorMsgs
		}
		logging.Logger.Info(h.serviceName + " service has stopped processing")
		return utils.InternalServerErrStatusCodeStr, nil, nil,
			msgHeadersBeforeProcessing, msgHeadersAfterProcessing, vendorMsgs
	}

	if isMal {
		logging.Logger.Debug(h.serviceName + ": file is not safe")
		if h.methodName == utils.ICAPModeResp {
			errPage := h.generalFunc.GenHtmlPage(utils.BlockPagePath, utils.ErrPageReasonFileIsNotSafe, h.serviceName, HashLookupIdentifier, h.httpMsg.Request.RequestURI)
			h.httpMsg.Response = h.generalFunc.ErrPageResp(http.StatusForbidden, errPage.Len())
			h.httpMsg.Response.Body = io.NopCloser(bytes.NewBuffer(errPage.Bytes()))
			logging.Logger.Info(h.serviceName + " service has stopped processing")
			msgHeadersAfterProcessing = h.generalFunc.LogHTTPMsgHeaders(h.methodName)
			return utils.OkStatusCodeStr, h.httpMsg.Response, serviceHeaders,
				msgHeadersBeforeProcessing, msgHeadersAfterProcessing, vendorMsgs
		} else {
			htmlPage, req, err := h.generalFunc.ReqModErrPage(utils.ErrPageReasonFileIsNotSafe, h.serviceName, HashLookupIdentifier)
			if err != nil {
				return utils.InternalServerErrStatusCodeStr, nil, nil,
					msgHeadersBeforeProcessing, msgHeadersAfterProcessing, vendorMsgs
			}
			req.Body = io.NopCloser(htmlPage)
			msgHeadersAfterProcessing = h.generalFunc.LogHTTPMsgHeaders(h.methodName)
			return utils.OkStatusCodeStr, req, serviceHeaders,
				msgHeadersBeforeProcessing, msgHeadersAfterProcessing, vendorMsgs
		}
	}

	//returning the scanned file if everything is ok
	scannedFile = h.generalFunc.PreparingFileAfterScanning(scannedFile, reqContentType, h.methodName)
	h.generalFunc.LogHTTPMsgHeaders(h.methodName)
	logging.Logger.Info(h.serviceName + " service has stopped processing")
	msgHeadersAfterProcessing = h.generalFunc.LogHTTPMsgHeaders(h.methodName)
	return utils.OkStatusCodeStr,
		h.generalFunc.ReturningHttpMessageWithFile(h.methodName, scannedFile), serviceHeaders,
		msgHeadersBeforeProcessing, msgHeadersAfterProcessing, vendorMsgs
}

// SendFileToScan is a function to send the file to API
func (h *Hashlookup) sendFileToScan(f *bytes.Buffer) (bool, error) {
	hash := md5.New()
	_, _ = io.Copy(hash, f)

	fileHash := hex.EncodeToString(hash.Sum([]byte(nil)))
	var jsonStr = []byte(`{"hash":"` + fileHash + `"}`)
	req, err := http.NewRequest("POST", h.ScanUrl, bytes.NewBuffer(jsonStr))
	client := &http.Client{}
	ctx, cancel := context.WithTimeout(context.Background(), h.Timeout)
	defer cancel()
	req = req.WithContext(ctx)
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	var data map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	parseBool, err := strconv.ParseBool(fmt.Sprint(data["isMalicious"]))
	if err != nil {
		return false, nil
	}
	return parseBool, nil
}

func (e *Hashlookup) ISTagValue() string {
	epochTime := strconv.FormatInt(time.Now().Unix(), 10)
	return "epoch-" + epochTime
}
