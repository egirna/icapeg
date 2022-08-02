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
	ScanUrl                     string
	ReportUrl                   string
	Timeout                     time.Duration
	APIKey                      string
	FailThreshold               int
	policy                      string
	returnOrigIfMaxSizeExc      bool
	return400IfFileExtRejected  bool
	generalFunc                 *general_functions.GeneralFunc
}
