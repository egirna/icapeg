package grayimages

import (
	"icapeg/http-message"
	"icapeg/readValues"
	services_utilities "icapeg/service/services-utilities"
	general_functions "icapeg/service/services-utilities/general-functions"
	"sync"
	"time"
)

var doOnce sync.Once
var grayimagesConfig *Grayimages

const GrayimagesIdentifier = "ECHO ID"

// Echo represents the information regarding the Echo service
type Grayimages struct {
	httpMsg                    *http_message.HttpMsg
	elapsed                    time.Duration
	serviceName                string
	methodName                 string
	maxFileSize                int
	bypassExts                 []string
	processExts                []string
	rejectExts                 []string
	extArrs                    []services_utilities.Extension
	returnOrigIfMaxSizeExc     bool
	return400IfFileExtRejected bool
	generalFunc                *general_functions.GeneralFunc
}

func InitEchoConfig(serviceName string) {
	doOnce.Do(func() {
		grayimagesConfig = &Grayimages{
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

// NewEchoService returns a new populated instance of the Echo service
func NewEchoService(serviceName, methodName string, httpMsg *http_message.HttpMsg) *Grayimages {
	return &Grayimages{
		httpMsg:                    httpMsg,
		serviceName:                serviceName,
		methodName:                 methodName,
		generalFunc:                general_functions.NewGeneralFunc(httpMsg),
		maxFileSize:                grayimagesConfig.maxFileSize,
		bypassExts:                 grayimagesConfig.bypassExts,
		processExts:                grayimagesConfig.processExts,
		rejectExts:                 grayimagesConfig.rejectExts,
		extArrs:                    grayimagesConfig.extArrs,
		returnOrigIfMaxSizeExc:     grayimagesConfig.returnOrigIfMaxSizeExc,
		return400IfFileExtRejected: grayimagesConfig.return400IfFileExtRejected,
	}
}
