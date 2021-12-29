package selectel

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccMKSKubeconfigV1DataSourceBasic(t *testing.T) {
	projectName := acctest.RandomWithPrefix("tf-acc")
	clusterName := acctest.RandomWithPrefix("tf-acc-cl")
	kubeVersion := testAccMKSClusterV1GetDefaultKubeVersion(t)
	maintenanceWindowStart := testAccMKSClusterV1GetMaintenanceWindowStart(12 * time.Hour)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCV2ProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMKSKubeconfigV1Basic(projectName, clusterName, kubeVersion, maintenanceWindowStart),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMKSKubeconfigV1("data.selectel_mks_kubeconfig_v1.kubeconfig_tf_acc_test_1"),
				),
			},
		},
	})
}

func testAccCheckMKSKubeconfigV1(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("can't find kubeconfig data source: %s", name)
		}

		if _, ok = rs.Primary.Attributes["raw_config"]; !ok {
			return errors.New("empty 'raw_config' field in kubeconfigs data source")
		}
		if _, ok = rs.Primary.Attributes["server"]; !ok {
			return errors.New("empty 'server' field in kubeconfigs data source")
		}
		if _, ok = rs.Primary.Attributes["cluster_ca_cert"]; !ok {
			return errors.New("empty 'cluster_ca_cert' field in kubeconfigs data source")
		}
		if _, ok = rs.Primary.Attributes["client_cert"]; !ok {
			return errors.New("empty 'client_cert' field in kubeconfigs data source")
		}
		if _, ok = rs.Primary.Attributes["client_key"]; !ok {
			return errors.New("empty 'client_key' field in kubeconfigs data source")
		}

		return nil
	}
}

func testAccMKSKubeconfigV1Basic(projectName, clusterName, kubeVersion, maintenanceWindowStart string) string {
	return fmt.Sprintf(`
%s

data "selectel_mks_kubeconfig_v1" "kubeconfig_tf_acc_test_1" {
  cluster_id    = "${selectel_mks_cluster_v1.cluster_tf_acc_test_1.id}"
  project_id    = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  region        = "ru-3"
}
`, testAccMKSClusterV1Basic(projectName, clusterName, kubeVersion, maintenanceWindowStart))
}
