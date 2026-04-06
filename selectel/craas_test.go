package selectel

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	v1 "github.com/selectel/craas-go/pkg"
	clientv1 "github.com/selectel/craas-go/pkg/v1/client"
	clientv2 "github.com/selectel/craas-go/pkg/v2/client"
	"github.com/stretchr/testify/assert"
)

func newCRaaSTestClient(rs *terraform.ResourceState, testAccProvider *schema.Provider) (*clientv1.ServiceClient, error) {
	config := testAccProvider.Meta().(*Config)

	var projectID string

	if id, ok := rs.Primary.Attributes["project_id"]; ok {
		projectID = id
	}

	selvpcClient, err := config.GetSelVPCClientWithProjectScope(projectID)
	if err != nil {
		return nil, fmt.Errorf("can't get selvpc client for craas acc tests: %w", err)
	}

	craasEndpoint, err := getEndpointForCRaaS(selvpcClient, CRaaS)
	if err != nil {
		return nil, fmt.Errorf("can't get endpoint for craas acc tests: %w", err)
	}

	craasClient := v1.NewCRaaSClientV1(selvpcClient.GetXAuthToken(), craasEndpoint)

	return craasClient, nil
}

func newCRaaSV2TestClient(rs *terraform.ResourceState, testAccProvider *schema.Provider) (*clientv2.ServiceClient, error) {
	config := testAccProvider.Meta().(*Config)

	var projectID string

	if id, ok := rs.Primary.Attributes["project_id"]; ok {
		projectID = id
	}

	selvpcClient, err := config.GetSelVPCClientWithProjectScope(projectID)
	if err != nil {
		return nil, fmt.Errorf("can't get selvpc client for craas v2 acc tests: %w", err)
	}

	craasEndpoint, err := getEndpointForCRaaS(selvpcClient, CRaaSV2)
	if err != nil {
		return nil, fmt.Errorf("can't get endpoint for craas v2 acc tests: %w", err)
	}

	craasClient, err := clientv2.NewCRaaSClientV2(selvpcClient.GetXAuthToken(), craasEndpoint)
	if err != nil {
		return nil, fmt.Errorf("can't create craas v2 client for acc tests: %w", err)
	}

	return craasClient, nil
}

func TestGetHostNameForCRaaS(t *testing.T) {
	expected := "https://cr.selcloud.ru"
	actual, err := getHostNameForCRaaS("https://cr.selcloud.ru/api/v1")
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}
