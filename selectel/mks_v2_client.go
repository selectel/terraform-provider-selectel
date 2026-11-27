package selectel

import (
	"context"
	"fmt"

	"github.com/selectel/go-selvpcclient/v5/selvpcclient"
	mksv2 "github.com/selectel/mks-go/v2/pkg"
)

// MKSV2Defaults supplies the scope for UUID-based imports.
func (c *Config) MKSV2Defaults() (string, string) { return c.ProjectID, c.Region }

// MKSV2Client authenticates for each operation, because mks-go takes a token string it cannot renew.
// The pool is checked against the catalog first to list the valid pools.
func (c *Config) MKSV2Client(ctx context.Context, projectID, pool string) (*mksv2.ServiceClient, error) {
	opts := c.selvpcClientOptions(projectID)
	opts.Context = ctx
	selvpcClient, err := selvpcclient.NewClient(opts)
	if err != nil {
		return nil, fmt.Errorf("authenticate project %s: %w", projectID, err)
	}
	err = validateRegion(selvpcClient, MKSV2, pool)
	if err != nil {
		return nil, fmt.Errorf("can't validate pool: %w", err)
	}
	endpoint, err := selvpcClient.Catalog.GetEndpoint(MKSV2, pool)
	if err != nil {
		return nil, err
	}
	client, err := mksv2.NewMKSClientV2(selvpcClient.GetXAuthToken(), endpoint.URL)
	if err != nil {
		return nil, err
	}
	client.UserAgent = c.UserAgent

	return client, nil
}
