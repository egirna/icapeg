package grayimages

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/kolesa-team/go-webp/decoder"
	"github.com/kolesa-team/go-webp/webp"
	"icapeg/utils"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

const GrayImagesIdentifier = "GRAYIMAGES ID"

func (g *GrayImages) Processing(partial bool) (int, interface{}, map[string]string) {
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
	//if it's compressed, we decompress it to send it to Glasswall service
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

	//check if the file extension is a bypass extension
	//if yes we will not modify the file, and we will return 204 No modifications
	isProcess, icapStatus, httpMsg := g.generalFunc.CheckTheExtension(fileExtension, g.extArrs,
		g.processExts, g.rejectExts, g.bypassExts, g.return400IfFileExtRejected, isGzip,
		g.serviceName, g.methodName, GrayImagesIdentifier, g.httpMsg.Request.RequestURI, reqContentType, file)
	if !isProcess {
		return icapStatus, httpMsg, serviceHeaders
	}

	//check if the file size is greater than max file size of the service
	//if yes we will return 200 ok or 204 no modification, it depends on the configuration of the service
	if g.maxFileSize != 0 && g.maxFileSize < file.Len() {
		status, file, httpMsg := g.generalFunc.IfMaxFileSizeExc(g.returnOrigIfMaxSizeExc, g.serviceName, file)
		fileAfterPrep, httpMsg := g.generalFunc.IfStatusIs204WithFile(g.methodName, status, file, isGzip, reqContentType, httpMsg)
		if fileAfterPrep == nil && httpMsg == nil {
			return utils.InternalServerErrStatusCodeStr, nil, serviceHeaders
		}
		switch msg := httpMsg.(type) {
		case *http.Request:
			msg.Body = io.NopCloser(bytes.NewBuffer(fileAfterPrep))
			return status, msg, nil
		case *http.Response:
			msg.Body = io.NopCloser(bytes.NewBuffer(fileAfterPrep))
			return status, msg, nil
		}
		return status, nil, nil
	}

	// convert the HTTP img to grayscale
	scale, err := g.ConvertImgToGrayScale(fileExtension, file)
	scannedFile := file.Bytes()
	if err != nil {
		if isGzip {
			// compress file again if it's decompressed
			scannedFile, err = g.generalFunc.CompressFileGzip(scannedFile)
			if err != nil {
				return utils.InternalServerErrStatusCodeStr, nil, serviceHeaders
			}
		}
		// return the same file if it can't be gray
		scannedFile = g.generalFunc.PreparingFileAfterScanning(scannedFile, reqContentType, g.methodName)
		return utils.NoModificationStatusCodeStr, g.generalFunc.ReturningHttpMessageWithFile(g.methodName, scannedFile), serviceHeaders
	}
	// convert grayImg into bytes
	scannedFile, err = os.ReadFile(scale.Name()) // just pass the file name
	// clean temp file on desk
	defer os.Remove(scale.Name())
	if err != nil {
		return utils.InternalServerErrStatusCodeStr, nil, serviceHeaders
	}
	// compress file again if it's decompressed
	if isGzip {
		scannedFile, err = g.generalFunc.CompressFileGzip(scannedFile)
		if err != nil {
			return utils.InternalServerErrStatusCodeStr, nil, serviceHeaders
		}
	}
	// return the gray image
	scannedFile = g.generalFunc.PreparingFileAfterScanning(scannedFile, reqContentType, g.methodName)
	return utils.OkStatusCodeStr, g.generalFunc.ReturningHttpMessageWithFile(g.methodName, scannedFile), serviceHeaders

}

func (g *GrayImages) ISTagValue() string {
	epochTime := strconv.FormatInt(time.Now().Unix(), 10)
	return "epoch-" + epochTime
}

