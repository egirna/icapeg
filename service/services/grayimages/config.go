package grayimages

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
var grayimagesConfig *Grayimages

const GrayimagesIdentifier = "GRAYIMAGES ID"

// Grayimages represents the information regarding the grayimages service
type Grayimages struct {
	httpMsg                    *http_message.HttpMsg
	elapsed                    time.Duration
	serviceName                string
	methodName                 string
	maxFileSize                int
	bypassExts                 []string
	processExts                []string
	rejectExts                 []string
	BaseURL                    string
	extArrs                    []services_utilities.Extension
	returnOrigIfMaxSizeExc     bool
	return400IfFileExtRejected bool
	generalFunc                *general_functions.GeneralFunc
}

func InitGrayimagesConfig(serviceName string) {
	logging.Logger.Debug("loading " + serviceName + " service configurations")
	doOnce.Do(func() {
		grayimagesConfig = &Grayimages{
			BaseURL:                    readValues.ReadValuesString(serviceName + ".base_url"),
			maxFileSize:                readValues.ReadValuesInt(serviceName + ".max_filesize"),
			bypassExts:                 readValues.ReadValuesSlice(serviceName + ".bypass_extensions"),
			processExts:                readValues.ReadValuesSlice(serviceName + ".process_extensions"),
			rejectExts:                 readValues.ReadValuesSlice(serviceName + ".reject_extensions"),
			returnOrigIfMaxSizeExc:     readValues.ReadValuesBool(serviceName + ".return_original_if_max_file_size_exceeded"),
			return400IfFileExtRejected: readValues.ReadValuesBool(serviceName + ".return_400_if_file_ext_rejected"),
		}
		grayimagesConfig.extArrs = services_utilities.InitExtsArr(grayimagesConfig.processExts, grayimagesConfig.rejectExts, grayimagesConfig.bypassExts)
	})
}

// NewGrayimagesService returns a new populated instance of the Grayimages service
func NewGrayimagesService(serviceName, methodName string, httpMsg *http_message.HttpMsg) *Grayimages {
	return &Grayimages{
		httpMsg:                    httpMsg,
		serviceName:                serviceName,
		methodName:                 methodName,
		generalFunc:                general_functions.NewGeneralFunc(httpMsg),
		BaseURL:                    grayimagesConfig.BaseURL,
		maxFileSize:                grayimagesConfig.maxFileSize,
		bypassExts:                 grayimagesConfig.bypassExts,
		processExts:                grayimagesConfig.processExts,
		rejectExts:                 grayimagesConfig.rejectExts,
		extArrs:                    grayimagesConfig.extArrs,
		returnOrigIfMaxSizeExc:     grayimagesConfig.returnOrigIfMaxSizeExc,
		return400IfFileExtRejected: grayimagesConfig.return400IfFileExtRejected,
	}
}
