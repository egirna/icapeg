package api

import (
	"encoding/json"
	"fmt"
	"icapeg/dtos"
	"icapeg/utils"
	"log"
	"net/http"
	"strconv"
)

// ErrorPageHanlder is the http handler for the error page
func ErrorPageHanlder(w http.ResponseWriter, r *http.Request) {

	data := dtos.TemplateData{}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		log.Println("Failed to decode template data for error page handler: ", err.Error())
		fmt.Fprint(w, "SOMETHING WENT WRONG")
		return
	}

	htmlBuf, _ := utils.GetTemplateBufferAndResponse(utils.BadFileTemplate, &data)

	w.WriteHeader(http.StatusForbidden)
	w.Header().Add("Content-Type", "text/html")
	w.Header().Add("Content-Length", strconv.Itoa(htmlBuf.Len()))

	w.Write(htmlBuf.Bytes())

}
