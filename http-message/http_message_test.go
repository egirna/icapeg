package http_message

import (
	"net/http/httptest"
	"testing"

	"icapeg/logging"

	"go.uber.org/zap"
)

func init() {
	logger, _ := zap.NewDevelopment()
	logging.Logger = logger
}

func TestNewHttpMsg(t *testing.T) {
	// Create a new HTTP request
	req := httptest.NewRequest("GET", "http://example.com", nil)

	// Create a new HTTP response recorder (which implements http.ResponseWriter)
	rec := httptest.NewRecorder()

	// Create a dummy HTTP response using the recorded response
	resp := rec.Result()

	// Create an instance of HttpMsg using the NewHttpMsg method
	httpMsg := (&HttpMsg{}).NewHttpMsg(req, resp)

	// Check if the HttpMsg struct is created correctly
	if httpMsg.Request != req {
		t.Errorf("Expected Request to be %v, got %v", req, httpMsg.Request)
	}

	if httpMsg.Response != resp {
		t.Errorf("Expected Response to be %v, got %v", resp, httpMsg.Response)
	}
}
