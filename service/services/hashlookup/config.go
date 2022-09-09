package hashlookuppackage

import (
	http_message "icapeg/http-message"
	"icapeg/logging"
	"icapeg/readValues"
	services_utilities "icapeg/service/services-utilities"
	general_functions "icapeg/service/services-utilities/general-functions"
	"sync"
	"time"
)

var doOnce sync.Once
var hashLookupConfig *Hashlookup

const HashLookupIdentifier = "HASHLOOKUP ID"

// Hashlookup represents the information regarding the Hashlookup service
type Hashlookup struct {
	httpMsg                    *http_message.HttpMsg
	serviceName                string
	methodName                 string
	maxFileSize                int
	bypassExts                 []string
	processExts                []string
	rejectExts                 []string
	extArrs                    []services_utilities.Extension
	ScanUrl                    string
	Timeout                    time.Duration
	returnOrigIfMaxSizeExc     bool
	return400IfFileExtRejected bool
	generalFunc                *general_functions.GeneralFunc
}

func InitHashlookupConfig(serviceName string) {
	logging.Logger.Debug("loading " + serviceName + " service configurations")
	doOnce.Do(func() {
		hashLookupConfig = &Hashlookup{
			maxFileSize:                readValues.ReadValuesInt(serviceName + ".max_filesize"),
			bypassExts:                 readValues.ReadValuesSlice(serviceName + ".bypass_extensions"),
			processExts:                readValues.ReadValuesSlice(serviceName + ".process_extensions"),
			rejectExts:                 readValues.ReadValuesSlice(serviceName + ".reject_extensions"),
			ScanUrl:                    readValues.ReadValuesString(serviceName + ".scan_url"),
			Timeout:                    readValues.ReadValuesDuration(serviceName + ".timeout"),
			returnOrigIfMaxSizeExc:     readValues.ReadValuesBool(serviceName + ".return_original_if_max_file_size_exceeded"),
			return400IfFileExtRejected: readValues.ReadValuesBool(serviceName + ".return_400_if_file_ext_rejected"),
		}
		hashLookupConfig.extArrs = services_utilities.InitExtsArr(hashLookupConfig.processExts, hashLookupConfig.rejectExts, hashLookupConfig.bypassExts)
	})
}

// NewHashlookupService returns a new populated instance of the Hashlookup service
func NewHashlookupService(serviceName, methodName string, httpMsg *http_message.HttpMsg) *Hashlookup {
	return &Hashlookup{
		httpMsg:                    httpMsg,
		serviceName:                serviceName,
		methodName:                 methodName,
		maxFileSize:                hashLookupConfig.maxFileSize,
		bypassExts:                 hashLookupConfig.bypassExts,
		processExts:                hashLookupConfig.processExts,
		rejectExts:                 hashLookupConfig.rejectExts,
		extArrs:                    hashLookupConfig.extArrs,
		ScanUrl:                    hashLookupConfig.ScanUrl,
		Timeout:                    hashLookupConfig.Timeout * time.Second,
		returnOrigIfMaxSizeExc:     hashLookupConfig.returnOrigIfMaxSizeExc,
		return400IfFileExtRejected: hashLookupConfig.return400IfFileExtRejected,
		generalFunc:                general_functions.NewGeneralFunc(httpMsg),
	}
}
