package helpers

import (
	"bytes"
	"fmt"
	"html/template"
	"icapeg/dtos"
	"net/http"
	"strconv"
	"strings"

	"github.com/gabriel-vasile/mimetype"
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
	mime := mimetype.Detect(data)
	return mime.Extension()
}

// GetFileName returns the filename from the http request
func GetFileName(req *http.Request) string {
	uu := strings.Split(req.RequestURI, "/")
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
			return ff[1]
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
func GetScannerVendorSpecificCfg(cfgField string) string {
	return fmt.Sprintf("%s.%s", viper.GetString("app.scanner_vendor"), cfgField)
}
