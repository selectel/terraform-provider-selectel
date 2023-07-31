package selectel

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	v1 "github.com/selectel/craas-go/pkg"
	"github.com/stretchr/testify/assert"
)

func newCRaaSTestClient(rs *terraform.ResourceState, testAccProvider *schema.Provider) (*v1.ServiceClient, error) {
	config := testAccProvider.Meta().(*Config)

	var projectID string

	if id, ok := rs.Primary.Attributes["project_id"]; ok {
		projectID = id
	}

	selvpcClient, err := config.GetSelVPCClientWithProjectScope(projectID)
	if err != nil {
		return nil, fmt.Errorf("can't get selvpc client for craas acc tests: %w", err)
	}

	craasEndpoint, err := getEndpointForCRaaS(selvpcClient)
	if err != nil {
		return nil, fmt.Errorf("can't get endpoint for craas acc tests: %w", err)
	}

	craasClient := v1.NewCRaaSClientV1(selvpcClient.GetXAuthToken(), craasEndpoint)

	return craasClient, nil
}

func TestGetHostNameForCRaaS(t *testing.T) {
	expected := "https://cr.selcloud.ru"
	actual, err := getHostNameForCRaaS("https://cr.selcloud.ru/api/v1")
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}
