package service

import (
	"fmt"
	"net/http"
	"time"

	ic "github.com/egirna/icap-client"
	"github.com/spf13/viper"
)

func init() {
	ic.SetDebugMode(false)
}

// RemoteICAPService represents the remote icap service informations
type RemoteICAPService struct {
	url             string
	respmodEndpoint string
	reqmodEndpoint  string
	optionsEndpoint string
	httpRequest     *http.Request
	httpResponse    *http.Response
	requestHeader   http.Header
	timeout         time.Duration
}

// NewRemoteICAPService populates and returns a new RemoteICAPService instance
func NewRemoteICAPService(name string) *RemoteICAPService {
	return &RemoteICAPService{
		url:             viper.GetString(fmt.Sprintf("%s.base_url", name)),
		respmodEndpoint: viper.GetString(fmt.Sprintf("%s.respmod_endpoint", name)),
		reqmodEndpoint:  viper.GetString(fmt.Sprintf("%s.reqmod_endpoint", name)),
		optionsEndpoint: viper.GetString(fmt.Sprintf("%s.options_endpoint", name)),
		timeout:         viper.GetDuration(fmt.Sprintf("%s.timeout", name)) * time.Second,
	}
}

// DoReqmod calls the remote icap server using REQMOD method
func (r *RemoteICAPService) DoReqmod() (*ic.Response, error) {

	urlStr := r.url + r.reqmodEndpoint

	req, err := ic.NewRequest(ic.MethodREQMOD, urlStr, r.httpRequest, nil)

	if err != nil {
		return nil, err
	}

	req.ExtendHeader(r.requestHeader)

	client := &ic.Client{
		Timeout: r.timeout,
	}

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	return resp, nil
}

// DoRespmod calls the remote icap server using RESPMOD method
func (r *RemoteICAPService) DoRespmod() (*ic.Response, error) {

	urlStr := r.url + r.respmodEndpoint

	req, err := ic.NewRequest(ic.MethodRESPMOD, urlStr, r.httpRequest, r.httpResponse)

	if err != nil {
		return nil, err
	}

	req.ExtendHeader(r.requestHeader)

	client := &ic.Client{
		Timeout: r.timeout,
	}

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	return resp, nil
}

// DoOptions calls the remote icap server using OPTIONS method
func (r *RemoteICAPService) DoOptions() (*ic.Response, error) {

	urlStr := r.url + r.optionsEndpoint

	req, err := ic.NewRequest(ic.MethodOPTIONS, urlStr, r.httpRequest, r.httpResponse)

	if err != nil {
		return nil, err
	}

	req.ExtendHeader(r.requestHeader)

	client := &ic.Client{
		Timeout: r.timeout,
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

// GetURL returns the base url of the remote icap service
func (r *RemoteICAPService) GetURL() string {
	return r.url
}

// GetRespmodEndpoint returns the respmod endpoint of the remote icap service
func (r *RemoteICAPService) GetRespmodEndpoint() string {
	return r.respmodEndpoint
}

// GetReqmodEndpoint returns the reqmod endpoint of the remote icap service
func (r *RemoteICAPService) GetReqmodEndpoint() string {
	return r.reqmodEndpoint
}

// GetOptionsEndpoint returns the options endpoint of the remote icap service
func (r *RemoteICAPService) GetOptionsEndpoint() string {
	return r.optionsEndpoint
}

// GetTimeout returns the timeout of the remote icap service
func (r *RemoteICAPService) GetTimeout() time.Duration {
	return r.timeout
}

// SetHTTPRequest sets the http request of the remote icap service
func (r *RemoteICAPService) SetHTTPRequest(req *http.Request) {
	r.httpRequest = req
}

// SetHTTPResponse sets the http response of the remote icap service
func (r *RemoteICAPService) SetHTTPResponse(resp *http.Response) {
	r.httpResponse = resp
}

// SetHeader sets the request header of the remote icap service
func (r *RemoteICAPService) SetHeader(hdr http.Header) {
	r.requestHeader = hdr
}

// ChangeOptionsEndpoint changes the options endpoint with the given one
func (r *RemoteICAPService) ChangeOptionsEndpoint(endpoint string) {
	r.optionsEndpoint = endpoint
}
