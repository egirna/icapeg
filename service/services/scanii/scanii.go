package scanii

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"icapeg/consts"
	"icapeg/logging"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Processing is a func used for to processing the http message
func (v *Scanii) Processing(partial bool) (int, interface{}, map[string]string) {
	logging.Logger.Info(v.serviceName + " service has stopped processing partially")
	serviceHeaders := make(map[string]string)

	// no need to scan part of the file, this service needs all the file at ine time
	if partial {
		logging.Logger.Info(v.serviceName + " service has stopped processing partially")
		return utils.Continue, nil, nil
	}

	isGzip := false

	//extracting the file from http message
	file, reqContentType, err := v.generalFunc.CopyingFileToTheBuffer(v.methodName)
	if err != nil {
		logging.Logger.Error(v.serviceName + " error: " + err.Error())
		logging.Logger.Info(v.serviceName + " service has stopped processing")
		return utils.InternalServerErrStatusCodeStr, nil, nil
	}

	//getting the extension of the file
	var contentType []string
	if len(contentType) == 0 {
		contentType = append(contentType, "")
	}
	var fileName string
	if v.methodName == utils.ICAPModeReq {
		contentType = v.httpMsg.Request.Header["Content-Type"]
		fileName = v.generalFunc.GetFileName()
	} else {
		contentType = v.httpMsg.Response.Header["Content-Type"]
		fileName = v.generalFunc.GetFileName()
	}
	if len(contentType) == 0 {
		contentType = append(contentType, "")
	}
	fileExtension := v.generalFunc.GetMimeExtension(file.Bytes(), contentType[0], fileName)

	isProcess, icapStatus, httpMsg := v.generalFunc.CheckTheExtension(fileExtension, v.extArrs,
		v.processExts, v.rejectExts, v.bypassExts, v.return400IfFileExtRejected, isGzip,
		v.serviceName, v.methodName, ScaniiIdentifier, v.httpMsg.Request.RequestURI, reqContentType, file)
	if !isProcess {
		logging.Logger.Info(v.serviceName + " service has stopped processing")
		return icapStatus, httpMsg, serviceHeaders
	}

	//check if the file size is greater than max file size of the service
	//if yes we will return 200 ok or 204 no modification, it depends on the configuration of the service
	if v.maxFileSize != 0 && v.maxFileSize < file.Len() {
		status, file, httpMsg := v.generalFunc.IfMaxFileSizeExc(v.returnOrigIfMaxSizeExc, v.serviceName, v.methodName, file, v.maxFileSize)
		fileAfterPrep, httpMsg := v.generalFunc.IfStatusIs204WithFile(v.methodName, status, file, isGzip, reqContentType, httpMsg, true)
		if fileAfterPrep == nil && httpMsg == nil {
			logging.Logger.Info(v.serviceName + " service has stopped processing")
			return utils.InternalServerErrStatusCodeStr, nil, nil
		}
		switch msg := httpMsg.(type) {
		case *http.Request:
			msg.Body = io.NopCloser(bytes.NewBuffer(fileAfterPrep))
			logging.Logger.Info(v.serviceName + " service has stopped processing")
			return status, msg, nil
		case *http.Response:
			msg.Body = io.NopCloser(bytes.NewBuffer(fileAfterPrep))
			logging.Logger.Info(v.serviceName + " service has stopped processing")
			return status, msg, nil
		}
		return status, nil, nil
	}

	scannedFile := file.Bytes()
	resource, score, total, err := v.SendFileToScan(file)
	if err != nil {
		logging.Logger.Error(v.serviceName + " error: " + err.Error())
		if strings.Contains(err.Error(), "context deadline exceeded") {
			logging.Logger.Info(v.serviceName + " service has stopped processing")
			return utils.RequestTimeOutStatusCodeStr, nil, nil
		}
		logging.Logger.Info(v.serviceName + " service has stopped processing")
		return utils.InternalServerErrStatusCodeStr, nil, nil
	}
	serviceHeaders["Scanii-Total"] = total
	serviceHeaders["Scanii-Positives"] = score

	scoreInt, err := strconv.Atoi(score)
	if scoreInt > 0 {
		logging.Logger.Debug(v.serviceName + ": file is not safe")
		if v.methodName == utils.ICAPModeResp {
			errPage := v.generalFunc.GenHtmlPage(utils.BlockPagePath, utils.ErrPageReasonFileIsNotSafe, v.serviceName, resource, v.httpMsg.Request.RequestURI)
			v.httpMsg.Response = v.generalFunc.ErrPageResp(http.StatusForbidden, errPage.Len())
			v.httpMsg.Response.Body = io.NopCloser(bytes.NewBuffer(errPage.Bytes()))
			logging.Logger.Info(v.serviceName + " service has stopped processing")
			return utils.OkStatusCodeStr, v.httpMsg.Response, serviceHeaders
		} else {
			htmlPage, req, err := v.generalFunc.ReqModErrPage(utils.ErrPageReasonFileIsNotSafe, v.serviceName, resource)
			if err != nil {
				return utils.InternalServerErrStatusCodeStr, nil, nil
			}
			req.Body = io.NopCloser(htmlPage)
			return utils.OkStatusCodeStr, req, serviceHeaders
		}
	}
	//returning the scanned file if everything is ok
	scannedFile = v.generalFunc.PreparingFileAfterScanning(scannedFile, reqContentType, v.methodName)
	logging.Logger.Info(v.serviceName + " service has stopped processing")
	return utils.OkStatusCodeStr, v.generalFunc.ReturningHttpMessageWithFile(v.methodName, scannedFile), serviceHeaders
}

