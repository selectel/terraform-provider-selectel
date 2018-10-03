package floatingips

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/selectel/go-selvpcclient/selvpcclient"
)

const resourceURL = "floatingips"

// Get returns a single floating ip by its id.
func Get(ctx context.Context, client *selvpcclient.ServiceClient, id string) (*FloatingIP, *selvpcclient.ResponseResult, error) {
	url := strings.Join([]string{client.Endpoint, resourceURL, id}, "/")
	responseResult, err := client.DoRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, err
	}
	if responseResult.Err != nil {
		return nil, responseResult, responseResult.Err
	}

	// Extract a floating ip from the response body.
	var result struct {
		FloatingIP *FloatingIP `json:"floatingip"`
	}
	err = responseResult.ExtractResult(&result)
	if err != nil {
		return nil, responseResult, err
	}

	return result.FloatingIP, responseResult, nil
}

// List gets a list of floating ips in the current domain.
func List(ctx context.Context, client *selvpcclient.ServiceClient, opts ListOpts) ([]*FloatingIP, *selvpcclient.ResponseResult, error) {
	url := strings.Join([]string{client.Endpoint, resourceURL}, "/")

	queryParams, err := selvpcclient.BuildQueryParameters(opts)
	if err != nil {
		return nil, nil, err
	}
	if queryParams != "" {
		url = strings.Join([]string{url, queryParams}, "?")
	}

	responseResult, err := client.DoRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, err
	}
	if responseResult.Err != nil {
		return nil, responseResult, responseResult.Err
	}

	// Extract floating ips from the response body.
	var result struct {
		FloatingIPs []*FloatingIP `json:"floatingips"`
	}
	err = responseResult.ExtractResult(&result)
	if err != nil {
		return nil, responseResult, err
	}

	return result.FloatingIPs, responseResult, nil
}

// Create requests a creation of the floating ip in the specified project.
func Create(ctx context.Context, client *selvpcclient.ServiceClient, projectID string, createOpts FloatingIPOpts) ([]*FloatingIP, *selvpcclient.ResponseResult, error) {
	createFloatingIPOpts := &createOpts
	requestBody, err := json.Marshal(createFloatingIPOpts)
	if err != nil {
		return nil, nil, err
	}

	url := strings.Join([]string{client.Endpoint, resourceURL, "projects", projectID}, "/")
	responseResult, err := client.DoRequest(ctx, http.MethodPost, url, bytes.NewReader(requestBody))
	if err != nil {
		return nil, nil, err
	}
	if responseResult.Err != nil {
		return nil, responseResult, responseResult.Err
	}

	// Extract floating ips from the response body.
	var result struct {
		FloatingIPs []*FloatingIP `json:"floatingips"`
	}
	err = responseResult.ExtractResult(&result)
	if err != nil {
		return nil, responseResult, err
	}

	return result.FloatingIPs, responseResult, nil
}

// Delete deletes a single floating ip by its id.
func Delete(ctx context.Context, client *selvpcclient.ServiceClient, id string) (*selvpcclient.ResponseResult, error) {
	url := strings.Join([]string{client.Endpoint, resourceURL, id}, "/")
	responseResult, err := client.DoRequest(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return nil, err
	}
	if responseResult.Err != nil {
		err = responseResult.Err
	}
	return responseResult, err
}
