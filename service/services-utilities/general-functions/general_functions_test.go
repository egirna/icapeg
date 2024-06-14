package general_functions

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"testing"

	utils "icapeg/consts"
	http_message "icapeg/http-message"
	"icapeg/logging"
	services_utilities "icapeg/service/services-utilities"
	"icapeg/service/services-utilities/ContentTypes"

	"go.uber.org/zap"
)

// Helper function to create an http_message.HttpMsg with request and response
func createHttpMsg() *http_message.HttpMsg {
	req := &http.Request{
		Header: http.Header{},
		Body:   io.NopCloser(bytes.NewBuffer([]byte("request body"))),
		URL:    &url.URL{Path: "/path/to/file.txt"},
	}
	resp := &http.Response{
		Header:  http.Header{},
		Body:    io.NopCloser(bytes.NewBuffer([]byte("response body"))),
		Request: req,
	}
	return &http_message.HttpMsg{
		Request:  req,
		Response: resp,
	}
}

// Helper function to initialize the logger
func initLogger() {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync() // flushes buffer, if any
	logging.Logger = logger
}

// Test for CopyingFileToTheBuffer method
func TestCopyingFileToTheBuffer(t *testing.T) {
	initLogger()
	httpMsg := createHttpMsg()
	gf := NewGeneralFunc(httpMsg, "test-xICAPMetadata")

	// Test for request mode
	file, reqContentType, err := gf.CopyingFileToTheBuffer(utils.ICAPModeReq)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if file == nil || reqContentType == nil {
		t.Errorf("Expected file and reqContentType to be non-nil")
	}

	// Test for response mode
	file, _, err = gf.CopyingFileToTheBuffer(utils.ICAPModeResp)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if file == nil {
		t.Errorf("Expected file to be non-nil")
	}
}

// Test for CheckTheExtension method
func TestCheckTheExtension(t *testing.T) {
	initLogger()
	httpMsg := createHttpMsg()
	gf := NewGeneralFunc(httpMsg, "test-xICAPMetadata")

	fileExtension := "txt"
	extArrs := []services_utilities.Extension{
		{Name: utils.ProcessExts},
		{Name: utils.RejectExts},
		{Name: utils.BypassExts},
	}
	processExts := []string{"txt"}
	rejectExts := []string{"exe"}
	bypassExts := []string{"jpg"}
	BlockPagePath := "block_page_path"
	fileSize := "1024"
	reqContentType := &ContentTypes.RegularFile{
		Buf:     bytes.NewBuffer([]byte("file content")),
		Encoded: false,
	}

	// Test for process extension
	shouldContinue, statusCode, _ := gf.CheckTheExtension(fileExtension, extArrs, processExts, rejectExts, bypassExts, false, false, "test-service", utils.ICAPModeReq, "id", "uri", reqContentType, bytes.NewBuffer([]byte("file content")), BlockPagePath, fileSize)
	if !shouldContinue || statusCode != 0 {
		t.Errorf("Expected process extension to continue, got shouldContinue: %v, statusCode: %v", shouldContinue, statusCode)
	}

	// Test for reject extension
	fileExtension = "exe"
	shouldContinue, statusCode, _ = gf.CheckTheExtension(fileExtension, extArrs, processExts, rejectExts, bypassExts, true, false, "test-service", utils.ICAPModeReq, "id", "uri", reqContentType, bytes.NewBuffer([]byte("file content")), BlockPagePath, fileSize)
	if shouldContinue || statusCode != utils.BadRequestStatusCodeStr {
		t.Errorf("Expected reject extension to stop, got shouldContinue: %v, statusCode: %v", shouldContinue, statusCode)
	}

	// Test for bypass extension
	fileExtension = "jpg"
	shouldContinue, statusCode, _ = gf.CheckTheExtension(fileExtension, extArrs, processExts, rejectExts, bypassExts, false, false, "test-service", utils.ICAPModeReq, "id", "uri", reqContentType, bytes.NewBuffer([]byte("file content")), BlockPagePath, fileSize)
	if shouldContinue || statusCode != utils.NoModificationStatusCodeStr {
		t.Errorf("Expected bypass extension to continue, got shouldContinue: %v, statusCode: %v", shouldContinue, statusCode)
	}
}

// Test for GetFileName method
func TestGetFileName(t *testing.T) {
	initLogger()
	httpMsg := createHttpMsg()
	gf := NewGeneralFunc(httpMsg, "test-xICAPMetadata")

	// Test when filename is in Content-Disposition header
	httpMsg.Response.Header.Set("Content-Disposition", "attachment; filename=testfile.txt")
	filename := gf.GetFileName("test-service", "test-xICAPMetadata")
	if filename != "testfile.txt" {
		t.Errorf("Expected filename 'testfile.txt', got %v", filename)
	}

	// Test when filename is in request URI
	httpMsg.Response.Header.Del("Content-Disposition")
	httpMsg.Request.URL.Path = "/path/to/testfile2.txt"
	t.Logf("URL path in request: %s", httpMsg.Request.URL.Path)
	filename = gf.GetFileName("test-service", "test-xICAPMetadata")
	if filename != "testfile2.txt" {
		t.Errorf("Expected filename 'testfile2.txt', got %v", filename)
	}

	// Test when no filename is found
	httpMsg.Request.URL.Path = ""
	t.Logf("URL path in request: %s", httpMsg.Request.URL.Path)
	filename = gf.GetFileName("test-service", "test-xICAPMetadata")
	if filename != "unnamed_file" {
		t.Errorf("Expected filename 'unnamed_file', got %v", filename)
	}
}
