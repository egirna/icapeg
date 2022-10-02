package general_functions

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"github.com/h2non/filetype"
	"html/template"
	"icapeg/consts"
	"icapeg/http-message"
	"icapeg/logging"
	"icapeg/readValues"
	services_utilities "icapeg/service/services-utilities"
	"icapeg/service/services-utilities/ContentTypes"
	"image"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// error page struct
type (
	ErrorPage struct {
		Reason       string `json:"reason"`
		ServiceName  string `json:"service_name"`
		RequestedURL string `json:"requested_url"`
		IdentifierId string `json:"identifier_id"`
	}
)

// GeneralFunc is a struct used for applying general functionalities that any service can apply
type GeneralFunc struct {
	httpMsg *http_message.HttpMsg
}

// NewGeneralFunc is used to create a new instance from the struct
func NewGeneralFunc(httpMsg *http_message.HttpMsg) *GeneralFunc {
	GeneralFunc := &GeneralFunc{
		httpMsg: httpMsg,
	}
	return GeneralFunc
}

// CopyingFileToTheBuffer is a func which used for extracting a file from the body of the http message
func (f *GeneralFunc) CopyingFileToTheBuffer(methodName string) (*bytes.Buffer, ContentTypes.ContentType, error) {
	logging.Logger.Info("extracting the body of HTTP message")
	file := &bytes.Buffer{}
	var err error
	var reqContentType ContentTypes.ContentType
	reqContentType = nil
	switch methodName {
	case utils.ICAPModeReq:
		file, reqContentType, err = f.copyingFileToTheBufferReq()
		break
	case utils.ICAPModeResp:
		file, err = f.copyingFileToTheBufferResp()
		break
	}
	if err != nil {
		return nil, nil, err
	}
	return file, reqContentType, nil
}

func (f *GeneralFunc) CheckTheExtension(fileExtension string, extArrs []services_utilities.Extension, processExts,
	rejectExts, bypassExts []string, return400IfFileExtRejected, isGzip bool, serviceName, methodName, identifier,
	requestURI string, reqContentType ContentTypes.ContentType, file *bytes.Buffer) (bool, int, interface{}) {
	logging.Logger.Info("checking the extension (reject or bypass or process)")
	for i := 0; i < 3; i++ {
		if extArrs[i].Name == utils.ProcessExts {
			if f.ifFileExtIsX(fileExtension, processExts) {
				logging.Logger.Debug("extension is process")
				break
			}
		} else if extArrs[i].Name == utils.RejectExts {
			if f.ifFileExtIsX(fileExtension, rejectExts) {
				logging.Logger.Debug("extension is reject")
				if return400IfFileExtRejected {
					return false, utils.BadRequestStatusCodeStr, nil
				}
				if methodName == "RESPMOD" {
					errPage := f.GenHtmlPage(utils.BlockPagePath, utils.ErrPageReasonFileRejected, serviceName, identifier, requestURI)
					f.httpMsg.Response = f.ErrPageResp(http.StatusForbidden, errPage.Len())
					f.httpMsg.Response.Body = io.NopCloser(bytes.NewBuffer(errPage.Bytes()))
					return false, utils.OkStatusCodeStr, f.httpMsg.Response
				} else {
					htmlPage, req, err := f.ReqModErrPage(utils.ErrPageReasonFileRejected, serviceName, "NO ID")
					if err != nil {
						return false, utils.InternalServerErrStatusCodeStr, nil
					}
					reqContentType = &ContentTypes.RegularFile{
						Buf:     file,
						Encoded: false,
					}
					fileAfterPrep := f.PreparingFileAfterScanning(htmlPage.Bytes(), reqContentType, methodName)
					req.Body = io.NopCloser(bytes.NewBuffer(fileAfterPrep))
					return false, utils.OkStatusCodeStr, req
				}
			}
		} else if extArrs[i].Name == utils.BypassExts {
			if f.ifFileExtIsX(fileExtension, bypassExts) {
				logging.Logger.Debug("extension is bypass")
				fileAfterPrep, httpMsg := f.IfICAPStatusIs204(methodName, utils.NoModificationStatusCodeStr,
					file, isGzip, reqContentType, f.httpMsg)
				if fileAfterPrep == nil && httpMsg == nil {
					return false, utils.InternalServerErrStatusCodeStr, nil
				}

				//returning the http message and the ICAP status code
				switch msg := httpMsg.(type) {
				case *http.Request:
					msg.Body = io.NopCloser(bytes.NewBuffer(fileAfterPrep))
					return false, utils.NoModificationStatusCodeStr, msg
				case *http.Response:
					msg.Body = io.NopCloser(bytes.NewBuffer(fileAfterPrep))
					return false, utils.NoModificationStatusCodeStr, msg
				}
				return false, utils.NoModificationStatusCodeStr, nil
			}
		}
	}
	return true, 0, nil
}

