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
	"net/url"
	"strconv"
	"strings"
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

type GeneralFunc struct {
	req     *http.Request
	resp    *http.Response
	elapsed time.Duration
	logger  *logger.ZLogger
}

func NewGeneralFunc(req *http.Request, resp *http.Response, elapsed time.Duration, logger *logger.ZLogger) *GeneralFunc {
	GeneralFunc := &GeneralFunc{
		req:     req,
		resp:    resp,
		logger:  logger,
		elapsed: elapsed,
	}
	return GeneralFunc
}

func (f *GeneralFunc) CopyingFileToTheBuffer(methodName string) (*bytes.Buffer, error) {
	buf := &bytes.Buffer{}
	var err error
	switch methodName {
	case utils.ICAPModeReq:
		_, err = io.Copy(buf, f.req.Body)
		break
	case utils.ICAPModeResp:
		_, err = io.Copy(buf, f.resp.Body)
		break
	}
	if err != nil {
		f.elapsed = time.Since(f.logger.LogStartTime)
		zLog.Error().Dur("duration", f.elapsed).Err(err).
			Str("value", "Failed to copy the response body to buffer").Msgf("read_request_body_error")
		return nil, err
	}
	return buf, nil
}

func (f *GeneralFunc) IsBodyGzipCompressed(methodName string) bool {
	switch methodName {
	case utils.ICAPModeReq:
		return f.req.Header.Get("Content-Encoding") == "gzip"
		break
	case utils.ICAPModeResp:
		return f.resp.Header.Get("Content-Encoding") == "gzip"
		break
	}
	return false
}

func (f *GeneralFunc) DecompressGzipBody(file *bytes.Buffer) (*bytes.Buffer, error) {
	reader, _ := gzip.NewReader(file)
	var result []byte
	result, err := ioutil.ReadAll(reader)
	if err != nil {
		f.elapsed = time.Since(f.logger.LogStartTime)
		zLog.Error().Dur("duration", f.elapsed).Err(err).Str("value", "failed to decompress input file").
			Msgf("decompress_gz_file_failed")
		return nil, err
	}
	return bytes.NewBuffer(result), nil
}

func (f *GeneralFunc) IfMaxFileSeizeExc(returnOrigIfMaxSizeExc bool, file *bytes.Buffer, maxFileSize int) (int, *bytes.Buffer, *http.Response) {
	zLog.Debug().Dur("duration", f.elapsed).Str("value",
		fmt.Sprintf("file size exceeds max filesize limit %d", maxFileSize)).
		Msgf("large_file_size")
	if returnOrigIfMaxSizeExc {
		return utils.NoModificationStatusCodeStr, file, nil
	} else {
		htmlErrPage := GenHtmlPage("service/unprocessable-file.html",
			"The Max file size is exceeded", f.req.RequestURI)
		f.resp = ErrPageResp(http.StatusForbidden, htmlErrPage.Len())
		fmt.Println(f.resp.StatusCode)
		return utils.OkStatusCodeStr, htmlErrPage, f.resp
	}
}

// GetFileName returns the filename from the http request
func (f *GeneralFunc) GetFileName() string {
	// req.RequestURI  inserting dummy response if the http request is nil
	var Requrl string
	if f.req == nil || f.req.RequestURI == "" {
		Requrl = "http://www.example/images/unnamed_file"

	} else {
		Requrl = f.req.RequestURI
		if Requrl[len(Requrl)-1] == '/' {
			Requrl = Requrl[0 : len(Requrl)-1]
		}

	}
	u, _ := url.Parse(Requrl)

	uu := strings.Split(u.EscapedPath(), "/")

	if len(uu) > 0 {
		return uu[len(uu)-1]
	}
	return "unnamed_file"
}

func (f *GeneralFunc) ExtractFileFromServiceResp(serviceResp *http.Response) ([]byte, error) {
	defer serviceResp.Body.Close()
	bodyByte, err := ioutil.ReadAll(serviceResp.Body)
	if err != nil {
		f.elapsed = time.Since(f.logger.LogStartTime)
		zLog.Error().Dur("duration", f.elapsed).Err(err).Str("value",
			"failed to read the response body from API response").
			Msgf("read_response_body_from_API_error")
		return nil, err
	}
	return bodyByte, nil
}

func (f *GeneralFunc) CompressFileGzip(scannedFile []byte) ([]byte, error) {
	var newBuf bytes.Buffer
	gz := gzip.NewWriter(&newBuf)
	if _, err := gz.Write(scannedFile); err != nil {
		f.elapsed = time.Since(f.logger.LogStartTime)
		zLog.Error().Dur("duration", f.elapsed).Err(err).
			Str("value", "failed to decompress input file").Msgf("decompress_gz_file_failed")
		return nil, err
	}
	gz.Close()
	return newBuf.Bytes(), nil
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
