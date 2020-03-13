package server

import (
	"fmt"
	"icapeg/api"
	"icapeg/config"
	"log"

	"github.com/egirna/icap"
)

// StartServer starts the icap server
func StartServer() error {

	//icap.HandleFunc("/reqmod-icapeg", toICAPEGReq)
	//icap.ListenAndServe(":1344", icap.HandlerFunc(toICAPEGReq))

	icap.HandleFunc("/respmod-icapeg", api.ToICAPEGResp)

	log.Println("Starting the ICAP server...")
	return icap.ListenAndServe(fmt.Sprintf(":%d", config.App().Port), icap.HandlerFunc(api.ToICAPEGResp))
}
