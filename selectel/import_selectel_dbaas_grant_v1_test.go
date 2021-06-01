package selectel

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDBaaSGrantV1ImportBasic(t *testing.T) {
	resourceName := "selectel_dbaas_grant_v1.grant_tf_acc_test_1"
	projectName := acctest.RandomWithPrefix("tf-acc")
	datastoreName := acctest.RandomWithPrefix("tf-acc-ds")
	userName := acctest.RandomWithPrefix("tf-acc-user")
	userPassword := acctest.RandomWithPrefix("tf-acc-pass")
	databaseName := acctest.RandomWithPrefix("tf-acc-db")
	nodeCount := 1

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCV2ProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDBaaSGrantV1Basic(projectName, datastoreName, userName, userPassword, databaseName, nodeCount),
				Check:  testAccCheckSelectelImportEnv(resourceName),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
