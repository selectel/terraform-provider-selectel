package selectel

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVPCPublicPortV1ImportBasic(t *testing.T) {
	region := os.Getenv("INFRA_REGION")
	projectID := os.Getenv("INFRA_PROJECT_ID")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheckWithProjectID(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccVPCPublicPortV1CheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCPublicPortV1(region, projectID),
			},
			{
				ResourceName:      "selectel_vpc_public_port_v1.port",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
