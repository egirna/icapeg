package utils

import (
	"bytes"
	"fmt"
	"html/template"
	"icapeg/dtos"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/h2non/filetype"
	"github.com/spf13/viper"
)

// The template names
const (
	BadFileTemplate = "templates/badfile.html"
)

// GetContentType returns the content type from a http response
func GetContentType(r *http.Response) string {
	ct := r.Header.Get("Content-Type")
	cc := strings.Split(ct, ";")
	if len(cc) > 0 {
		return cc[0]
	}
	return cc[0]
}

// GetMimeExtension returns the mime type extension of the data
func GetMimeExtension(data []byte) string {
	kind, _ := filetype.Match(data)

	if kind == filetype.Unknown {
		return Unknown
	}

	return kind.Extension

}

// GetFileName returns the filename from the http request
func GetFileName(req *http.Request) string {

	u, _ := url.Parse(req.RequestURI)

	uu := strings.Split(u.EscapedPath(), "/")

	if len(uu) > 0 {
		return uu[len(uu)-1]
	}
	return ""
}

// GetFileExtension returns the file extension of the concerned file of the http request
func GetFileExtension(req *http.Request) string {
	filenameWithExt := GetFileName(req)

	if filenameWithExt != "" {
		ff := strings.Split(filenameWithExt, ".")
		if len(ff) > 1 {
			return ff[len(ff)-1]
		}
	}

	return ""
}

// InStringSlice determines whether a string slices contains the data
func InStringSlice(data string, ss []string) bool {
	for _, s := range ss {
		if data == s {
			return true
		}
	}
	return false
}

// GetTemplateBufferAndResponse returns the html template buffer and the response to be returned for RESPMOD
func GetTemplateBufferAndResponse(templateName string, data *dtos.TemplateData) (*bytes.Buffer, *http.Response) {
	tmpl, _ := template.ParseFiles(templateName)
	htmlBuf := &bytes.Buffer{}
	tmpl.Execute(htmlBuf, data)
	newResp := &http.Response{
		StatusCode: http.StatusForbidden,
		Status:     http.StatusText(http.StatusForbidden),
		Header: http.Header{
			"Content-Type":   []string{"text/html"},
			"Content-Length": []string{strconv.Itoa(htmlBuf.Len())},
		},
	}

	return htmlBuf, newResp
}

// ByteToMegaBytes returns the mega-byte equivalence of the byte
func ByteToMegaBytes(b int) float64 {
	return float64(b) / 1000000
}

// BreakHTTPURL breaks the http url to hxxp so that its known to be malicious and can't be used
func BreakHTTPURL(url string) string {
	uu := strings.Split(url, ":")
	newURL := ""
	if uu[0] == "http" || uu[0] == "https" {
		brokenProtocol := strings.Replace(uu[0], "t", "x", 2)
		uu[0] = brokenProtocol
		newURL = strings.Join(uu, ":")
	} else {
		newURL = url
	}
	return newURL
}

// GetScannerVendorSpecificCfg returns the current scanner vendor specific configuration field
func GetScannerVendorSpecificCfg(mode, cfgField string) string {

	absoluteCfgField := ""

	switch mode {
	case ICAPModeResp:
		absoluteCfgField = fmt.Sprintf("%s.%s", viper.GetString("app.resp_scanner_vendor"), cfgField)
	case ICAPModeReq:
		absoluteCfgField = fmt.Sprintf("%s.%s", viper.GetString("app.req_scanner_vendor"), cfgField)
	}

	return absoluteCfgField
}

// IfPropagateError returns one of the given parameter depending on the propagate error configuration
func IfPropagateError(thenStatus, elseStatus int) int {
	if viper.GetBool("app.propagate_error") {
		return thenStatus
	}

	return elseStatus
}

// GetHTTPResponseCopy creates a new http.Response for the given one, including the body
func GetHTTPResponseCopy(resp *http.Response) http.Response {
	b, _ := ioutil.ReadAll(resp.Body)
	copyResp := *resp
	copyResp.Body = ioutil.NopCloser(bytes.NewBuffer(b))
	resp.Body = ioutil.NopCloser(bytes.NewBuffer(b))
	return copyResp
}

// CopyHeaders copy all the headers from the source(src) to destination(dest), without the provided header(if any)
func CopyHeaders(src map[string][]string, dest http.Header, without string) {
	for header, values := range src {
		if header == without {
			continue
		}

		for _, value := range values {
			dest.Add(header, value)
		}
	}
}

// GetNewURL generates a new URL for a http request with a URL with no scheme
func GetNewURL(req *http.Request) *url.URL {
	u, _ := url.Parse("http://" + req.Host + req.URL.Path)
	return u
}

// CopyBuffer creates a new buffer from the given one
func CopyBuffer(buf *bytes.Buffer) *bytes.Buffer {
	if buf != nil {
		return bytes.NewBuffer(buf.Bytes())
	}
	return nil
}
