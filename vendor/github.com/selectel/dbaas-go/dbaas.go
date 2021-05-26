// Package dbaas implements the Selectel DBaaS v1 API
package dbaas

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
)

const (
	// appName specifies an application name.
	appName = "dbaas-go"

	// appVersion specifies an application version.
	appVersion = "0.1.0"

	// userAgent contains a basic user agent that will be used in queries.
	userAgent = appName + "/" + appVersion
)

// Status represents custom type for various DBaaS objects statuses.
type Status string

const (
	StatusActive        Status = "ACTIVE"
	StatusDeleted       Status = "DELETED"
	StatusDegraded      Status = "DEGRADED"
	StatusDiskFull      Status = "DISK_FULL"
	StatusError         Status = "ERROR"
	StatusPendingCreate Status = "PENDING_CREATE"
	StatusPendingUpdate Status = "PENDING_UPDATE"
	StatusPendingDelete Status = "PENDING_DELETE"
	StatusDown          Status = "DOWN"
	StatusResizing      Status = "RESIZING"
)

// API stores details that are needed to work with Selectel DBaaS API.
type API struct {
	HTTPClient *http.Client
	Token      string
	Endpoint   string
	UserAgent  string
}

// NewDBAASClient initializes a new DBaaS client for the V1 API.
func NewDBAASClient(token, endpoint string) (*API, error) {
	return &API{
		HTTPClient: http.DefaultClient,
		Token:      token,
		Endpoint:   endpoint,
		UserAgent:  userAgent,
	}, nil
}

// NewDBAASClientV1WithCustomHTTP initializes a new DBaaS client for the V1 API using custom HTTP client.
// If custom HTTP client is nil - default HTTP client will be used.
func NewDBAASClientV1WithCustomHTTP(customHTTPClient *http.Client, token, endpoint string) (*API, error) {
	if customHTTPClient == nil {
		customHTTPClient = http.DefaultClient
	}
	return &API{
		HTTPClient: customHTTPClient,
		Token:      token,
		Endpoint:   endpoint,
		UserAgent:  userAgent,
	}, nil
}

// NewDBAASClientV1WithOpenstackCredentials initializes a new DBaaS client for the V1 API using openstack credentials.
// You need to provide identityEndpoint, region and serviceType to get correct service endpoint.
func NewDBAASClientV1WithOpenstackCredentials(token, identityEndpoint, region, serviceType string) (*API, error) {
	opts := gophercloud.AuthOptions{
		IdentityEndpoint: identityEndpoint,
		TokenID:          token,
	}

	provider, err := openstack.AuthenticatedClient(opts)
	if err != nil {
		return nil, fmt.Errorf("could not authenticate to openstack, %w", err)
	}
	endpointOpts := gophercloud.EndpointOpts{Region: region}
	endpointOpts.ApplyDefaults(serviceType)
	endpoint, err := provider.EndpointLocator(endpointOpts)
	if err != nil {
		return nil, fmt.Errorf("could not locate an endpoint, %w", err)
	}

	return &API{
		HTTPClient: http.DefaultClient,
		Token:      token,
		Endpoint:   endpoint,
		UserAgent:  userAgent,
	}, nil
}

// makeRequest makes a HTTP request and returns the body as a byte slice.
// Params will be serialized to JSON.
func (api *API) makeRequest(ctx context.Context, method, uri string, params interface{}) ([]byte, error) {
	jsonBody, err := handleParams(params)
	if err != nil {
		return nil, err
	}

	var resp *http.Response
	var respErr error
	var reqBody io.Reader
	var respBody []byte
	if jsonBody != nil {
		reqBody = bytes.NewReader(jsonBody)
	}

	resp, respErr = api.request(ctx, method, uri, reqBody)
	if respErr != nil || resp.StatusCode >= http.StatusInternalServerError {
		if respErr == nil {
			respBody, err = ioutil.ReadAll(resp.Body)
			resp.Body.Close()

			respErr = fmt.Errorf("could not read response body, %w", err)
			fmt.Printf("Request: %s %s got an error response %d", method, uri, resp.StatusCode)
		} else {
			fmt.Printf("Error performing request: %s %s : %s \n", method, uri, respErr.Error())
		}
	} else {
		respBody, err = ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("could not read response body, %w", err)
		}
	}
	if respErr != nil {
		return nil, respErr
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return nil, handleStatusCode(resp.StatusCode, respBody, uri)
	}

	return respBody, nil
}

// request makes a HTTP request to the given API endpoint, returning the raw
// *http.Response, or an error if one occurred.
// Authentication and optional headers will be added automatically.
func (api *API) request(ctx context.Context, method, uri string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, api.Endpoint+uri, body)
	if err != nil {
		return nil, fmt.Errorf("HTTP request creation failed, %w", err)
	}

	req.Header.Set("User-Agent", api.UserAgent)
	req.Header.Set("X-Auth-Token", api.Token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := api.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed, %w", err)
	}

	return resp, nil
}

// handleParams converts body params to slice of bytes if they are not nil.
func handleParams(params interface{}) ([]byte, error) {
	var jsonBody []byte
	var err error

	if params == nil {
		return nil, nil
	}

	if paramBytes, ok := params.([]byte); ok {
		jsonBody = paramBytes
	} else {
		jsonBody, err = json.Marshal(params)
		if err != nil {
			return nil, fmt.Errorf("error marshalling params to JSON, %w", err)
		}
	}

	return jsonBody, nil
}

// handleStatusCode checks status code and returns corresponding error.
func handleStatusCode(statusCode int, body []byte, uri string) error {
	if statusCode >= http.StatusInternalServerError {
		return fmt.Errorf("http status %d: service failed.\n%v\n%v", statusCode, body, uri) //nolint
	}

	errBody := &DBaaSAPIError{}
	err := json.Unmarshal(body, &errBody)
	if err != nil {
		return fmt.Errorf("can't unmarshal response:\n%s, %w", body, err)
	}
	return errBody
}

// setQueryParams updates uri string with query parameters.
func setQueryParams(uri string, params interface{}) (string, error) {
	v := url.Values{}

	var queryParams map[string]interface{}
	jsonParams, err := json.Marshal(params)
	if err != nil {
		return "", fmt.Errorf("Error marshalling params to JSON, %w", err)
	}

	err = json.Unmarshal(jsonParams, &queryParams)
	if err != nil {
		return "", fmt.Errorf("Error during Unmarshal, %w", err)
	}
	for key, value := range queryParams {
		v.Set(key, fmt.Sprintf("%v", value))
	}

	if len(v) > 0 {
		uri = uri + "?" + v.Encode()
	}

	return uri, nil
}
