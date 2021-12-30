package readValues

import (
	str2duration "github.com/xhit/go-str2duration/v2"
	"os"
	"strconv"
	"strings"
	"time"
)

//ReadIntFromEnv is used to get int value from env vars
func ReadIntFromEnv(varName string) int {
	result, _ := strconv.Atoi(os.Getenv(varName))
	return result
}

//ReadStringFromEnv is used to get string value from env vars
func ReadStringFromEnv(varName string) string {
	return os.Getenv(varName)
}

//ReadBoolFromEnv is used to get bool value from env vars
func ReadBoolFromEnv(varName string) bool {
	result, _ := strconv.ParseBool(os.Getenv(varName))
	return result
}

//ReadDurationFromEnv is used to get time.Duration value from env vars
func ReadDurationFromEnv(varName string) time.Duration {
	result, _ := str2duration.ParseDuration(os.Getenv(varName))
	return result
}

//ReadSliceFromEnv is used to get string slice value from env vars
func ReadSliceFromEnv(varName string) []string {
	result:= os.Getenv(varName)
	result=strings.ReplaceAll(result, " ", "")
	result=strings.ReplaceAll(result, "\"", "")
	arr := strings.Split(result, ",")
	arr[0]=arr[0][1:len(arr[0])]
	temp:=arr[len(arr)-1]
	temp=temp[0:len(temp)-1]
	arr[len(arr)-1] = temp
	return arr
}