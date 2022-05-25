package glasswall

import (
	"encoding/json"
	"icapeg/readValues"
	general_functions "icapeg/service/services-utilities/general-functions"
	"icapeg/utils"
	"sync"
	"time"
)

var doOnce sync.Once
var glasswallConfig *Glasswall

type AuthTokens struct {
	Tokens []Tokens `json:"gw-auth-tokens"`
}

type Tokens struct {
	Id           string `json:"id"`
	Role         string `json:"role"`
	Enabled      bool   `json:"enabled"`
	CreationDate int64  `json:"creation_date"`
	ExpiryDate   int64  `json:"expiry_date"`
}

// Glasswall represents the information regarding the Glasswall service
type Glasswall struct {
	httpMsg                           *utils.HttpMsg
	elapsed                           time.Duration
	serviceName                       string
	methodName                        string
	maxFileSize                       int
	bypassExts                        []string
	processExts                       []string
	BaseURL                           string
	Timeout                           time.Duration
	APIKey                            string
	ScanEndpoint                      string
	ReportEndpoint                    string
	FailThreshold                     int
	statusCheckInterval               time.Duration
	statusCheckTimeout                time.Duration
	badFileStatus                     []string
	okFileStatus                      []string
	statusEndPointExists              bool
	policy                            string
	returnOrigIfMaxSizeExc            bool
	returnOrigIfUnprocessableFileType bool
	returnOrigIf400                   bool
	authID                            string
	generalFunc                       *general_functions.GeneralFunc
}

func InitGlasswallConfig(serviceName string) {
	doOnce.Do(func() {
		glasswallConfig = &Glasswall{
			maxFileSize:                       readValues.ReadValuesInt(serviceName + ".max_filesize"),
			bypassExts:                        readValues.ReadValuesSlice(serviceName + ".bypass_extensions"),
			processExts:                       readValues.ReadValuesSlice(serviceName + ".process_extensions"),
			BaseURL:                           readValues.ReadValuesString(serviceName + ".base_url"),
			Timeout:                           readValues.ReadValuesDuration(serviceName + ".timeout"),
			APIKey:                            readValues.ReadValuesString(serviceName + ".api_key"),
			ScanEndpoint:                      readValues.ReadValuesString(serviceName + ".scan_endpoint"),
			FailThreshold:                     readValues.ReadValuesInt(serviceName + ".fail_threshold"),
			policy:                            readValues.ReadValuesString(serviceName + ".policy"),
			returnOrigIfMaxSizeExc:            readValues.ReadValuesBool(serviceName + ".return_original_if_max_file_size_exceeded"),
			returnOrigIfUnprocessableFileType: readValues.ReadValuesBool(serviceName + ".return_original_if_unprocessable_file_type"),
			returnOrigIf400:                   readValues.ReadValuesBool(serviceName + ".return_original_if_400_response"),
		}
	})
}

// NewGlasswallService returns a new populated instance of the Glasswall service
func NewGlasswallService(serviceName, methodName string, httpMsg *utils.HttpMsg) *Glasswall {
	gw := &Glasswall{
		httpMsg:                           httpMsg,
		serviceName:                       serviceName,
		methodName:                        methodName,
		maxFileSize:                       glasswallConfig.maxFileSize,
		bypassExts:                        glasswallConfig.bypassExts,
		processExts:                       glasswallConfig.processExts,
		BaseURL:                           glasswallConfig.BaseURL,
		Timeout:                           glasswallConfig.Timeout * time.Second,
		APIKey:                            glasswallConfig.APIKey,
		ScanEndpoint:                      glasswallConfig.ScanEndpoint,
		ReportEndpoint:                    "/",
		FailThreshold:                     glasswallConfig.FailThreshold,
		statusCheckInterval:               2 * time.Second,
		policy:                            glasswallConfig.policy,
		returnOrigIfMaxSizeExc:            glasswallConfig.returnOrigIfMaxSizeExc,
		returnOrigIfUnprocessableFileType: glasswallConfig.returnOrigIfUnprocessableFileType,
		returnOrigIf400:                   glasswallConfig.returnOrigIf400,
		generalFunc:                       general_functions.NewGeneralFunc(httpMsg),
	}
	authTokens := new(AuthTokens)
	err := json.Unmarshal([]byte(gw.APIKey), authTokens)
	if err != nil {
		gw.authID = ""
		return gw
	}
	for _, token := range authTokens.Tokens {
		if token.Role == "file_operations" {
			gw.authID = token.Id
		}
	}
	return gw
}
