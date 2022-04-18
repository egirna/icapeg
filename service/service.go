package service

import (
	"icapeg/logger"
	"icapeg/service/services/clamav"
	"icapeg/service/services/echo"
	"icapeg/service/services/glasswall"
	"icapeg/utils"
	"time"
)

//Vendors names
const (
	VendorGlasswall = "glasswall"
	VendorEcho      = "echo"
	VendorClamav    = "clamav"
)

type (
	// Service holds the info to distinguish a service
	Service interface {
		Processing() (int, interface{}, map[string]string)
	}
)

// GetService returns a service based on the service name
// change name to vendor and add parameter service name
func GetService(vendor, serviceName, methodName string, httpMsg *utils.HttpMsg, elapsed time.Duration, logger *logger.ZLogger) Service {
	switch vendor {
	case VendorGlasswall:
		return glasswall.NewGlasswallService(serviceName, methodName, httpMsg, elapsed, logger)
	case VendorEcho:
		return echo.NewEchoService(serviceName, methodName, httpMsg, elapsed, logger)
	case VendorClamav:
		return clamav.NewClamavService(serviceName, methodName, httpMsg, elapsed, logger)

	}
	return nil
}
