package grayimages

import (
	"bytes"
	"icapeg/utils"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"strconv"
	"time"
)

const GrayImagesIdentifier = "GRAYIMAGES ID"

func (g *GrayImages) Processing(partial bool) (int, interface{}, map[string]string) {
	log.Println("processing")
	serviceHeaders := make(map[string]string)
	// no need to scan part of the file, this service needs all the file at ine time
	if partial {
		return utils.Continue, nil, nil
	}

	isGzip := false

	//extracting the file from http message
	file, reqContentType, err := g.generalFunc.CopyingFileToTheBuffer(g.methodName)
	if err != nil {
		log.Println("30")
		return utils.InternalServerErrStatusCodeStr, nil, serviceHeaders
	}
	// check if file is compressed
	isGzip = g.generalFunc.IsBodyGzipCompressed(g.methodName)
	if isGzip {
		if file, err = g.generalFunc.DecompressGzipBody(file); err != nil {
			return utils.InternalServerErrStatusCodeStr, nil, nil
		}
	}
	//getting the extension of the file
	var contentType []string
	if len(contentType) == 0 {
		contentType = append(contentType, "")
	}
	// getting file extension
	var fileName string
	if g.methodName == utils.ICAPModeReq {
		contentType = g.httpMsg.Request.Header["Content-Type"]
		fileName = utils.GetFileName(g.httpMsg.Request)
	} else {
		contentType = g.httpMsg.Response.Header["Content-Type"]
		fileName = utils.GetFileName(g.httpMsg.Response)
	}
	if len(contentType) == 0 {
		contentType = append(contentType, "")
	}
	fileExtension := utils.GetMimeExtension(file.Bytes(), contentType[0], fileName)

	/*isProcess, icapStatus, httpMsg := g.generalFunc.CheckTheExtension(fileExtension, g.extArrs,
		g.processExts, g.rejectExts, g.bypassExts, g.return400IfFileExtRejected, isGzip,
		g.serviceName, g.methodName, GrayImagesIdentifier, g.httpMsg.Request.RequestURI, reqContentType, file)
	if !isProcess {
		return icapStatus, httpMsg, serviceHeaders
	}*/

	if fileExtension != "png" && fileExtension != "jpeg" && fileExtension != "jpg" && fileExtension != "webp" {
		originalFile := file.Bytes()
		if isGzip {
			originalFile, err = g.generalFunc.CompressFileGzip(originalFile)
			if err != nil {
				return utils.InternalServerErrStatusCodeStr, nil, serviceHeaders
			}
		}
		originalFile = g.generalFunc.PreparingFileAfterScanning(originalFile, reqContentType, g.methodName)
		return utils.NoModificationStatusCodeStr, g.generalFunc.ReturningHttpMessageWithFile(g.methodName, originalFile), serviceHeaders
	}
	// sending request to gray images api
	scannedFile := file.Bytes()
	serviceResp, err := g.SendFileToAPI(file, fileExtension, fileName)
	if err != nil || serviceResp.StatusCode != 200 {
		if serviceResp != nil {
			log.Println(serviceResp.StatusCode)
			log.Println(serviceResp.Body)
		}
		// if file recompress the file if it was compressed
		if isGzip {
			scannedFile, err = g.generalFunc.CompressFileGzip(scannedFile)
			if err != nil {
				return utils.InternalServerErrStatusCodeStr, nil, serviceHeaders
			}
		}
		// send file with no modification code
		scannedFile = g.generalFunc.PreparingFileAfterScanning(scannedFile, reqContentType, g.methodName)
		return utils.NoModificationStatusCodeStr, g.generalFunc.ReturningHttpMessageWithFile(g.methodName, scannedFile), serviceHeaders
	}
	// get the byte array from response body
	scannedFile, err = ioutil.ReadAll(serviceResp.Body)
	if err != nil {
		if isGzip {
			scannedFile, err = g.generalFunc.CompressFileGzip(scannedFile)
			if err != nil {
				return utils.InternalServerErrStatusCodeStr, nil, serviceHeaders
			}
		}
		scannedFile = g.generalFunc.PreparingFileAfterScanning(scannedFile, reqContentType, g.methodName)
		return utils.NoModificationStatusCodeStr, g.generalFunc.ReturningHttpMessageWithFile(g.methodName, scannedFile), serviceHeaders
	}
	// compress file again if it's decompressed
	if isGzip {
		scannedFile, err = g.generalFunc.CompressFileGzip(scannedFile)
		if err != nil {
			return utils.InternalServerErrStatusCodeStr, nil, serviceHeaders
		}
	}
	// send the image after gray scale
	scannedFile = g.generalFunc.PreparingFileAfterScanning(scannedFile, reqContentType, g.methodName)
	return utils.OkStatusCodeStr, g.generalFunc.ReturningHttpMessageWithFile(g.methodName, scannedFile), serviceHeaders
}

func (g *GrayImages) SendFileToAPI(f *bytes.Buffer, fileType string, fileName string) (*http.Response, error) {
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
		return nil, err
	}
	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	//defer res.Body.Close()
	//body, err := ioutil.ReadAll(res.Body)
	//log.Println(string(body))
	return res, nil
}

func (g *GrayImages) ISTagValue() string {
	epochTime := strconv.FormatInt(time.Now().Unix(), 10)
	return "epoch-" + epochTime
}
