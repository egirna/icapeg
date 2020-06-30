package utils

// The icap protocol modes
const (
	ICAPModeResp    = "RESPMOD"
	ICAPModeOptions = "OPTIONS"
	ICAPModeReq     = "REQMOD"
)

// the common constants
const (
	ISTag                       = "\"ICAPEG\""
	Unknown                     = "unknown"
	Any                         = "*"
	NoModificationStatusCodeStr = "204"
	HeaderEncapsulated          = "Encapsulated"
	ICAPPrefix                  = "icap_"
	NoVendor                    = "none"
)
