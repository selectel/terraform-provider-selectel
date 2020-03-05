package selvpcclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode"
)

const (
	// AppVersion is a version of the application.
	AppVersion = "1.10.0"

	// AppName is a global application name.
	AppName = "go-selvpcclient"

	// DefaultEndpoint contains basic endpoint for queries.
	DefaultEndpoint = "https://api.selectel.ru/vpc"

	// DefaultUserAgent contains basic user agent that will be used in queries.
	DefaultUserAgent = AppName + "/" + AppVersion

	// defaultHTTPTimeout represents the default timeout (in seconds) for HTTP
	// requests.
	defaultHTTPTimeout = 120

	// defaultDialTimeout represents the default timeout (in seconds) for HTTP
	// connection establishments.
	defaultDialTimeout = 60

	// defaultKeepalive represents the default keep-alive period for an active
	// network connection.
	defaultKeepaliveTimeout = 60

	// defaultMaxIdleConns represents the maximum number of idle (keep-alive)
	// connections.
	defaultMaxIdleConns = 100

	// defaultIdleConnTimeout represents the maximum amount of time an idle
	// (keep-alive) connection will remain idle before closing itself.
	defaultIdleConnTimeout = 100

	// defaultTLSHandshakeTimeout represents the default timeout (in seconds)
	// for TLS handshake.
	defaultTLSHandshakeTimeout = 60

	// defaultExpectContinueTimeout represents the default amount of time to
	// wait for a server's first response headers.
	defaultExpectContinueTimeout = 1
)

// NewHTTPClient returns a reference to an initialized configured HTTP client.
func NewHTTPClient() *http.Client {
	return &http.Client{
		Timeout:   time.Second * defaultHTTPTimeout,
		Transport: newHTTPTransport(),
	}
}

// newHTTPTransport returns a reference to an initialized configured HTTP
// transport.
func newHTTPTransport() *http.Transport {
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   defaultDialTimeout * time.Second,
			KeepAlive: defaultKeepaliveTimeout * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          defaultMaxIdleConns,
		IdleConnTimeout:       defaultIdleConnTimeout * time.Second,
		TLSHandshakeTimeout:   defaultTLSHandshakeTimeout * time.Second,
		ExpectContinueTimeout: defaultExpectContinueTimeout * time.Second,
	}
}

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

// ExtractErr build a string without whitespaces from the error body.
// We don't unmarshal it into some struct because there are no strict error definition in the API.
func (result *ResponseResult) ExtractErr() (string, error) {
	body, err := ioutil.ReadAll(result.Body)
	defer result.Body.Close()
	if err != nil {
		return "", err
	}

	resp := string(body)

	var builder strings.Builder
	builder.Grow(len(resp))
	for _, ch := range resp {
		if !unicode.IsSpace(ch) {
			builder.WriteRune(ch)
		}
	}

	return builder.String(), nil
}

// DoRequest performs the HTTP request with the current ServiceClient's HTTPClient.
// Authentication and optional headers will be added automatically.
func (client *ServiceClient) DoRequest(ctx context.Context, method, path string, body io.Reader) (*ResponseResult, error) {
	// Prepare a HTTP request with the provided context.
	request, err := http.NewRequest(method, path, body)
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

	// Check status code and populate extended error message if it's possible.
	if response.StatusCode >= 400 && response.StatusCode <= 599 {
		extendedError, err := responseResult.ExtractErr()
		if err != nil {
			responseResult.Err = fmt.Errorf("selvpcclient: got the %d status code from the server", response.StatusCode)
		} else {
			responseResult.Err = fmt.Errorf("selvpcclient: got the %d status code from the server: %s", response.StatusCode, extendedError)
		}
	}

	return responseResult, nil
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

// BuildQueryParameters converts provided options struct to the string of URL parameters.
func BuildQueryParameters(opts interface{}) (string, error) {
	optsValue := reflect.ValueOf(opts)
	if optsValue.Kind() != reflect.Struct {
		return "", errors.New("provided options is not a structure")
	}
	optsType := reflect.TypeOf(opts)

	params := url.Values{}

	for i := 0; i < optsValue.NumField(); i++ {
		fieldValue := optsValue.Field(i)
		fieldType := optsType.Field(i)

		queryTag := fieldType.Tag.Get("param")
		if queryTag != "" {
			if isZero(fieldValue) {
				continue
			}

			tags := strings.Split(queryTag, ",")
		loop:
			switch fieldValue.Kind() {
			case reflect.Ptr:
				fieldValue = fieldValue.Elem()
				goto loop
			case reflect.String:
				params.Add(tags[0], fieldValue.String())
			case reflect.Int:
				params.Add(tags[0], strconv.FormatInt(fieldValue.Int(), 10))
			case reflect.Bool:
				params.Add(tags[0], strconv.FormatBool(fieldValue.Bool()))
			}
		}
	}

	return params.Encode(), nil
}

// isZero checks if provided value is zero.
func isZero(v reflect.Value) bool {
	if v.Kind() == reflect.Ptr {
		return v.IsNil()
	}
	z := reflect.Zero(v.Type())

	return v.Interface() == z.Interface()
}
