package service

import (
	"icapeg/logger"
	"icapeg/service/glasswall"
	"net/http"
	"time"
)

//Vendors names
const (
	VendorGlasswall = "glasswall"
)

type (
	// Service holds the info to distinguish a service
	Service interface {
		Processing() (int, []byte, *http.Response, map[string]string)
	}
)

// GetService returns a service based on the service name
// change name to vendor and add parameter service name
func GetService(vendor, serviceName, methodName string, req *http.Request, resp *http.Response, elapsed time.Duration, logger *logger.ZLogger) Service {
	switch vendor {
	case VendorGlasswall:
		return glasswall.NewGlasswallService(serviceName, methodName, req, resp, elapsed, logger)
	}
	return nil
}
