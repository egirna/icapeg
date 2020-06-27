package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// RemoteICAPConfig represents the remote icap configuration settings
type RemoteICAPConfig struct {
	BaseURL         string
	ReqmodEndpoint  string
	RespmodEndpoint string
	OptionsEndpoint string
	Timeout         time.Duration
}

var riCfg RemoteICAPConfig

// LoadRemoteICAP populated the remote icap configuration instance with values from the config
func LoadRemoteICAP(name string) {
	riCfg = RemoteICAPConfig{
		BaseURL:         viper.GetString(fmt.Sprintf("%s.base_url", name)),
		ReqmodEndpoint:  viper.GetString(fmt.Sprintf("%s.reqmod_endpoint", name)),
		RespmodEndpoint: viper.GetString(fmt.Sprintf("%s.respmod_endpoint", name)),
		OptionsEndpoint: viper.GetString(fmt.Sprintf("%s.options_endpoint", name)),
		Timeout:         viper.GetDuration(fmt.Sprintf("%s.timeout", name)) * time.Second,
	}
}

// RemoteICAP returns the remote icap configuration instance
func RemoteICAP() RemoteICAPConfig {
	return riCfg
}
