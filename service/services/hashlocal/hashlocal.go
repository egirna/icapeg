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
	"icapeg/service/services-utilities/ContentTypes"
	"io"
	"log"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func (h *Hashlocal) Processing(partial bool, IcapHeader textproto.MIMEHeader) (int, interface{}, map[string]string, map[string]interface{}, map[string]interface{}, map[string]interface{}) {
	h.initProcessing(IcapHeader)

	if partial {
		return h.handlePartialProcessing()
	}

	file, reqContentType, err := h.extractFile()
	if err != nil {
		return h.handleError(utils.InternalServerErrStatusCodeStr, err)
	}

	if h.methodName == utils.ICAPModeReq && h.httpMsg.Request.Method == http.MethodConnect {
		return h.handleConnectMethod(file)
	}

	fileName, contentType := h.getFileNameAndContentType()
	fileSize := h.getFileSize(file)
	/////////////////////////
	ExceptionPagePath := utils.BlockPagePath
	// fileExtension := h.generalFunc.GetMimeExtension(file, contentType[0], fileName)
	var fileExtension string
	filehead, err := h.httpMsg.StorageClient.ReadFileHeader(h.httpMsg.StorageKey)
	if err != nil {
		// Handle the error by extracting the file extension from the filename
		logging.Logger.Warn(utils.PrepareLogMsg(h.xICAPMetadata,
			"failed to read file header, falling back to file extension from filename: "+err.Error()))
		fileExtension = filepath.Ext(fileName)[1:]
	} else {
		// Determine the file extension using the header data
		fileExtension = h.generalFunc.GetMimeExtension(filehead, contentType[0], fileName, false)
		// If the file extension is "zip-partial", we need to send all file to handle .docx disguised as zip
		if fileExtension == "zip-partial" {
			fileExtension = h.generalFunc.GetMimeExtension(file, contentType[0], fileName, true)
		}
	}
	h.FileHash, err = h.calculateFileHash(file)
	if err != nil {
		logging.Logger.Error(utils.PrepareLogMsg(h.xICAPMetadata, h.serviceName+" calculateFileHash error : "+err.Error()))
	}
	/////////////////////////////////
	isProcess, StatusCodeStr, _ := h.generalFunc.CheckTheExtension(fileExtension, h.extArrs,
		h.processExts, h.rejectExts, h.bypassExts, h.return400IfFileExtRejected, h.isGzip(),
		h.serviceName, h.methodName, h.FileHash, h.httpMsg.Request.RequestURI, reqContentType, bytes.NewBuffer(file), ExceptionPagePath, fmt.Sprint(fileSize))

	///////////////////////////////////

	if h.isMaxFileSizeExceeded(fileSize) {
		return h.handleMaxFileSizeExceeded(bytes.NewBuffer(file), fileSize, reqContentType)
	}

	if fileSize == 0 {
		return h.handleEmptyFile(file)
	}

	isMal, err := h.sendFileToScan(h.FileHash, isProcess)
	if err != nil {
		return h.handleScanningError(err, file)
	}

	if isMal {
		return h.handleMaliciousFile(fileSize)
	}
	return h.handleSuccessfulScan(StatusCodeStr, file, fileSize, reqContentType)
}

func (h *Hashlocal) initProcessing(IcapHeader textproto.MIMEHeader) {
	h.serviceHeaders = map[string]string{"X-ICAP-Metadata": h.xICAPMetadata}
	h.msgHeadersBeforeProcessing = h.generalFunc.LogHTTPMsgHeaders(h.methodName)
	h.msgHeadersAfterProcessing = make(map[string]interface{})
	h.vendorMsgs = make(map[string]interface{})
	h.IcapHeaders = IcapHeader
	h.IcapHeaders.Add("X-ICAP-Metadata", h.xICAPMetadata)
	logging.Logger.Info(utils.PrepareLogMsg(h.xICAPMetadata, h.serviceName+" service has started processing"))
}

func (h *Hashlocal) handlePartialProcessing() (int, interface{}, map[string]string, map[string]interface{}, map[string]interface{}, map[string]interface{}) {
	logging.Logger.Info(utils.PrepareLogMsg(h.xICAPMetadata, h.serviceName+" service has stopped processing partially"))
	return utils.Continue, nil, nil, h.msgHeadersBeforeProcessing, h.msgHeadersAfterProcessing, h.vendorMsgs
}

