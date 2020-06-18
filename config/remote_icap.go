package config

import (
	"time"

	"github.com/spf13/viper"
)

// ICAPCfg represents the remote icap configuration settings
type ICAPCfg struct {
	Enabled         bool
	BaseURL         string
	ReqmodEndpoint  string
	RespmodEndpoint string
	OptionsEndpoint string
	Timeout         time.Duration
}

var iCfg ICAPCfg

type iotaType int

const (
	remote iotaType = iota
	shadow
)

func (i iotaType) toString() string {
	switch i {
	case remote:
		return "remote_icap"
	case shadow:
		return "shadow_icap"
	default:
		// Default to remote
		return "remote_icap"
	}
}

// loadICAP populated the remote icap configuration instance with values from the config
func loadICAP(t iotaType) *ICAPCfg {
	configMap := viper.GetStringMapString(t.toString())

	return &ICAPCfg{
		Enabled:         viper.GetBool(configMap["enabled"]),
		BaseURL:         configMap["base_url"],
		ReqmodEndpoint:  configMap["reqmod_endpoint"],
		RespmodEndpoint: configMap["respmod_endpoint"],
		OptionsEndpoint: configMap["options_endpoint"],
		Timeout:         viper.GetDuration(configMap["timeout"]) * time.Second,
	}
}

// RemoteICAP returns the remote icap configuration instance
func RemoteICAP() *ICAPCfg {
	return loadICAP(remote)
}

// ShadowICAP returns the shadow icap configuration instance
func ShadowICAP() *ICAPCfg {
	return loadICAP(shadow)
}
