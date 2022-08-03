package cloudmersive

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"icapeg/utils"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func (c CloudMersive) Processing(partial bool) (int, interface{}, map[string]string) {
	//TODO implement me

	// no need to scan part of the file, this service needs all the file at one time
	if partial {
		return utils.Continue, nil, nil
	}

	//extracting the file from http message
	file, reqContentType, err := c.generalFunc.CopyingFileToTheBuffer(c.methodName)
	if err != nil {
		return utils.InternalServerErrStatusCodeStr, nil, nil
	}
	// comparing file extension with restrictFileTypes list
	fileExtension := utils.GetMimeExtension(file.Bytes())
	fileAllowed := false
	if strings.Contains(c.restrictFileTypes, fileExtension) || c.restrictFileTypes == "" || fileExtension == utils.Unknown {
		fileAllowed = true
	}
	if !fileAllowed {
		reason := "File rejected"
		if c.return400IfFileExtRejected {
			return utils.BadRequestStatusCodeStr, nil, nil
		}
		errPage := c.generalFunc.GenHtmlPage("service/unprocessable-file.html", reason, c.httpMsg.Request.RequestURI)
		c.httpMsg.Response = c.generalFunc.ErrPageResp(http.StatusForbidden, errPage.Len())
		c.httpMsg.Response.Body = io.NopCloser(bytes.NewBuffer(errPage.Bytes()))
		return utils.OkStatusCodeStr, c.httpMsg.Response, nil
	}

}

func (c *CloudMersive) SendFileToAPI(f *bytes.Buffer, filename string) *http.Response {
	url := c.BaseURL + c.ScanEndPoint
	bodyBuf := &bytes.Buffer{}

	bodyWriter := multipart.NewWriter(bodyBuf)

	// adding policy in the request
	bodyWriter.WriteField("contentManagementFlagJson", c.policy)

	part, err := bodyWriter.CreateFormFile("file", filename)
	if err != nil {
		return nil
	}

	io.Copy(part, bytes.NewReader(f.Bytes()))
	if err := bodyWriter.Close(); err != nil {
		return nil
	}
	req, err := http.NewRequest(http.MethodPost, url, bodyBuf)
	if err != nil {
		return nil
	}
	req.Header.Add("allowExecutables", strconv.FormatBool(c.allowExecutables))
	req.Header.Add("allowInvalidFiles", strconv.FormatBool(c.allowInvalidFiles))
	req.Header.Add("allowScripts", strconv.FormatBool(c.allowScripts))
	req.Header.Add("allowPasswordProtectedFiles", strconv.FormatBool(c.allowPasswordProtectedFiles))
	req.Header.Add("allowMacros", strconv.FormatBool(c.allowMacros))
	req.Header.Add("allowXmlExternalEntities", strconv.FormatBool(c.allowXmlExternalEntities))
	req.Header.Add("restrictFileTypes", c.restrictFileTypes)
	req.Header.Add("Content-Type", "multipart/form-data")
	// TODO how to read environment variable from app.env?
	req.Header.Add("Apikey", os.Getenv("_AUTH_TOKENS"))
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
		return nil
	}
	defer res.Body.Close()

	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	return resp
}
