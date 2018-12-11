package selvpc

import (
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccResellV2RoleImportBasic(t *testing.T) {
	resourceName := "selvpc_resell_role_v2.role_tf_acc_test_1"
	projectName := acctest.RandomWithPrefix("tf-acc")
	userName := acctest.RandomWithPrefix("tf-acc")
	userPassword := acctest.RandString(8)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccSelVPCPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckResellV2RoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResellV2RoleBasic(projectName, userName, userPassword),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