// copyingFileToTheBufferResp is a utility function for CopyingFileToTheBuffer func
// it's used for extracting a file from the body of the http response
func (f *GeneralFunc) copyingFileToTheBufferResp() (*bytes.Buffer, error) {
	file := &bytes.Buffer{}
	_, err := io.Copy(file, f.httpMsg.Response.Body)
	return file, err
}

// copyingFileToTheBufferReq is a utility function for CopyingFileToTheBuffer func
// it's used for extracting a file from the body of the http request
func (f *GeneralFunc) copyingFileToTheBufferReq() (*bytes.Buffer, ContentTypes.ContentType, error) {
	reqContentType := ContentTypes.GetContentType(f.httpMsg.Request)
	// getting the file from request and store it in buf as a type of bytes.Buffer
	file := reqContentType.GetFileFromRequest()
	return file, reqContentType, nil

}

// InStringSlice determines whether a string slices contains the data
func (f *GeneralFunc) inStringSlice(data string, ss []string) bool {
	for _, s := range ss {
		if data == s {
			return true
		}
	}
	return false
}

func (f *GeneralFunc) ifFileExtIsX(fileExtension string, arr []string) bool {
	if len(arr) == 1 && arr[0] == utils.Any {
		return true
	}
	if f.inStringSlice(fileExtension, arr) {
		return true
	}
	return false
}

// IsBodyGzipCompressed is a func used for checking if the body of
// the http message is compressed ing Gzip or not
func (f *GeneralFunc) IsBodyGzipCompressed(methodName string) bool {
	logging.Logger.Info("checking if the body is compressed in GZIP")
	switch methodName {
	case utils.ICAPModeReq:
		return f.httpMsg.Request.Header.Get("Content-Encoding") == "gzip"
		break
	case utils.ICAPModeResp:
		return f.httpMsg.Response.Header.Get("Content-Encoding") == "gzip"
		break
	}
	return false
}

// DecompressGzipBody is a func used for decompress files which compressed in Gzip
func (f *GeneralFunc) DecompressGzipBody(file *bytes.Buffer) (*bytes.Buffer, error) {
	logging.Logger.Info("decompressing the HTTP message body")
	reader, err := gzip.NewReader(file)
	defer reader.Close()
	result, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return bytes.NewBuffer(result), nil
}

func (f *GeneralFunc) ReqModErrPage(reason, serviceName, IdentifierId string) (*bytes.Buffer, *http.Request, error) {
	host := readValues.ReadValuesString("app.web_server_host")
	endpoint := readValues.ReadValuesString("app.web_server_endpoint")
	url := host + endpoint
	f.httpMsg.Request.URL.Scheme = ""
	f.httpMsg.Request.URL.Opaque = url
	f.httpMsg.Request.URL.Path = ""
	f.httpMsg.Request.URL.Host = host
	for key, _ := range f.httpMsg.Request.Header {
		f.httpMsg.Request.Header.Del(key)
	}
	reqUri := f.httpMsg.Request.RequestURI
	f.httpMsg.Request.Header.Set("Host", host)
	f.httpMsg.Request.Method = http.MethodGet
	f.httpMsg.Request.RequestURI = url
	f.httpMsg.Request.Body = io.NopCloser(strings.NewReader(""))
	htmlPageStructInstance := &ErrorPage{
		Reason:       reason,
		ServiceName:  serviceName,
		RequestedURL: reqUri,
		IdentifierId: IdentifierId,
	}
	req := &http.Request{
		URL:        f.httpMsg.Request.URL,
		RequestURI: f.httpMsg.Request.RequestURI,
		Header:     f.httpMsg.Request.Header,
		RemoteAddr: f.httpMsg.Request.RemoteAddr,
		Host:       f.httpMsg.Request.Host,
		Method:     f.httpMsg.Request.Method,
		Proto:      f.httpMsg.Request.Proto,
		ProtoMajor: f.httpMsg.Request.ProtoMajor,
		ProtoMinor: f.httpMsg.Request.ProtoMinor,
	}
	body, err := json.Marshal(htmlPageStructInstance)
	if err != nil {
		return nil, nil, err
	}
	return bytes.NewBuffer(body), req, err
}