func (h *Hashlocal) extractFile() ([]byte, ContentTypes.ContentType, error) {
	file, reqContentType, err := h.generalFunc.CopyingFileToTheBuffer(h.methodName)
	if err != nil {
		logging.Logger.Error(utils.PrepareLogMsg(h.xICAPMetadata, h.serviceName+" error: "+err.Error()))
		logging.Logger.Info(utils.PrepareLogMsg(h.xICAPMetadata, h.serviceName+" service has stopped processing"))
	}
	return file, reqContentType, err
}

func (h *Hashlocal) handleError(status int, err error) (int, interface{}, map[string]string, map[string]interface{}, map[string]interface{}, map[string]interface{}) {
	logging.Logger.Error(utils.PrepareLogMsg(h.xICAPMetadata, h.serviceName+" error: "+err.Error()))
	logging.Logger.Info(utils.PrepareLogMsg(h.xICAPMetadata, h.serviceName+" service has stopped processing"))
	return status, nil, h.serviceHeaders, h.msgHeadersBeforeProcessing, h.msgHeadersAfterProcessing, h.vendorMsgs
}

func (h *Hashlocal) handleConnectMethod(file []byte) (int, interface{}, map[string]string, map[string]interface{}, map[string]interface{}, map[string]interface{}) {
	return utils.OkStatusCodeStr, h.generalFunc.ReturningHttpMessageWithFile(h.methodName, file), h.serviceHeaders, h.msgHeadersBeforeProcessing, h.msgHeadersAfterProcessing, h.vendorMsgs
}

func (h *Hashlocal) getFileNameAndContentType() (string, []string) {
	var contentType []string
	var fileName string
	if h.methodName == utils.ICAPModeReq {
		contentType = h.httpMsg.Request.Header["Content-Type"]
		fileName = h.generalFunc.GetFileName(h.serviceName, h.xICAPMetadata)
	} else {
		contentType = h.httpMsg.Response.Header["Content-Type"]
		fileName = h.generalFunc.GetFileName(h.serviceName, h.xICAPMetadata)
	}
	if len(contentType) == 0 {
		contentType = append(contentType, "")
	}
	if filepath.Ext(fileName) == "" {
		fileName += "." + utils.Unknown
	}
	if filepath.Ext(fileName) == "." {
		fileName += utils.Unknown
	}
	logging.Logger.Info(utils.PrepareLogMsg(h.xICAPMetadata, h.serviceName+" file name : "+fileName))
	if h.httpMsg.Response != nil {
		logging.Logger.Info(utils.PrepareLogMsg(h.xICAPMetadata,
			h.serviceName+" full response headers: "+fmt.Sprintf("%v", h.httpMsg.Response.Header)))
	}
	return fileName, contentType
}

func (h *Hashlocal) getFileSize(file []byte) int {
	if h.methodName == utils.ICAPModeResp {
		size, _ := h.httpMsg.StorageClient.Size(h.httpMsg.StorageKey)
		return int(size)
	}
	return len(file)
}

func (h *Hashlocal) isMaxFileSizeExceeded(fileSize int) bool {
	return h.maxFileSize != 0 && h.maxFileSize < fileSize
}

func (h *Hashlocal) handleMaxFileSizeExceeded(file *bytes.Buffer, fileSize int, reqContentType ContentTypes.ContentType) (int, interface{}, map[string]string, map[string]interface{}, map[string]interface{}, map[string]interface{}) {
	status, file, httpMsg := h.generalFunc.IfMaxFileSizeExc(h.returnOrigIfMaxSizeExc, h.serviceName, h.methodName, file, h.maxFileSize, h.ExceptionPagePath(), fmt.Sprint(fileSize))
	fileAfterPrep, httpMsg := h.generalFunc.IfStatusIs204WithFile(h.methodName, status, file, h.isGzip(), reqContentType, httpMsg, true)
	if fileAfterPrep == nil && httpMsg == nil {
		return h.handleError(utils.InternalServerErrStatusCodeStr, fmt.Errorf("fileAfterPrep bytes is null"))
	}
	return h.prepareHttpResponse(status, httpMsg, fileAfterPrep)
}

