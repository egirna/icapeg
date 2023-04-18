package cloudmersive

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	utils "icapeg/consts"
	"icapeg/logging"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strconv"
	"time"
)

func (c *CloudMersive) Processing(partial bool) (int, interface{}, map[string]string, map[string]interface{},
	map[string]interface{}, map[string]interface{}) {
	serviceHeaders := make(map[string]string)
	serviceHeaders["X-ICAP-Metadata"] = c.xICAPMetadata
	msgHeadersBeforeProcessing := c.generalFunc.LogHTTPMsgHeaders(c.methodName)
	msgHeadersAfterProcessing := make(map[string]interface{})
	vendorMsgs := make(map[string]interface{})
	logging.Logger.Info(utils.PrepareLogMsg(c.xICAPMetadata, c.serviceName+" service has started processing"))

	// no need to scan part of the file, this service needs all the file at ine time
	if partial {
		logging.Logger.Info(utils.PrepareLogMsg(c.xICAPMetadata, c.serviceName+" service has stopped processing partially"))
		return utils.Continue, nil, nil, msgHeadersBeforeProcessing, msgHeadersAfterProcessing, vendorMsgs
	}
	isGzip := false

	//extracting the file from http message
	file, reqContentType, err := c.generalFunc.CopyingFileToTheBuffer(c.methodName)
	if err != nil {
		logging.Logger.Error(utils.PrepareLogMsg(c.xICAPMetadata, c.serviceName+" error: "+err.Error()))
		logging.Logger.Info(utils.PrepareLogMsg(c.xICAPMetadata, c.serviceName+" service has stopped processing"))
		return utils.InternalServerErrStatusCodeStr, nil, serviceHeaders,
			msgHeadersBeforeProcessing, msgHeadersAfterProcessing, vendorMsgs
	}

	//if the http method is Connect, return the request as it is because it has no body
	if c.httpMsg.Request.Method == http.MethodConnect {
		return utils.OkStatusCodeStr, c.generalFunc.ReturningHttpMessageWithFile(c.methodName, file.Bytes()),
			serviceHeaders, msgHeadersBeforeProcessing, msgHeadersAfterProcessing, vendorMsgs
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
		c.serviceName, c.methodName, CloudMersiveIdentifier, c.httpMsg.Request.RequestURI, reqContentType, file)
	if !isProcess {
		logging.Logger.Info(utils.PrepareLogMsg(c.xICAPMetadata, c.serviceName+" service has stopped processing"))
		msgHeadersAfterProcessing = c.generalFunc.LogHTTPMsgHeaders(c.methodName)
		return icapStatus, httpMsg, serviceHeaders, msgHeadersBeforeProcessing,
			msgHeadersAfterProcessing, vendorMsgs
	}

	//check if the file size is greater than max file size of the service
	//if yes we will return 200 ok or 204 no modification, it depends on the configuration of the service
	if c.maxFileSize != 0 && c.maxFileSize < file.Len() {
		status, file, httpMsgAfter := c.generalFunc.IfMaxFileSizeExc(c.returnOrigIfMaxSizeExc, c.serviceName, c.methodName, file, c.maxFileSize)
		fileAfterPrep, httpMsgAfter := c.generalFunc.IfStatusIs204WithFile(c.methodName, status, file, isGzip, reqContentType, httpMsgAfter, true)
		if fileAfterPrep == nil && httpMsgAfter == nil {
			logging.Logger.Info(utils.PrepareLogMsg(c.xICAPMetadata, c.serviceName+" service has stopped processing"))
			return utils.InternalServerErrStatusCodeStr, nil, serviceHeaders,
				msgHeadersBeforeProcessing, msgHeadersAfterProcessing, vendorMsgs
		}
		switch msg := httpMsgAfter.(type) {

		case *http.Request:
			msg.Body = io.NopCloser(bytes.NewBuffer(fileAfterPrep))
			logging.Logger.Info(utils.PrepareLogMsg(c.xICAPMetadata, c.serviceName+" service has stopped processing"))
			msgHeadersAfterProcessing = c.generalFunc.LogHTTPMsgHeaders(c.methodName)
			return status, msg, nil, msgHeadersBeforeProcessing, msgHeadersAfterProcessing, vendorMsgs
		case *http.Response:
			msg.Body = io.NopCloser(bytes.NewBuffer(fileAfterPrep))
			logging.Logger.Info(utils.PrepareLogMsg(c.xICAPMetadata, c.serviceName+" service has stopped processing"))
			msgHeadersAfterProcessing = c.generalFunc.LogHTTPMsgHeaders(c.methodName)
			return status, msg, nil, msgHeadersBeforeProcessing, msgHeadersAfterProcessing, vendorMsgs
		}
		msgHeadersAfterProcessing = c.generalFunc.LogHTTPMsgHeaders(c.methodName)
		return status, nil, nil, msgHeadersBeforeProcessing, msgHeadersAfterProcessing, vendorMsgs
	}

	// sending request to cloudmersive api
	serviceResp, err := c.SendFileToAPI(file, fileName)
	if err != nil {
		logging.Logger.Error(utils.PrepareLogMsg(c.xICAPMetadata, c.serviceName+" error: "+err.Error()))
		logging.Logger.Info(utils.PrepareLogMsg(c.xICAPMetadata, c.serviceName+" service has stopped processing"))
		return serviceResp.StatusCode, nil, serviceHeaders, msgHeadersBeforeProcessing, msgHeadersAfterProcessing, vendorMsgs
	}
	// getting response body
	body, err := ioutil.ReadAll(serviceResp.Body)
	if err != nil {
		logging.Logger.Error(utils.PrepareLogMsg(c.xICAPMetadata, c.serviceName+" error: "+err.Error()))
		logging.Logger.Info(utils.PrepareLogMsg(c.xICAPMetadata, c.serviceName+" service has stopped processing"))
		return serviceResp.StatusCode, nil, serviceHeaders, msgHeadersBeforeProcessing, msgHeadersAfterProcessing, vendorMsgs
	}
	var data map[string]interface{}
	json.Unmarshal(body, &data)
	// msg used to read error messages when status is not 200
	msg := string(body)
	if serviceResp.StatusCode == 400 && msg == "Invalid input: Input file was empty." {
		logging.Logger.Error(utils.PrepareLogMsg(c.xICAPMetadata, c.serviceName+" error: Invalid input: Input file was empty"))
		errPage := c.generalFunc.GenHtmlPage(utils.BlockPagePath, msg, c.serviceName, serviceResp.Header["Request-Context"][0], c.httpMsg.Request.RequestURI)
		c.httpMsg.Response = c.generalFunc.ErrPageResp(http.StatusForbidden, errPage.Len())
		c.httpMsg.Response.Body = io.NopCloser(bytes.NewBuffer(errPage.Bytes()))
		logging.Logger.Info(utils.PrepareLogMsg(c.xICAPMetadata, c.serviceName+" service has stopped processing"))
		return utils.OkStatusCodeStr, c.httpMsg.Response, serviceHeaders, msgHeadersBeforeProcessing, msgHeadersAfterProcessing, vendorMsgs
	}
	// check CleanResult, if false detect why
	var reason string
	reason = ""
	if data["CleanResult"].(bool) == false {
		serviceHeaders["CleanResult"] = "false"
		logging.Logger.Debug(utils.PrepareLogMsg(c.xICAPMetadata, c.serviceName+": CleanResult equals false"))
		if data["ContainsExecutable"].(bool) == true && !c.allowExecutables {
			reason = "executable files not allowed"
		} else if data["ContainsScript"].(bool) == true && !c.allowScripts {
			reason = "scripts not allowed"
		} else if data["ContainsPasswordProtectedFile"].(bool) == true && !c.allowPasswordProtectedFiles {
			reason = "password protected files not allowed"
		} else if data["ContainsMacros"].(bool) == true && !c.allowMacros {
			reason = "macros not allowed"
		} else if data["ContainsXmlExternalEntities"].(bool) == true && !c.allowXmlExternalEntities {
			reason = "xml external entities not allowed"
		} else if data["ContainsInsecureDeserialization"].(bool) == true && !c.allowInsecureDeserialization {
			reason = "insecure deserialization not allowed"
		} else if data["ContainsHtml"].(bool) == true && !c.allowHtml {
			reason = "html not allowed"
		} else if data["FoundViruses"] == nil {
			reason = fmt.Sprintln("File is not safe")
		}
		if reason != "" {
			if c.return400IfFileExtRejected {
				logging.Logger.Info(utils.PrepareLogMsg(c.xICAPMetadata, c.serviceName+" service has stopped processing"))
				return utils.BadRequestStatusCodeStr, nil, serviceHeaders, msgHeadersBeforeProcessing, msgHeadersAfterProcessing, vendorMsgs
			}
			errPage := c.generalFunc.GenHtmlPage(utils.BlockPagePath, reason, c.serviceName, serviceResp.Header["Request-Context"][0], c.httpMsg.Request.RequestURI)
			c.httpMsg.Response = c.generalFunc.ErrPageResp(http.StatusForbidden, errPage.Len())
			c.httpMsg.Response.Body = io.NopCloser(bytes.NewBuffer(errPage.Bytes()))
			logging.Logger.Info(utils.PrepareLogMsg(c.xICAPMetadata, c.serviceName+" service has stopped processing"))
			return utils.OkStatusCodeStr, c.httpMsg.Response, serviceHeaders, msgHeadersBeforeProcessing, msgHeadersAfterProcessing, vendorMsgs
		}
		logging.Logger.Debug(utils.PrepareLogMsg(c.xICAPMetadata, c.serviceName+" error: "+reason))
	}
	// check if viruses found, and return virus information in error page
	if data["FoundViruses"] != nil {
		serviceHeaders["FoundViruses"] = ""
		var v string
		reason = fmt.Sprintln("File contains virus, viruses found: ")
		for _, item := range data["FoundViruses"].([]interface{}) {
			v = fmt.Sprintf("%v \n", item.(map[string]interface{})["VirusName"])
			reason += v
			serviceHeaders["FoundViruses"] += v
		}
		logging.Logger.Debug(utils.PrepareLogMsg(c.xICAPMetadata, c.serviceName+" error: "+reason))
		if c.methodName == utils.ICAPModeResp {
			errPage := c.generalFunc.GenHtmlPage(utils.BlockPagePath, reason, c.serviceName, serviceResp.Header["Request-Context"][0], c.httpMsg.Request.RequestURI)
			c.httpMsg.Response = c.generalFunc.ErrPageResp(http.StatusForbidden, errPage.Len())
			c.httpMsg.Response.Body = io.NopCloser(bytes.NewBuffer(errPage.Bytes()))
			logging.Logger.Info(utils.PrepareLogMsg(c.xICAPMetadata, c.serviceName+" service has stopped processing"))
			return utils.OkStatusCodeStr, c.httpMsg.Response,
				serviceHeaders, msgHeadersBeforeProcessing, msgHeadersAfterProcessing, vendorMsgs
		} else {
			htmlPage, req, err := c.generalFunc.ReqModErrPage(reason, c.serviceName, CloudMersiveIdentifier)
			if err != nil {
				return utils.InternalServerErrStatusCodeStr, nil, nil, msgHeadersBeforeProcessing, msgHeadersAfterProcessing, vendorMsgs
			}
			req.Body = io.NopCloser(htmlPage)
			return utils.OkStatusCodeStr, req, serviceHeaders, msgHeadersBeforeProcessing, msgHeadersAfterProcessing, vendorMsgs
		}
	}
	serviceHeaders["CleanResult"] = "true"
	scannedFile := c.generalFunc.PreparingFileAfterScanning(file.Bytes(), reqContentType, c.methodName)
	logging.Logger.Info(utils.PrepareLogMsg(c.xICAPMetadata, c.serviceName+" service has stopped processing"))
	return utils.OkStatusCodeStr, c.generalFunc.ReturningHttpMessageWithFile(c.methodName, scannedFile),
		serviceHeaders, msgHeadersBeforeProcessing, msgHeadersAfterProcessing, vendorMsgs
}

