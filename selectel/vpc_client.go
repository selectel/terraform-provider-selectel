package selectel

import (
	"fmt"

	vpc "github.com/selectel/vpc-go/pkg/v2"
)

const vpcV2ServiceType = "network"

func vpcV2Client(meta any, projectID, region string) (*vpc.Client, error) {
	config := meta.(*Config)

	selvpcClient, err := config.GetSelVPCClientWithProjectScope(projectID)
	if err != nil {
		return nil, fmt.Errorf("can't get project-scope selvpc client for VPC api: %w", err)
	}

	endpoint, err := selvpcClient.Catalog.GetEndpoint(vpcV2ServiceType, region)
	if err != nil {
		return nil, fmt.Errorf("can't resolve the %s service endpoint in region '%s': %w",
			vpcV2ServiceType, region, err)
	}

	client, err := vpc.NewClient(vpc.Config{
		Endpoint:  endpoint.URL,
		Token:     selvpcClient.GetXAuthToken(),
		UserAgent: config.UserAgent,
	})
	if err != nil {
		return nil, fmt.Errorf("can't build the VPC client for region '%s': %w", region, err)
	}

	return client, nil
}
