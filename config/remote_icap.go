package config

import (
	"time"

	"github.com/spf13/viper"
)

// RemoteICAPCfg represents the remote icap configuration settings
type RemoteICAPCfg struct {
	Enabled         bool
	BaseURL         string
	ReqmodEndpoint  string
	RespmodEndpoint string
	OptionsEndpoint string
	Timeout         time.Duration
}

var riCfg RemoteICAPCfg

// LoadRemoteICAP populated the remote icap configuration instance with values from the config
func LoadRemoteICAP() {
	riCfg = RemoteICAPCfg{
		Enabled:         viper.GetBool("remote_icap.enabled"),
		BaseURL:         viper.GetString("remote_icap.base_url"),
		ReqmodEndpoint:  viper.GetString("remote_icap.reqmod_endpoint"),
		RespmodEndpoint: viper.GetString("remote_icap.respmod_endpoint"),
		OptionsEndpoint: viper.GetString("remote_icap.options_endpoint"),
		Timeout:         viper.GetDuration("remote_icap.timeout") * time.Second,
	}
}

// RemoteICAP returns the remote icap configuration instance
func RemoteICAP() RemoteICAPCfg {
	return riCfg
}
