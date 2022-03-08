package logger

import (
	"net/http"

	"icapeg/icap"
)

// LoggingHTTPHandlerFunc : type that looks a lot like http.HandlerFunc
// but which also accepts an argument that implements our logger interface
type LoggingHTTPHandlerFunc = func(w http.ResponseWriter, r *http.Request, l *ZLogger)

// LoggingICAPHandlerFunc : type that looks a lot like icap.HandlerFunc
// but which also accepts an argument that implements our logger interface
type LoggingICAPHandlerFunc = func(w icap.ResponseWriter, r *icap.Request, l *ZLogger)

// LoggingHTTPHandler : implements http.Handler
type LoggingHTTPHandler struct {
	logger      *ZLogger
	HandlerFunc LoggingHTTPHandlerFunc
}

// LoggingICAPHandler : implements icap.Handler
type LoggingICAPHandler struct {
	logger      *ZLogger
	HandlerFunc LoggingICAPHandlerFunc
}

func (lh *LoggingICAPHandler) ServeICAP(w icap.ResponseWriter, r *icap.Request) {
	lh.HandlerFunc(w, r, lh.logger)
}

func (lh *LoggingHTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	lh.HandlerFunc(w, r, lh.logger)
}

// LoggingHandlerHTTPFactory : takes in a logger and returns a closure that,
// when called with a LoggingHTTPHandlerFunc, returns a new *LoggingHTTPHandler with the correct logger baked in.
func LoggingHandlerHTTPFactory(l *ZLogger) func(LoggingHTTPHandlerFunc) *LoggingHTTPHandler {
	return func(hf LoggingHTTPHandlerFunc) *LoggingHTTPHandler {
		return &LoggingHTTPHandler{l, hf}
	}
}

// LoggingHandlerICAPFactory : takes in a logger and returns a closure that,
// when called with a LoggingICAPHandlerFunc, returns a new *LoggingICAPHandler with the correct logger baked in.
func LoggingHandlerICAPFactory(l *ZLogger) func(LoggingICAPHandlerFunc) *LoggingICAPHandler {
	return func(hf LoggingICAPHandlerFunc) *LoggingICAPHandler {
		return &LoggingICAPHandler{l, hf}
	}
}
