package hashlocal

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	utils "icapeg/consts"
	"icapeg/logging"
	"io"
	"net/http"
	"net/textproto"
	"os"
	"strconv"
	"strings"
	"time"
)

// Processing is a func used for to processing the http message
func (h *Hashlocal) Processing(partial bool, IcapHeader textproto.MIMEHeader) (int, interface{}, map[string]string, map[string]interface{},
	map[string]interface{}, map[string]interface{}) {
	serviceHeaders := make(map[string]string)
	serviceHeaders["X-ICAP-Metadata"] = h.xICAPMetadata
	logging.Logger.Info(utils.PrepareLogMsg(h.xICAPMetadata, h.serviceName+" service has started processing"))
	msgHeadersBeforeProcessing := h.generalFunc.LogHTTPMsgHeaders(h.methodName)
	msgHeadersAfterProcessing := make(map[string]interface{})
	vendorMsgs := make(map[string]interface{})
	h.IcapHeaders = IcapHeader
	h.IcapHeaders.Add("X-ICAP-Metadata", h.xICAPMetadata)
	// no need to scan part of the file, this service needs all the file at ine time
	if partial {
		logging.Logger.Info(utils.PrepareLogMsg(h.xICAPMetadata,
			h.serviceName+" service has stopped processing partially"))
		return utils.Continue, nil, nil,
			msgHeadersBeforeProcessing, msgHeadersAfterProcessing, vendorMsgs
	}
	if h.methodName == utils.ICAPModeResp {
		if h.httpMsg.Response != nil {
			if h.httpMsg.Response.StatusCode == 206 {
				logging.Logger.Info(utils.PrepareLogMsg(h.xICAPMetadata, h.serviceName+" service has stopped processing byte range received"))
				return utils.NoModificationStatusCodeStr, h.httpMsg, serviceHeaders,
					msgHeadersBeforeProcessing, msgHeadersAfterProcessing, vendorMsgs
			}
		}
	}
	isGzip := false
	ExceptionPagePath := utils.BlockPagePath

	if h.ExceptionPage != "" {
		ExceptionPagePath = h.ExceptionPage
	}
	//extracting the file from http message

	file, reqContentType, err := h.generalFunc.CopyingFileToTheBuffer(h.methodName)

	if err != nil {
		logging.Logger.Error(utils.PrepareLogMsg(h.xICAPMetadata, h.serviceName+" error: "+err.Error()))
		logging.Logger.Info(utils.PrepareLogMsg(h.xICAPMetadata, h.serviceName+" service has stopped processing"))
		return utils.InternalServerErrStatusCodeStr, nil, serviceHeaders,
			msgHeadersBeforeProcessing, msgHeadersAfterProcessing, vendorMsgs
	}

	//if the http method is Connect, return the request as it is because it has no body
	if h.methodName == utils.ICAPModeReq {
		if h.httpMsg.Request.Method == http.MethodConnect {
			return utils.OkStatusCodeStr, h.generalFunc.ReturningHttpMessageWithFile(h.methodName, file.Bytes()),
				serviceHeaders, msgHeadersBeforeProcessing, msgHeadersAfterProcessing, vendorMsgs
		}
	}

	//getting the extension of the file
	var contentType []string

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

	logging.Logger.Info(utils.PrepareLogMsg(h.xICAPMetadata, h.serviceName+" file name : "+fileName))

	fileExtension := h.generalFunc.GetMimeExtension(file.Bytes(), contentType[0], fileName)
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
	logging.Logger.Info(utils.PrepareLogMsg(h.xICAPMetadata, h.serviceName+" file hash : "+fileHash))

	//check if the file extension is a bypass extension
	//if yes we will not modify the file, and we will return 204 No modifications
	isProcess, icapStatus, httpMsg := h.generalFunc.CheckTheExtension(fileExtension, h.extArrs,
		h.processExts, h.rejectExts, h.bypassExts, h.return400IfFileExtRejected, isGzip,
		h.serviceName, h.methodName, fileHash, h.httpMsg.Request.RequestURI, reqContentType, file, ExceptionPagePath, fileSize)
	if !isProcess {
		logging.Logger.Info(utils.PrepareLogMsg(h.xICAPMetadata, h.serviceName+" service has stopped processing"))
		msgHeadersAfterProcessing = h.generalFunc.LogHTTPMsgHeaders(h.methodName)
		return icapStatus, httpMsg, serviceHeaders,
			msgHeadersBeforeProcessing, msgHeadersAfterProcessing, vendorMsgs
	}
	//check if the file size is greater than max file size of the service
	//if yes we will return 200 ok or 204 no modification, it depends on the configuration of the service
	if h.maxFileSize != 0 && h.maxFileSize < file.Len() {
		status, file, httpMsg := h.generalFunc.IfMaxFileSizeExc(h.returnOrigIfMaxSizeExc, h.serviceName, h.methodName, file, h.maxFileSize, ExceptionPagePath, fileSize)
		fileAfterPrep, httpMsg := h.generalFunc.IfStatusIs204WithFile(h.methodName, status, file, isGzip, reqContentType, httpMsg, true)
		if fileAfterPrep == nil && httpMsg == nil {
			logging.Logger.Info(utils.PrepareLogMsg(h.xICAPMetadata, h.serviceName+" service has stopped processing"))
			return utils.InternalServerErrStatusCodeStr, nil, serviceHeaders,
				msgHeadersBeforeProcessing, msgHeadersAfterProcessing, vendorMsgs
		}
		switch msg := httpMsg.(type) {
		case *http.Request:
			msg.Body = io.NopCloser(bytes.NewBuffer(fileAfterPrep))
			logging.Logger.Info(utils.PrepareLogMsg(h.xICAPMetadata, h.serviceName+" service has stopped processing"))
			msgHeadersAfterProcessing = h.generalFunc.LogHTTPMsgHeaders(h.methodName)
			return status, msg, nil,
				msgHeadersBeforeProcessing, msgHeadersAfterProcessing, vendorMsgs
		case *http.Response:
			msg.Body = io.NopCloser(bytes.NewBuffer(fileAfterPrep))
			msgHeadersAfterProcessing = h.generalFunc.LogHTTPMsgHeaders(h.methodName)
			logging.Logger.Info(utils.PrepareLogMsg(h.xICAPMetadata, h.serviceName+" service has stopped processing"))
			return status, msg, nil,
				msgHeadersBeforeProcessing, msgHeadersAfterProcessing, vendorMsgs
		}
		msgHeadersAfterProcessing = h.generalFunc.LogHTTPMsgHeaders(h.methodName)
		return status, nil, nil,
			msgHeadersBeforeProcessing, msgHeadersAfterProcessing, vendorMsgs
	}

	scannedFile := file.Bytes()
	isMal, err := h.sendFileToScan(file)
	if err != nil && !h.BypassOnApiError {
		logging.Logger.Error(utils.PrepareLogMsg(h.xICAPMetadata, h.serviceName+" error: "+err.Error()))
		if strings.Contains(err.Error(), "context deadline exceeded") {
			logging.Logger.Info(utils.PrepareLogMsg(h.xICAPMetadata, h.serviceName+" service has stopped processing"))
			return utils.RequestTimeOutStatusCodeStr, nil, nil,
				msgHeadersBeforeProcessing, msgHeadersAfterProcessing, vendorMsgs
		}
		// its suppose to be InternalServerErrStatusCodeStr but need to be handled
		logging.Logger.Info(utils.PrepareLogMsg(h.xICAPMetadata, h.serviceName+" service has stopped processing"))
		return utils.BadRequestStatusCodeStr, nil, nil,
			msgHeadersBeforeProcessing, msgHeadersAfterProcessing, vendorMsgs
	}

	if isMal {
		logging.Logger.Debug(utils.PrepareLogMsg(h.xICAPMetadata, h.serviceName+": file is not safe"))
		if h.methodName == utils.ICAPModeResp {

			errPage := h.generalFunc.GenHtmlPage(ExceptionPagePath, utils.ErrPageReasonFileIsNotSafe, h.serviceName, h.FileHash, h.httpMsg.Request.RequestURI, fileSize, h.xICAPMetadata)

			h.httpMsg.Response = h.generalFunc.ErrPageResp(h.CaseBlockHttpResponseCode, errPage.Len())
			if h.CaseBlockHttpBody {
				h.httpMsg.Response.Body = io.NopCloser(bytes.NewBuffer(errPage.Bytes()))
			} else {
				var r []byte
				h.httpMsg.Response.Body = io.NopCloser(bytes.NewBuffer(r))
				delete(h.httpMsg.Response.Header, "Content-Type")
				delete(h.httpMsg.Response.Header, "Content-Length")
			}
			logging.Logger.Info(utils.PrepareLogMsg(h.xICAPMetadata, h.serviceName+" service has stopped processing"))
			msgHeadersAfterProcessing = h.generalFunc.LogHTTPMsgHeaders(h.methodName)
			return utils.OkStatusCodeStr, h.httpMsg.Response, serviceHeaders,
				msgHeadersBeforeProcessing, msgHeadersAfterProcessing, vendorMsgs
		} else {
			htmlPage, req, err := h.generalFunc.ReqModErrPage(utils.ErrPageReasonFileIsNotSafe, h.serviceName, h.FileHash, fileSize)
			if err != nil {
				logging.Logger.Error(utils.PrepareLogMsg(h.xICAPMetadata, h.serviceName+" error: "+err.Error()))

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
	//scannedFile = h.generalFunc.PreparingFileAfterScanning(scannedFile, reqContentType, h.methodName)
	h.generalFunc.LogHTTPMsgHeaders(h.methodName)
	logging.Logger.Info(utils.PrepareLogMsg(h.xICAPMetadata, h.serviceName+" service has stopped processing"))
	msgHeadersAfterProcessing = h.generalFunc.LogHTTPMsgHeaders(h.methodName)
	// OkStatusCodeStr => NoModificationStatusCodeStr
	/*
		return utils.NoModificationStatusCodeStr,
			h.generalFunc.ReturningHttpMessageWithFile(h.methodName, scannedFile), serviceHeaders,
			msgHeadersBeforeProcessing, msgHeadersAfterProcessing, vendorMsgs
		/*return utils.NoModificationStatusCodeStr, nil, serviceHeaders, msgHeadersBeforeProcessing,
		msgHeadersAfterProcessing, vendorMsgs	*/
	/*return utils.NoModificationStatusCodeStr,
	h.ReturningHttpMessageWithFile(h.methodName, scannedFile, h.OriginalMsg), serviceHeaders,
	msgHeadersBeforeProcessing, msgHeadersAfterProcessing, vendorMsgs*/
	scannedFile = h.generalFunc.PreparingFileAfterScanning(scannedFile, reqContentType, h.methodName)

	return utils.NoModificationStatusCodeStr, h.generalFunc.ReturningHttpMessageWithFile(h.methodName, scannedFile),
		serviceHeaders, msgHeadersBeforeProcessing, msgHeadersAfterProcessing, vendorMsgs

}

// new functions
func checkValueInFile(filePath, targetValue string) (bool, error) {
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return false, err
	}
	defer file.Close()

	// Create a scanner to read the file line by line
	scanner := bufio.NewScanner(file)

	// Iterate through each line
	for scanner.Scan() {
		line := scanner.Text()
		// for triming white space
		trimmedLine := strings.TrimSpace(line)
		//for converting into lowwercase
		convtolowercase := strings.ToLower(targetValue)
		// Check if the target value is present in the line
		if subtle.ConstantTimeCompare([]byte(strings.ToLower(trimmedLine)), []byte(convtolowercase)) == 1 {
			return true, nil
		}
	}

	// Check for errors during scanning
	if err := scanner.Err(); err != nil {
		return false, err
	}

	// If the value is not found in any line
	return false, nil
}
func (h *Hashlocal) sendFileToScan(f *bytes.Buffer) (bool, error) {

	//hash code
	hash := sha256.New()
	_, _ = io.Copy(hash, f)
	bs := hash.Sum(nil)
	pass := hex.EncodeToString(bs[:])
	h.FileHash = pass
	//  the file path
	filePath := "./test/testhash.txt"
	// Check if the target value is present in the file
	found, err := checkValueInFile(filePath, pass)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return false, nil
	}
	// Display the result
	if found {
		fmt.Printf("Value '%s' found in the file.\n", pass)
		return false, nil

	} else {
		fmt.Printf("Value '%s' not found in the file.\n", pass)
		return false, nil
	}
}
func (e *Hashlocal) ISTagValue() string {
	epochTime := strconv.FormatInt(time.Now().Unix(), 10)
	return "epoch-" + epochTime
}
