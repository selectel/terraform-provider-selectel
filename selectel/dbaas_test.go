package selectel

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/selectel/dbaas-go"
)

func newTestDBaaSClient(_ context.Context, rs *terraform.ResourceState, testAccProvider *schema.Provider) (*dbaas.API, error) {
	config := testAccProvider.Meta().(*Config)

	var projectID string
	var endpoint string

	if id, ok := rs.Primary.Attributes["project_id"]; ok {
		projectID = id
	}

	selvpcClient, err := config.GetSelVPCClientWithProjectScope(projectID)
	if err != nil {
		return nil, fmt.Errorf("can't get selvpc client for dbaas acc tests: %w", err)
	}

	if region, ok := rs.Primary.Attributes["region"]; ok {
		dbaasEndpoint, err := selvpcClient.Catalog.GetEndpoint(DBaaS, region)
		if err != nil {
			return nil, fmt.Errorf("can't get endpoint for dbaas acc tests: %w", err)
		}
		endpoint = dbaasEndpoint.URL
	}

	dbaasClient, err := dbaas.NewDBAASClient(selvpcClient.GetXAuthToken(), endpoint)
	if err != nil {
		return nil, fmt.Errorf("can't get dbaas client for dbaas acc tests: %w", err)
	}

	return dbaasClient, nil
}
