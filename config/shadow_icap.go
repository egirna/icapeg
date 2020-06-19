package config

import (
	"time"

	"github.com/spf13/viper"
)

// ShadowICAPCfg represents the shadow icap configuration settings
type ShadowICAPCfg struct {
	Enabled         bool
	BaseURL         string
	ReqmodEndpoint  string
	RespmodEndpoint string
	OptionsEndpoint string
	Timeout         time.Duration
}

var siCfg ShadowICAPCfg

// LoadShadowICAP populated the shadow icap configuration instance with values from the config
func LoadShadowICAP() {
	siCfg = ShadowICAPCfg{
		Enabled:         viper.GetBool("shadow_icap.enabled"),
		BaseURL:         viper.GetString("shadow_icap.base_url"),
		ReqmodEndpoint:  viper.GetString("shadow_icap.reqmod_endpoint"),
		RespmodEndpoint: viper.GetString("shadow_icap.respmod_endpoint"),
		OptionsEndpoint: viper.GetString("shadow_icap.options_endpoint"),
		Timeout:         viper.GetDuration("shadow_icap.timeout") * time.Second,
	}
}

// ShadowICAP returns the shadow icap configuration instance
func ShadowICAP() ShadowICAPCfg {
	return siCfg
}
