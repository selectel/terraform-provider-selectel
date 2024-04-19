package selectel

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccIAMV1ServiceUserImportBasic(t *testing.T) {
	resourceName := "selectel_iam_serviceuser_v1.serviceuser_tf_acc_test_1"
	serviceUserName := acctest.RandomWithPrefix("tf-acc")
	serviceUserPassword := "A" + acctest.RandString(8) + "1"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckIAMV1ServiceUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIAMV1ServiceUserBasic(serviceUserName, serviceUserPassword),
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