func (c *CloudMersive) SendFileToAPI(f *bytes.Buffer, filename string) (*http.Response, error) {
	logging.Logger.Debug(utils.PrepareLogMsg(c.xICAPMetadata, "sending the HTTP message body to "+c.serviceName+" API"))
	url := c.BaseURL + c.ScanEndPoint
	bodyBuf := &bytes.Buffer{}

	bodyWriter := multipart.NewWriter(bodyBuf)

	part, err := bodyWriter.CreateFormFile("file", filename)
	if err != nil {
		logging.Logger.Error(utils.PrepareLogMsg(c.xICAPMetadata, c.serviceName+" error: "+err.Error()))
		return nil, err
	}

	io.Copy(part, bytes.NewReader(f.Bytes()))
	if err := bodyWriter.Close(); err != nil {
		logging.Logger.Error(utils.PrepareLogMsg(c.xICAPMetadata, c.serviceName+" error: "+err.Error()))
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, url, bodyBuf)
	if err != nil {
		logging.Logger.Error(utils.PrepareLogMsg(c.xICAPMetadata, c.serviceName+" error: "+err.Error()))
		return nil, err
	}
	req.Header.Add("allowExecutables", strconv.FormatBool(c.allowExecutables))
	req.Header.Add("allowInvalidFiles", strconv.FormatBool(c.allowInvalidFiles))
	req.Header.Add("allowScripts", strconv.FormatBool(c.allowScripts))
	req.Header.Add("allowPasswordProtectedFiles", strconv.FormatBool(c.allowPasswordProtectedFiles))
	req.Header.Add("allowMacros", strconv.FormatBool(c.allowMacros))
	req.Header.Add("allowXmlExternalEntities", strconv.FormatBool(c.allowXmlExternalEntities))
	req.Header.Add("allowHtml", strconv.FormatBool(c.allowHtml))
	req.Header.Add("allowInsecureDeserialization", strconv.FormatBool(c.allowInsecureDeserialization))
	req.Header.Add("Content-Type", "multipart/form-data")
	req.Header.Add("Apikey", c.APIKey)
	req.Header.Set("Content-Type", bodyWriter.FormDataContentType())

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: c.generalFunc.InitSecure(c.verifyServerCert)},
	}
	client := &http.Client{Transport: tr}
	ctx, _ := context.WithTimeout(context.Background(), c.Timeout)

	// defer cancel() commit cancel must be on goroutine
	req = req.WithContext(ctx)

	res, err := client.Do(req)
	if err != nil {
		logging.Logger.Error(utils.PrepareLogMsg(c.xICAPMetadata, c.serviceName+" error: "+err.Error()))
		return nil, err
	}
	return res, nil
}

func (c *CloudMersive) ISTagValue() string {
	epochTime := strconv.FormatInt(time.Now().Unix(), 10)
	return "epoch-" + epochTime
}
