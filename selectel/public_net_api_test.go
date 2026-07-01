package selectel

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	publicnetapi "github.com/selectel/public-net-api-go/pkg/v1"
)

func newTestPublicNetAPI(
	rs *terraform.ResourceState,
	testAccProvider *schema.Provider,
) (*publicnetapi.PublicNetAPIClient, error) {
	config := testAccProvider.Meta().(*Config)

	projectID, ok := rs.Primary.Attributes["project_id"]
	if !ok || projectID == "" {
		return nil, fmt.Errorf("project_id is required for %s client initialization", PublicNetAPI)
	}

	region, ok := rs.Primary.Attributes["region"]
	if !ok || region == "" {
		return nil, fmt.Errorf("region is required for for %s client initialization", PublicNetAPI)
	}

	selectelVPCClient, err := config.GetSelVPCClientWithProjectScope(projectID)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to get project-scope Selectel VPC client (project_id=%s): %w",
			projectID,
			err,
		)
	}

	endpoint, err := selectelVPCClient.Catalog.GetEndpoint(PublicNetAPI, region)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to get endpoint for %s (region=%s, project_id=%s): %w",
			PublicNetAPI,
			region,
			projectID,
			err,
		)
	}

	if endpoint.URL == "" {
		return nil, fmt.Errorf(
			"empty endpoint URL returned for %s (region=%s, project_id=%s)",
			PublicNetAPI,
			region,
			projectID,
		)
	}

	cfg := &publicnetapi.Config{
		URL:       endpoint.URL,
		AuthToken: selectelVPCClient.GetXAuthToken(),
	}

	client, err := publicnetapi.NewPublicNetAPIClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize %s client: %w", PublicNetAPI, err)
	}

	return client, nil
}
