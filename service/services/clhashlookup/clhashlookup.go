package clhashlookup

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	utils "icapeg/consts"
	"icapeg/logging"
	"icapeg/service/services-utilities/ContentTypes"
	"io"
	"net/http"
	"net/textproto"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func (d *Hashlookup) Processing(partial bool, IcapHeader textproto.MIMEHeader) (int, interface{}, map[string]string, map[string]interface{}, map[string]interface{}, map[string]interface{}) {
	d.initProcessing(IcapHeader)

	if partial {
		return d.handlePartialProcessing()
	}

	file, reqContentType, err := d.extractFile()
	if err != nil {
		return d.handleError(utils.InternalServerErrStatusCodeStr, err)
	}

	if d.methodName == utils.ICAPModeReq && d.httpMsg.Request.Method == http.MethodConnect {
		return d.handleConnectMethod(file)
	}

	fileName, contentType := d.getFileNameAndContentType()
	fileSize := d.getFileSize(file)
	/////////////////////////
	ExceptionPagePath := utils.BlockPagePath
	// fileExtension := d.generalFunc.GetMimeExtension(file, contentType[0], fileName)
	var fileExtension string
	if d.methodName == utils.ICAPModeReq {
		fileExtension = d.generalFunc.GetMimeExtension(file, contentType[0], fileName)
	} else {
		filehead, err := d.httpMsg.StorageClient.ReadFileHeader(d.httpMsg.StorageKey)
		if err != nil {
			// Handle the error by extracting the file extension from the filename
			logging.Logger.Warn(utils.PrepareLogMsg(d.xICAPMetadata,
				"failed to read file header, falling back to file extension from filename: "+err.Error()))
			fileExtension = filepath.Ext(fileName)[1:]
		} else {
			// Determine the file extension using the header data
			fileExtension = d.generalFunc.GetMimeExtension(filehead, contentType[0], fileName)
		}
	}

	d.FileHash, err = d.calculateFileHash(file)
	if err != nil {
		logging.Logger.Error(utils.PrepareLogMsg(d.xICAPMetadata, d.serviceName+" calculateFileHash error : "+err.Error()))
	}
	/////////////////////////////////
	isProcess, _, _ := d.generalFunc.CheckTheExtension(fileExtension, d.extArrs,
		d.processExts, d.rejectExts, d.bypassExts, d.return400IfFileExtRejected, d.isGzip(),
		d.serviceName, d.methodName, d.FileHash, d.httpMsg.Request.RequestURI, reqContentType, bytes.NewBuffer(file), ExceptionPagePath, fmt.Sprint(fileSize))

	///////////////////////////////////

	if d.isMaxFileSizeExceeded(fileSize) {
		return d.handleMaxFileSizeExceeded(bytes.NewBuffer(file), fileSize, reqContentType)
	}

	if fileSize == 0 {
		return d.handleEmptyFile(file)
	}

	isMal, err := d.sendFileToScan(d.FileHash, isProcess)
	if err != nil {
		return d.handleScanningError(err, file)
	}

	if isMal {
		return d.handleMaliciousFile(fileSize)
	}
	return d.handleSuccessfulScan(file, fileSize, reqContentType)
}

func (d *Hashlookup) initProcessing(IcapHeader textproto.MIMEHeader) {
	d.serviceHeaders = map[string]string{"X-ICAP-Metadata": d.xICAPMetadata}
	d.msgHeadersBeforeProcessing = d.generalFunc.LogHTTPMsgHeaders(d.methodName)
	d.msgHeadersAfterProcessing = make(map[string]interface{})
	d.vendorMsgs = make(map[string]interface{})
	d.IcapHeaders = IcapHeader
	d.IcapHeaders.Add("X-ICAP-Metadata", d.xICAPMetadata)
	logging.Logger.Info(utils.PrepareLogMsg(d.xICAPMetadata, d.serviceName+" service has started processing"))
}

func (d *Hashlookup) handlePartialProcessing() (int, interface{}, map[string]string, map[string]interface{}, map[string]interface{}, map[string]interface{}) {
	logging.Logger.Info(utils.PrepareLogMsg(d.xICAPMetadata, d.serviceName+" service has stopped processing partially"))
	return utils.Continue, nil, nil, d.msgHeadersBeforeProcessing, d.msgHeadersAfterProcessing, d.vendorMsgs
}

func (d *Hashlookup) extractFile() ([]byte, ContentTypes.ContentType, error) {
	file, reqContentType, err := d.generalFunc.CopyingFileToTheBuffer(d.methodName)
	if err != nil {
		logging.Logger.Error(utils.PrepareLogMsg(d.xICAPMetadata, d.serviceName+" error: "+err.Error()))
		logging.Logger.Info(utils.PrepareLogMsg(d.xICAPMetadata, d.serviceName+" service has stopped processing"))
	}
	return file, reqContentType, err
}

