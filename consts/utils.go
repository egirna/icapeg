package utils

import (
	"encoding/json"
	"strings"
)

func PrepareLogMsg(xICAPMetadata, msg string) string {
	logPlaceolder := make(map[string]interface{})
	logPlaceolder["X-ICAP-Metadata"] = xICAPMetadata
	logPlaceolder["log"] = msg
	jsonHeaders, _ := json.Marshal(logPlaceolder)
	final := string(jsonHeaders)
	final = strings.ReplaceAll(final, `\`, "")
	return final
}
