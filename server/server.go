package server

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"icapeg/api"
	"icapeg/config"
	"icapeg/icap"
	"icapeg/logger"

	zlog "github.com/rs/zerolog/log"
)

// https://github.com/k8-proxy/k8-rebuild-rest-api
// StartServer starts the icap server

func StartServer() error {
	// any request even the service doesn't exist in toml file, it will go to api.ToICAPEGServe
	// and there, the request will be filtered to check if the service exists or not

	config.Init()
	// initialize zerolog
	zLogger, err := logger.NewZLogger(config.App())
	if err != nil {
		return fmt.Errorf("could not start logger service %w", err)
	}

	withHTTPLogger := logger.LoggingHandlerHTTPFactory(zLogger)
	withICAPLogger := logger.LoggingHandlerICAPFactory(zLogger)

	icap.Handle("/", withICAPLogger(api.ToICAPEGServe))
	http.Handle("/", withHTTPLogger(api.ErrorPageHandler))
	zlog.Debug().Msg("starting the ICAP server")

	stop := make(chan os.Signal, 1)

	signal.Notify(stop, syscall.SIGKILL, syscall.SIGINT, syscall.SIGQUIT)

	go func() {
		if err = icap.ListenAndServe(fmt.Sprintf(":%d", config.App().Port), nil); err != nil {
			zlog.Error().Err(err)
			zlog.Fatal()
		}
	}()

	ticker := time.NewTicker(10 * time.Second)
	go func() {
		for {
			select {
			case _ = <-ticker.C:
				zLogger.FlushLogs()
			}
		}
	}()

	time.Sleep(5 * time.Millisecond)

	zlog.Debug().Msgf("ICAP server is running on localhost: %d", config.App().Port)

	<-stop
	ticker.Stop()

	zlog.Debug().Msg("ICAP server gracefully shut down")

	return nil
}
