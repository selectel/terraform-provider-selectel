package selectel

import (
	"context"

	blockstorage "github.com/selectel/blockstorage-go/pkg/v1"
)

// BlockStorageDefaults supplies the scope for UUID-based imports.
func (c *Config) BlockStorageDefaults() (string, string) { return c.ProjectID, c.Region }

// BlockStorageClient authenticates for each operation. blockstorage-go takes a
// token string that it cannot renew, so the cached selvpc client is not reused.
func (c *Config) BlockStorageClient(ctx context.Context, projectID, region string) (*blockstorage.Client, error) {
	endpoint, token, err := c.projectEndpoint(ctx, projectID, region, BlockStorage)
	if err != nil {
		return nil, err
	}

	return blockstorage.NewClient(blockstorage.Config{Endpoint: endpoint, Token: token, UserAgent: c.UserAgent})
}
