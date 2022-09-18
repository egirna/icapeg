package http_server

import (
	"bytes"
	"encoding/json"
	"html/template"
	utils "icapeg/consts"
	general_functions "icapeg/service/services-utilities/general-functions"
	"net/http"
)

func HtmlMessage(w http.ResponseWriter, r *http.Request) {
	htmlTmpl, _ := template.ParseFiles(utils.BlockPagePath)
	htmlErrPage := &bytes.Buffer{}
	var errPageStruct general_functions.ErrorPage
	_ = json.NewDecoder(r.Body).Decode(&errPageStruct)
	htmlTmpl.Execute(htmlErrPage, &errPageStruct)
	w.Write(htmlErrPage.Bytes())
}
