package cloudmersive

import (
	general_functions "icapeg/service/services-utilities/general-functions"
	"icapeg/utils"
	"sync"
	"time"
)

var doOnce sync.Once
var cloudMersiveConfig *CloudMersive

type CloudMersive struct {
	httpMsg                     *utils.HttpMsg
	serviceName                 string
	methodName                  string
	allowExecutables            bool
	allowInvalidFiles           bool
	allowScripts                bool
	allowPasswordProtectedFiles bool
	allowMacros                 bool
	allowEmlExternalEntities    bool
	restrictFileTypes           bool
	maxFileSize                 int
	bypassExts                  []string
	processExts                 []string
	rejectExts                  []string
	ScanEndPoint                string
	Timeout                     time.Duration
	APIKey                      string
	FailThreshold               int
	policy                      string
	returnOrigIfMaxSizeExc      bool
	return400IfFileExtRejected  bool
	generalFunc                 *general_functions.GeneralFunc
}

func NewCloudMersiveService(serviceName, methodName string, httpMsg *utils.HttpMsg) *CloudMersive {
	return &CloudMersive{
		httpMsg:                     httpMsg,
		serviceName:                 serviceName,
		methodName:                  methodName,
		allowExecutables:            cloudMersiveConfig.allowExecutables,
		allowEmlExternalEntities:    cloudMersiveConfig.allowEmlExternalEntities,
		allowMacros:                 cloudMersiveConfig.allowMacros,
		allowScripts:                cloudMersiveConfig.allowScripts,
		allowInvalidFiles:           cloudMersiveConfig.allowInvalidFiles,
		allowPasswordProtectedFiles: cloudMersiveConfig.allowPasswordProtectedFiles,
		restrictFileTypes:           cloudMersiveConfig.restrictFileTypes,
		maxFileSize:                 cloudMersiveConfig.maxFileSize,
		bypassExts:                  cloudMersiveConfig.bypassExts,
		processExts:                 cloudMersiveConfig.processExts,
		rejectExts:                  cloudMersiveConfig.rejectExts,
		ScanEndPoint:                cloudMersiveConfig.ScanEndPoint,
		Timeout:                     cloudMersiveConfig.Timeout * time.Second,
		APIKey:                      cloudMersiveConfig.APIKey,
		FailThreshold:               cloudMersiveConfig.FailThreshold,
		policy:                      cloudMersiveConfig.policy,
		returnOrigIfMaxSizeExc:      cloudMersiveConfig.returnOrigIfMaxSizeExc,
		return400IfFileExtRejected:  cloudMersiveConfig.return400IfFileExtRejected,
		generalFunc:                 general_functions.NewGeneralFunc(httpMsg),
	}
}
