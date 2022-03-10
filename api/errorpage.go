package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"icapeg/dtos"
	"icapeg/logger"
	"icapeg/utils"

	zLog "github.com/rs/zerolog/log"
)

// ErrorPageHandler is the http handler for the error page
func ErrorPageHandler(w http.ResponseWriter, r *http.Request, logger *logger.ZLogger) {

	data := dtos.TemplateData{}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {

		zLog.Logger.Error().Msgf("failed to decode template data for error page handler: %s", err.Error())
		fmt.Fprint(w, "SOMETHING WENT WRONG")
		return
	}

	htmlBuf, _ := utils.GetTemplateBufferAndResponse(utils.BadFileTemplate, &data)

	w.WriteHeader(http.StatusForbidden)
	w.Header().Add("Content-Type", "text/html")
	w.Header().Add("Content-Length", strconv.Itoa(htmlBuf.Len()))

	w.Write(htmlBuf.Bytes())

}

type (
	errorPage struct {
		Reason                    string `json:"reason"`
		XAdaptationFileId         string `json:"x-adaptation-file-id"`
		XSdkEngineVersion         string `json:"x-sdk-engine-version"`
		RequestedURL              string `json:"requested_url"`
		XGlasswallCloudApiVersion string `json:"x-glasswall-cloud-api-version"`
	}
)
