package selvpc

import (
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccResellV2ProjectImportBasic(t *testing.T) {
	resourceName := "selvpc_project_v2.project_tf_acc_test_1"
	projectName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccSelVPCPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckResellV2ProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResellV2ProjectBasic(projectName),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"all_quotas"},
			},
		},
	})
}
