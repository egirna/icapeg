package server

import (
	"fmt"
	"icapeg/api"
	"icapeg/config"
	"icapeg/icap"
	"icapeg/logger"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

//https://github.com/k8-proxy/k8-rebuild-rest-api
// StartServer starts the icap server

func StartServer() error {

	//any request even the service doesn't exist in toml file, it will go to api.ToICAPEGServe
	//and there, the request will be filtered to check if the service exists or not
	config.Init()
	icap.HandleFunc("/", api.ToICAPEGServe)

	http.HandleFunc("/", api.ErrorPageHanlder)
	logger.SetLogLevel(config.App().LogLevel)
	logr := logger.NewLogger()

	if err := logger.SetLogFile("logs.txt"); err != nil {
		logr.LogToScreen("Failed to prepare log file: ", err.Error())
	} else {
		defer logger.LogFile().Close()
	}

	if config.App().LogLevel == logger.LogLevelDebug {
		logr.LogToAll("Starting the ICAP server in DEBUG mode...")
	} else {
		logr.LogToAll("Starting the ICAP server...")
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGKILL, syscall.SIGINT, syscall.SIGQUIT)

	go func() {
		if err := icap.ListenAndServe(fmt.Sprintf(":%d", config.App().Port), nil); err != nil {
			logger.LogFatalToScreen(err.Error())
		}
	}()

	time.Sleep(5 * time.Millisecond)

	logr.LogfToAll("ICAP server is running on localhost:%d ...\n", config.App().Port)

	<-stop

	logr.LogToAll("ICAP server gracefully shut down")

	return nil
}
