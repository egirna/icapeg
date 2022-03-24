package service

import (
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
		Processing() (int, []byte, map[string]string)
		//SendReqToAPI(f *bytes.Buffer, filename string) *http.Response
		//RespMode(req *http.Request, resp *http.Response)
	}
)

// GetService returns a service based on the service name
// change name to vendor and add parameter service name
func GetService(vendor, serviceName, methodName string, req *http.Request, resp *http.Response, elapsed time.Duration, logger *logger.ZLogger) Service {
	switch vendor {
	case SVCGlasswall:
		return glasswall.NewGlasswallService(serviceName, methodName, req, resp, elapsed, logger)
	}
	return nil
}
