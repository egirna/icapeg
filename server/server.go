package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	utils "icapeg/consts"
	"icapeg/logging"
	general_functions "icapeg/service/services-utilities/general-functions"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"icapeg/api"
	"icapeg/config"
	"icapeg/icap"
)

// https://github.com/k8-proxy/k8-rebuild-rest-api
// StartServer starts the icap server

func world(w http.ResponseWriter, r *http.Request) {
	htmlTmpl, _ := template.ParseFiles(utils.BlockPagePath)
	htmlErrPage := &bytes.Buffer{}
	var errPageStruct general_functions.ErrorPage
	_ = json.NewDecoder(r.Body).Decode(&errPageStruct)
	htmlTmpl.Execute(htmlErrPage, &errPageStruct)
	w.Write(htmlErrPage.Bytes())

}
func StartServer() error {
	// any request even the service doesn't exist in toml file, it will go to api.ToICAPEGServe
	// and there, the request will be filtered to check if the service exists or not

	config.Init()
	//HTTP server
	serverMuxB := http.NewServeMux()
	serverMuxB.HandleFunc("/service/message", world)
	go func() {
		http.ListenAndServe(":8081", serverMuxB)
	}()
	icap.HandleFunc("/", api.ToICAPEGServe)

	logging.Logger.Info("starting the ICAP server")

	stop := make(chan os.Signal, 1)

	signal.Notify(stop, syscall.SIGKILL, syscall.SIGINT, syscall.SIGQUIT)

	go func() {
		if err := icap.ListenAndServe(fmt.Sprintf(":%d", config.App().Port), nil); err != nil {
			logging.Logger.Fatal(err.Error())
		}
	}()

	ticker := time.NewTicker(10 * time.Second)
	go func() {
		for {
			select {
			case _ = <-ticker.C:
			}
		}
	}()

	time.Sleep(5 * time.Millisecond)
	port := strconv.Itoa(config.App().Port)
	logging.Logger.Info("ICAP server is running on localhost: " + port)

	<-stop
	ticker.Stop()

	logging.Logger.Info("ICAP server gracefully shut down")

	return nil
}
