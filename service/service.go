package service

import (
	"bytes"
	"icapeg/icap"
	"net/http"
	"time"

	_ "icapeg/icap-client"
	"icapeg/logger"
	"icapeg/service/glasswall"
)

//Services names
const (
	SVCGlasswall = "glasswall"
)

type (
	// Service holds the info to distinguish a service
	Service interface {
		SendReqToAPI(f *bytes.Buffer, filename string) *http.Response
		RespMode(req *http.Request, resp *http.Response)
	}
)

// GetService returns a service based on the service name
// change name to vendor and add parameter service name
func GetService(vendor string, serviceName string, w icap.ResponseWriter, req *http.Request, resp *http.Response, elapsed time.Duration, Is204Allowed bool, methodName string, logger *logger.ZLogger) Service {
	switch vendor {
	case SVCGlasswall:
		return glasswall.NewGlasswallService(w, req, resp, elapsed, Is204Allowed, serviceName, methodName, logger)
	}
	return nil
}
