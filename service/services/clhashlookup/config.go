package clhashlookup

import (
	"net/textproto"
	"sync"
	"time"

	http_message "github.com/egirna/icapeg/http-message"
	"github.com/egirna/icapeg/logging"
	"github.com/egirna/icapeg/readValues"
	services_utilities "github.com/egirna/icapeg/service/services-utilities"
	general_functions "github.com/egirna/icapeg/service/services-utilities/general-functions"
)

var doOnce sync.Once
var HashLookupConfig *Hashlookup

// Hashlookup represents the information regarding the Hashlookup service
type Hashlookup struct {
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
	Timeout                    time.Duration
	returnOrigIfMaxSizeExc     bool
	return400IfFileExtRejected bool
	generalFunc                *general_functions.GeneralFunc
	BypassOnApiError           bool
	verifyServerCert           bool
	FileHash                   string
	CaseBlockHttpResponseCode  int
	CaseBlockHttpBody          bool
	ExceptionPage              string
	IcapHeaders                textproto.MIMEHeader
}

func InitHashlookupConfig(serviceName string) {
	logging.Logger.Debug("loading " + serviceName + " service configurations")
	doOnce.Do(func() {
		HashLookupConfig = &Hashlookup{
			maxFileSize:                readValues.ReadValuesInt(serviceName + ".max_filesize"),
			bypassExts:                 readValues.ReadValuesSlice(serviceName + ".bypass_extensions"),
			processExts:                readValues.ReadValuesSlice(serviceName + ".process_extensions"),
			rejectExts:                 readValues.ReadValuesSlice(serviceName + ".reject_extensions"),
			ScanUrl:                    readValues.ReadValuesString(serviceName + ".scan_url"),
			Timeout:                    readValues.ReadValuesDuration(serviceName + ".timeout"),
			returnOrigIfMaxSizeExc:     readValues.ReadValuesBool(serviceName + ".return_original_if_max_file_size_exceeded"),
			return400IfFileExtRejected: readValues.ReadValuesBool(serviceName + ".return_400_if_file_ext_rejected"),
			BypassOnApiError:           readValues.ReadBoolFromEnv(serviceName + ".bypass_on_api_error"),
			verifyServerCert:           readValues.ReadValuesBool(serviceName + ".verify_server_cert"),
			CaseBlockHttpResponseCode:  readValues.ReadValuesInt(serviceName + ".http_exception_response_code"),
			CaseBlockHttpBody:          readValues.ReadValuesBool(serviceName + ".http_exception_has_body"),
			ExceptionPage:              readValues.ReadValuesString(serviceName + ".exception_page"),
		}
		HashLookupConfig.extArrs = services_utilities.InitExtsArr(HashLookupConfig.processExts, HashLookupConfig.rejectExts, HashLookupConfig.bypassExts)
	})
}

// NewHashlookupService returns a new populated instance of the Hashlookup service
func NewHashlookupService(serviceName, methodName string, httpMsg *http_message.HttpMsg, xICAPMetadata string) *Hashlookup {
	return &Hashlookup{
		xICAPMetadata:              xICAPMetadata,
		httpMsg:                    httpMsg,
		serviceName:                serviceName,
		methodName:                 methodName,
		maxFileSize:                HashLookupConfig.maxFileSize,
		bypassExts:                 HashLookupConfig.bypassExts,
		processExts:                HashLookupConfig.processExts,
		rejectExts:                 HashLookupConfig.rejectExts,
		extArrs:                    HashLookupConfig.extArrs,
		ScanUrl:                    HashLookupConfig.ScanUrl,
		Timeout:                    HashLookupConfig.Timeout * time.Second,
		returnOrigIfMaxSizeExc:     HashLookupConfig.returnOrigIfMaxSizeExc,
		return400IfFileExtRejected: HashLookupConfig.return400IfFileExtRejected,
		generalFunc:                general_functions.NewGeneralFunc(httpMsg, xICAPMetadata),
		verifyServerCert:           HashLookupConfig.verifyServerCert,
		BypassOnApiError:           HashLookupConfig.BypassOnApiError,
		CaseBlockHttpResponseCode:  HashLookupConfig.CaseBlockHttpResponseCode,
		CaseBlockHttpBody:          HashLookupConfig.CaseBlockHttpBody,
		ExceptionPage:              HashLookupConfig.ExceptionPage,
	}
}
