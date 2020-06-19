package config

import (
	"log"

	"github.com/spf13/viper"
)

// AppConfig represents the app configuration
type AppConfig struct {
	Port              int
	MaxFileSize       int
	Debug             bool
	RespScannerVendor string
	ReqScannerVendor  string
	BypassExtensions  []string
	ProcessExtensions []string
	PreviewBytes      string
	PropagateError    bool
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
		Port:              viper.GetInt("app.port"),
		MaxFileSize:       viper.GetInt("app.max_filesize"),
		Debug:             viper.GetBool("app.debug"),
		RespScannerVendor: viper.GetString("app.resp_scanner_vendor"),
		ReqScannerVendor:  viper.GetString("app.req_scanner_vendor"),
		BypassExtensions:  viper.GetStringSlice("app.bypass_extensions"),
		ProcessExtensions: viper.GetStringSlice("app.process_extensions"),
		PreviewBytes:      viper.GetString("app.preview_bytes"),
		PropagateError:    viper.GetBool("app.propagate_error"),
	}

	LoadRemoteICAP()
	LoadShadowICAP()

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
		Port:              viper.GetInt("app.port"),
		MaxFileSize:       viper.GetInt("app.max_filesize"),
		Debug:             viper.GetBool("app.debug"),
		RespScannerVendor: viper.GetString("app.resp_scanner_vendor"),
		ReqScannerVendor:  viper.GetString("app.req_scanner_vendor"),
		BypassExtensions:  viper.GetStringSlice("app.bypass_extensions"),
		ProcessExtensions: viper.GetStringSlice("app.process_extensions"),
		PreviewBytes:      viper.GetString("app.preview_bytes"),
		PropagateError:    viper.GetBool("app.propagate_error"),
	}
}

// App returns the the app configuration instance
func App() AppConfig {
	return appCfg
}
