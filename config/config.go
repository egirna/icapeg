package config

import (
	"fmt"
	"icapeg/readValues"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// AppConfig represents the app configuration
type AppConfig struct {
	Port                 int
	LogLevel             string
	LoggingServerURL     string
	LoggingFlushDuration float64
	WriteLogsToConsole   bool
	//RespScannerVendor       string
	//ReqScannerVendor        string
	RespScannerVendorShadow string
	ReqScannerVendorShadow  string
	BypassExtensions        []string
	ProcessExtensions       []string
	PreviewBytes            string
	PreviewEnabled          bool
	PropagateError          bool
	VerifyServerCert        bool
	services                []string
}

var appCfg AppConfig

// Init initializes the configuration
func Init() {
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")
	if readValues.IsSecExists("app") {
		fmt.Println("app section doesn't exist in config file")
	}
	appCfg = AppConfig{
		Port:                    readValues.ReadValuesInt("app.port"),
		LogLevel:                readValues.ReadValuesString("app.log_level"),
		LoggingServerURL:        readValues.ReadValuesString("app.log_service_url"),
		WriteLogsToConsole:      readValues.ReadValuesBool("app.write_logs_to_console"),
		RespScannerVendorShadow: strings.ToLower(readValues.ReadValuesString("app.resp_scanner_vendor_shadow")),
		ReqScannerVendorShadow:  strings.ToLower(readValues.ReadValuesString("app.req_scanner_vendor_shadow")),
		PreviewBytes:            readValues.ReadValuesString("app.preview_bytes"),
		PreviewEnabled:          readValues.ReadValuesBool("app.preview_enabled"),
		PropagateError:          readValues.ReadValuesBool("app.propagate_error"),
		VerifyServerCert:        readValues.ReadValuesBool("app.verify_server_cert"),
		services:                readValues.ReadValuesSlice("app.services"),
	}
	for i := 0; i < len(appCfg.services); i++ {
		if readValues.ReadValuesInt(appCfg.services[i]+".max_filesize") < 0 {
			fmt.Println("max_filesize value in config.toml file is not valid")
			os.Exit(1)
		}
	}
	//this loop to make sure that all services in the array of services has sections in the config file and from request mode and response mode
	//there is one at least from them are enabled in every service
	for i := 0; i < len(appCfg.services); i++ {
		if !readValues.IsSecExists(appCfg.services[i]) {
			fmt.Println(appCfg.services[i] + " section doesn't exist")
			os.Exit(1)
		}
		if !readValues.ReadValuesBool(appCfg.services[i]+".req_mode") && !readValues.ReadValuesBool(appCfg.services[i]+".resp_mode") {
			fmt.Println("Request mode and response mode are disabled together in " + appCfg.services[i] + " service")
			os.Exit(1)
		}
	}

}

// InitTestConfig initializes the app with the test config file (for integration test)
func InitTestConfig() {
	appCfg = AppConfig{
		Port:                 readValues.ReadValuesInt("app.port"),
		LogLevel:             readValues.ReadValuesString("app.log_level"),
		LoggingServerURL:     readValues.ReadValuesString("app.log_service_url"),
		LoggingFlushDuration: float64(readValues.ReadValuesInt("app.log_flush_duration")),
		//RespScannerVendor:       strings.ToLower(readValues.ReadValuesString("app.resp_scanner_vendor")),
		//ReqScannerVendor:        strings.ToLower(readValues.ReadValuesString("app.req_scanner_vendor")),
		RespScannerVendorShadow: strings.ToLower(readValues.ReadValuesString("app.resp_scanner_vendor_shadow")),
		ReqScannerVendorShadow:  strings.ToLower(readValues.ReadValuesString("app.req_scanner_vendor_shadow")),
		PreviewBytes:            readValues.ReadValuesString("app.preview_bytes"),
		PropagateError:          readValues.ReadValuesBool("app.propagate_error"),
	}
}

// App returns the the app configuration instance
func App() *AppConfig {
	return &appCfg
}
