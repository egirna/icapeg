package api

import (
	"icapeg/icap"
	"icapeg/logger"
)

// ToICAPEGServe is the ICAsP Request Handler for all modes and services:
func ToICAPEGServe(w icap.ResponseWriter, req *icap.Request, zlogger *logger.ZLogger) {

	ICAPRequest := NewICAPRequest(w, req, zlogger)

	err := ICAPRequest.RequestInitialization()
	if err != nil {
		return
	}
	ICAPRequest.RequestProcessing()
}
