package selectel

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccCloudPrivateDNSServiceV1ImportBasic(t *testing.T) {
	region := os.Getenv("INFRA_REGION")
	projectID := os.Getenv("INFRA_PROJECT_ID")
	resourceName := "selectel_private_dns_service_v1.service"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectID(t) },
		ProviderFactories: testAccProvidersWithOpenStack,
		CheckDestroy:      testAccCloudPrivateDNSServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudPrivateDNSServiceV1Basic(region, projectID),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
