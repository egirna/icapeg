package services_utilities

type Extension struct {
	Name string
	Exts []string
}

func InitExtsArr(processExts, rejectExts, bypassExts []string) []Extension {
	process := Extension{Name: "process", Exts: processExts}
	reject := Extension{Name: "reject", Exts: rejectExts}
	bypass := Extension{Name: "bypass", Exts: bypassExts}
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