func (h *Hashlocal) handleEmptyFile(file []byte) (int, interface{}, map[string]string, map[string]interface{}, map[string]interface{}, map[string]interface{}) {
	logging.Logger.Info(utils.PrepareLogMsg(h.xICAPMetadata, h.serviceName+" service has stopped processing zero file length"))
	return utils.NoModificationStatusCodeStr, h.generalFunc.ReturningHttpMessageWithFile(h.methodName, file), h.serviceHeaders, h.msgHeadersBeforeProcessing, h.msgHeadersAfterProcessing, h.vendorMsgs
}

func (h *Hashlocal) handleScanningError(err error, file []byte) (int, interface{}, map[string]string, map[string]interface{}, map[string]interface{}, map[string]interface{}) {
	logging.Logger.Error(utils.PrepareLogMsg(h.xICAPMetadata, h.serviceName+" error: "+err.Error()))
	if !h.BypassOnApiError {
		if strings.Contains(err.Error(), "context deadline exceeded") {
			return h.handleError(utils.RequestTimeOutStatusCodeStr, err)
		}
		return h.handleError(utils.BadRequestStatusCodeStr, err)
	}
	logging.Logger.Info(utils.PrepareLogMsg(h.xICAPMetadata, h.serviceName+" BypassOnApiError true"))
	return utils.NoModificationStatusCodeStr, h.generalFunc.ReturningHttpMessageWithFile(h.methodName, file), h.serviceHeaders, h.msgHeadersBeforeProcessing, h.msgHeadersAfterProcessing, h.vendorMsgs
}

func (h *Hashlocal) handleMaliciousFile(fileSize int) (int, interface{}, map[string]string, map[string]interface{}, map[string]interface{}, map[string]interface{}) {
	logging.Logger.Debug(utils.PrepareLogMsg(h.xICAPMetadata, h.serviceName+": file is not safe"))
	if h.methodName == utils.ICAPModeResp {
		serviceHeadersStr := mapToString(h.serviceHeaders)
		errPage := h.generalFunc.GenHtmlPage(h.ExceptionPagePath(), utils.ErrPageReasonFileIsNotSafe, h.serviceName, h.FileHash, h.httpMsg.Request.RequestURI, fmt.Sprint(fileSize), serviceHeadersStr)
		h.httpMsg.Response = h.generalFunc.ErrPageResp(h.CaseBlockHttpResponseCode, errPage.Len())
		h.saveErrorPageBody(errPage.Bytes())
		return h.prepareHttpResponse(utils.OkStatusCodeStr, h.httpMsg.Response, nil)
	} else {
		htmlPage, req, err := h.generalFunc.ReqModErrPage(utils.ErrPageReasonFileIsNotSafe, h.serviceName, h.FileHash, fmt.Sprint(fileSize))
		if err != nil {
			return h.handleError(utils.InternalServerErrStatusCodeStr, err)
		}
		req.Body = io.NopCloser(htmlPage)
		return h.prepareHttpResponse(utils.OkStatusCodeStr, req, nil)
	}
}

func (h *Hashlocal) handleSuccessfulScan(StatusCodeStr int, scannedFile []byte, fileSize int, reqContentType ContentTypes.ContentType) (int, interface{}, map[string]string, map[string]interface{}, map[string]interface{}, map[string]interface{}) {
	// Log headers and processing stop
	h.generalFunc.LogHTTPMsgHeaders(h.methodName)

	h.msgHeadersAfterProcessing = h.generalFunc.LogHTTPMsgHeaders(h.methodName)

	// Process the scanned file
	scannedFile = h.generalFunc.PreparingFileAfterScanning(scannedFile, reqContentType, h.methodName)
	logging.Logger.Info(utils.PrepareLogMsg(h.xICAPMetadata, h.serviceName+" service has stopped processing"))

	// The file was processed successfully with no changes
	if StatusCodeStr == 0 {
		StatusCodeStr = utils.NoModificationStatusCodeStr
	}
	return StatusCodeStr, h.generalFunc.ReturningHttpMessageWithFile(h.methodName, scannedFile),
		h.serviceHeaders, h.msgHeadersBeforeProcessing, h.msgHeadersAfterProcessing, h.vendorMsgs

}

