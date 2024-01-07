package hashlocal

import (
	http_message "icapeg/http-message"
	"icapeg/logging"
	"icapeg/readValues"
	services_utilities "icapeg/service/services-utilities"
	general_functions "icapeg/service/services-utilities/general-functions"
	"net/textproto"
	"sync"
	//"time"
)

var doOnce sync.Once
var HashlocalConfig *Hashlocal

// Hashlookup represents the information regarding the Hashlookup service
type Hashlocal struct {
	xICAPMetadata string
	httpMsg       *http_message.HttpMsg
	serviceName   string
	methodName    string
	maxFileSize   int
	bypassExts    []string
	processExts   []string
	rejectExts    []string
	extArrs       []services_utilities.Extension
	ScanUrl       string
	//Timeout                    time.Duration
	returnOrigIfMaxSizeExc     bool
	return400IfFileExtRejected bool
	generalFunc                *general_functions.GeneralFunc
	BypassOnApiError           bool
	//verifyServerCert           bool
	FileHash                  string
	CaseBlockHttpResponseCode int
	CaseBlockHttpBody         bool
	ExceptionPage             string
	IcapHeaders               textproto.MIMEHeader
}

func InitHashlocalConfig(serviceName string) {
	logging.Logger.Debug("loading " + serviceName + " service configurations")
	doOnce.Do(func() {
		HashlocalConfig = &Hashlocal{
			maxFileSize: readValues.ReadValuesInt(serviceName + ".max_filesize"),
			bypassExts:  readValues.ReadValuesSlice(serviceName + ".bypass_extensions"),
			processExts: readValues.ReadValuesSlice(serviceName + ".process_extensions"),
			rejectExts:  readValues.ReadValuesSlice(serviceName + ".reject_extensions"),
			ScanUrl:     readValues.ReadValuesString(serviceName + ".scan_url"),
			//Timeout:                    readValues.ReadValuesDuration(serviceName + ".timeout"),
			returnOrigIfMaxSizeExc:     readValues.ReadValuesBool(serviceName + ".return_original_if_max_file_size_exceeded"),
			return400IfFileExtRejected: readValues.ReadValuesBool(serviceName + ".return_400_if_file_ext_rejected"),
			//BypassOnApiError:           readValues.ReadBoolFromEnv(serviceName + ".bypass_on_api_error"),
			//verifyServerCert:           readValues.ReadValuesBool(serviceName + ".verify_server_cert"),
			CaseBlockHttpResponseCode: readValues.ReadValuesInt(serviceName + ".http_exception_response_code"),
			CaseBlockHttpBody:         readValues.ReadValuesBool(serviceName + ".http_exception_has_body"),
			ExceptionPage:             readValues.ReadValuesString(serviceName + ".exception_page"),
		}
		HashlocalConfig.extArrs = services_utilities.InitExtsArr(HashlocalConfig.processExts, HashlocalConfig.rejectExts, HashlocalConfig.bypassExts)
	})
}

// NewHashlookupService returns a new populated instance of the Hashlookup service
func NewHashlocalService(serviceName, methodName string, httpMsg *http_message.HttpMsg, xICAPMetadata string) *Hashlocal {
	return &Hashlocal{
		xICAPMetadata: xICAPMetadata,
		httpMsg:       httpMsg,
		serviceName:   serviceName,
		methodName:    methodName,
		maxFileSize:   HashlocalConfig.maxFileSize,
		bypassExts:    HashlocalConfig.bypassExts,
		processExts:   HashlocalConfig.processExts,
		rejectExts:    HashlocalConfig.rejectExts,
		extArrs:       HashlocalConfig.extArrs,
		ScanUrl:       HashlocalConfig.ScanUrl,
		//Timeout:                    HashlocalConfig.Timeout * time.Second,
		returnOrigIfMaxSizeExc:     HashlocalConfig.returnOrigIfMaxSizeExc,
		return400IfFileExtRejected: HashlocalConfig.return400IfFileExtRejected,
		generalFunc:                general_functions.NewGeneralFunc(httpMsg, xICAPMetadata),
		//verifyServerCert:           HashlocalConfig.verifyServerCert,
		BypassOnApiError:          HashlocalConfig.BypassOnApiError,
		CaseBlockHttpResponseCode: HashlocalConfig.CaseBlockHttpResponseCode,
		CaseBlockHttpBody:         HashlocalConfig.CaseBlockHttpBody,
		ExceptionPage:             HashlocalConfig.ExceptionPage,
	}
}
