package glasswall

import (
	"bytes"
	"compress/gzip"
	"fmt"
	zLog "github.com/rs/zerolog/log"
	"html/template"
	"icapeg/icap"
	"icapeg/logger"
	"icapeg/utils"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

type (
	errorPage struct {
		Reason                    string `json:"reason"`
		RequestedURL              string `json:"requested_url"`
		XAdaptationFileId         string `json:"x-adaptation-file-id"`
		XSdkEngineVersion         string `json:"x-sdk-engine-version"`
		XGlasswallCloudApiVersion string `json:"x-glasswall-cloud-api-version"`
	}
)

func CopyingFileToTheBuffer(req *http.Response, w icap.ResponseWriter,
	elapsed time.Duration, zlogger *logger.ZLogger) (*bytes.Buffer, error) {
	buf := &bytes.Buffer{}
	if _, err := io.Copy(buf, req.Body); err != nil {
		elapsed = time.Since(zlogger.LogStartTime)
		zLog.Error().Dur("duration", elapsed).Err(err).
			Str("value", "Failed to copy the response body to buffer").Msgf("read_request_body_error")
		w.WriteHeader(http.StatusNoContent, nil, false)
		return nil, err
	}
	return buf, nil
}

func DecodeGzip(buf *bytes.Buffer, w icap.ResponseWriter,
	elapsed time.Duration, zlogger *logger.ZLogger) (*bytes.Buffer, error) {
	reader, _ := gzip.NewReader(buf)
	var result []byte
	result, err := ioutil.ReadAll(reader)
	if err != nil {
		elapsed = time.Since(zlogger.LogStartTime)
		zLog.Error().Dur("duration", elapsed).Err(err).Str("value", "failed to decompress input file").
			Msgf("decompress_gz_file_failed")
		w.WriteHeader(http.StatusBadRequest, nil, false)
		return nil, err
	}
	return bytes.NewBuffer(result), nil
}

func MaxFileSeizeExc(returnOrigIfMaxSizeExc, is204Allowed bool, w icap.ResponseWriter, req *http.Request,
	resp *http.Response, file *bytes.Buffer, maxFileSize int, elapsed time.Duration, zlogger *logger.ZLogger) {
	zLog.Debug().Dur("duration", elapsed).Str("value",
		fmt.Sprintf("file size exceeds max filesize limit %d", maxFileSize)).
		Msgf("large_file_size")
	if returnOrigIfMaxSizeExc {
		if is204Allowed {
			w.WriteHeader(utils.NoModificationStatusCodeStr, nil, false)
		} else {
			resp.Body = io.NopCloser(file)
			w.WriteHeader(utils.OkStatusCodeStr, resp, true)
			w.Write(file.Bytes())
		}
	} else {
		htmlErrPage := GenHtmlPage("service/unprocessable-file.html",
			"The Max file size is exceeded", req.RequestURI)
		newResp := ErrPageResp(http.StatusForbidden, htmlErrPage.Len())
		w.WriteHeader(utils.OkStatusCodeStr, newResp, true)
		w.Write(htmlErrPage.Bytes())
	}
}

func ErrPageResp(status int, pageContentLength int) *http.Response {
	return &http.Response{
		StatusCode: status,
		Status:     strconv.Itoa(status) + " " + http.StatusText(status),
		Header: http.Header{
			utils.ContentType:   []string{utils.HTMLContentType},
			utils.ContentLength: []string{strconv.Itoa(pageContentLength)},
		},
	}
}

func GenHtmlPage(path, reason, reqUrl string) *bytes.Buffer {
	htmlTmpl, _ := template.ParseFiles(path)
	htmlErrPage := &bytes.Buffer{}
	htmlTmpl.Execute(htmlErrPage, &errorPage{
		Reason:       reason,
		RequestedURL: reqUrl,
	})
	return htmlErrPage
}

func ApiRespAnalysis(serviceResp *http.Response, w icap.ResponseWriter, isGzip bool,
	elapsed time.Duration, zlogger *logger.ZLogger) ([]byte, error) {
	defer serviceResp.Body.Close()
	bodyByte, err := ioutil.ReadAll(serviceResp.Body)
	if err != nil {
		elapsed = time.Since(zlogger.LogStartTime)
		zLog.Error().Dur("duration", elapsed).Err(err).Str("value",
			"failed to read the response body from API response").
			Msgf("read_response_body_from_API_error")
		w.WriteHeader(http.StatusInternalServerError, nil, false)
		return nil, err
	}
	if isGzip {
		var newBuf bytes.Buffer
		gz := gzip.NewWriter(&newBuf)
		if _, err := gz.Write(bodyByte); err != nil {
			elapsed = time.Since(zlogger.LogStartTime)
			zLog.Error().Dur("duration", elapsed).Err(err).
				Str("value", "failed to decompress input file").Msgf("decompress_gz_file_failed")
			w.WriteHeader(http.StatusInternalServerError, nil, false)
			return nil, err
		}
		gz.Close()
		bodyByte = newBuf.Bytes()
	}
	zLog.Info().Dur("duration", elapsed).Err(err).Str("value", "file was processed").
		Msgf("file_processed_successfully")
	return bodyByte, nil
}

func compressGzip(scannedFile []byte, w icap.ResponseWriter,
	elapsed time.Duration, zlogger *logger.ZLogger) ([]byte, error) {
	var newBuf bytes.Buffer
	gz := gzip.NewWriter(&newBuf)
	if _, err := gz.Write(scannedFile); err != nil {
		elapsed = time.Since(zlogger.LogStartTime)
		zLog.Error().Dur("duration", elapsed).Err(err).
			Str("value", "failed to decompress input file").Msgf("decompress_gz_file_failed")
		w.WriteHeader(http.StatusInternalServerError, nil, false)
		return nil, err
	}
	gz.Close()
	return newBuf.Bytes(), nil
}
