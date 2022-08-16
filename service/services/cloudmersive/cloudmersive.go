package cloudmersive

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"icapeg/utils"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strconv"
	"time"
)

func (c *CloudMersive) Processing(partial bool) (int, interface{}, map[string]string) {
	//TODO implement me

	// no need to scan part of the file, this service needs all the file at one time
	if partial {
		return utils.Continue, nil, nil
	}

	// ICAP response headers
	serviceHeaders := make(map[string]string)
	//extracting the file from http message
	file, reqContentType, err := c.generalFunc.CopyingFileToTheBuffer(c.methodName)
	if err != nil {
		return utils.InternalServerErrStatusCodeStr, nil, serviceHeaders
	}

	var contentType []string
	if len(contentType) == 0 {
		contentType = append(contentType, "")
	}
	var fileName string
	if c.methodName == utils.ICAPModeReq {
		contentType = c.httpMsg.Request.Header["Content-Type"]
		fileName = utils.GetFileName(c.httpMsg.Request)
	} else {
		contentType = c.httpMsg.Response.Header["Content-Type"]
		fileName = utils.GetFileName(c.httpMsg.Response)
	}
	if len(contentType) == 0 {
		contentType = append(contentType, "")
	}
	fileExtension := utils.GetMimeExtension(file.Bytes(), contentType[0], fileName)
	fmt.Println(fileExtension)
	//check if the file extension is a bypass extension
	//if yes we will not modify the file, and we will return 204 No modifications
	fmt.Println(c.extArrs)
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
				errPage := c.generalFunc.GenHtmlPage("service/unprocessable-file.html", reason, c.serviceName, "NO ID", c.httpMsg.Request.RequestURI)
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

	//check if the file size is greater than max file size of the service or 3M size, according to account payment plans ,etc
	if c.maxFileSize != 0 && c.maxFileSize < file.Len() || file.Len() > 3e6 {
		errPage := c.generalFunc.GenHtmlPage("service/unprocessable-file.html", "file size exceeded maximum allowed size", "NO ID", c.serviceName, c.httpMsg.Request.RequestURI)
		c.httpMsg.Response = c.generalFunc.ErrPageResp(http.StatusForbidden, errPage.Len())
		c.httpMsg.Response.Body = io.NopCloser(bytes.NewBuffer(errPage.Bytes()))
		return utils.OkStatusCodeStr, c.httpMsg.Response, serviceHeaders
	}
	// sending request to cloudmersive api
	serviceResp, err := c.SendFileToAPI(file, fileName)
	if err != nil {
		fmt.Println("error line 44\n", err.Error())
		return serviceResp.StatusCode, nil, serviceHeaders
	}
	// getting response body
	body, err := ioutil.ReadAll(serviceResp.Body)
	if err != nil {
		fmt.Println(err)
		return serviceResp.StatusCode, nil, serviceHeaders
	}
	var data map[string]interface{}
	json.Unmarshal(body, &data)
	// msg used to read error messages when status is not 200
	msg := string(body)
	if serviceResp.StatusCode == 400 && msg == "Invalid input: Input file was empty." {
		fmt.Println(msg)
		errPage := c.generalFunc.GenHtmlPage("service/unprocessable-file.html", msg, c.serviceName, serviceResp.Header["Request-Context"][0], c.httpMsg.Request.RequestURI)
		c.httpMsg.Response = c.generalFunc.ErrPageResp(http.StatusForbidden, errPage.Len())
		c.httpMsg.Response.Body = io.NopCloser(bytes.NewBuffer(errPage.Bytes()))
		return utils.OkStatusCodeStr, c.httpMsg.Response, serviceHeaders
	}
	// check CleanResult, if false detect why
	var reason string
	reason = ""
	fmt.Println(serviceResp.Header)
	if data["CleanResult"].(bool) == false {
		serviceHeaders["CleanResult"] = "false"
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
		} else if data["FoundViuses"] == nil {
			reason = fmt.Sprintln("File is not safe")
		}
		if reason != "" {
			if c.return400IfFileExtRejected {
				return utils.BadRequestStatusCodeStr, nil, serviceHeaders
			}
			fmt.Println(reason)
			errPage := c.generalFunc.GenHtmlPage("service/unprocessable-file.html", reason, c.serviceName, serviceResp.Header["Request-Context"][0], c.httpMsg.Request.RequestURI)
			c.httpMsg.Response = c.generalFunc.ErrPageResp(http.StatusForbidden, errPage.Len())
			c.httpMsg.Response.Body = io.NopCloser(bytes.NewBuffer(errPage.Bytes()))
			return utils.OkStatusCodeStr, c.httpMsg.Response, serviceHeaders
		}
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
		errPage := c.generalFunc.GenHtmlPage("service/unprocessable-file.html", reason, c.serviceName, serviceResp.Header["Request-Context"][0], c.httpMsg.Request.RequestURI)
		c.httpMsg.Response = c.generalFunc.ErrPageResp(http.StatusForbidden, errPage.Len())
		c.httpMsg.Response.Body = io.NopCloser(bytes.NewBuffer(errPage.Bytes()))
		return utils.OkStatusCodeStr, c.httpMsg.Response, serviceHeaders
	}
	serviceHeaders["CleanResult"] = "true"
	scannedFile := c.generalFunc.PreparingFileAfterScanning(file.Bytes(), reqContentType, c.methodName)
	return utils.OkStatusCodeStr, c.generalFunc.ReturningHttpMessageWithFile(c.methodName, scannedFile), serviceHeaders
}

func (c *CloudMersive) SendFileToAPI(f *bytes.Buffer, filename string) (*http.Response, error) {
	url := c.BaseURL + c.ScanEndPoint
	bodyBuf := &bytes.Buffer{}

	bodyWriter := multipart.NewWriter(bodyBuf)

	// adding policy in the request
	bodyWriter.WriteField("contentManagementFlagJson", c.policy)

	part, err := bodyWriter.CreateFormFile("file", filename)
	if err != nil {
		return nil, err
	}

	io.Copy(part, bytes.NewReader(f.Bytes()))
	if err := bodyWriter.Close(); err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, url, bodyBuf)
	if err != nil {
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
	req.Header.Add("restrictFileTypes", c.restrictFileTypes)
	fmt.Println("restrictFileTypes: ", c.restrictFileTypes)
	req.Header.Add("Content-Type", "multipart/form-data")
	req.Header.Add("Apikey", c.APIKey)
	req.Header.Set("Content-Type", bodyWriter.FormDataContentType())

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: utils.InitSecure()},
	}
	client := &http.Client{Transport: tr}
	ctx, _ := context.WithTimeout(context.Background(), c.Timeout)

	// defer cancel() commit cancel must be on goroutine
	req = req.WithContext(ctx)

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return res, nil
}

func (c *CloudMersive) ISTagValue() string {
	epochTime := strconv.FormatInt(time.Now().Unix(), 10)
	return "epoch-" + epochTime
}
