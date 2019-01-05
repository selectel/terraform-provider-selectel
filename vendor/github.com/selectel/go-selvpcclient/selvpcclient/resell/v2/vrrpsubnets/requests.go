package vrrpsubnets

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/selectel/go-selvpcclient/selvpcclient"
)

const resourceURL = "vrrp_subnets"

// Get returns a single VRRP subnet by its id.
func Get(ctx context.Context, client *selvpcclient.ServiceClient, id string) (*VRRPSubnet, *selvpcclient.ResponseResult, error) {
	url := strings.Join([]string{client.Endpoint, resourceURL, id}, "/")
	responseResult, err := client.DoRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, err
	}
	if responseResult.Err != nil {
		return nil, responseResult, responseResult.Err
	}

	// Extract a VRRP subnet from the response body.
	var result struct {
		VRRPSubnet *VRRPSubnet `json:"vrrp_subnet"`
	}
	err = responseResult.ExtractResult(&result)
	if err != nil {
		return nil, responseResult, err
	}

	return result.VRRPSubnet, responseResult, nil
}

// List gets a list of VRRP subnets in the current domain.
func List(ctx context.Context, client *selvpcclient.ServiceClient, opts ListOpts) ([]*VRRPSubnet, *selvpcclient.ResponseResult, error) {
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

	// Extract VRRP subnets from the response body.
	var result struct {
		VRRPSubnets []*VRRPSubnet `json:"vrrp_subnets"`
	}
	err = responseResult.ExtractResult(&result)
	if err != nil {
		return nil, responseResult, err
	}

	return result.VRRPSubnets, responseResult, nil
}

// Create requests a creation of the VRRP subnets in the specified project.
func Create(ctx context.Context, client *selvpcclient.ServiceClient, projectID string, createOpts VRRPSubnetOpts) ([]*VRRPSubnet, *selvpcclient.ResponseResult, error) {
	createVRRPSubnetsOpts := &createOpts
	requestBody, err := json.Marshal(createVRRPSubnetsOpts)
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

	// Extract VRRP subnets from the response body.
	var result struct {
		VRRPSubnets []*VRRPSubnet `json:"vrrp_subnets"`
	}
	err = responseResult.ExtractResult(&result)
	if err != nil {
		return nil, responseResult, err
	}

	return result.VRRPSubnets, responseResult, nil
}

// Delete deletes a single VRRP subnet by its id.
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
