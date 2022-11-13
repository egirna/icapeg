package echo

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
var echoConfig *Echo

const EchoIdentifier = "ECHO ID"

// Echo represents the information regarding the Echo service
type Echo struct {
	xICAPMetadata              string
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
	logging.Logger.Debug("loading " + serviceName + " service configurations")
	doOnce.Do(func() {
		echoConfig = &Echo{
			maxFileSize:                readValues.ReadValuesInt(serviceName + ".max_filesize"),
			bypassExts:                 readValues.ReadValuesSlice(serviceName + ".bypass_extensions"),
			processExts:                readValues.ReadValuesSlice(serviceName + ".process_extensions"),
			rejectExts:                 readValues.ReadValuesSlice(serviceName + ".reject_extensions"),
			returnOrigIfMaxSizeExc:     readValues.ReadValuesBool(serviceName + ".return_original_if_max_file_size_exceeded"),
			return400IfFileExtRejected: readValues.ReadValuesBool(serviceName + ".return_400_if_file_ext_rejected"),
		}
		echoConfig.extArrs = services_utilities.InitExtsArr(echoConfig.processExts, echoConfig.rejectExts, echoConfig.bypassExts)
	})
}

// NewEchoService returns a new populated instance of the Echo service
func NewEchoService(serviceName, methodName string, httpMsg *http_message.HttpMsg, xICAPMetadata string) *Echo {
	return &Echo{
		xICAPMetadata:              xICAPMetadata,
		httpMsg:                    httpMsg,
		serviceName:                serviceName,
		methodName:                 methodName,
		generalFunc:                general_functions.NewGeneralFunc(httpMsg, xICAPMetadata),
		maxFileSize:                echoConfig.maxFileSize,
		bypassExts:                 echoConfig.bypassExts,
		processExts:                echoConfig.processExts,
		rejectExts:                 echoConfig.rejectExts,
		extArrs:                    echoConfig.extArrs,
		returnOrigIfMaxSizeExc:     echoConfig.returnOrigIfMaxSizeExc,
		return400IfFileExtRejected: echoConfig.return400IfFileExtRejected,
	}
}
