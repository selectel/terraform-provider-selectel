package selectel

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDBaaSRedisDatastoreV1ImportBasic(t *testing.T) {
	resourceName := "selectel_dbaas_redis_datastore_v1.datastore_tf_acc_test_1"
	projectName := acctest.RandomWithPrefix("tf-acc")
	datastoreName := acctest.RandomWithPrefix("tf-acc-ds")
	nodeCount := 1

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCV2ProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDBaaSRedisDatastoreV1Basic(projectName, datastoreName, nodeCount),
				Check:  testAccCheckSelectelImportEnv(resourceName),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"redis_password"},
			},
		},
	})
}
