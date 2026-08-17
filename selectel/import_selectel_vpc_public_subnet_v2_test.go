package selectel

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVPCV2SubnetImportBasic(t *testing.T) {
	resourceName := "selectel_vpc_public_subnet_v2.subnet_tf_acc_test_1"
	projectID := os.Getenv("INFRA_PROJECT_ID")
	_, region := vpcV2TestEnv()
	prefixLength := testAccVPCPublicSubnetPrefixLength(t)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCV2SubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCV2SubnetExistingProject(projectID, region, prefixLength),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
