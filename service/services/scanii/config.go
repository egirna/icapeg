package scanii

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
var scaniiConfig *Scanii

const ScaniiIdentifier = "SCANII ID"

// Scanii represents the information regarding the Scanii service
type Scanii struct {
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

func InitScaniiConfig(serviceName string) {
	logging.Logger.Debug("loading " + serviceName + " service configurations")
	doOnce.Do(func() {
		scaniiConfig = &Scanii{
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
		scaniiConfig.extArrs = services_utilities.InitExtsArr(scaniiConfig.processExts, scaniiConfig.rejectExts, scaniiConfig.bypassExts)
	})
}

// NewScaniiService returns a new populated instance of the Scanii service
func NewScaniiService(serviceName, methodName string, httpMsg *http_message.HttpMsg) *Scanii {
	return &Scanii{
		httpMsg:                    httpMsg,
		serviceName:                serviceName,
		methodName:                 methodName,
		maxFileSize:                scaniiConfig.maxFileSize,
		bypassExts:                 scaniiConfig.bypassExts,
		processExts:                scaniiConfig.processExts,
		rejectExts:                 scaniiConfig.rejectExts,
		extArrs:                    scaniiConfig.extArrs,
		ScanUrl:                    scaniiConfig.ScanUrl,
		ReportUrl:                  scaniiConfig.ReportUrl,
		Timeout:                    scaniiConfig.Timeout * time.Second,
		APIKey:                     scaniiConfig.APIKey,
		returnOrigIfMaxSizeExc:     scaniiConfig.returnOrigIfMaxSizeExc,
		return400IfFileExtRejected: scaniiConfig.return400IfFileExtRejected,
		generalFunc:                general_functions.NewGeneralFunc(httpMsg),
	}
}
