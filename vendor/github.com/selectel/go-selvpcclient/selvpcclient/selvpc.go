package selvpcclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	// AppVersion is a version of the application.
	AppVersion = "1.0.0"

	// AppName is a global application name.
	AppName = "selvpcclient"

	// DefaultEndpoint contains basic endpoint for queries.
	DefaultEndpoint = "https://api.selectel.ru/vpc"

	// DefaultUserAgent contains basic user agent that will be used in queries.
	DefaultUserAgent = AppName + "/" + AppVersion
)

// ServiceClient stores details that are needed to work with different Selectel VPC APIs.
type ServiceClient struct {
	// HTTPClient represents an initialized HTTP client that will be used to do requests.
	HTTPClient *http.Client

	// Endpoint represents an endpoint that will be used in all requests.
	Endpoint string

	// TokenID is a client authentication token.
	TokenID string

	// UserAgent contains user agent that will be used in all requests.
	UserAgent string
}

// ResponseResult represents a result of a HTTP request.
// It embeddes standard http.Response and adds a custom error description.
type ResponseResult struct {
	*http.Response

	// Err contains error that can be provided to a caller.
	Err error
}

// DoRequest performs the HTTP request with the current ServiceClient's HTTPClient.
// Authentication and optional headers will be automatically added.
func (client *ServiceClient) DoRequest(ctx context.Context, method, url string, body io.Reader) (*ResponseResult, error) {
	// Prepare a HTTP request with the provided context.
	request, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	request.Header.Set("User-Agent", client.UserAgent)
	request.Header.Set("X-token", client.TokenID)
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	request = request.WithContext(ctx)

	// Send HTTP request and populate the ResponseResult.
	response, err := client.HTTPClient.Do(request)
	if err != nil {
		return nil, err
	}
	responseResult := &ResponseResult{
		response,
		nil,
	}
	if response.StatusCode >= 400 && response.StatusCode <= 599 {
		err = fmt.Errorf("selvpcclient: got the %d error status code from the server", response.StatusCode)
		responseResult.Err = err
	}

	return responseResult, nil
}

// ExtractResult allows to provide an object into which ResponseResult body will be extracted.
func (result *ResponseResult) ExtractResult(to interface{}) error {
	body, err := ioutil.ReadAll(result.Body)
	defer result.Body.Close()
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, to)
	return err
}

// RFC3339NoZ describes a timestamp format used by some SelVPC responses.
const RFC3339NoZ = "2006-01-02T15:04:05"

// JSONRFC3339NoZTimezone is a type for timestamps SelVPC responses with the RFC3339NoZ format.
type JSONRFC3339NoZTimezone time.Time

// UnmarshalJSON helps to unmarshal timestamps from SelVPC responses to the
// JSONRFC3339NoZTimezone type.
func (jt *JSONRFC3339NoZTimezone) UnmarshalJSON(data []byte) error {
	b := bytes.NewBuffer(data)
	dec := json.NewDecoder(b)
	var s string
	if err := dec.Decode(&s); err != nil {
		return err
	}
	if s == "" {
		return nil
	}
	t, err := time.Parse(RFC3339NoZ, s)
	if err != nil {
		return err
	}
	*jt = JSONRFC3339NoZTimezone(t)
	return nil
}

const (
	// IPv4 represents IP version 4.
	IPv4 IPVersion = "ipv4"

	// IPv6 represents IP version 6.
	IPv6 IPVersion = "ipv6"
)

// IPVersion represents a type for the IP versions of the different Selectel VPC APIs.
type IPVersion string
