package logger

import (
	"crypto/tls"
	"net/http"

	"icapeg/utils"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type LoggingServer struct {
	client HTTPClient
}

func (l *LoggingServer) Do(req *http.Request) (*http.Response, error) {
	return l.client.Do(req)
}

func NewLoggerClient() *LoggingServer {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: utils.InitSecure()},
	}
	return &LoggingServer{
		client: &http.Client{Transport: tr},
	}
}
