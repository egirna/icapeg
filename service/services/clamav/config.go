package clamav

import (
	"icapeg/http-message"
	"icapeg/readValues"
	services_utilities "icapeg/service/services-utilities"
	general_functions "icapeg/service/services-utilities/general-functions"
	"sync"
	"time"
)

// the clamav constants
const (
	ClamavMalStatus  = "FOUND"
	ClamavIdentifier = "CLAMAV ID"
)

var doOnce sync.Once
var clamavConfig *Clamav

// Clamav represents the information regarding the clamav service
type Clamav struct {
	httpMsg                    *http_message.HttpMsg
	elapsed                    time.Duration
	serviceName                string
	methodName                 string
	maxFileSize                int
	bypassExts                 []string
	processExts                []string
	rejectExts                 []string
	extArrs                    []services_utilities.Extension
	SocketPath                 string
	Timeout                    time.Duration
	badFileStatus              []string
	okFileStatus               []string
	returnOrigIfMaxSizeExc     bool
	return400IfFileExtRejected bool
	generalFunc                *general_functions.GeneralFunc
}

func InitClamavConfig(serviceName string) {
	doOnce.Do(func() {
		clamavConfig = &Clamav{
			maxFileSize:                readValues.ReadValuesInt(serviceName + ".max_filesize"),
			bypassExts:                 readValues.ReadValuesSlice(serviceName + ".bypass_extensions"),
			processExts:                readValues.ReadValuesSlice(serviceName + ".process_extensions"),
			rejectExts:                 readValues.ReadValuesSlice(serviceName + ".reject_extensions"),
			returnOrigIfMaxSizeExc:     readValues.ReadValuesBool(serviceName + ".return_original_if_max_file_size_exceeded"),
			SocketPath:                 readValues.ReadValuesString(serviceName + ".socket_path"),
			Timeout:                    readValues.ReadValuesDuration(serviceName+".timeout") * time.Second,
			return400IfFileExtRejected: readValues.ReadValuesBool(serviceName + ".return_400_if_file_ext_rejected"),
		}

		clamavConfig.extArrs = services_utilities.InitExtsArr(clamavConfig.processExts, clamavConfig.rejectExts, clamavConfig.bypassExts)
	})
}

func NewClamavService(serviceName, methodName string, httpMsg *http_message.HttpMsg) *Clamav {
	return &Clamav{
		httpMsg:                    httpMsg,
		serviceName:                serviceName,
		methodName:                 methodName,
		generalFunc:                general_functions.NewGeneralFunc(httpMsg),
		maxFileSize:                clamavConfig.maxFileSize,
		bypassExts:                 clamavConfig.bypassExts,
		processExts:                clamavConfig.processExts,
		rejectExts:                 clamavConfig.rejectExts,
		extArrs:                    clamavConfig.extArrs,
		Timeout:                    clamavConfig.Timeout * time.Second,
		SocketPath:                 clamavConfig.SocketPath,
		returnOrigIfMaxSizeExc:     clamavConfig.returnOrigIfMaxSizeExc,
		return400IfFileExtRejected: clamavConfig.return400IfFileExtRejected,
	}
}
