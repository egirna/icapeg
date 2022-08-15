package virustotal

import (
	"fmt"
	"icapeg/config"
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
	httpMsg                    *utils.HttpMsg
	serviceName                string
	methodName                 string
	maxFileSize                int
	bypassExts                 []string
	processExts                []string
	rejectExts                 []string
	extArrs                    []config.Extension
	ScanUrl                    string
	ReportUrl                  string
	Timeout                    time.Duration
	APIKey                     string
	FailThreshold              int
	policy                     string
	returnOrigIfMaxSizeExc     bool
	return400IfFileExtRejected bool
	generalFunc                *general_functions.GeneralFunc
}

func InitVirustotalConfig(serviceName string) {
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
			FailThreshold:              readValues.ReadValuesInt(serviceName + ".fail_threshold"),
			policy:                     readValues.ReadValuesString(serviceName + ".policy"),
			returnOrigIfMaxSizeExc:     readValues.ReadValuesBool(serviceName + ".return_original_if_max_file_size_exceeded"),
			return400IfFileExtRejected: readValues.ReadValuesBool(serviceName + ".return_400_if_file_ext_rejected"),
		}
		process := config.Extension{Name: "process", Exts: virustoalConfig.processExts}
		reject := config.Extension{Name: "reject", Exts: virustoalConfig.rejectExts}
		bypass := config.Extension{Name: "bypass", Exts: virustoalConfig.bypassExts}
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
		fmt.Println(extArrs)
		virustoalConfig.extArrs = extArrs
	})
}

// NewVirustotalService returns a new populated instance of the Virustotal service
func NewVirustotalService(serviceName, methodName string, httpMsg *utils.HttpMsg) *Virustotal {
	return &Virustotal{
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
		FailThreshold:              virustoalConfig.FailThreshold,
		policy:                     virustoalConfig.policy,
		returnOrigIfMaxSizeExc:     virustoalConfig.returnOrigIfMaxSizeExc,
		return400IfFileExtRejected: virustoalConfig.return400IfFileExtRejected,
		generalFunc:                general_functions.NewGeneralFunc(httpMsg),
	}
}
