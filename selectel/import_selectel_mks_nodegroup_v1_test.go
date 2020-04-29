package selectel

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccMKSNodegroupV1ImportBasic(t *testing.T) {
	resourceName := "selectel_mks_nodegroup_v1.nodegroup_tf_acc_test_1"
	projectName := acctest.RandomWithPrefix("tf-acc")
	clusterName := acctest.RandomWithPrefix("tf-acc-cl")
	kubeVersion := testAccMKSClusterV1GetDefaultKubeVersion(t)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccSelectelPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVPCV2ProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMKSNodegroupV1Basic(projectName, clusterName, kubeVersion),
				Check:  testAccCheckSelectelImportEnv(resourceName),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"nodes_count", "cpus", "ram_mb"},
			},
		},
	})
}
