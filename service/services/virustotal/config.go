package virustotal

import (
	"icapeg/http-message"
	"icapeg/logging"
	"icapeg/readValues"
	services_utilities "icapeg/service/services-utilities"
	general_functions "icapeg/service/services-utilities/general-functions"
	"sync"
	"time"
)

var doOnce sync.Once
var virustoalConfig *Virustotal

const VirustotalIdentifier = "VIRUSTOTAL ID"

// Virustotal represents the information regarding the Virustotal service
type Virustotal struct {
	xICAPMetadata              string
	httpMsg                    *http_message.HttpMsg
	serviceName                string
	methodName                 string
	maxFileSize                int
	bypassExts                 []string
	processExts                []string
	rejectExts                 []string
	extArrs                    []services_utilities.Extension
	ScanUrl                    string
	ReportUrl                  string
	Timeout                    time.Duration
	APIKey                     string
	returnOrigIfMaxSizeExc     bool
	return400IfFileExtRejected bool
	generalFunc                *general_functions.GeneralFunc
}

func InitVirustotalConfig(serviceName string) {
	logging.Logger.Debug("loading " + serviceName + " service configurations")
	doOnce.Do(func() {
		virustoalConfig = &Virustotal{
			maxFileSize:                readValues.ReadValuesInt(serviceName + ".max_filesize"),
			bypassExts:                 readValues.ReadValuesSlice(serviceName + ".bypass_extensions"),
			processExts:                readValues.ReadValuesSlice(serviceName + ".process_extensions"),
			rejectExts:                 readValues.ReadValuesSlice(serviceName + ".reject_extensions"),
			ScanUrl:                    readValues.ReadValuesString(serviceName + ".scan_url"),
			ReportUrl:                  readValues.ReadValuesString(serviceName + ".report_url"),
			Timeout:                    readValues.ReadValuesDuration(serviceName + ".timeout"),
			APIKey:                     readValues.ReadValuesString(serviceName + ".api_key"),
			returnOrigIfMaxSizeExc:     readValues.ReadValuesBool(serviceName + ".return_original_if_max_file_size_exceeded"),
			return400IfFileExtRejected: readValues.ReadValuesBool(serviceName + ".return_400_if_file_ext_rejected"),
		}
		virustoalConfig.extArrs = services_utilities.InitExtsArr(virustoalConfig.processExts, virustoalConfig.rejectExts, virustoalConfig.bypassExts)
	})
}

// NewVirustotalService returns a new populated instance of the Virustotal service
func NewVirustotalService(serviceName, methodName string, httpMsg *http_message.HttpMsg, xICAPMetadata string) *Virustotal {
	return &Virustotal{
		xICAPMetadata:              xICAPMetadata,
		httpMsg:                    httpMsg,
		serviceName:                serviceName,
		methodName:                 methodName,
		maxFileSize:                virustoalConfig.maxFileSize,
		bypassExts:                 virustoalConfig.bypassExts,
		processExts:                virustoalConfig.processExts,
		rejectExts:                 virustoalConfig.rejectExts,
		extArrs:                    virustoalConfig.extArrs,
		ScanUrl:                    virustoalConfig.ScanUrl,
		ReportUrl:                  virustoalConfig.ReportUrl,
		Timeout:                    virustoalConfig.Timeout * time.Second,
		APIKey:                     virustoalConfig.APIKey,
		returnOrigIfMaxSizeExc:     virustoalConfig.returnOrigIfMaxSizeExc,
		return400IfFileExtRejected: virustoalConfig.return400IfFileExtRejected,
		generalFunc:                general_functions.NewGeneralFunc(httpMsg, xICAPMetadata),
	}
}
