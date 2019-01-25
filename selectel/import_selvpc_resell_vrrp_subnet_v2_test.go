package selvpc

import (
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccResellV2VRRPSubnetImportBasic(t *testing.T) {
	resourceName := "selvpc_resell_vrrp_subnet_v2.vrrp_subnet_tf_acc_test_1"
	projectName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccSelVPCPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckResellV2VRRPSubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResellV2VRRPSubnetBasic(projectName),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
