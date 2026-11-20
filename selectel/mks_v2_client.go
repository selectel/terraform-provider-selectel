package selectel

import (
	"context"
	"fmt"
	"strings"

	"github.com/selectel/go-selvpcclient/v5/selvpcclient"
	mksv2 "github.com/selectel/mks-go/v2/pkg"
)

// MKSV2Defaults supplies the scope for UUID-based imports.
func (c *Config) MKSV2Defaults() (string, string) { return c.ProjectID, c.Region }

// MKSV2Client authenticates for each operation. mks-go takes a token string
// that it cannot renew, so the cached selvpc client is not reused. The pool is
// checked against the catalog first, like _v1 does, to list the valid pools.
func (c *Config) MKSV2Client(ctx context.Context, projectID, pool string) (*mksv2.ServiceClient, error) {
	selvpcClient, err := c.projectClient(ctx, projectID)
	if err != nil {
		return nil, err
	}
	err = validateRegion(selvpcClient, MKSV2, pool)
	if err != nil {
		return nil, fmt.Errorf("can't validate pool: %w", err)
	}
	endpoint, err := selvpcClient.Catalog.GetEndpoint(MKSV2, pool)
	if err != nil {
		return nil, err
	}
	client, err := mksv2.NewMKSClientV2(selvpcClient.GetXAuthToken(), mksV2BaseURL(endpoint.URL))
	if err != nil {
		return nil, err
	}
	client.UserAgent = c.UserAgent

	return client, nil
}

// projectClient authenticates a fresh selvpc client scoped to the project,
// the way projectEndpoint does, but without the cache.
func (c *Config) projectClient(ctx context.Context, projectID string) (*selvpcclient.Client, error) {
	opts := c.selvpcClientOptions(projectID)
	opts.Context = ctx
	client, err := selvpcclient.NewClient(opts)
	if err != nil {
		return nil, fmt.Errorf("authenticate project %s: %w", projectID, err)
	}

	return client, nil
}

// mksV2BaseURL turns the catalog endpoint into the base URL of the MKS API
// v2 client: the endpoint ends with /v2 (https://ru-7.mks.selcloud.ru/v2),
// while the mks-go v2 request paths start with /v2/, so the suffix is trimmed.
func mksV2BaseURL(endpoint string) string {
	return strings.TrimSuffix(strings.TrimSuffix(endpoint, "/"), "/v2")
}
