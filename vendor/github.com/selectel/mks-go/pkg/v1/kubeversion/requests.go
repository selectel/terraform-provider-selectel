package kubeversion

import (
	"context"
	"net/http"
	"strings"

	v1 "github.com/selectel/mks-go/pkg/v1"
)

// List gets a list of all supported Kubernetes versions.
func List(ctx context.Context, client *v1.ServiceClient) ([]*View, *v1.ResponseResult, error) {
	url := strings.Join([]string{client.Endpoint, v1.ResourceURLKubeversion}, "/")
	responseResult, err := client.DoRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, err
	}
	if responseResult.Err != nil {
		return nil, responseResult, responseResult.Err
	}

	// Extract Kubernetes versions from the response body.
	var result struct {
		KubeVersions []*View `json:"kube_versions"`
	}
	err = responseResult.ExtractResult(&result)
	if err != nil {
		return nil, responseResult, err
	}

	return result.KubeVersions, responseResult, nil
}
