package service

import (
	http_message "icapeg/http-message"
	"icapeg/logging"
	"icapeg/service/services/clamav"
	"icapeg/service/services/clhashlookup"
	"icapeg/service/services/echo"
	"icapeg/service/services/hashlocal"
	"net/textproto"
)

// Vendors names
const (
	VendorEcho       = "echo"
	VendorClamav     = "clamav"
	VendorHashlookup = "clhashlookup"
	VendorHashlocal  = "hashlocal"
)

type (
	// Service holds the info to distinguish a service
	Service interface {
		Processing(bool, textproto.MIMEHeader) (int, interface{}, map[string]string,
			map[string]interface{}, map[string]interface{}, map[string]interface{})
		ISTagValue() string
	}
)

// GetService returns a service based on the service name
// change name to vendor and add parameter service name
func GetService(vendor, serviceName, methodName string, httpMsg *http_message.HttpMsg, xICAPMetadata string) Service {
	logging.Logger.Info("getting instance from " + serviceName + " struct")
	switch vendor {
	case VendorEcho:
		return echo.NewEchoService(serviceName, methodName, httpMsg, xICAPMetadata)
	case VendorHashlookup:
		return clhashlookup.NewHashlookupService(serviceName, methodName, httpMsg, xICAPMetadata)
	case VendorClamav:
		return clamav.NewClamavService(serviceName, methodName, httpMsg, xICAPMetadata)
	case VendorHashlocal:
		return hashlocal.NewHashlocalService(serviceName, methodName, httpMsg, xICAPMetadata)
	}
	return nil
}

// InitServiceConfig is used to load the services configuration
func InitServiceConfig(vendor, serviceName string) {
	logging.Logger.Info("loading all the services configuration")
	switch vendor {
	case VendorEcho:
		echo.InitEchoConfig(serviceName)
	case VendorHashlookup:
		clhashlookup.InitHashlookupConfig(serviceName)
	case VendorClamav:
		clamav.InitClamavConfig(serviceName)
	case VendorHashlocal:
		hashlocal.InitHashlocalConfig(serviceName)
	}
}
