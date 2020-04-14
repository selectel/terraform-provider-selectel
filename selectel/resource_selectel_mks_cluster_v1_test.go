package selectel

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/selectel/go-selvpcclient/selvpcclient/resell/v2/projects"
	"github.com/selectel/go-selvpcclient/selvpcclient/resell/v2/tokens"
	v1 "github.com/selectel/mks-go/pkg/v1"
	"github.com/selectel/mks-go/pkg/v1/cluster"
)

func TestAccMKSClusterV1Basic(t *testing.T) {
	var (
		mksCluster cluster.View
		project    projects.Project
	)

	projectName := acctest.RandomWithPrefix("tf-acc")
	clusterName := acctest.RandomWithPrefix("tf-acc-cl")
	nodegroupName := acctest.RandomWithPrefix("tf-acc-ng")

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccSelectelPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccMKSClusterV1Basic(projectName, clusterName, nodegroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					testAccCheckMKSClusterV1Exists("selectel_mks_cluster_v1.cluster_tf_acc_test_1", &mksCluster),
					resource.TestCheckResourceAttr("selectel_mks_cluster_v1.cluster_tf_acc_test_1", "kube_version", "1.16.8"),
					resource.TestCheckResourceAttr("selectel_mks_cluster_v1.cluster_tf_acc_test_1", "region", "ru-3"),
					resource.TestCheckResourceAttr("selectel_mks_cluster_v1.cluster_tf_acc_test_1", "nodegroups.#", "1"),
					resource.TestCheckResourceAttr("selectel_mks_cluster_v1.cluster_tf_acc_test_1", "enable_autorepair", "true"),
					resource.TestCheckResourceAttr("selectel_mks_cluster_v1.cluster_tf_acc_test_1", "enable_patch_version_auto_upgrade", "true"),
					resource.TestCheckResourceAttr("selectel_mks_cluster_v1.cluster_tf_acc_test_1", "maintenance_window_start", "01:00:00"),
					resource.TestCheckResourceAttr("selectel_mks_cluster_v1.cluster_tf_acc_test_1", "maintenance_window_end", "03:00:00"),
					resource.TestCheckResourceAttr("selectel_mks_cluster_v1.cluster_tf_acc_test_1", "status", "ACTIVE"),
				),
			},
		},
	})
}

func testAccCheckMKSClusterV1Exists(name string, mksCluster *cluster.View) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return errors.New("no ID is set")
		}

		var projectID, endpoint string
		if id, ok := rs.Primary.Attributes["project_id"]; ok {
			projectID = id
		}
		if region, ok := rs.Primary.Attributes["region"]; ok {
			endpoint = getMKSClusterV1Endpoint(region)
		}

		config := testAccProvider.Meta().(*Config)
		resellV2Client := config.resellV2Client()
		ctx := context.Background()

		tokenOpts := tokens.TokenOpts{
			ProjectID: projectID,
		}
		token, _, err := tokens.Create(ctx, resellV2Client, tokenOpts)
		if err != nil {
			return errCreatingObject(objectToken, err)
		}

		mksClient := v1.NewMKSClientV1(token.ID, endpoint)
		foundCluster, _, err := cluster.Get(ctx, mksClient, rs.Primary.ID)
		if err != nil {
			return err
		}

		if foundCluster.ID != rs.Primary.ID {
			return errors.New("cluster not found")
		}

		*mksCluster = *foundCluster

		return nil
	}
}

func testAccMKSClusterV1Basic(projectName, clusterName, nodegroupName string) string {
	return fmt.Sprintf(`
resource "selectel_vpc_project_v2" "project_tf_acc_test_1" {
  name        = "%s"
  auto_quotas = true
}

resource "selectel_mks_cluster_v1" "cluster_tf_acc_test_1" {
  name         = "%s"
  kube_version = "1.16.8"
  project_id   = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  region       = "ru-3"

  nodegroups {
    name              = "%s"
    availability_zone = "ru-3a"
    count             = 1
    cpus              = 1
    ram_mb            = 1024
    volume_gb         = 10
    volume_type       = "fast.ru-3a"
  }
}`, projectName, clusterName, nodegroupName)
}