func (h *Hashlocal) saveErrorPageBody(body []byte) {
	if h.CaseBlockHttpBody {
		h.httpMsg.StorageClient.Save(h.httpMsg.StorageKey, body)
	} else {
		var r []byte
		h.httpMsg.StorageClient.Save(h.httpMsg.StorageKey, r)
		delete(h.httpMsg.Response.Header, "Content-Type")
		delete(h.httpMsg.Response.Header, "Content-Length")
	}
}
func (h *Hashlocal) prepareHttpResponse(status int, httpMsg interface{}, fileAfterPrep []byte) (int, interface{}, map[string]string, map[string]interface{}, map[string]interface{}, map[string]interface{}) {
	if httpMsg != nil {
		switch msg := httpMsg.(type) {
		case *http.Request:
			if fileAfterPrep != nil {
				msg.Body = io.NopCloser(bytes.NewBuffer(fileAfterPrep))
			}
		case *http.Response:
			if fileAfterPrep != nil {
				msg.Body = io.NopCloser(bytes.NewBuffer(fileAfterPrep))
			}
		}
	}
	h.msgHeadersAfterProcessing = h.generalFunc.LogHTTPMsgHeaders(h.methodName)
	logging.Logger.Info(utils.PrepareLogMsg(h.xICAPMetadata, h.serviceName+" service has stopped processing"))
	return status, httpMsg, h.serviceHeaders, h.msgHeadersBeforeProcessing, h.msgHeadersAfterProcessing, h.vendorMsgs
}

func (h *Hashlocal) ExceptionPagePath() string {
	if h.ExceptionPage != "" {
		return h.ExceptionPage
	}
	return utils.BlockPagePath
}

func (h *Hashlocal) isGzip() bool {
	return false
}

// New function to calculate file hash without consuming memory
func (h *Hashlocal) calculateFileHash(file []byte) (string, error) {
	var fileHash string
	if h.methodName == utils.ICAPModeReq || h.httpMsg.StorageClient == nil {

		hash := sha256.New()
		_, err := hash.Write(file)
		if err != nil {
			logging.Logger.Error(utils.PrepareLogMsg(h.xICAPMetadata, h.serviceName+" calculateFileHash error : "+err.Error()))

		}
		fileHash = hex.EncodeToString(hash.Sum([]byte(nil)))
		return fileHash, nil
	}

	return h.httpMsg.StorageClient.ComputeHash(h.httpMsg.StorageKey)
}

func mapToString(m map[string]string) string {
	var sb strings.Builder
	for key, value := range m {
		sb.WriteString("\r\n")
		sb.WriteString(fmt.Sprintf("%s: %s", key, value))
	}
	return sb.String()
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
		// to help userr write comment as he wish , so just make code to skip ; at all
		if (strings.HasPrefix)(trimmedLine, ";") || (strings.HasSuffix)(trimmedLine, ";") {
			continue
		}
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
func (h *Hashlocal) sendFileToScan(f string, isProcess bool) (bool, error) {
	if !isProcess {
		return false, nil
	}
	h.FileHash = f

	//hash code
	// hash := sha256.New()
	// _, _ = io.Copy(hash, f)
	// bs := hash.Sum(nil)
	// pass := hex.EncodeToString(bs[:])

	//  the file path
	filePath := "./hash_file/hash_file_path.txt"
	// Check if the target value is present in the file
	found, err := checkValueInFile(filePath, h.FileHash)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return false, nil
	}
	logfile, err := os.OpenFile("logs/HashlocalLog.json", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer logfile.Close()
	log.SetOutput(logfile)
	if found {
		log.Printf("Value '%s'  found in the file.\n", h.FileHash)
		return true, nil
	} else {
		log.Printf("Value '%s' not found in the file.\n", h.FileHash)
		return false, nil
	}

}

func (e *Hashlocal) ISTagValue() string {
	epochTime := strconv.FormatInt(time.Now().Unix(), 10)
	return "epoch-" + epochTime
}