// SendFileToScan is a function to send the file to API
func (v *Scanii) SendFileToScan(f *bytes.Buffer) (string, string, string, error) {
	logging.Logger.Debug("sending the HTTP message body to " + v.serviceName + " API to be scanned")
	urlStr := v.ScanUrl

	//form-data
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)
	part, err := bodyWriter.CreateFormFile("file", "filename")
	io.Copy(part, bytes.NewReader(f.Bytes()))

	part, _ = bodyWriter.CreateFormField("apikey")
	io.Copy(part, strings.NewReader(v.APIKey))

	req, err := http.NewRequest(http.MethodPost, urlStr, bodyBuf)
	if err != nil {
		logging.Logger.Error(v.serviceName + " error: " + err.Error())
		return "", "", "", err
	}

	client := http.Client{}
	ctx, cancel := context.WithTimeout(context.Background(), v.Timeout)
	defer cancel()
	req = req.WithContext(ctx)

	//headers
	req.Header.Add("Content-Type", bodyWriter.FormDataContentType())
	resp, err := client.Do(req)
	if err != nil {
		logging.Logger.Error(v.serviceName + " error: " + err.Error())
		return "", "", "", err
	}
	var data map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	resource := fmt.Sprint(data["resource"])
	return v.SendFileToGetReport(resource)

}

func (v *Scanii) SendFileToGetReport(resource string) (string, string, string, error) {
	logging.Logger.Debug("sending the resource of HTTP message body which scanned" +
		" to " + v.serviceName + " API to get the report")
	score, total := "", ""
	for {
		urlStr := v.ReportUrl
		//form-data
		bodyBuf := &bytes.Buffer{}
		bodyWriter := multipart.NewWriter(bodyBuf)
		// part, _ := bodyWriter.CreateFormField("apikey")
		// io.Copy(part, strings.NewReader(v.APIKey))
		part, _ := bodyWriter.CreateFormField("")
		io.Copy(part, strings.NewReader(resource))

		req, err := http.NewRequest(http.MethodPost, urlStr, bodyBuf)
		if err != nil {
			logging.Logger.Error(v.serviceName + " error: " + err.Error())
			return "", "", "", err
		}

		client := http.Client{}
		ctx, cancel := context.WithTimeout(context.Background(), v.Timeout)
		defer cancel()
		req = req.WithContext(ctx)

		//headers
		req.Header.Add("Content-Type", bodyWriter.FormDataContentType())
		resp, err := client.Do(req)
		var data map[string]interface{}
		if err != nil {
			logging.Logger.Error(v.serviceName + " error: " + err.Error())
			return "", "", "", err
		}
		err = json.NewDecoder(resp.Body).Decode(&data)
		if data["positives"] == nil {
			time.Sleep(10 * time.Second)
			continue
		}
		total = fmt.Sprint(data["total"])
		score = fmt.Sprint(data["positives"])
		break
	}
	return resource, score, total, nil
}

func (v *Scanii) ISTagValue() string {
	epochTime := strconv.FormatInt(time.Now().Unix(), 10)
	return "epoch-" + epochTime
}
