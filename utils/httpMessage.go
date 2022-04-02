package utils

import "net/http"

type HttpMsg struct {
	Request  *http.Request
	Response *http.Response
}

func (h *HttpMsg) NewHttpMsg(Request *http.Request, Response *http.Response) *HttpMsg {
	return &HttpMsg{
		Request:  Request,
		Response: Response,
	}
}
