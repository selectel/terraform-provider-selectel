package selectel

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccVPCV2RoleImportBasic(t *testing.T) {
	resourceName := "selectel_vpc_role_v2.role_tf_acc_test_1"
	projectName := acctest.RandomWithPrefix("tf-acc")
	userName := acctest.RandomWithPrefix("tf-acc")
	userPassword := acctest.RandString(8)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccSelectelPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVPCV2RoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCV2RoleBasic(projectName, userName, userPassword),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
