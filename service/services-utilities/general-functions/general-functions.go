package general_functions

import (
	"bytes"
	"compress/gzip"
	"errors"
	"html/template"
	services_utilities "icapeg/service/services-utilities"
	"icapeg/service/services-utilities/ContentTypes"
	"icapeg/utils"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// error page struct
type (
	errorPage struct {
		Reason       string `json:"reason"`
		ServiceName  string `json:"service_name"`
		RequestedURL string `json:"requested_url"`
		IdentifierId string `json:"identifier_id"`
	}
)

// GeneralFunc is a struct used for applying general functionalities that any service can apply
type GeneralFunc struct {
	httpMsg *utils.HttpMsg
}

// NewGeneralFunc is used to create a new instance from the struct
func NewGeneralFunc(httpMsg *utils.HttpMsg) *GeneralFunc {
	GeneralFunc := &GeneralFunc{
		httpMsg: httpMsg,
	}
	return GeneralFunc
}

// CopyingFileToTheBuffer is a func which used for extracting a file from the body of the http message
func (f *GeneralFunc) CopyingFileToTheBuffer(methodName string) (*bytes.Buffer, ContentTypes.ContentType, error) {
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
	for i := 0; i < 3; i++ {
		if extArrs[i].Name == utils.ProcessExts {
			if f.IfFileExtIsX(fileExtension, processExts) {
				break
			}
		} else if extArrs[i].Name == utils.RejectExts {
			if f.IfFileExtIsX(fileExtension, rejectExts) {
				reason := "File rejected"
				if return400IfFileExtRejected {
					return false, utils.BadRequestStatusCodeStr, nil
				}
				errPage := f.GenHtmlPage(utils.BlockPagePath, reason, serviceName, identifier, requestURI)
				f.httpMsg.Response = f.ErrPageResp(http.StatusForbidden, errPage.Len())
				f.httpMsg.Response.Body = io.NopCloser(bytes.NewBuffer(errPage.Bytes()))
				return false, utils.OkStatusCodeStr, f.httpMsg.Response
			}
		} else if extArrs[i].Name == utils.BypassExts {
			if f.IfFileExtIsX(fileExtension, bypassExts) {
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

// inStringSlice is a func which used for checking if a string element exists in a slice or not
func (f *GeneralFunc) inStringSlice(data string, ss []string) bool {
	for _, s := range ss {
		if data == s {
			return true
		}
	}
	return false
}

// IfFileExtIsBypass is a func to check if a file extension is bypass extension or not
func (f *GeneralFunc) IfFileExtIsBypass(fileExtension string, bypassExts []string) error {
	if utils.InStringSlice(fileExtension, bypassExts) {
		return errors.New("processing not required for file type")
	}
	return nil
}

func (f *GeneralFunc) IfFileExtIsX(fileExtension string, arr []string) bool {
	if len(arr) == 1 && arr[0] == "*" {
		return true
	}
	if utils.InStringSlice(fileExtension, arr) {
		return true
	}
	return false
}

// IfFileExtIsReject is a func to check if a file extension is bypass extension or not
func (f *GeneralFunc) IfFileExtIsReject(fileExtension string, rejectExts []string) error {
	if utils.InStringSlice(fileExtension, rejectExts) {
		return errors.New("processing rejected for file type")
	}
	return nil
}

// IfFileExtIsBypassAndNotProcess is a func to check if a file extension is bypass extension and not a process extension
func (f *GeneralFunc) IfFileExtIsBypassAndNotProcess(fileExtension string, bypassExts []string, processExts []string) error {
	if utils.InStringSlice(utils.Any, bypassExts) && !utils.InStringSlice(fileExtension, processExts) {
		// if extension does not belong to "All bypassable except the processable ones" group
		return errors.New("processing not required for file type")
	}
	return nil
}

// IsBodyGzipCompressed is a func used for checking if the body of
// the http message is compressed ing Gzip or not
func (f *GeneralFunc) IsBodyGzipCompressed(methodName string) bool {
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
	reader, err := gzip.NewReader(file)
	defer reader.Close()
	result, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return bytes.NewBuffer(result), nil
}

// IfMaxFileSeizeExc is a functions which used for deciding the right http message should be returned
// if the file size is greater than the max file size of the service
func (f *GeneralFunc) IfMaxFileSeizeExc(returnOrigIfMaxSizeExc bool, serviceName string, file *bytes.Buffer, maxFileSize int) (int, *bytes.Buffer, interface{}) {
	//check if returning the original file option is enabled in this case or not
	//if yes, return no modification status code
	//if not, return an error page
	if returnOrigIfMaxSizeExc {
		return utils.NoModificationStatusCodeStr, file, nil
	} else {
		htmlErrPage := f.GenHtmlPage(utils.BlockPagePath,
			"The Max file size is exceeded", serviceName, "NO ID", f.httpMsg.Request.RequestURI)
		f.httpMsg.Response = f.ErrPageResp(http.StatusForbidden, htmlErrPage.Len())
		return utils.OkStatusCodeStr, htmlErrPage, f.httpMsg.Response
	}
}

// GetFileName returns the filename from the http request
func (f *GeneralFunc) GetFileName() string {
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

// ExtractFileFromServiceResp is a function which used for extracting file from
// the response of the API of the service
func (f *GeneralFunc) ExtractFileFromServiceResp(serviceResp *http.Response) ([]byte, error) {
	defer serviceResp.Body.Close()
	bodyByte, err := ioutil.ReadAll(serviceResp.Body)
	if err != nil {
		return nil, err
	}
	return bodyByte, nil
}

// CompressFileGzip is a func which used for compress files in gzip
func (f *GeneralFunc) CompressFileGzip(scannedFile []byte) ([]byte, error) {
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
	htmlTmpl, _ := template.ParseFiles(path)
	htmlErrPage := &bytes.Buffer{}
	htmlTmpl.Execute(htmlErrPage, &errorPage{
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
	switch methodName {
	case utils.ICAPModeReq:
		return []byte(reqContentType.BodyAfterScanning(scannedFile))
	}
	return scannedFile
}

// IfStatusIs204WithFile handling the HTTP message if the status should be 204 no modifications
func (f *GeneralFunc) IfStatusIs204WithFile(methodName string, status int, file *bytes.Buffer, isGzip bool, reqContentType ContentTypes.ContentType, httpMessage interface{}) ([]byte,
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
		return fileAfterPrep, f.ReturningHttpMessageWithFile(methodName, fileAfterPrep)
	}
	return fileAfterPrep, httpMessage
}

// ReturningHttpMessageWithFile function to return the suitable http message (http request, http response)
func (f *GeneralFunc) ReturningHttpMessageWithFile(methodName string, file []byte) interface{} {
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

func (f *GeneralFunc) IfICAPStatusIs204(methodName string, status int, file *bytes.Buffer, isGzip bool, reqContentType ContentTypes.ContentType, httpMessage interface{}) ([]byte,
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
