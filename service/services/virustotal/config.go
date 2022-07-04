package virustotal

import (
	"icapeg/readValues"
	general_functions "icapeg/service/services-utilities/general-functions"
	"icapeg/utils"
	"sync"
	"time"
)

var doOnce sync.Once
var virustoalConfig *Virustotal

// Virustotal represents the information regarding the Virustotal service
type Virustotal struct {
	httpMsg                *utils.HttpMsg
	serviceName            string
	methodName             string
	maxFileSize            int
	bypassExts             []string
	processExts            []string
	ScanUrl                string
	ReportUrl              string
	Timeout                time.Duration
	APIKey                 string
	FailThreshold          int
	policy                 string
	returnOrigIfMaxSizeExc bool
	generalFunc            *general_functions.GeneralFunc
}

func InitVirustotalConfig(serviceName string) {
	doOnce.Do(func() {
		virustoalConfig = &Virustotal{
			maxFileSize:            readValues.ReadValuesInt(serviceName + ".max_filesize"),
			bypassExts:             readValues.ReadValuesSlice(serviceName + ".bypass_extensions"),
			processExts:            readValues.ReadValuesSlice(serviceName + ".process_extensions"),
			ScanUrl:                readValues.ReadValuesString(serviceName + ".scan_url"),
			ReportUrl:              readValues.ReadValuesString(serviceName + ".report_url"),
			Timeout:                readValues.ReadValuesDuration(serviceName + ".timeout"),
			APIKey:                 readValues.ReadValuesString(serviceName + ".api_key"),
			FailThreshold:          readValues.ReadValuesInt(serviceName + ".fail_threshold"),
			policy:                 readValues.ReadValuesString(serviceName + ".policy"),
			returnOrigIfMaxSizeExc: readValues.ReadValuesBool(serviceName + ".return_original_if_max_file_size_exceeded"),
		}
	})
}

// NewVirustotalService returns a new populated instance of the Virustotal service
func NewVirustotalService(serviceName, methodName string, httpMsg *utils.HttpMsg) *Virustotal {
	return &Virustotal{
		httpMsg:                httpMsg,
		serviceName:            serviceName,
		methodName:             methodName,
		maxFileSize:            virustoalConfig.maxFileSize,
		bypassExts:             virustoalConfig.bypassExts,
		processExts:            virustoalConfig.processExts,
		ScanUrl:                virustoalConfig.ScanUrl,
		ReportUrl:              virustoalConfig.ReportUrl,
		Timeout:                virustoalConfig.Timeout * time.Second,
		APIKey:                 virustoalConfig.APIKey,
		FailThreshold:          virustoalConfig.FailThreshold,
		policy:                 virustoalConfig.policy,
		returnOrigIfMaxSizeExc: virustoalConfig.returnOrigIfMaxSizeExc,
		generalFunc:            general_functions.NewGeneralFunc(httpMsg),
	}
}
