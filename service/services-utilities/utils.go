package services_utilities

import "icapeg/utils"

// Extension struct is used for storing the name of the extension array (bypass, reject, process)
// in Name field and the content of the array (for example ["pdf", "zip", "com"])
// in Exts field
type Extension struct {
	Name string
	Exts []string
}

// InitExtsArr function helps in preparing the order of checking extensions array
// it returns array of Extension type with size 3
// it stores the Extensions arrays in a specific order
// the array which has just an asterisk is stored as the last element in the array
// that's because we should check it at the end because asterisk means everything except the ones in bypass
// but the first element and the second element doesn't matter if they don't have an asterisk
// let's have an example
// suppose that the extensions arrays are: bypass_extensions = ["*"], reject_extensions = ["docx"] and
// process_extensions = ["pdf", "zip", "com"]
// there are two valid order of the Extension array which will be returned by InitExtsArr function
// First valid order:
// {{Name: "reject", Exts: ["docx"]}, {Name: "process", Exts: ["pdf", "zip", "com"]}, {Name: "bypass", Exts: ["*"]}}
// so when the service want to check about if the extension of the current HTTP body is process, bypass or reject
// it will search first in reject array, then in process array, then in bypass array
// Second valid order:
// {{Name: "process", Exts: ["pdf", "zip", "com"]}, {Name: "reject", Exts: ["docx"]}, {Name: "bypass", Exts: ["*"]}}
// so when the service want to check about if the extension of the current HTTP body is process, bypass or reject
// it will search first in process array, then in reject array, then in bypass array
func InitExtsArr(processExts, rejectExts, bypassExts []string) []Extension {
	process := Extension{Name: utils.ProcessExts, Exts: processExts}
	reject := Extension{Name: utils.RejectExts, Exts: rejectExts}
	bypass := Extension{Name: utils.BypassExts, Exts: bypassExts}
	extArrs := make([]Extension, 3)
	ind := 0
	if len(process.Exts) == 1 && process.Exts[0] == "*" {
		extArrs[2] = process
	} else {
		extArrs[ind] = process
		ind++
	}
	if len(reject.Exts) == 1 && reject.Exts[0] == "*" {
		extArrs[2] = reject
	} else {
		extArrs[ind] = reject
		ind++
	}
	if len(bypass.Exts) == 1 && bypass.Exts[0] == "*" {
		extArrs[2] = bypass
	} else {
		extArrs[ind] = bypass
		ind++
	}
	return extArrs
}