// IfMaxFileSizeExc is a functions which used for deciding the right http message should be returned
// if the file size is greater than the max file size of the service
func (f *GeneralFunc) IfMaxFileSizeExc(returnOrigIfMaxSizeExc bool, serviceName, methodName string,
	file *bytes.Buffer, maxFileSize int) (int, *bytes.Buffer, interface{}) {
	logging.Logger.Debug("HTTP message body size exceeds the limit")
	logging.Logger.Debug("HTTP message body size: " + strconv.Itoa(file.Len()) +
		" MB, the allowed max file size: " + strconv.Itoa(maxFileSize) + " MB") //check if returning the original file option is enabled in this case or not
	//if yes, return no modification status code
	//if not, return an error page
	if returnOrigIfMaxSizeExc {
		return utils.NoModificationStatusCodeStr, file, nil
	} else {
		if methodName == utils.ICAPModeResp {
			htmlErrPage := f.GenHtmlPage(utils.BlockPagePath,
				utils.ErrPageReasonMaxFileExceeded, serviceName, "NO ID", f.httpMsg.Request.RequestURI)
			f.httpMsg.Response = f.ErrPageResp(http.StatusForbidden, htmlErrPage.Len())
			return utils.OkStatusCodeStr, htmlErrPage, f.httpMsg.Response
		} else {
			htmlPage, req, err := f.ReqModErrPage(utils.ErrPageReasonMaxFileExceeded, serviceName, "NO ID")
			if err != nil {
				return utils.InternalServerErrStatusCodeStr, htmlPage, req
			}
			return utils.OkStatusCodeStr, htmlPage, req
		}
	}
}

// GetFileName returns the filename from the http request
func (f *GeneralFunc) GetFileName() string {
	logging.Logger.Info("getting the file name")

	//req.RequestURI  inserting dummy response if the http request is nil
	var Requrl string
	if f.httpMsg.Request == nil || f.httpMsg.Request.RequestURI == "" {
		Requrl = "http://www.example/images/unnamed_file"

	} else {
		Requrl = f.httpMsg.Request.RequestURI
		if Requrl[len(Requrl)-1] == '/' {
			Requrl = Requrl[0 : len(Requrl)-1]
		} else {
			Requrl = "http://www.example/images/unnamed_file"
		}
	}
	u, _ := url.Parse(Requrl)

	uu := strings.Split(u.EscapedPath(), "/")

	if len(uu) > 0 {
		return uu[len(uu)-1]
	}
	return "unnamed_file"
}

// CompressFileGzip is a func which used for compress files in gzip
func (f *GeneralFunc) CompressFileGzip(scannedFile []byte) ([]byte, error) {
	logging.Logger.Info("compressing the file in GZIP")
	var newBuf bytes.Buffer
	gz := gzip.NewWriter(&newBuf)
	if _, err := gz.Write(scannedFile); err != nil {
		return nil, err
	}
	gz.Close()
	return newBuf.Bytes(), nil
}

// ErrPageResp is a func used for creating http response for returning an error page
func (f *GeneralFunc) ErrPageResp(status int, pageContentLength int) *http.Response {
	logging.Logger.Info("preparing http response with the block page")
	return &http.Response{
		StatusCode: status,
		Status:     strconv.Itoa(status) + " " + http.StatusText(status),
		Header: http.Header{
			utils.ContentType:   []string{utils.HTMLContentType},
			utils.ContentLength: []string{strconv.Itoa(pageContentLength)},
		},
	}
}

// GenHtmlPage is a func used for generating an error page
func (f *GeneralFunc) GenHtmlPage(path, reason, serviceName, identifierId, reqUrl string) *bytes.Buffer {
	logging.Logger.Info("preparing a block page")
	htmlTmpl, _ := template.ParseFiles(path)
	htmlErrPage := &bytes.Buffer{}
	htmlTmpl.Execute(htmlErrPage, &ErrorPage{
		Reason:       reason,
		ServiceName:  serviceName,
		RequestedURL: reqUrl,
		IdentifierId: identifierId,
	})
	return htmlErrPage
}

// PreparingFileAfterScanning is a func used for preparing the http response before returning it
// preparing means converting the file to the original structure before scanning
func (f *GeneralFunc) PreparingFileAfterScanning(scannedFile []byte, reqContentType ContentTypes.ContentType, methodName string) []byte {
	logging.Logger.Info("preparing body after the service finished processing")
	switch methodName {
	case utils.ICAPModeReq:
		return []byte(reqContentType.BodyAfterScanning(scannedFile))
	}
	return scannedFile
}

// IfStatusIs204WithFile handling the HTTP message if the status should be 204 no modifications
func (f *GeneralFunc) IfStatusIs204WithFile(methodName string, status int, file *bytes.Buffer, isGzip bool,
	reqContentType ContentTypes.ContentType, httpMessage interface{}, isErr bool) ([]byte,
	interface{}) {
	var fileAfterPrep []byte
	var err error
	if isGzip {
		fileAfterPrep, err = f.CompressFileGzip(file.Bytes())
		if err != nil {
			return nil, nil
		}
	}

	if methodName == utils.ICAPModeReq {
		if isErr {
			reqContentType = &ContentTypes.RegularFile{
				Buf:     file,
				Encoded: false,
			}
		}
		fileAfterPrep = f.PreparingFileAfterScanning(file.Bytes(), reqContentType, methodName)
	} else {
		fileAfterPrep = file.Bytes()
	}
	if status == utils.NoModificationStatusCodeStr {
		return fileAfterPrep, f.ReturningHttpMessageWithFile(methodName, fileAfterPrep)
	}
	return fileAfterPrep, httpMessage
}

