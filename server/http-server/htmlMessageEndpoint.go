package http_server

import (
	"bytes"
	"encoding/json"
	"html/template"
	"net/http"

	general_functions "github.com/egirna/icapeg/service/services-utilities/general-functions"
	utils "github.com/egirna/icapeg/utils"
)

func HtmlMessage(w http.ResponseWriter, r *http.Request) {

	htmlTmpl, _ := template.ParseFiles(utils.BlockPagePath)
	htmlErrPage := &bytes.Buffer{}
	var errPageStruct general_functions.ErrorPage
	_ = json.NewDecoder(r.Body).Decode(&errPageStruct)
	if errPageStruct.ExceptionPage != "" {
		tmpl, err := template.ParseFiles(utils.BlockPagePath)
		if err == nil {
			htmlTmpl = tmpl
		}
	}
	htmlTmpl.Execute(htmlErrPage, &errPageStruct)
	w.Write(htmlErrPage.Bytes())
}
