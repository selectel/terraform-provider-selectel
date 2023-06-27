package selectel

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccCRaaSRegistryV1ImportBasic(t *testing.T) {
	resourceName := "selectel_craas_registry_v1.registry_tf_acc_test_1"
	projectName := acctest.RandomWithPrefix("tf-acc")
	registryName := acctest.RandomWithPrefix("tf-acc-reg")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCV2ProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCRaaSRegistryV1Basic(projectName, registryName),
				Check:  testAccCheckSelectelCRaaSImportEnv(resourceName),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
