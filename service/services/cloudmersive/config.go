package cloudmersive

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
var cloudMersiveConfig *CloudMersive

const CloudMersiveIdentifier = "CLOUDMERSIVE ID"

type CloudMersive struct {
	xICAPMetadata                string
	httpMsg                      *http_message.HttpMsg
	serviceName                  string
	methodName                   string
	bypassExts                   []string
	processExts                  []string
	rejectExts                   []string
	extArrs                      []services_utilities.Extension
	verifyServerCert             bool
	allowExecutables             bool
	allowInvalidFiles            bool
	allowScripts                 bool
	allowPasswordProtectedFiles  bool
	allowMacros                  bool
	allowXmlExternalEntities     bool
	allowHtml                    bool
	allowInsecureDeserialization bool
	maxFileSize                  int
	BaseURL                      string
	ScanEndPoint                 string
	Timeout                      time.Duration
	APIKey                       string
	returnOrigIfMaxSizeExc       bool
	return400IfFileExtRejected   bool
	generalFunc                  *general_functions.GeneralFunc
}

func InitCloudMersiveConfig(serviceName string) {
	logging.Logger.Debug("loading " + serviceName + " service configurations")
	doOnce.Do(func() {
		cloudMersiveConfig = &CloudMersive{
			maxFileSize:                  readValues.ReadValuesInt(serviceName + ".max_filesize"),
			BaseURL:                      readValues.ReadValuesString(serviceName + ".base_url"),
			ScanEndPoint:                 readValues.ReadValuesString(serviceName + ".scan_endpoint"),
			Timeout:                      readValues.ReadValuesDuration(serviceName + ".timeout"),
			verifyServerCert:             readValues.ReadValuesBool(serviceName + ".verify_server_cert"),
			APIKey:                       readValues.ReadValuesString(serviceName + ".api_key"),
			returnOrigIfMaxSizeExc:       readValues.ReadValuesBool(serviceName + ".return_original_if_max_file_size_exceeded"),
			return400IfFileExtRejected:   readValues.ReadValuesBool(serviceName + ".return_400_if_file_ext_rejected"),
			allowScripts:                 readValues.ReadValuesBool(serviceName + ".allow_scripts"),
			allowExecutables:             readValues.ReadValuesBool(serviceName + ".allow_executables"),
			allowMacros:                  readValues.ReadValuesBool(serviceName + ".allow_macros"),
			allowInvalidFiles:            readValues.ReadValuesBool(serviceName + ".allow_invalid_files"),
			allowXmlExternalEntities:     readValues.ReadValuesBool(serviceName + ".allow_xml_external_entities"),
			allowPasswordProtectedFiles:  readValues.ReadValuesBool(serviceName + ".allow_password_protected_files"),
			allowInsecureDeserialization: readValues.ReadValuesBool(serviceName + ".allow_insecure_deserialization"),
			allowHtml:                    readValues.ReadValuesBool(serviceName + ".allow_html"),
			bypassExts:                   readValues.ReadValuesSlice(serviceName + ".bypass_extensions"),
			processExts:                  readValues.ReadValuesSlice(serviceName + ".process_extensions"),
			rejectExts:                   readValues.ReadValuesSlice(serviceName + ".reject_extensions"),
		}
		cloudMersiveConfig.extArrs = services_utilities.InitExtsArr(cloudMersiveConfig.processExts, cloudMersiveConfig.rejectExts, cloudMersiveConfig.bypassExts)
	})
}

func NewCloudMersiveService(serviceName, methodName string, httpMsg *http_message.HttpMsg, xICAPMetadata string) *CloudMersive {
	return &CloudMersive{
		xICAPMetadata:               xICAPMetadata,
		httpMsg:                     httpMsg,
		serviceName:                 serviceName,
		methodName:                  methodName,
		allowExecutables:            cloudMersiveConfig.allowExecutables,
		allowXmlExternalEntities:    cloudMersiveConfig.allowXmlExternalEntities,
		allowMacros:                 cloudMersiveConfig.allowMacros,
		allowScripts:                cloudMersiveConfig.allowScripts,
		allowInvalidFiles:           cloudMersiveConfig.allowInvalidFiles,
		allowPasswordProtectedFiles: cloudMersiveConfig.allowPasswordProtectedFiles,
		maxFileSize:                 cloudMersiveConfig.maxFileSize,
		BaseURL:                     cloudMersiveConfig.BaseURL,
		ScanEndPoint:                cloudMersiveConfig.ScanEndPoint,
		Timeout:                     cloudMersiveConfig.Timeout * time.Second,
		verifyServerCert:            cloudMersiveConfig.verifyServerCert,
		APIKey:                      cloudMersiveConfig.APIKey,
		returnOrigIfMaxSizeExc:      cloudMersiveConfig.returnOrigIfMaxSizeExc,
		return400IfFileExtRejected:  cloudMersiveConfig.return400IfFileExtRejected,
		generalFunc:                 general_functions.NewGeneralFunc(httpMsg, xICAPMetadata),
		bypassExts:                  cloudMersiveConfig.bypassExts,
		processExts:                 cloudMersiveConfig.processExts,
		rejectExts:                  cloudMersiveConfig.rejectExts,
		extArrs:                     cloudMersiveConfig.extArrs,
	}
}
