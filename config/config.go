package config

import (
	"fmt"
	"os"

	"github.com/egirna/icapeg/logging"
	"github.com/egirna/icapeg/readValues"

	"github.com/spf13/viper"
)

type serviceIcapInfo struct {
	Vendor         string
	ServiceCaption string
	ServiceTag     string
	ReqMode        bool
	RespMode       bool
	ShadowService  bool
	PreviewEnabled bool
	PreviewBytes   string
}

// AppConfig represents the app configuration
type AppConfig struct {
	Port               int
	LogLevel           string
	WriteLogsToConsole bool
	BypassExtensions   []string
	ProcessExtensions  []string
	PreviewBytes       string
	PreviewEnabled     bool
	DebuggingHeaders   bool
	Services           []string
	ServicesInstances  map[string]*serviceIcapInfo
}

var AppCfg AppConfig

// Init initializes the configuration
func Init() {
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath("/etc/icapeg/")
	viper.AddConfigPath("/usr/local/etc/icapeg/")
	viper.AddConfigPath("$HOME/.config/icapeg")
	viper.AddConfigPath(".")
	if readValues.IsSecExists("app") {
		fmt.Println("app section doesn't exist in config file")
	}
	AppCfg = AppConfig{
		Port:               readValues.ReadValuesInt("app.port"),
		LogLevel:           readValues.ReadValuesString("app.log_level"),
		WriteLogsToConsole: readValues.ReadValuesBool("app.write_logs_to_console"),
		DebuggingHeaders:   readValues.ReadValuesBool("app.debugging_headers"),
		Services:           readValues.ReadValuesSlice("app.services"),
	}
	logging.InitializeLogger(AppCfg.LogLevel, AppCfg.WriteLogsToConsole)
	logging.Logger.Info("Reading config.toml file")

	//this loop to make sure that all services in the array of services has sections in the config file and from request mode and response mode
	//there is one at least from them are enabled in every service
	AppCfg.ServicesInstances = make(map[string]*serviceIcapInfo)
	logging.Logger.Debug("checking that all services in the array of services has sections in the config file and from request mode and response mode")
	for i := 0; i < len(AppCfg.Services); i++ {
		serviceName := AppCfg.Services[i]
		if !readValues.IsSecExists(serviceName) {
			logging.Logger.Fatal(serviceName + " section doesn't exist")
			fmt.Println(serviceName + " section doesn't exist")
			os.Exit(1)
		}
		if !readValues.ReadValuesBool(serviceName+".req_mode") && !readValues.ReadValuesBool(serviceName+".resp_mode") {
			logging.Logger.Fatal("Request mode and response mode are disabled together in " + serviceName + " service")
			fmt.Println("Request mode and response mode are disabled together in " + serviceName + " service")
			os.Exit(1)
		}
		if readValues.ReadValuesInt(serviceName+".max_filesize") < 0 {
			logging.Logger.Fatal("max_filesize value in config.toml file is not valid")
			fmt.Println("max_filesize value in config.toml file is not valid")
			os.Exit(1)
		}
		//checking if extensions arrays are valid in every service
		//arrays are valid if there is only one array has asterisk and no two arrays has same file type
		logging.Logger.Debug("checking if extensions arrays are valid in every service")
		ext := make(map[string]bool)
		asterisks := 0
		//bypass
		bypass := readValues.ReadValuesSlice(serviceName + ".bypass_extensions")
		for i := 0; i < len(bypass); i++ {
			if bypass[i] == "*" && len(bypass) != 1 {
				logging.Logger.Fatal("bypass_extensions array has one asterisk \"*\"" +
					" and other extensions but asterisk should be the only element in the array otherwise add extensions as you want")
				fmt.Println("bypass_extensions array has one asterisk \"*\"" +
					" and other extensions but asterisk should be the only element in the array otherwise add extensions as you want")
				os.Exit(1)
			}
			if bypass[i] == "*" {
				asterisks++
			}
			if ext[bypass[i]] == false {
				ext[bypass[i]] = true
			} else {
				logging.Logger.Fatal("This extension \"" + bypass[i] + "\" was " +
					"stored in multiple arrays (bypass_extensions or reject_extensions)")
				fmt.Println("This extension \"" + bypass[i] + "\" was " +
					"stored in multiple arrays (bypass_extensions or reject_extensions)")
				os.Exit(1)
			}
		}
		//process
		process := readValues.ReadValuesSlice(serviceName + ".process_extensions")
		for i := 0; i < len(process); i++ {
			if process[i] == "*" && len(process) != 1 {
				logging.Logger.Fatal("process_extensions array has one asterisk \"*\" and other extensions " +
					"but asterisk should be the only element in the array otherwise add extensions as you want")
				fmt.Println("process_extensions array has one asterisk \"*\" and other extensions " +
					"but asterisk should be the only element in the array otherwise add extensions as you want")
				os.Exit(1)
			}
			if process[i] == "*" {
				asterisks++
			}
			if ext[process[i]] == false {
				ext[process[i]] = true
			} else {
				logging.Logger.Fatal("This extension \"" + process[i] + "\" is stored in multiple arrays")
				fmt.Println("This extension \"" + process[i] + "\" is stored in multiple arrays")
				os.Exit(1)
			}
		}
		//reject
		reject := readValues.ReadValuesSlice(serviceName + ".reject_extensions")
		for i := 0; i < len(reject); i++ {
			if reject[i] == "*" && len(reject) != 1 {
				logging.Logger.Fatal("reject_extensions array has one asterisk \"*\" and other extensions but asterisk " +
					"should be the only element in the array otherwise add extensions as you want")
				fmt.Println("reject_extensions array has one asterisk \"*\" and other extensions but asterisk " +
					"should be the only element in the array otherwise add extensions as you want")
				os.Exit(1)
			}
			if reject[i] == "*" {
				asterisks++
			}
			if ext[reject[i]] == false {
				ext[reject[i]] = true
			} else {
				logging.Logger.Fatal("This extension \"" + reject[i] + "\" is stored in multiple arrays")
				fmt.Println("This extension \"" + reject[i] + "\" is stored in multiple arrays")
				os.Exit(1)
			}
		}
		if asterisks != 1 {
			logging.Logger.Fatal("There is no \"*\" stored in any extension arrays")
			fmt.Println("There is no \"*\" stored in any extension arrays")
			os.Exit(1)
		}

		AppCfg.ServicesInstances[serviceName] = &serviceIcapInfo{
			Vendor:         readValues.ReadValuesString(serviceName + ".vendor"),
			ServiceTag:     readValues.ReadValuesString(serviceName + ".service_tag"),
			ServiceCaption: readValues.ReadValuesString(serviceName + ".service_caption"),
			ReqMode:        readValues.ReadValuesBool(serviceName + ".req_mode"),
			RespMode:       readValues.ReadValuesBool(serviceName + ".resp_mode"),
			ShadowService:  readValues.ReadValuesBool(serviceName + ".shadow_service"),
			PreviewBytes:   readValues.ReadValuesString(serviceName + ".preview_bytes"),
			PreviewEnabled: readValues.ReadValuesBool(serviceName + ".preview_enabled"),
		}
	}
}

// App returns the app configuration instance
func App() *AppConfig {
	return &AppCfg
}
