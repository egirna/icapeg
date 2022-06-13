package api

import (
	"fmt"
	"icapeg/icap"
)

// ToICAPEGServe is the ICAsP Request Handler for all modes and services:
func ToICAPEGServe(w icap.ResponseWriter, req *icap.Request) {
	fmt.Println("inside enpoint")

	//Creating new instance from struct IcapRequest yo handle upcoming ICAP requests
	ICAPRequest := NewICAPRequest(w, req)

	//calling RequestInitialization to retrieve the important information from the ICAP request
	//and initialize the ICAP response
	err := ICAPRequest.RequestInitialization()
	if err != nil {
		return
	}
	// after initialization we call RequestProcessing func to process the ICAP request with a service
	ICAPRequest.RequestProcessing()
}
