package server

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/egirna/icapeg/logging"
	http_server "github.com/egirna/icapeg/server/http-server"

	"github.com/egirna/icapeg/api"
	"github.com/egirna/icapeg/config"
	"github.com/egirna/icapeg/icap"
)

// https://github.com/k8-proxy/k8-rebuild-rest-api
// StartServer starts the icap server

func StartServer() error {
	// any request even the service doesn't exist in toml file, it will go to api.ToICAPEGServe
	// and there, the request will be filtered to check if the service exists or not

	config.Init()

	//HTTP server
	htmlWebServer := http.NewServeMux()
	htmlWebServer.HandleFunc("/service/message", http_server.HtmlMessage)
	go func() {
		http.ListenAndServe(":8081", htmlWebServer)
	}()

	icap.HandleFunc("/", api.ToICAPEGServe)

	logging.Logger.Info("starting the ICAP server")

	stop := make(chan os.Signal, 1)

	signal.Notify(stop, syscall.SIGKILL, syscall.SIGINT, syscall.SIGQUIT)

	go func() {
		fmt.Println("Starting the ICAP server")
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
