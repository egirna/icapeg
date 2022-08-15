package clamav

import (
	"icapeg/config"
	"icapeg/readValues"
	general_functions "icapeg/service/services-utilities/general-functions"
	"icapeg/utils"
	"sync"
	"time"
)

// the clamav constants
const (
	ClamavMalStatus = "FOUND"
)

var doOnce sync.Once
var clamavConfig *Clamav

// Clamav represents the informations regarding the clamav service
type Clamav struct {
	httpMsg                    *utils.HttpMsg
	elapsed                    time.Duration
	serviceName                string
	methodName                 string
	maxFileSize                int
	bypassExts                 []string
	processExts                []string
	rejectExts                 []string
	extArrs                    []config.Extension
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

		process := config.Extension{Name: "process", Exts: clamavConfig.processExts}
		reject := config.Extension{Name: "reject", Exts: clamavConfig.rejectExts}
		bypass := config.Extension{Name: "bypass", Exts: clamavConfig.bypassExts}
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
		clamavConfig.extArrs = extArrs
	})
}

func NewClamavService(serviceName, methodName string, httpMsg *utils.HttpMsg) *Clamav {
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
