package v2

import (
	"github.com/selectel/go-selvpcclient/selvpcclient"
	"github.com/selectel/go-selvpcclient/selvpcclient/resell"
)

// APIVersion sets the version of the Resell client.
const (
	APIVersion = "v2"
)

// NewV2ResellClient initializes a new Resell client for the V2 API.
func NewV2ResellClient(tokenID string) *selvpcclient.ServiceClient {
	return &selvpcclient.ServiceClient{
		HTTPClient: selvpcclient.NewHTTPClient(),
		Endpoint:   resell.Endpoint + "/" + APIVersion,
		TokenID:    tokenID,
		UserAgent:  resell.UserAgent,
	}
}

// NewV2ResellClientWithEndpoint initializes a new Resell client for the V2 API with a custom endpoint.
func NewV2ResellClientWithEndpoint(tokenID, endpoint string) *selvpcclient.ServiceClient {
	resellClient := &selvpcclient.ServiceClient{
		HTTPClient: selvpcclient.NewHTTPClient(),
		Endpoint:   endpoint,
		TokenID:    tokenID,
		UserAgent:  resell.UserAgent,
	}

	return resellClient
}
