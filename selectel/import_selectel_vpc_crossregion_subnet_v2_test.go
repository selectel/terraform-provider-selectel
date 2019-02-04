package selectel

import (
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccVPCV2CrossRegionSubnetImportBasic(t *testing.T) {
	resourceName := "selectel_vpc_crossregion_subnet_v2.crossregion_subnet_tf_acc_test_1"
	projectName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccSelectelPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVPCV2CrossRegionSubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCV2CrossRegionSubnetBasic(projectName),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
