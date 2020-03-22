package server

import (
	"fmt"
	"icapeg/api"
	"icapeg/config"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/egirna/icap"
)

// StartServer starts the icap server
func StartServer() error {

	//icap.HandleFunc("/reqmod-icapeg", toICAPEGReq)
	//icap.ListenAndServe(":1344", icap.HandlerFunc(toICAPEGReq))

	icap.HandleFunc("/respmod-icapeg", api.ToICAPEGResp)

	log.Println("Starting the ICAP server...")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGKILL, syscall.SIGINT, syscall.SIGQUIT)

	go func() {
		if err := icap.ListenAndServe(fmt.Sprintf(":%d", config.App().Port), icap.HandlerFunc(api.ToICAPEGResp)); err != nil {
			log.Fatal(err.Error())
		}
	}()
	time.Sleep(5 * time.Millisecond)

	log.Printf("ICAP server is running on localhost:%d ...\n", config.App().Port)

	<-stop

	log.Println("ICAP server gracefully shut down")

	return nil
}
