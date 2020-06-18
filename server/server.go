package server

import (
	"fmt"
	"icapeg/api"
	"icapeg/config"
	"log"
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

	if config.App().Debug {
		log.Println("Starting the ICAP server in DEBUG MODE...")
	} else {
		log.Println("Starting the ICAP server...")
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGKILL, syscall.SIGINT, syscall.SIGQUIT)

	go func() {
		if err := icap.ListenAndServe(fmt.Sprintf(":%d", config.App().Port), nil); err != nil {
			log.Fatal(err.Error())
		}
	}()

	time.Sleep(5 * time.Millisecond)

	log.Printf("ICAP server is running on localhost:%d ...\n", config.App().Port)

	<-stop

	log.Println("ICAP server gracefully shut down")

	return nil
}
