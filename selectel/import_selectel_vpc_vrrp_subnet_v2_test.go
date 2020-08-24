package selectel

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVPCV2VRRPSubnetImportBasic(t *testing.T) {
	resourceName := "selectel_vpc_vrrp_subnet_v2.vrrp_subnet_tf_acc_test_1"
	projectName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCV2VRRPSubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCV2VRRPSubnetBasic(projectName),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
