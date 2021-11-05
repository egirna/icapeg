package examples

import (
	"fmt"
	"log"
	"net/http"
	"time"

	ic "github.com/egirna/icap-client"
)

func makeReqmodCall() {

	/* preparing the http request required for the REQMOD */
	httpReq, err := http.NewRequest(http.MethodGet, "http://localhost:8000/sample.pdf", nil)

	if err != nil {
		log.Fatal(err)
	}

	/* making a icap request with OPTIONS method */
	optReq, err := ic.NewRequest(ic.MethodOPTIONS, "icap://127.0.0.1:1344/reqmod", nil, nil)

	if err != nil {
		log.Fatal(err)
		return
	}

	/* making the icap client responsible for making the requests */
	client := &ic.Client{
		Timeout: 5 * time.Second,
	}

	/* making the OPTIONS request call */
	optResp, err := client.Do(optReq)

	if err != nil {
		log.Fatal(err)
		return
	}

	/* making a icap request with REQMOD method */
	req, err := ic.NewRequest(ic.MethodREQMOD, "icap://127.0.0.1:1344/reqmod", httpReq, nil)

	if err != nil {
		log.Fatal(err)
	}

	req.SetPreview(optResp.PreviewBytes) // setting the preview bytes obtained from the OPTIONS call

	/* making the REQMOD request call */
	resp, err := client.Do(req)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(resp.Status)

}
