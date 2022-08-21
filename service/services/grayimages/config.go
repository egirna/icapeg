package grayimages

import (
	"icapeg/config"
	"icapeg/readValues"
	general_functions "icapeg/service/services-utilities/general-functions"
	"icapeg/utils"
	"sync"
	"time"
)

var doOnce sync.Once
var grayimagesConfig *GrayImages

// Echo represents the information regarding the Echo service
type GrayImages struct {
	httpMsg                    *utils.HttpMsg
	elapsed                    time.Duration
	serviceName                string
	methodName                 string
	maxFileSize                int
	bypassExts                 []string
	processExts                []string
	rejectExts                 []string
	extArrs                    []config.Extension
	BaseURL                    string
	Timeout                    time.Duration
	APIKey                     string
	ScanEndpoint               string
	FailThreshold              int
	returnOrigIfMaxSizeExc     bool
	return400IfFileExtRejected bool
	generalFunc                *general_functions.GeneralFunc
}

func InitGrayImagesConfig(serviceName string) {
	doOnce.Do(func() {
		grayimagesConfig = &GrayImages{
			maxFileSize:                readValues.ReadValuesInt(serviceName + ".max_filesize"),
			bypassExts:                 readValues.ReadValuesSlice(serviceName + ".bypass_extensions"),
			processExts:                readValues.ReadValuesSlice(serviceName + ".process_extensions"),
			rejectExts:                 readValues.ReadValuesSlice(serviceName + ".reject_extensions"),
			BaseURL:                    readValues.ReadValuesString(serviceName + ".base_url"),
			Timeout:                    readValues.ReadValuesDuration(serviceName+".timeout") * time.Second,
			APIKey:                     readValues.ReadValuesString(serviceName + ".api_key"),
			ScanEndpoint:               readValues.ReadValuesString(serviceName + ".scan_endpoint"),
			FailThreshold:              readValues.ReadValuesInt(serviceName + ".fail_threshold"),
			returnOrigIfMaxSizeExc:     readValues.ReadValuesBool(serviceName + ".return_original_if_max_file_size_exceeded"),
			return400IfFileExtRejected: readValues.ReadValuesBool(serviceName + ".return_400_if_file_ext_rejected"),
		}

		process := config.Extension{Name: "process", Exts: grayimagesConfig.processExts}
		reject := config.Extension{Name: "reject", Exts: grayimagesConfig.rejectExts}
		bypass := config.Extension{Name: "bypass", Exts: grayimagesConfig.bypassExts}
		extArrs := make([]config.Extension, 3)
		ind := 0
		if len(process.Exts) == 1 && process.Exts[0] == "*" {
			extArrs[2] = process
		} else {
			extArrs[ind] = process
			ind++
		}
		if len(reject.Exts) == 1 && reject.Exts[0] == "*" {
			extArrs[2] = reject
		} else {
			extArrs[ind] = reject
			ind++
		}
		if len(bypass.Exts) == 1 && bypass.Exts[0] == "*" {
			extArrs[2] = bypass
		} else {
			extArrs[ind] = bypass
			ind++
		}
		grayimagesConfig.extArrs = extArrs
	})
}

// NewEchoService returns a new populated instance of the Echo service
func NewEchoService(serviceName, methodName string, httpMsg *utils.HttpMsg) *GrayImages {
	return &GrayImages{
		httpMsg:                    httpMsg,
		serviceName:                serviceName,
		methodName:                 methodName,
		generalFunc:                general_functions.NewGeneralFunc(httpMsg),
		maxFileSize:                grayimagesConfig.maxFileSize,
		bypassExts:                 grayimagesConfig.bypassExts,
		processExts:                grayimagesConfig.processExts,
		rejectExts:                 grayimagesConfig.rejectExts,
		extArrs:                    grayimagesConfig.extArrs,
		BaseURL:                    grayimagesConfig.BaseURL,
		Timeout:                    grayimagesConfig.Timeout * time.Second,
		APIKey:                     grayimagesConfig.APIKey,
		ScanEndpoint:               grayimagesConfig.ScanEndpoint,
		FailThreshold:              grayimagesConfig.FailThreshold,
		returnOrigIfMaxSizeExc:     grayimagesConfig.returnOrigIfMaxSizeExc,
		return400IfFileExtRejected: grayimagesConfig.return400IfFileExtRejected,
	}
}
