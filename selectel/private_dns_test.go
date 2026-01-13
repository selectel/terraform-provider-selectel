package selectel

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	privatedns "github.com/selectel/private-dns-go/pkg/v1"
)

func newTestPrivateDNSClient(rs *terraform.ResourceState, testAccProvider *schema.Provider) (*privatedns.PrivateDNSClient, error) {
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
		dnsEndpoint, err := selvpcClient.Catalog.GetEndpoint(PrivateDNS, region)
		if err != nil {
			return nil, fmt.Errorf("can't get endpoint for dbaas acc tests: %w", err)
		}
		endpoint = dnsEndpoint.URL
	}

	cfg := &privatedns.Config{
		URL:       endpoint,
		AuthToken: selvpcClient.GetXAuthToken(),
	}
	client := privatedns.NewPrivateDNSClient(cfg)

	return client, nil
}
