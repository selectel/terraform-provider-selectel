package nodegroup

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"

	v1 "github.com/selectel/mks-go/pkg/v1"
)

// Get returns a cluster nodegroup by its id.
func Get(ctx context.Context, client *v1.ServiceClient, clusterID, nodegroupID string) (*View, *v1.ResponseResult, error) {
	url := strings.Join([]string{client.Endpoint, v1.ResourceURLCluster, clusterID, v1.ResourceURLNodegroup, nodegroupID}, "/")
	responseResult, err := client.DoRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, err
	}
	if responseResult.Err != nil {
		return nil, responseResult, responseResult.Err
	}

	// Extract a nodegroup to the response body.
	var result struct {
		Nodegroup *View `json:"nodegroup"`
	}
	err = responseResult.ExtractResult(&result)
	if err != nil {
		return nil, responseResult, err
	}

	return result.Nodegroup, responseResult, err
}

// List gets a list of all cluster nodegroups.
func List(ctx context.Context, client *v1.ServiceClient, clusterID string) ([]*View, *v1.ResponseResult, error) {
	url := strings.Join([]string{client.Endpoint, v1.ResourceURLCluster, clusterID, v1.ResourceURLNodegroup}, "/")
	responseResult, err := client.DoRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, err
	}
	if responseResult.Err != nil {
		return nil, responseResult, responseResult.Err
	}

	// Extract nodegroups from the response body.
	var result struct {
		Nodegroups []*View `json:"nodegroups"`
	}
	err = responseResult.ExtractResult(&result)
	if err != nil {
		return nil, responseResult, err
	}

	return result.Nodegroups, responseResult, err
}

// Create requests a creation of a new cluster nodegroup.
func Create(ctx context.Context, client *v1.ServiceClient, clusterID string, opts *CreateOpts) (*v1.ResponseResult, error) {
	createNodegroupOpts := struct {
		Nodegroup *CreateOpts `json:"nodegroup"`
	}{
		Nodegroup: opts,
	}
	requestBody, err := json.Marshal(createNodegroupOpts)
	if err != nil {
		return nil, err
	}

	url := strings.Join([]string{client.Endpoint, v1.ResourceURLCluster, clusterID, v1.ResourceURLNodegroup}, "/")
	responseResult, err := client.DoRequest(ctx, http.MethodPost, url, bytes.NewReader(requestBody))
	if err != nil {
		return nil, err
	}
	if responseResult.Err != nil {
		err = responseResult.Err
	}

	return responseResult, err
}

// Delete deletes a cluster nodegroup by its id.
func Delete(ctx context.Context, client *v1.ServiceClient, clusterID, nodegroupID string) (*v1.ResponseResult, error) {
	url := strings.Join([]string{client.Endpoint, v1.ResourceURLCluster, clusterID, v1.ResourceURLNodegroup, nodegroupID}, "/")
	responseResult, err := client.DoRequest(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return nil, err
	}
	if responseResult.Err != nil {
		err = responseResult.Err
	}

	return responseResult, err
}

// Resize requests a resize of a cluster nodegroup by its id.
func Resize(ctx context.Context, client *v1.ServiceClient, clusterID, nodegroupID string, opts *ResizeOpts) (*v1.ResponseResult, error) {
	resizeNodegroupOpts := struct {
		Nodegroup *ResizeOpts `json:"nodegroup"`
	}{
		Nodegroup: opts,
	}
	requestBody, err := json.Marshal(resizeNodegroupOpts)
	if err != nil {
		return nil, err
	}

	url := strings.Join([]string{client.Endpoint, v1.ResourceURLCluster, clusterID, v1.ResourceURLNodegroup, nodegroupID, v1.ResourceURLResize}, "/")
	responseResult, err := client.DoRequest(ctx, http.MethodPost, url, bytes.NewReader(requestBody))
	if err != nil {
		return nil, err
	}
	if responseResult.Err != nil {
		err = responseResult.Err
	}

	return responseResult, err
}

// Update requests a update of a cluster nodegroup by its id.
func Update(ctx context.Context, client *v1.ServiceClient, clusterID, nodegroupID string, opts *UpdateOpts) (*v1.ResponseResult, error) {
	updateNodegroupOpts := struct {
		Nodegroup *UpdateOpts `json:"nodegroup"`
	}{
		Nodegroup: opts,
	}
	requestBody, err := json.Marshal(updateNodegroupOpts)
	if err != nil {
		return nil, err
	}

	url := strings.Join([]string{client.Endpoint, v1.ResourceURLCluster, clusterID, v1.ResourceURLNodegroup, nodegroupID}, "/")
	responseResult, err := client.DoRequest(ctx, http.MethodPut, url, bytes.NewReader(requestBody))
	if err != nil {
		return nil, err
	}
	if responseResult.Err != nil {
		err = responseResult.Err
	}

	return responseResult, err
}
