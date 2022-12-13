package hashlookuppackage

import (
	"bytes"
	"context"
	"crypto/sha256"
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
func (h *Hashlookup) Processing(partial bool) (int, interface{}, map[string]string) {
	logging.Logger.Info(h.serviceName + " service has started processing")
	serviceHeaders := make(map[string]string)
	// no need to scan part of the file, this service needs all the file at ine time
	if partial {
		logging.Logger.Info(h.serviceName + " service has stopped processing partially")
		return utils.Continue, nil, nil
	}

	isGzip := false

	//extracting the file from http message
	file, reqContentType, err := h.generalFunc.CopyingFileToTheBuffer(h.methodName)
	if err != nil {
		logging.Logger.Error(h.serviceName + " error: " + err.Error())
		logging.Logger.Info(h.serviceName + " service has stopped processing")
		return utils.InternalServerErrStatusCodeStr, nil, serviceHeaders
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
		return icapStatus, httpMsg, serviceHeaders
	}

	//check if the file size is greater than max file size of the service
	//if yes we will return 200 ok or 204 no modification, it depends on the configuration of the service
	if h.maxFileSize != 0 && h.maxFileSize < file.Len() {
		status, file, httpMsg := h.generalFunc.IfMaxFileSizeExc(h.returnOrigIfMaxSizeExc, h.serviceName, h.methodName, file, h.maxFileSize)
		fileAfterPrep, httpMsg := h.generalFunc.IfStatusIs204WithFile(h.methodName, status, file, isGzip, reqContentType, httpMsg, true)
		if fileAfterPrep == nil && httpMsg == nil {
			logging.Logger.Info(h.serviceName + " service has stopped processing")
			return utils.InternalServerErrStatusCodeStr, nil, serviceHeaders
		}
		switch msg := httpMsg.(type) {
		case *http.Request:
			msg.Body = io.NopCloser(bytes.NewBuffer(fileAfterPrep))
			logging.Logger.Info(h.serviceName + " service has stopped processing")
			return status, msg, nil
		case *http.Response:
			msg.Body = io.NopCloser(bytes.NewBuffer(fileAfterPrep))
			logging.Logger.Info(h.serviceName + " service has stopped processing")
			return status, msg, nil
		}
		return status, nil, nil
	}

	scannedFile := file.Bytes()
	isMal, err := h.sendFileToScan(file)
	if err != nil {
		logging.Logger.Error(h.serviceName + " error: " + err.Error())
		if strings.Contains(err.Error(), "context deadline exceeded") {
			logging.Logger.Info(h.serviceName + " service has stopped processing")
			return utils.RequestTimeOutStatusCodeStr, nil, nil
		}
		logging.Logger.Info(h.serviceName + " service has stopped processing")
		return utils.InternalServerErrStatusCodeStr, nil, nil
	}

	if isMal {
		logging.Logger.Debug(h.serviceName + ": file is not safe")
		if h.methodName == utils.ICAPModeResp {
			errPage := h.generalFunc.GenHtmlPage(utils.BlockPagePath, utils.ErrPageReasonFileIsNotSafe, h.serviceName, HashLookupIdentifier, h.httpMsg.Request.RequestURI)
			h.httpMsg.Response = h.generalFunc.ErrPageResp(http.StatusForbidden, errPage.Len())
			h.httpMsg.Response.Body = io.NopCloser(bytes.NewBuffer(errPage.Bytes()))
			logging.Logger.Info(h.serviceName + " service has stopped processing")
			return utils.OkStatusCodeStr, h.httpMsg.Response, serviceHeaders
		} else {
			htmlPage, req, err := h.generalFunc.ReqModErrPage(utils.ErrPageReasonFileIsNotSafe, h.serviceName, HashLookupIdentifier)
			if err != nil {
				return utils.InternalServerErrStatusCodeStr, nil, nil
			}
			req.Body = io.NopCloser(htmlPage)
			return utils.OkStatusCodeStr, req, serviceHeaders
		}
	}

	//returning the scanned file if everything is ok
	scannedFile = h.generalFunc.PreparingFileAfterScanning(scannedFile, reqContentType, h.methodName)
	logging.Logger.Info(h.serviceName + " service has stopped processing")
	return utils.OkStatusCodeStr, h.generalFunc.ReturningHttpMessageWithFile(h.methodName, scannedFile), serviceHeaders
}

// SendFileToScan is a function to send the file to API
func (h *Hashlookup) sendFileToScan(f *bytes.Buffer) (bool, error) {
	hash := sha256.New()
	_, _ = io.Copy(hash, f)
	fileHash := hex.EncodeToString(hash.Sum([]byte(nil)))
	req, err := http.NewRequest("GET", h.ScanUrl+fileHash, nil)
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
	y, err := (fmt.Sprint(data["KnownMalicious"])), nil
	if len(y) > 0 {
		return true, nil
	} else {
		return false, nil

	}

}

func (e *Hashlookup) ISTagValue() string {
	epochTime := strconv.FormatInt(time.Now().Unix(), 10)
	return "epoch-" + epochTime
}
