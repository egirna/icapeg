package config

import (
	"log"
	"strings"

	"github.com/spf13/viper"
)

// AppConfig represents the app configuration
type AppConfig struct {
	Port                    int
	MaxFileSize             int
	LogLevel                string
	RespScannerVendor       string
	ReqScannerVendor        string
	RespScannerVendorShadow string
	ReqScannerVendorShadow  string
	BypassExtensions        []string
	ProcessExtensions       []string
	PreviewBytes            string
	PreviewEnabled          bool
	PropagateError          bool
}

var appCfg AppConfig

// Init initializes the configuration
func Init() {
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(err.Error())
	}

	appCfg = AppConfig{
		Port:                    viper.GetInt("app.port"),
		MaxFileSize:             viper.GetInt("app.max_filesize"),
		LogLevel:                viper.GetString("app.log_level"),
		RespScannerVendor:       strings.ToLower(viper.GetString("app.resp_scanner_vendor")),
		ReqScannerVendor:        strings.ToLower(viper.GetString("app.req_scanner_vendor")),
		RespScannerVendorShadow: strings.ToLower(viper.GetString("app.resp_scanner_vendor_shadow")),
		ReqScannerVendorShadow:  strings.ToLower(viper.GetString("app.req_scanner_vendor_shadow")),
		BypassExtensions:        viper.GetStringSlice("app.bypass_extensions"),
		ProcessExtensions:       viper.GetStringSlice("app.process_extensions"),
		PreviewBytes:            viper.GetString("app.preview_bytes"),
		PreviewEnabled:          viper.GetBool("app.preview_enabled"),
		PropagateError:          viper.GetBool("app.propagate_error"),
	}

}

// InitTestConfig initializes the app with the test config file (for integration test)
func InitTestConfig() {
	viper.SetConfigName("config.test")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(err.Error())
	}

	appCfg = AppConfig{
		Port:                    viper.GetInt("app.port"),
		MaxFileSize:             viper.GetInt("app.max_filesize"),
		LogLevel:                viper.GetString("app.log_level"),
		RespScannerVendor:       strings.ToLower(viper.GetString("app.resp_scanner_vendor")),
		ReqScannerVendor:        strings.ToLower(viper.GetString("app.req_scanner_vendor")),
		RespScannerVendorShadow: strings.ToLower(viper.GetString("app.resp_scanner_vendor_shadow")),
		ReqScannerVendorShadow:  strings.ToLower(viper.GetString("app.req_scanner_vendor_shadow")),
		BypassExtensions:        viper.GetStringSlice("app.bypass_extensions"),
		ProcessExtensions:       viper.GetStringSlice("app.process_extensions"),
		PreviewBytes:            viper.GetString("app.preview_bytes"),
		PropagateError:          viper.GetBool("app.propagate_error"),
	}
}

// App returns the the app configuration instance
func App() *AppConfig {
	return &appCfg
}
