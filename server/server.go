package server

import (
	"fmt"
	"icapeg/api"
	"icapeg/config"
	"icapeg/logger"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/egirna/icap"
)

// StartServer starts the icap server
func StartServer() error {

	config.Init()

	icap.HandleFunc("/respmod-icapeg", api.ToICAPEGResp)
	icap.HandleFunc("/reqmod-icapeg", api.ToICAPEGReq)

	http.HandleFunc("/", api.ErrorPageHanlder)

	if err := logger.SetLogFile("logs.txt"); err != nil {
		logger.LogToScreen("Failed to prepare log file: ", err.Error())
	} else {
		defer logger.LogFile().Close()
	}

	if config.App().Debug {
		logger.LogToAll("Starting the ICAP server in DEBUG MODE...")
	} else {
		logger.LogToAll("Starting the ICAP server...")
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGKILL, syscall.SIGINT, syscall.SIGQUIT)

	go func() {
		if err := icap.ListenAndServe(fmt.Sprintf(":%d", config.App().Port), nil); err != nil {
			logger.LogFatalToScreen(err.Error())
		}
	}()

	time.Sleep(5 * time.Millisecond)

	logger.LogfToAll("ICAP server is running on localhost:%d ...\n", config.App().Port)

	<-stop

	logger.LogToAll("ICAP server gracefully shut down")

	return nil
}
