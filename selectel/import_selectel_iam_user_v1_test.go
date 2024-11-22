package selectel

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccIAMV1UserImportBasic(t *testing.T) {
	resourceName := "selectel_iam_user_v1.user_tf_acc_test_1"
	userEmail := acctest.RandomWithPrefix("tf-acc") + "@example.com"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckIAMV1UserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIAMV1UserBasic(userEmail),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"email"},
			},
		},
	})
}
