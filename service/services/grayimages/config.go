package grayimages

import (
	"icapeg/readValues"
	services_utilities "icapeg/service/services-utilities"
	general_functions "icapeg/service/services-utilities/general-functions"
	"icapeg/utils"
	"sync"
	"time"
)

var doOnce sync.Once
var grayImagesConfig *GrayImages

// Echo represents the information regarding the Echo service
type GrayImages struct {
	httpMsg                    *utils.HttpMsg
	elapsed                    time.Duration
	serviceName                string
	methodName                 string
	maxFileSize                int
	bypassExts                 []string
	processExts                []string
	rejectExts                 []string
	extArrs                    []services_utilities.Extension
	BaseURL                    string
	Timeout                    time.Duration
	verifyServerCert           bool
	APIKey                     string
	ScanEndpoint               string
	FailThreshold              int
	returnOrigIfMaxSizeExc     bool
	return400IfFileExtRejected bool
	generalFunc                *general_functions.GeneralFunc
	imagesDir                  string
}

func InitGrayImagesConfig(serviceName string) {
	doOnce.Do(func() {
		grayImagesConfig = &GrayImages{
			maxFileSize:                readValues.ReadValuesInt(serviceName + ".max_filesize"),
			bypassExts:                 readValues.ReadValuesSlice(serviceName + ".bypass_extensions"),
			processExts:                readValues.ReadValuesSlice(serviceName + ".process_extensions"),
			rejectExts:                 readValues.ReadValuesSlice(serviceName + ".reject_extensions"),
			BaseURL:                    readValues.ReadValuesString(serviceName + ".base_url"),
			Timeout:                    readValues.ReadValuesDuration(serviceName+".timeout") * time.Second,
			verifyServerCert:           readValues.ReadValuesBool(serviceName + ".verify_server_cert"),
			APIKey:                     readValues.ReadValuesString(serviceName + ".api_key"),
			ScanEndpoint:               readValues.ReadValuesString(serviceName + ".scan_endpoint"),
			FailThreshold:              readValues.ReadValuesInt(serviceName + ".fail_threshold"),
			returnOrigIfMaxSizeExc:     readValues.ReadValuesBool(serviceName + ".return_original_if_max_file_size_exceeded"),
			return400IfFileExtRejected: readValues.ReadValuesBool(serviceName + ".return_400_if_file_ext_rejected"),
			imagesDir:                  readValues.ReadValuesString(serviceName + ".images_dir"),
		}

		grayImagesConfig.extArrs = services_utilities.InitExtsArr(grayImagesConfig.processExts, grayImagesConfig.rejectExts, grayImagesConfig.bypassExts)
	})
}

// NewEchoService returns a new populated instance of the Echo service
func NewGrayImagesService(serviceName, methodName string, httpMsg *utils.HttpMsg) *GrayImages {
	return &GrayImages{
		httpMsg:                    httpMsg,
		serviceName:                serviceName,
		methodName:                 methodName,
		generalFunc:                general_functions.NewGeneralFunc(httpMsg),
		maxFileSize:                grayImagesConfig.maxFileSize,
		bypassExts:                 grayImagesConfig.bypassExts,
		processExts:                grayImagesConfig.processExts,
		rejectExts:                 grayImagesConfig.rejectExts,
		extArrs:                    grayImagesConfig.extArrs,
		BaseURL:                    grayImagesConfig.BaseURL,
		Timeout:                    grayImagesConfig.Timeout * time.Second,
		verifyServerCert:           grayImagesConfig.verifyServerCert,
		APIKey:                     grayImagesConfig.APIKey,
		ScanEndpoint:               grayImagesConfig.ScanEndpoint,
		FailThreshold:              grayImagesConfig.FailThreshold,
		returnOrigIfMaxSizeExc:     grayImagesConfig.returnOrigIfMaxSizeExc,
		return400IfFileExtRejected: grayImagesConfig.return400IfFileExtRejected,
		imagesDir:                  grayImagesConfig.imagesDir,
	}
}
