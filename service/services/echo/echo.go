package echo

import (
	"icapeg/utils"
)

//Processing is a func used for to processing the http message
func (e *Echo) Processing(partial bool) (int, interface{}, map[string]string) {
	// no need to scan part of the file, this service needs all the file at ine time
	if partial {
		return utils.Continue, nil, nil
	}

	return utils.NoModificationStatusCodeStr, e.httpMsg.Response, nil
}
