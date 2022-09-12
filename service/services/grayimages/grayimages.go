package grayimages

import (
	"bytes"
	utils "icapeg/consts"
	"icapeg/logging"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"strconv"
	"time"
)

func (g *Grayimages) Processing(partial bool) (int, interface{}, map[string]string) {
	logging.Logger.Info(g.serviceName + " service has started processing")
	serviceHeaders := make(map[string]string)

	// no need to scan part of the file, this service needs all the file at ine time
	if partial {
		logging.Logger.Info(g.serviceName + " service has stopped processing partially")
		return utils.Continue, nil, nil
	}

	//extracting the file from http message
	file, reqContentType, err := g.generalFunc.CopyingFileToTheBuffer(g.methodName)
	if err != nil {
		logging.Logger.Error(g.serviceName + " error: " + err.Error())
		logging.Logger.Info(g.serviceName + " service has stopped processing")
		return utils.InternalServerErrStatusCodeStr, nil, nil
	}

	//getting the extension of the file
	var contentType []string
	if len(contentType) == 0 {
		contentType = append(contentType, "")
	}
	var fileName string
	if g.methodName == utils.ICAPModeReq {
		contentType = g.httpMsg.Request.Header["Content-Type"]
		fileName = g.generalFunc.GetFileName()
	} else {
		contentType = g.httpMsg.Response.Header["Content-Type"]
		fileName = g.generalFunc.GetFileName()
	}
	if len(contentType) == 0 {
		contentType = append(contentType, "")
	}
	isGzip := g.generalFunc.IsBodyGzipCompressed(g.methodName)

	fileExtension := g.generalFunc.GetMimeExtension(file.Bytes(), contentType[0], fileName)

	isProcess, icapStatus, httpMsg := g.generalFunc.CheckTheExtension(fileExtension, g.extArrs,
		g.processExts, g.rejectExts, g.bypassExts, g.return400IfFileExtRejected, isGzip,
		g.serviceName, g.methodName, GrayimagesIdentifier, g.httpMsg.Request.RequestURI, reqContentType, file)
	if !isProcess {
		return icapStatus, httpMsg, serviceHeaders
	}
	// sending request to gray images api
	scannedFile := file.Bytes()
	serviceResp, err := g.SendFileToAPI(file, fileExtension, fileName)
	if err != nil || serviceResp.StatusCode != 200 {
		logging.Logger.Error(g.serviceName + " error: " + err.Error())
		// if file recompress the file if it was compressed
		if isGzip {
			scannedFile, err = g.generalFunc.CompressFileGzip(scannedFile)
			if err != nil {
				logging.Logger.Error(g.serviceName + " error: " + err.Error())
				logging.Logger.Info(g.serviceName + " service has stopped processing")
				return utils.InternalServerErrStatusCodeStr, nil, serviceHeaders
			}
		}
		// send file with no modification code
		scannedFile = g.generalFunc.PreparingFileAfterScanning(scannedFile, reqContentType, g.methodName)
		logging.Logger.Info(g.serviceName + " service has stopped processing")
		return utils.NoModificationStatusCodeStr, g.generalFunc.ReturningHttpMessageWithFile(g.methodName, scannedFile), serviceHeaders
	}
	// get the byte array from response body
	scannedFile, err = ioutil.ReadAll(serviceResp.Body)
	if err != nil {
		if isGzip {
			scannedFile, err = g.generalFunc.CompressFileGzip(scannedFile)
			if err != nil {
				logging.Logger.Error(g.serviceName + " error: " + err.Error())
				logging.Logger.Info(g.serviceName + " service has stopped processing")
				return utils.InternalServerErrStatusCodeStr, nil, serviceHeaders
			}
		}
		scannedFile = g.generalFunc.PreparingFileAfterScanning(scannedFile, reqContentType, g.methodName)
		logging.Logger.Info(g.serviceName + " service has stopped processing")
		return utils.NoModificationStatusCodeStr, g.generalFunc.ReturningHttpMessageWithFile(g.methodName, scannedFile), serviceHeaders
	}
	// compress file again if it's decompressed
	if isGzip {
		scannedFile, err = g.generalFunc.CompressFileGzip(scannedFile)
		if err != nil {
			logging.Logger.Error(g.serviceName + " error: " + err.Error())
			logging.Logger.Info(g.serviceName + " service has stopped processing")
			return utils.InternalServerErrStatusCodeStr, nil, serviceHeaders
		}
	}
	// send the image after gray scale
	scannedFile = g.generalFunc.PreparingFileAfterScanning(scannedFile, reqContentType, g.methodName)
	logging.Logger.Info(g.serviceName + " service has stopped processing")
	return utils.OkStatusCodeStr, g.generalFunc.ReturningHttpMessageWithFile(g.methodName, scannedFile), serviceHeaders
}

func (g *Grayimages) SendFileToAPI(f *bytes.Buffer, fileType string, fileName string) (*http.Response, error) {
	logging.Logger.Debug("sending the image to API to process it")
	var url string
	switch fileType {
	case "png":
		url = g.BaseURL + "/png"
	case "webp":
		url = g.BaseURL + "/webp"
	case "jpeg":
		url = g.BaseURL + "/jpeg"
	case "jpg":
		url = g.BaseURL + "/jpeg"
	}
	log.Println("113, url: ", url)
	method := "POST"

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	part1,
		errFile1 := writer.CreateFormFile("img", fileName)
	_, errFile1 = io.Copy(part1, bytes.NewReader(f.Bytes()))
	if errFile1 != nil {
		return nil, errFile1
	}
	err := writer.Close()
	if err != nil {
		logging.Logger.Error(g.serviceName + " error: " + err.Error())
		return nil, err
	}
	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		logging.Logger.Error(g.serviceName + " error: " + err.Error())
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := client.Do(req)
	if err != nil {
		logging.Logger.Error(g.serviceName + " error: " + err.Error())
		return nil, err
	}
	return res, nil
}

func (g *Grayimages) ISTagValue() string {
	epochTime := strconv.FormatInt(time.Now().Unix(), 10)
	return "epoch-" + epochTime
}
