package main

import (
	"icapeg/server"
	"log"
)

func main() {

	if err := server.StartServer(); err != nil {
		log.Fatal(err.Error())
	}
}
