package utils

// The icap protocol modes
const (
	ICAPModeResp    = "RESPMOD"
	ICAPModeOptions = "OPTIONS"
	ICAPModeReq     = "REQMOD"
)

// the sample severity constants
const (
	SampleSeverityOk        = "ok"
	SampleSeverityMalicious = "malicious"
)

// the common constants
const (
	ISTag                       = "\"ICAPEG\""
	Unknown                     = "unknown"
	Any                         = "*"
	NoModificationStatusCodeStr = 204
	BadRequestStatusCodeStr     = 400
	OkStatusCodeStr             = 200
	HeaderEncapsulated          = "Encapsulated"
	ICAPPrefix                  = "icap_"
	NoVendor                    = "none"
	ContentLength               = "Content-Length"
	ContentType                 = "Content-Type"
	HTMLContentType             = "text/html"
)