func (d *Hashlookup) handleError(status int, err error) (int, interface{}, map[string]string, map[string]interface{}, map[string]interface{}, map[string]interface{}) {
	logging.Logger.Error(utils.PrepareLogMsg(d.xICAPMetadata, d.serviceName+" error: "+err.Error()))
	logging.Logger.Info(utils.PrepareLogMsg(d.xICAPMetadata, d.serviceName+" service has stopped processing"))
	return status, nil, d.serviceHeaders, d.msgHeadersBeforeProcessing, d.msgHeadersAfterProcessing, d.vendorMsgs
}

func (d *Hashlookup) handleConnectMethod(file []byte) (int, interface{}, map[string]string, map[string]interface{}, map[string]interface{}, map[string]interface{}) {
	return utils.OkStatusCodeStr, d.generalFunc.ReturningHttpMessageWithFile(d.methodName, file), d.serviceHeaders, d.msgHeadersBeforeProcessing, d.msgHeadersAfterProcessing, d.vendorMsgs
}

func (d *Hashlookup) getFileNameAndContentType() (string, []string) {
	var contentType []string
	var fileName string
	if d.methodName == utils.ICAPModeReq {
		contentType = d.httpMsg.Request.Header["Content-Type"]
		fileName = d.generalFunc.GetFileName(d.serviceName, d.xICAPMetadata)
	} else {
		contentType = d.httpMsg.Response.Header["Content-Type"]
		fileName = d.generalFunc.GetFileName(d.serviceName, d.xICAPMetadata)
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
	logging.Logger.Info(utils.PrepareLogMsg(d.xICAPMetadata, d.serviceName+" file name : "+fileName))
	return fileName, contentType
}

func (d *Hashlookup) getFileSize(file []byte) int {
	if d.methodName == utils.ICAPModeResp {
		size, _ := d.httpMsg.StorageClient.Size(d.httpMsg.StorageKey)
		return int(size)
	}
	return len(file)
}

func (d *Hashlookup) isMaxFileSizeExceeded(fileSize int) bool {
	return d.maxFileSize != 0 && d.maxFileSize < fileSize
}

func (d *Hashlookup) handleMaxFileSizeExceeded(file *bytes.Buffer, fileSize int, reqContentType ContentTypes.ContentType) (int, interface{}, map[string]string, map[string]interface{}, map[string]interface{}, map[string]interface{}) {
	status, file, httpMsg := d.generalFunc.IfMaxFileSizeExc(d.returnOrigIfMaxSizeExc, d.serviceName, d.methodName, file, d.maxFileSize, d.ExceptionPagePath(), fmt.Sprint(fileSize))
	fileAfterPrep, httpMsg := d.generalFunc.IfStatusIs204WithFile(d.methodName, status, file, d.isGzip(), reqContentType, httpMsg, true)
	if fileAfterPrep == nil && httpMsg == nil {
		return d.handleError(utils.InternalServerErrStatusCodeStr, fmt.Errorf("fileAfterPrep bytes is null"))
	}
	return d.prepareHttpResponse(status, httpMsg, fileAfterPrep)
}

func (d *Hashlookup) handleEmptyFile(file []byte) (int, interface{}, map[string]string, map[string]interface{}, map[string]interface{}, map[string]interface{}) {
	logging.Logger.Info(utils.PrepareLogMsg(d.xICAPMetadata, d.serviceName+" service has stopped processing zero file length"))
	return utils.NoModificationStatusCodeStr, d.generalFunc.ReturningHttpMessageWithFile(d.methodName, file), d.serviceHeaders, d.msgHeadersBeforeProcessing, d.msgHeadersAfterProcessing, d.vendorMsgs
}

func (d *Hashlookup) handleScanningError(err error, file []byte) (int, interface{}, map[string]string, map[string]interface{}, map[string]interface{}, map[string]interface{}) {
	logging.Logger.Error(utils.PrepareLogMsg(d.xICAPMetadata, d.serviceName+" error: "+err.Error()))
	if !d.BypassOnApiError {
		if strings.Contains(err.Error(), "context deadline exceeded") {
			return d.handleError(utils.RequestTimeOutStatusCodeStr, err)
		}
		return d.handleError(utils.BadRequestStatusCodeStr, err)
	}
	logging.Logger.Info(utils.PrepareLogMsg(d.xICAPMetadata, d.serviceName+" BypassOnApiError true"))
	return utils.NoModificationStatusCodeStr, d.generalFunc.ReturningHttpMessageWithFile(d.methodName, file), d.serviceHeaders, d.msgHeadersBeforeProcessing, d.msgHeadersAfterProcessing, d.vendorMsgs
}

func (d *Hashlookup) handleMaliciousFile(fileSize int) (int, interface{}, map[string]string, map[string]interface{}, map[string]interface{}, map[string]interface{}) {
	logging.Logger.Debug(utils.PrepareLogMsg(d.xICAPMetadata, d.serviceName+": file is not safe"))
	if d.methodName == utils.ICAPModeResp {
		serviceHeadersStr := mapToString(d.serviceHeaders)
		errPage := d.generalFunc.GenHtmlPage(d.ExceptionPagePath(), utils.ErrPageReasonFileIsNotSafe, d.serviceName, d.FileHash, d.httpMsg.Request.RequestURI, fmt.Sprint(fileSize), serviceHeadersStr)
		d.httpMsg.Response = d.generalFunc.ErrPageResp(d.CaseBlockHttpResponseCode, errPage.Len())
		d.saveErrorPageBody(errPage.Bytes())
		return d.prepareHttpResponse(utils.OkStatusCodeStr, d.httpMsg.Response, nil)
	} else {
		htmlPage, req, err := d.generalFunc.ReqModErrPage(utils.ErrPageReasonFileIsNotSafe, d.serviceName, d.FileHash, fmt.Sprint(fileSize))
		if err != nil {
			return d.handleError(utils.InternalServerErrStatusCodeStr, err)
		}
		req.Body = io.NopCloser(htmlPage)
		return d.prepareHttpResponse(utils.OkStatusCodeStr, req, nil)
	}
}

func (d *Hashlookup) handleSuccessfulScan(scannedFile []byte, fileSize int, reqContentType ContentTypes.ContentType) (int, interface{}, map[string]string, map[string]interface{}, map[string]interface{}, map[string]interface{}) {
	// Log headers and processing stop
	d.generalFunc.LogHTTPMsgHeaders(d.methodName)
	logging.Logger.Info(utils.PrepareLogMsg(d.xICAPMetadata, d.serviceName+" service has stopped processing"))
	d.msgHeadersAfterProcessing = d.generalFunc.LogHTTPMsgHeaders(d.methodName)

	// Process the scanned file
	scannedFile = d.generalFunc.PreparingFileAfterScanning(scannedFile, reqContentType, d.methodName)

	return utils.NoModificationStatusCodeStr, d.generalFunc.ReturningHttpMessageWithFile(d.methodName, scannedFile),
		d.serviceHeaders, d.msgHeadersBeforeProcessing, d.msgHeadersAfterProcessing, d.vendorMsgs
}

func (d *Hashlookup) saveErrorPageBody(body []byte) {
	if d.CaseBlockHttpBody {
		d.httpMsg.StorageClient.Save(d.httpMsg.StorageKey, body)
	} else {
		var r []byte
		d.httpMsg.StorageClient.Save(d.httpMsg.StorageKey, r)
		delete(d.httpMsg.Response.Header, "Content-Type")
		delete(d.httpMsg.Response.Header, "Content-Length")
	}
}
func (d *Hashlookup) prepareHttpResponse(status int, httpMsg interface{}, fileAfterPrep []byte) (int, interface{}, map[string]string, map[string]interface{}, map[string]interface{}, map[string]interface{}) {
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
	d.msgHeadersAfterProcessing = d.generalFunc.LogHTTPMsgHeaders(d.methodName)
	logging.Logger.Info(utils.PrepareLogMsg(d.xICAPMetadata, d.serviceName+" service has stopped processing"))
	return status, httpMsg, d.serviceHeaders, d.msgHeadersBeforeProcessing, d.msgHeadersAfterProcessing, d.vendorMsgs
}

func (d *Hashlookup) ExceptionPagePath() string {
	if d.ExceptionPage != "" {
		return d.ExceptionPage
	}
	return utils.BlockPagePath
}

func (d *Hashlookup) isGzip() bool {
	return false
}

// New function to calculate file hash without consuming memory
func (d *Hashlookup) calculateFileHash(file []byte) (string, error) {
	var fileHash string
	if d.methodName == utils.ICAPModeReq || d.httpMsg.StorageClient == nil {

		hash := sha256.New()
		_, err := hash.Write(file)
		if err != nil {
			logging.Logger.Error(utils.PrepareLogMsg(d.xICAPMetadata, d.serviceName+" calculateFileHash error : "+err.Error()))

		}
		fileHash = hex.EncodeToString(hash.Sum([]byte(nil)))
		return fileHash, nil
	}

	return d.httpMsg.StorageClient.ComputeHash(d.httpMsg.StorageKey)
}

func mapToString(m map[string]string) string {
	var sb strings.Builder
	for key, value := range m {
		sb.WriteString(fmt.Sprintf("\r\n"))
		sb.WriteString(fmt.Sprintf("%s: %s", key, value))
	}
	return sb.String()
}

// SendFileToScan is a function to send the file to API
func (h *Hashlookup) sendFileToScan(f string, isProcess bool) (bool, error) {

	if !isProcess {
		return false, nil
	}
	h.FileHash = f
	//var jsonStr = []byte(`{"hash":"` + fileHash + `"}`)
	req, err := http.NewRequest("GET", h.ScanUrl+h.FileHash, nil)
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
	if len(y) > 0 && y != "<nil>" {
		return true, nil
	} else {
		return false, nil

	}

}

func (e *Hashlookup) ISTagValue() string {
	epochTime := strconv.FormatInt(time.Now().Unix(), 10)
	return "epoch-" + epochTime
}
