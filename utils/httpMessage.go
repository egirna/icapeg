package utils

import "net/http"

//HttpMsg is a struct used for encapsulating http message (http request, http response)
// to facilitate passing them together throw functions
type HttpMsg struct {
	Request  *http.Request
	Response *http.Response
}

//NewHttpMsg is a func used for creating an instance from HttpMsg struct
func (h *HttpMsg) NewHttpMsg(Request *http.Request, Response *http.Response) *HttpMsg {
	return &HttpMsg{
		Request:  Request,
		Response: Response,
	}
}
