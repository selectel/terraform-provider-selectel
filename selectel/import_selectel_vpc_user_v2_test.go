package selectel

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVPCV2UserImportBasic(t *testing.T) {
	resourceName := "selectel_vpc_user_v2.user_tf_acc_test_1"
	userName := acctest.RandomWithPrefix("tf-acc")
	userPassword := acctest.RandString(8)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCV2UserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCV2UserBasic(userName, userPassword),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
	})
}
