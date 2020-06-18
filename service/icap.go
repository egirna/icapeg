package service

import (
	"fmt"
	"icapeg/config"
	"net/http"
	"time"

	ic "github.com/egirna/icap-client"
)

func init() {
	ic.SetDebugMode(config.App().Debug)
}

// RemoteICAPService represents the remote icap service informations
type RemoteICAPService struct {
	URL           string
	Endpoint      string
	HTTPRequest   *http.Request
	HTTPResponse  *http.Response
	RequestHeader http.Header
	Timeout       time.Duration
}

// RemoteICAPReqmod calls the remote icap server using REQMOD method
func RemoteICAPReqmod(rs RemoteICAPService) (*ic.Response, error) {

	urlStr := rs.URL + rs.Endpoint

	req, err := ic.NewRequest(ic.MethodREQMOD, urlStr, rs.HTTPRequest, nil)

	if err != nil {
		return nil, err
	}

	req.ExtendHeader(rs.RequestHeader)

	client := &ic.Client{
		Timeout: rs.Timeout,
	}

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	return resp, nil
}

// RemoteICAPRespmod calls the remote icap server using RESPMOD method
func RemoteICAPRespmod(rs RemoteICAPService) (*ic.Response, error) {

	urlStr := rs.URL + rs.Endpoint

	req, err := ic.NewRequest(ic.MethodRESPMOD, urlStr, rs.HTTPRequest, rs.HTTPResponse)

	if err != nil {
		return nil, err
	}

	req.ExtendHeader(rs.RequestHeader)

	client := &ic.Client{
		Timeout: rs.Timeout,
	}

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	return resp, nil
}

// RemoteICAPOptions calls the remote icap server using OPTIONS method
func RemoteICAPOptions(rs RemoteICAPService) (*ic.Response, error) {

	urlStr := rs.URL + rs.Endpoint
	req, err := ic.NewRequest(ic.MethodOPTIONS, urlStr, rs.HTTPRequest, rs.HTTPResponse)

	if err != nil {
		return nil, err
	}

	req.ExtendHeader(rs.RequestHeader)

	client := &ic.Client{
		Timeout: rs.Timeout,
	}

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Invalid status(%d) received from server", resp.StatusCode)
	}

	return resp, nil
}
