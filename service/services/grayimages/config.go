package grayimages

import (
	"icapeg/http-message"
	services_utilities "icapeg/service/services-utilities"
	general_functions "icapeg/service/services-utilities/general-functions"
	"sync"
	"time"
)

var doOnce sync.Once
var grayimagesConfig *Grayimages

const GrayimagesIdentifier = "ECHO ID"

// Echo represents the information regarding the Echo service
type Grayimages struct {
	httpMsg                    *http_message.HttpMsg
	elapsed                    time.Duration
	serviceName                string
	methodName                 string
	maxFileSize                int
	bypassExts                 []string
	processExts                []string
	rejectExts                 []string
	extArrs                    []services_utilities.Extension
	returnOrigIfMaxSizeExc     bool
	return400IfFileExtRejected bool
	generalFunc                *general_functions.GeneralFunc
}
