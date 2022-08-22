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
	"log"
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
	contentType := g.httpMsg.Response.Header["Content-Type"]
	var fileName string
	if g.methodName == utils.ICAPModeReq {
		fileName = utils.GetFileName(g.httpMsg.Request)
	} else {
		fileName = utils.GetFileName(g.httpMsg.Response)
	}
	fileExtension := utils.GetMimeExtension(file.Bytes(), contentType[0], fileName)

	isProcess, icapStatus, httpMsg := g.generalFunc.CheckTheExtension(fileExtension, g.extArrs,
		g.processExts, g.rejectExts, g.bypassExts, g.return400IfFileExtRejected, isGzip,
		g.serviceName, g.methodName, GrayImagesIdentifier, g.httpMsg.Request.RequestURI, reqContentType, file)
	if !isProcess {
		return icapStatus, httpMsg, serviceHeaders
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
		grayImg, err = g.webpImagehandler(file)
	} else if imgExtension == "png" {
		grayImg, err = g.pngImagehandler(file)
	} else if imgExtension == "jpeg" || imgExtension == "jpg" {
		grayImg, err = g.pngImagehandler(file)
	} else {
		return nil, errors.New("file is not a supported image")
	}
	return grayImg, err
}

func (g *GrayImages) webpImagehandler(file *bytes.Buffer) (*os.File, error) {
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

func (g *GrayImages) pngImagehandler(file *bytes.Buffer) (*os.File, error) {
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
