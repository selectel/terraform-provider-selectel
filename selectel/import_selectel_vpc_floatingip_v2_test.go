package selectel

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccVPCV2FloatingIPImportBasic(t *testing.T) {
	resourceName := "selectel_vpc_floatingip_v2.floatingip_tf_acc_test_1"
	projectName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccSelectelPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVPCV2FloatingIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCV2FloatingIPBasic(projectName),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
