package examples

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	ic "github.com/egirna/icap-client"
)

func reqmodInDebug() {

	/* setting a text file to dump my icap-client Debug logs into */
	f, _ := os.OpenFile("logs.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	ic.SetDebugMode(true)
	ic.SetDebugOutput(f)

	/* making the http request required for the REQMOD */
	httpReq, err := http.NewRequest(http.MethodGet, "http://localhost:8000/sample.pdf", nil)

	if err != nil {
		log.Fatal(err)
	}

	/* making the icap client & the icap request that'll be made by the client */
	client := &ic.Client{
		Timeout: 500000 * time.Second,
	}

	req, err := ic.NewRequest(ic.MethodREQMOD, "icap://127.0.0.1:1344/reqmod", httpReq, nil)

	if err != nil {
		log.Fatal(err)
	}

	/* making the request call & getting the response */
	resp, err := client.Do(req)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(resp.Status)

}
