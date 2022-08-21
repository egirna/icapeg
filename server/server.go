package server

import (
	"fmt"
	"icapeg/utils"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"icapeg/api"
	"icapeg/config"
	"icapeg/icap"
)

// https://github.com/k8-proxy/k8-rebuild-rest-api
// StartServer starts the icap server

func StartServer() error {

	utils.InitializeLogger()

	// any request even the service doesn't exist in toml file, it will go to api.ToICAPEGServe
	// and there, the request will be filtered to check if the service exists or not

	config.Init()

	icap.HandleFunc("/", api.ToICAPEGServe)
	//http.HandleFunc("/", api.ErrorPageHanlder)

	log.Println("starting the ICAP server")

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

	log.Printf("ICAP server is running on localhost: %d", config.App().Port)

	<-stop
	ticker.Stop()

	log.Printf("ICAP server gracefully shut down")

	return nil
}
