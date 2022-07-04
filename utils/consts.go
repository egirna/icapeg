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
	Unknown                        = "unknown"
	Any                            = "*"
	NoModificationStatusCodeStr    = 204
	BadRequestStatusCodeStr        = 400
	OkStatusCodeStr                = 200
	InternalServerErrStatusCodeStr = 500
	Continue                       = 100
	RequestTimeOutStatusCodeStr    = 408
	HeaderEncapsulated             = "Encapsulated"
	ICAPPrefix                     = "icap_"
	NoVendor                       = "none"
	ContentLength                  = "Content-Length"
	ContentType                    = "Content-Type"
	HTMLContentType                = "text/html"
)
