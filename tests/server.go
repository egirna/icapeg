package tests

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"icapeg/api"
	"icapeg/config"
	"icapeg/logger"

	"icapeg/icap"

	"github.com/rs/zerolog"
	zLog "github.com/rs/zerolog/log"
)

const (
	badFileURL  = "http://www.eicar.org/download/eicar.com"
	goodFileURL = "https://file-examples.com/wp-content/uploads/2017/10/file-example_PDF_1MB.pdf"
	// goodFileURL = "http://localhost:8000/sample.pdf"
)

// startTestServer starts a test server
func startTestServer(stop chan os.Signal) error {

	lr := zLog.Logger
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zLogger := &logger.ZLogger{Logger: lr}
	withICAPLogger := logger.LoggingHandlerICAPFactory(zLogger)
	withHTTPLogger := logger.LoggingHandlerHTTPFactory(zLogger)

	icap.Handle("/", withICAPLogger(api.ToICAPEGServe))
	http.Handle("/", withHTTPLogger(api.ErrorPageHandler))

	log.Println("Starting the ICAP server...")

	signal.Notify(stop, syscall.SIGKILL, syscall.SIGINT, syscall.SIGQUIT)

	go func() {
		if err := icap.ListenAndServe(fmt.Sprintf(":%d", config.App().Port), nil); err != nil {
			log.Fatal(err.Error())
		}
	}()

	time.Sleep(20 * time.Millisecond)

	log.Printf("ICAP server is running on localhost:%d ...\n", config.App().Port)

	<-stop

	log.Println("ICAP server gracefully shut down")

	return nil
}

func stopTestServer(stop chan os.Signal) {
	stop <- syscall.SIGKILL
}
