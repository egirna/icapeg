package tests

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

const (
	badFileURL  = "http://www.eicar.org/download/eicar.com"
	goodFileURL = "https://file-examples.com/wp-content/uploads/2017/10/file-example_PDF_1MB.pdf"
	// goodFileURL = "http://localhost:8000/sample.pdf"
)

// startTestServer starts a test server
func startTestServer(stop chan os.Signal) error {

	icap.HandleFunc("/respmod-icapeg", api.ToICAPEGResp)
	icap.HandleFunc("/reqmod-icapeg", api.ToICAPEGReq)

	http.HandleFunc("/", api.ErrorPageHanlder)

	log.Println("Starting the ICAP server...")

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

func stopTestServer(stop chan os.Signal) {
	stop <- syscall.SIGKILL
}
