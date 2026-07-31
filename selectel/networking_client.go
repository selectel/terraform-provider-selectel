package selectel

import (
	"fmt"

	"github.com/selectel/go-selvpcclient/v4/selvpcclient"
	vpc "github.com/selectel/vpc-go/pkg/v2"
)

// networkingV2ServiceType is the Keystone catalog service type of Neutron.
const networkingV2ServiceType = "network"

// networkingTokenCatalog is the part of go-selvpcclient that building a
// networking client depends on: a project-scoped token and a regional endpoint.
type networkingTokenCatalog interface {
	// GetXAuthToken returns the project-scoped Keystone token to authenticate with.
	GetXAuthToken() string

	// GetEndpointURL resolves the URL of serviceType in regionName.
	GetEndpointURL(serviceType, regionName string) (string, error)
}

// selvpcTokenCatalog adapts *selvpcclient.Client, whose catalog is a struct
// field rather than a method, to networkingTokenCatalog.
type selvpcTokenCatalog struct {
	client *selvpcclient.Client
}

func (s selvpcTokenCatalog) GetXAuthToken() string {
	return s.client.GetXAuthToken()
}

func (s selvpcTokenCatalog) GetEndpointURL(serviceType, regionName string) (string, error) {
	endpoint, err := s.client.Catalog.GetEndpoint(serviceType, regionName)
	if err != nil {
		return "", err
	}

	return endpoint.URL, nil
}

// networkingV2Client returns a networking client for the given project and
// region. It reuses the project-scoped token that go-selvpcclient has already
// issued instead of authenticating against Keystone a second time, and hands it
// to the Selectel networking SDK together with the regional endpoint.
func networkingV2Client(meta interface{}, projectID, region string) (*vpc.Client, error) {
	config := meta.(*Config)

	selvpcClient, err := config.GetSelVPCClientWithProjectScope(projectID)
	if err != nil {
		return nil, fmt.Errorf("can't get project-scope selvpc client for networking api: %w", err)
	}

	return newNetworkingV2Client(selvpcTokenCatalog{client: selvpcClient}, region, config.UserAgent)
}

// newNetworkingV2Client builds the networking SDK client from an already
// authenticated token/catalog source.
func newNetworkingV2Client(source networkingTokenCatalog, region, userAgent string) (*vpc.Client, error) {
	if region == "" {
		return nil, fmt.Errorf("region must be set to use the networking api")
	}

	token := source.GetXAuthToken()
	if token == "" {
		return nil, fmt.Errorf("got an empty authentication token for the networking api")
	}

	endpoint, err := source.GetEndpointURL(networkingV2ServiceType, region)
	if err != nil {
		return nil, fmt.Errorf("can't resolve the %s service endpoint in region '%s': %w",
			networkingV2ServiceType, region, err)
	}
	if endpoint == "" {
		return nil, fmt.Errorf("the %s service endpoint in region '%s' is empty",
			networkingV2ServiceType, region)
	}

	client, err := vpc.NewClient(vpc.Config{
		Endpoint:  endpoint,
		Token:     token,
		UserAgent: userAgent,
	})
	if err != nil {
		return nil, fmt.Errorf("can't build the networking client for region '%s': %w", region, err)
	}

	return client, nil
}
