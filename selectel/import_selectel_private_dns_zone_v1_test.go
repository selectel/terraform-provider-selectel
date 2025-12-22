package selectel

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccCloudPrivateDNSZoneV1ImportBasic(t *testing.T) {
	region := os.Getenv("INFRA_REGION")
	projectID := os.Getenv("INFRA_PROJECT_ID")
	resourceName := "selectel_private_dns_zone_v1.zone"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectID(t) },
		ProviderFactories: testAccProvidersWithOpenStack,
		CheckDestroy:      testAccCloudPrivateDNSZoneV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudPrivateDNSZoneV1WithRecords(region, 1800, projectID),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
