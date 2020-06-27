package config

import "github.com/spf13/viper"

// ShadowConfig represents the shadow configuration
type ShadowConfig struct {
	RespScannerVendor string
	ReqScannerVendor  string
	RemoteICAP        string
}

var shadowCfg ShadowConfig

// LoadShadow populates the shadow config instance with the config values
func LoadShadow() {
	shadowCfg = ShadowConfig{
		RespScannerVendor: viper.GetString("shadow.resp_scanner_vendor"),
		ReqScannerVendor:  viper.GetString("shadow.req_scanner.vendor"),
		RemoteICAP:        viper.GetString("shadow.remote_icap"),
	}
}

// Shadow returns the shadow config instance
func Shadow() ShadowConfig {
	return shadowCfg
}