func (g *GrayImages) ConvertImgToGrayScale(imgExtension string, file *bytes.Buffer) (*os.File, error) {
	var grayImg *os.File
	var err error
	if imgExtension == "webp" {
		grayImg, err = g.webpImageHandler(file)
	} else if imgExtension == "png" {
		grayImg, err = g.pngImageHandler(file)
	} else if imgExtension == "jpeg" || imgExtension == "jpg" {
		grayImg, err = g.jpegImageHandler(imgExtension, file)
	} else {
		return nil, errors.New("file is not a supported image")
	}
	return grayImg, err
}

// webpImageHandler convert webp images to gray images
func (g *GrayImages) webpImageHandler(file *bytes.Buffer) (*os.File, error) {
	// create temporarily file jpg file, used to convert original img data to gray
	tmpJpeg, err := os.CreateTemp(g.imagesDir, "*.jpg")
	if err != nil {
		return nil, err
	}
	// read webp image data
	webpDecode, err := webp.Decode(file, &decoder.Options{})
	if err != nil {
		return nil, err
	}
	// encode webp image data to a jpg file
	if err = jpeg.Encode(tmpJpeg, webpDecode, &jpeg.Options{Quality: 80}); err != nil {
		return nil, err
	}
	// read image bytes
	jpegBytes, err := os.ReadFile(tmpJpeg.Name())
	// clean the file on desk
	defer os.Remove(tmpJpeg.Name())
	// put files to buffer
	jpegBuffer := bytes.NewBuffer(jpegBytes)
	// get image object from jpeg data
	jpegImg, err := g.generalFunc.GetDecodedImage(jpegBuffer)
	if err != nil {
		return nil, err
	}
	// convert image to gray
	grayImg := image.NewGray(jpegImg.Bounds())
	for y := jpegImg.Bounds().Min.Y; y < jpegImg.Bounds().Max.Y; y++ {
		for x := jpegImg.Bounds().Min.X; x < jpegImg.Bounds().Max.X; x++ {
			grayImg.Set(x, y, jpegImg.At(x, y))
		}
	}
	// create temporarily file for the gray image
	grayWebp, err := os.CreateTemp(g.imagesDir, "*.jpg")
	if err != nil {
		return nil, err
	}
	// encode gray image data into the file
	if err = jpeg.Encode(grayWebp, grayImg, nil); err != nil {
		return nil, err
	}
	return grayWebp, nil
}

// pngImageHandler convert png images to gray images
func (g *GrayImages) pngImageHandler(file *bytes.Buffer) (*os.File, error) {
	// convert HTTP file to image object
	img, err := g.generalFunc.GetDecodedImage(file)
	if err != nil {
		return nil, err
	}
	// convert the image to grayscale
	grayImg := image.NewGray(img.Bounds())
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			grayImg.Set(x, y, img.At(x, y))
		}
	}
	// Working with grayscale image, convert to png
	// create new temporarily png file
	newImg, err := os.CreateTemp(g.imagesDir, "*.png")
	log.Println(newImg.Name())
	if err != nil {
		return nil, err
	}
	// encode gray image data and save it into the created png file
	if err = png.Encode(newImg, grayImg); err != nil {
		return nil, err
	}
	// return the png file after converting it to gray image
	return newImg, nil

}

// jpegImageHandler convert jpeg/jpg images to gray images
func (g *GrayImages) jpegImageHandler(imgExtension string, file *bytes.Buffer) (*os.File, error) {
	// convert HTTP file to image object
	img, err := g.generalFunc.GetDecodedImage(file)
	if err != nil {
		return nil, err
	}
	// convert the image to grayscale
	grayImg := image.NewGray(img.Bounds())
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			grayImg.Set(x, y, img.At(x, y))
		}
	}
	pattern := fmt.Sprintf("*.%s", imgExtension)
	// create new temporarily (jpeg or jpg) file
	newImg, err := os.CreateTemp(g.imagesDir, pattern)
	if err != nil {
		return nil, err
	}
	// encode gray image data and save it into the created jpeg/jpg file
	if err = jpeg.Encode(newImg, grayImg, nil); err != nil {
		return nil, err
	}
	// return the png file after converting it to gray image
	return newImg, nil
}
