package selectel

import (
	"context"

	filestorage "github.com/selectel/filestorage-go/pkg/v1"
)

// FileStorageDefaults supplies the scope for UUID-based imports.
func (c *Config) FileStorageDefaults() (string, string) { return c.ProjectID, c.Region }

// FileStorageClient authenticates for each operation. filestorage-go takes a
// token string that it cannot renew, so the cached selvpc client is not reused.
func (c *Config) FileStorageClient(ctx context.Context, projectID, region string) (*filestorage.Client, error) {
	endpoint, token, err := c.projectEndpoint(ctx, projectID, region, FileStorage)
	if err != nil {
		return nil, err
	}

	return filestorage.NewClient(filestorage.Config{Endpoint: endpoint, Token: token, UserAgent: c.UserAgent})
}
