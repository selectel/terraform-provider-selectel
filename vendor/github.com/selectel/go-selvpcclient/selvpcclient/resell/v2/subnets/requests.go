package subnets

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/selectel/go-selvpcclient/selvpcclient"
)

const resourceURL = "subnets"

// Get returns a single subnet by its id.
func Get(ctx context.Context, client *selvpcclient.ServiceClient, id string) (*Subnet, *selvpcclient.ResponseResult, error) {
	url := strings.Join([]string{client.Endpoint, resourceURL, id}, "/")
	responseResult, err := client.DoRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, err
	}
	if responseResult.Err != nil {
		return nil, responseResult, responseResult.Err
	}

	// Extract a subnet from the response body.
	var result struct {
		Subnet *Subnet `json:"subnet"`
	}
	err = responseResult.ExtractResult(&result)
	if err != nil {
		return nil, responseResult, err
	}

	return result.Subnet, responseResult, nil
}

// List gets a list of subnets in the current domain.
func List(ctx context.Context, client *selvpcclient.ServiceClient, opts ListOpts) ([]*Subnet, *selvpcclient.ResponseResult, error) {
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

	// Extract subnets from the response body.
	var result struct {
		Subnets []*Subnet `json:"subnets"`
	}
	err = responseResult.ExtractResult(&result)
	if err != nil {
		return nil, responseResult, err
	}

	return result.Subnets, responseResult, nil
}

// Create requests a creation of the subnets in the specified project.
func Create(ctx context.Context, client *selvpcclient.ServiceClient, projectID string, createOpts SubnetOpts) ([]*Subnet, *selvpcclient.ResponseResult, error) {
	createSubnetsOpts := &createOpts
	requestBody, err := json.Marshal(createSubnetsOpts)
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

	// Extract subnets from the response body.
	var result struct {
		Subnets []*Subnet `json:"subnets"`
	}
	err = responseResult.ExtractResult(&result)
	if err != nil {
		return nil, responseResult, err
	}

	return result.Subnets, responseResult, nil
}

// Delete deletes a single subnet by its id.
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
