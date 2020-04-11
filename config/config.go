package config

import (
	"log"

	"github.com/spf13/viper"
)

// AppConfig represents the app configuration
type AppConfig struct {
	Port        int
	HTTPPort    int
	MaxFileSize int
}

var appCfg AppConfig

func init() {
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(err.Error())
	}

	appCfg = AppConfig{
		Port:        viper.GetInt("app.port"),
		HTTPPort:    viper.GetInt("app.http_port"),
		MaxFileSize: viper.GetInt("app.max_filesize"),
	}
}

// App returns the the app configuration instance
func App() AppConfig {
	return appCfg
}
