package node

import (
	"context"
	"net/http"
	"strings"

	v1 "github.com/selectel/mks-go/pkg/v1"
)

// Get returns a node of a cluster nodegroup by its id.
func Get(ctx context.Context, client *v1.ServiceClient, clusterID, nodegroupID, nodeID string) (*View, *v1.ResponseResult, error) {
	url := strings.Join([]string{client.Endpoint, v1.ResourceURLCluster, clusterID, v1.ResourceURLNodegroup, nodegroupID, nodeID}, "/")
	responseResult, err := client.DoRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, err
	}
	if responseResult.Err != nil {
		return nil, responseResult, responseResult.Err
	}

	// Extract a node to the response body.
	var result struct {
		Node *View `json:"node"`
	}
	err = responseResult.ExtractResult(&result)
	if err != nil {
		return nil, responseResult, err
	}

	return result.Node, responseResult, err
}

// Reinstall requests to make reinstall of a single node of a cluster nodegroup by its id.
func Reinstall(ctx context.Context, client *v1.ServiceClient, clusterID, nodegroupID, nodeID string) (*v1.ResponseResult, error) {
	url := strings.Join([]string{client.Endpoint, v1.ResourceURLCluster, clusterID, v1.ResourceURLNodegroup, nodegroupID, nodeID, v1.ResourceURLReinstall}, "/")
	responseResult, err := client.DoRequest(ctx, http.MethodPost, url, nil)
	if err != nil {
		return nil, err
	}
	if responseResult.Err != nil {
		err = responseResult.Err
	}

	return responseResult, err
}
