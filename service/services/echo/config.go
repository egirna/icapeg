package echo

import (
	"icapeg/readValues"
	general_functions "icapeg/service/services-utilities/general-functions"
	"icapeg/utils"
	"sync"
	"time"
)

var doOnce sync.Once
var echoConfig *Echo

// Echo represents the information regarding the Echo service
type Echo struct {
	serviceCaption         string
	serviceTag             string
	reqMode                bool
	respMode               bool
	shadowService          bool
	httpMsg                *utils.HttpMsg
	elapsed                time.Duration
	serviceName            string
	methodName             string
	maxFileSize            int
	previewEnabled         bool
	previewBytes           string
	bypassExts             []string
	processExts            []string
	BaseURL                string
	Timeout                time.Duration
	APIKey                 string
	ScanEndpoint           string
	FailThreshold          int
	returnOrigIfMaxSizeExc bool
	returnOrigIf400        bool
	generalFunc            *general_functions.GeneralFunc
}

func initEchoConfig(serviceName string) {
	doOnce.Do(func() {
		echoConfig = &Echo{
			serviceCaption:         readValues.ReadValuesString(serviceName + ".service_caption"),
			serviceTag:             readValues.ReadValuesString(serviceName + ".service_tag"),
			reqMode:                readValues.ReadValuesBool(serviceName + ".req_mode"),
			respMode:               readValues.ReadValuesBool(serviceName + ".resp_mode"),
			shadowService:          readValues.ReadValuesBool(serviceName + ".shadow_service"),
			previewEnabled:         readValues.ReadValuesBool(serviceName + ".preview_enabled"),
			previewBytes:           readValues.ReadValuesString(serviceName + ".preview_bytes"),
			maxFileSize:            readValues.ReadValuesInt(serviceName + ".max_filesize"),
			bypassExts:             readValues.ReadValuesSlice(serviceName + ".bypass_extensions"),
			processExts:            readValues.ReadValuesSlice(serviceName + ".process_extensions"),
			BaseURL:                readValues.ReadValuesString(serviceName + ".base_url"),
			Timeout:                readValues.ReadValuesDuration(serviceName+".timeout") * time.Second,
			APIKey:                 readValues.ReadValuesString(serviceName + ".api_key"),
			ScanEndpoint:           readValues.ReadValuesString(serviceName + ".scan_endpoint"),
			FailThreshold:          readValues.ReadValuesInt(serviceName + ".fail_threshold"),
			returnOrigIfMaxSizeExc: readValues.ReadValuesBool(serviceName + ".return_original_if_max_file_size_exceeded"),
		}
	})

}

// NewEchoService returns a new populated instance of the Echo service
func NewEchoService(serviceName, methodName string, httpMsg *utils.HttpMsg) *Echo {
	initEchoConfig(serviceName)
	return &Echo{
		httpMsg:                httpMsg,
		serviceName:            serviceName,
		methodName:             methodName,
		generalFunc:            general_functions.NewGeneralFunc(httpMsg),
		previewEnabled:         echoConfig.previewEnabled,
		previewBytes:           echoConfig.previewBytes,
		maxFileSize:            echoConfig.maxFileSize,
		bypassExts:             echoConfig.bypassExts,
		processExts:            echoConfig.processExts,
		BaseURL:                echoConfig.BaseURL,
		Timeout:                echoConfig.Timeout * time.Second,
		APIKey:                 echoConfig.APIKey,
		ScanEndpoint:           echoConfig.ScanEndpoint,
		FailThreshold:          echoConfig.FailThreshold,
		returnOrigIfMaxSizeExc: echoConfig.returnOrigIfMaxSizeExc,
	}
}
