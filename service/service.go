package service

import (
	"icapeg/service/services/clamav"
	"icapeg/service/services/cloudmersive"
	"icapeg/service/services/echo"
	"icapeg/service/services/glasswall"
	"icapeg/utils"
)

//Vendors names
const (
	VendorGlasswall    = "glasswall"
	VendorEcho         = "echo"
	VendorClamav       = "clamav"
	VendorCloudMersive = "cloudmersive"
)

type (
	// Service holds the info to distinguish a service
	Service interface {
		Processing(bool) (int, interface{}, map[string]string)
	}
)

// GetService returns a service based on the service name
// change name to vendor and add parameter service name
func GetService(vendor, serviceName, methodName string, httpMsg *utils.HttpMsg) Service {
	switch vendor {
	case VendorGlasswall:
		return glasswall.NewGlasswallService(serviceName, methodName, httpMsg)
	case VendorEcho:
		return echo.NewEchoService(serviceName, methodName, httpMsg)
	case VendorClamav:
		return clamav.NewClamavService(serviceName, methodName, httpMsg)
	case VendorCloudMersive:
		return cloudmersive.NewCloudMersiveService(serviceName, methodName, httpMsg)

	}

	return nil
}

func InitServiceConfig(vendor, serviceName string) {
	switch vendor {
	case VendorGlasswall:
		glasswall.InitGlasswallConfig(serviceName)
	case VendorEcho:
		echo.InitEchoConfig(serviceName)
	case VendorCloudMersive:
		cloudmersive.InitCloudMersiveConfig(serviceName)
	}
}
