package clamav

import (
	http_message "icapeg/http-message"
	"icapeg/logging"
	"icapeg/readValues"
	services_utilities "icapeg/service/services-utilities"
	general_functions "icapeg/service/services-utilities/general-functions"
	"net/textproto"
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
	xICAPMetadata string
	httpMsg       *http_message.HttpMsg
	//elapsed                    time.Duration
	serviceName string
	methodName  string
	maxFileSize int
	bypassExts  []string
	processExts []string
	rejectExts  []string
	extArrs     []services_utilities.Extension
	SocketPath  string
	Timeout     time.Duration
	//badFileStatus              []string
	//okFileStatus               []string
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

func InitClamavConfig(serviceName string) {
	logging.Logger.Debug("loading " + serviceName + " service configurations")
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
			BypassOnApiError:           readValues.ReadBoolFromEnv(serviceName + ".bypass_on_api_error"),
			verifyServerCert:           readValues.ReadValuesBool(serviceName + ".verify_server_cert"),
			CaseBlockHttpResponseCode:  readValues.ReadValuesInt(serviceName + ".http_exception_response_code"),
			CaseBlockHttpBody:          readValues.ReadValuesBool(serviceName + ".http_exception_has_body"),
			ExceptionPage:              readValues.ReadValuesString(serviceName + ".exception_page"),
		}

		clamavConfig.extArrs = services_utilities.InitExtsArr(clamavConfig.processExts, clamavConfig.rejectExts, clamavConfig.bypassExts)
	})
}

func NewClamavService(serviceName, methodName string, httpMsg *http_message.HttpMsg, xICAPMetadata string) *Clamav {
	return &Clamav{
		xICAPMetadata:              xICAPMetadata,
		httpMsg:                    httpMsg,
		serviceName:                serviceName,
		methodName:                 methodName,
		generalFunc:                general_functions.NewGeneralFunc(httpMsg, xICAPMetadata),
		maxFileSize:                clamavConfig.maxFileSize,
		bypassExts:                 clamavConfig.bypassExts,
		processExts:                clamavConfig.processExts,
		rejectExts:                 clamavConfig.rejectExts,
		extArrs:                    clamavConfig.extArrs,
		Timeout:                    clamavConfig.Timeout * time.Second,
		SocketPath:                 clamavConfig.SocketPath,
		returnOrigIfMaxSizeExc:     clamavConfig.returnOrigIfMaxSizeExc,
		return400IfFileExtRejected: clamavConfig.return400IfFileExtRejected,
		verifyServerCert:           clamavConfig.verifyServerCert,
		BypassOnApiError:           clamavConfig.BypassOnApiError,
		CaseBlockHttpResponseCode:  clamavConfig.CaseBlockHttpResponseCode,
		CaseBlockHttpBody:          clamavConfig.CaseBlockHttpBody,
		ExceptionPage:              clamavConfig.ExceptionPage,
	}
}