// ReturningHttpMessageWithFile function to return the suitable http message (http request, http response)
func (f *GeneralFunc) ReturningHttpMessageWithFile(methodName string, file []byte) interface{} {
	logging.Logger.Info("returning the HTTP message after processing by the service")
	switch methodName {
	case utils.ICAPModeReq:
		f.httpMsg.Request.Header.Set(utils.ContentLength, strconv.Itoa(len(string(file))))
		f.httpMsg.Request.Body = io.NopCloser(bytes.NewBuffer(file))
		if f.httpMsg.Request.URL.Scheme == "" {
			f.httpMsg.Request.URL.Opaque = f.httpMsg.Request.URL.Host
		}
		return f.httpMsg.Request
	case utils.ICAPModeResp:
		f.httpMsg.Response.Header.Set(utils.ContentLength, strconv.Itoa(len(string(file))))
		f.httpMsg.Response.Body = io.NopCloser(bytes.NewBuffer(file))
		return f.httpMsg.Response
	}
	return nil
}

func (f *GeneralFunc) IfICAPStatusIs204(methodName string, status int, file *bytes.Buffer, isGzip bool,
	reqContentType ContentTypes.ContentType, httpMessage interface{}) ([]byte,
	interface{}) {
	var fileAfterPrep []byte
	var err error
	if isGzip {
		fileAfterPrep, err = f.CompressFileGzip(file.Bytes())
		if err != nil {
			return nil, nil
		}
	}

	if methodName == utils.ICAPModeReq {
		fileAfterPrep = f.PreparingFileAfterScanning(file.Bytes(), reqContentType, methodName)
	} else {
		fileAfterPrep = file.Bytes()
	}
	if status == utils.NoModificationStatusCodeStr {
		return fileAfterPrep, f.returningHttpMessage(methodName, fileAfterPrep)
	}
	return fileAfterPrep, httpMessage
}

// function to return the suitable http message (http request, http response)
func (f *GeneralFunc) returningHttpMessage(methodName string, file []byte) interface{} {
	logging.Logger.Info("returning the HTTP message after processing by the service")
	switch methodName {
	case utils.ICAPModeReq:
		f.httpMsg.Request.Header.Set(utils.ContentLength, strconv.Itoa(len(string(file))))
		f.httpMsg.Request.Body = io.NopCloser(bytes.NewBuffer(file))
		return f.httpMsg.Request
	case utils.ICAPModeResp:
		f.httpMsg.Response.Header.Set(utils.ContentLength, strconv.Itoa(len(string(file))))
		f.httpMsg.Response.Body = io.NopCloser(bytes.NewBuffer(file))
		return f.httpMsg.Response
	}
	return nil
}

// GetDecodedImage takes the HTTP file and converts it to an image object
func (f *GeneralFunc) GetDecodedImage(file *bytes.Buffer) (image.Image, error) {
	logging.Logger.Info("getting decoded image")
	img, _, err := image.Decode(file)
	return img, err
}

// InitSecure set insecure flag based on user input
func (f *GeneralFunc) InitSecure(VerifyServerCert bool) bool {
	if !VerifyServerCert {
		return true
	}
	return false
}

// GetMimeExtension returns the mime type extension of the data
func (f *GeneralFunc) GetMimeExtension(data []byte, contentType string, filename string) string {
	logging.Logger.Info("getting the mime extension of the HTTP message body")
	kind, _ := filetype.Match(data)
	exts := map[string]string{"application/xml": "xml", "application/html": "html", "text/html": "html", "text/json": "html", "application/json": "json", "text/plain": "txt"}
	contentType = strings.Split(contentType, ";")[0]
	if kind == filetype.Unknown {
		if _, ok := exts[contentType]; ok {
			logging.Logger.Debug("HTTP message body mime extension is " + kind.Extension)
			return exts[contentType]
		}
	}
	if kind == filetype.Unknown {
		filenameArr := strings.Split(filename, ".")
		if len(filenameArr) > 1 {
			return filenameArr[len(filenameArr)-1]
		}
	}
	if kind == filetype.Unknown {
		logging.Logger.Debug("HTTP message body mime extension is " + kind.Extension)
		return utils.Unknown
	}
	logging.Logger.Debug("HTTP message body mime extension is " + kind.Extension)
	return kind.Extension

}
