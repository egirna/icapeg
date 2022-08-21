package server

import (
	"fmt"
	"icapeg/logging"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"icapeg/api"
	"icapeg/config"
	"icapeg/icap"
)

// StartServer starts the icap server

func StartServer() error {

	// any request even the service doesn't exist in toml file, it will go to api.ToICAPEGServe
	// and there, the request will be filtered to check if the service exists or not

	config.Init()
	utils.InitializeLogger()

	icap.HandleFunc("/", api.ToICAPEGServe)
	//http.HandleFunc("/", api.ErrorPageHanlder)

	utils.Logger.Info("starting the ICAP server")

	stop := make(chan os.Signal, 1)

	signal.Notify(stop, syscall.SIGKILL, syscall.SIGINT, syscall.SIGQUIT)

	go func() {
		if err := icap.ListenAndServe(fmt.Sprintf(":%d", config.App().Port), nil); err != nil {
			log.Println(err)
			log.Fatal(err)
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
	utils.Logger.Info("ICAP server is running on localhost: " + strconv.Itoa(config.App().Port))

	<-stop
	ticker.Stop()

	utils.Logger.Info("ICAP server gracefully shut down")

	return nil
}
