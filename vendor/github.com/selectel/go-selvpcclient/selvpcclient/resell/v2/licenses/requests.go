package licenses

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/selectel/go-selvpcclient/selvpcclient"
)

const resourceURL = "licenses"

// Get returns a single license by its id.
func Get(ctx context.Context, client *selvpcclient.ServiceClient, id string) (*License, *selvpcclient.ResponseResult, error) {
	url := strings.Join([]string{client.Endpoint, resourceURL, id}, "/")
	responseResult, err := client.DoRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, err
	}
	if responseResult.Err != nil {
		return nil, responseResult, responseResult.Err
	}

	// Extract a license from the response body.
	var result struct {
		License *License `json:"license"`
	}
	err = responseResult.ExtractResult(&result)
	if err != nil {
		return nil, responseResult, err
	}

	return result.License, responseResult, nil
}

// List gets a list of licenses in the current domain.
func List(ctx context.Context, client *selvpcclient.ServiceClient, opts ListOpts) ([]*License, *selvpcclient.ResponseResult, error) {
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

	// Extract licenses from the response body.
	var result struct {
		Licenses []*License `json:"licenses"`
	}
	err = responseResult.ExtractResult(&result)
	if err != nil {
		return nil, responseResult, err
	}

	return result.Licenses, responseResult, nil
}

// Create requests a creation of the licenses in the specified project.
func Create(ctx context.Context, client *selvpcclient.ServiceClient, projectID string, createOpts LicenseOpts) ([]*License, *selvpcclient.ResponseResult, error) {
	createLicenseOpts := &createOpts
	requestBody, err := json.Marshal(createLicenseOpts)
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

	// Extract licenses from the response body.
	var result struct {
		Licenses []*License `json:"licenses"`
	}
	err = responseResult.ExtractResult(&result)
	if err != nil {
		return nil, responseResult, err
	}

	return result.Licenses, responseResult, nil
}

// Delete deletes a single license by its id.
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
