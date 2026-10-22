package selectel

import (
	"context"

	images "github.com/selectel/images-go/pkg/v1"
)

// ImagesDefaults supplies the scope for imports.
func (c *Config) ImagesDefaults() (string, string) { return c.ProjectID, c.Region }

// ImagesClient authenticates for each operation: images-go takes a token
// string that it cannot renew, and a wait may outlive the token.
func (c *Config) ImagesClient(ctx context.Context, projectID, region string) (*images.Client, error) {
	endpoint, token, err := c.projectEndpoint(ctx, projectID, region, Image)
	if err != nil {
		return nil, err
	}

	return images.NewClient(images.Config{Endpoint: endpoint, Token: token, UserAgent: c.UserAgent})
}
