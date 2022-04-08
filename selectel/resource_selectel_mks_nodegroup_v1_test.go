package selectel

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/selectel/go-selvpcclient/selvpcclient/resell/v2/projects"
	"github.com/selectel/go-selvpcclient/selvpcclient/resell/v2/tokens"
	v1 "github.com/selectel/mks-go/pkg/v1"
	"github.com/selectel/mks-go/pkg/v1/nodegroup"
)

func TestAccMKSNodegroupV1Basic(t *testing.T) {
	var (
		mksNodegroup nodegroup.View
		project      projects.Project
	)

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
				Config: testAccMKSNodegroupV1Basic(projectName, clusterName, kubeVersion, maintenanceWindowStart),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					testAccCheckMKSNodegroupV1Exists("selectel_mks_nodegroup_v1.nodegroup_tf_acc_test_1", &mksNodegroup),
					resource.TestCheckResourceAttr("selectel_mks_nodegroup_v1.nodegroup_tf_acc_test_1", "availability_zone", "ru-3a"),
					resource.TestCheckResourceAttr("selectel_mks_nodegroup_v1.nodegroup_tf_acc_test_1", "nodes_count", "2"),
					resource.TestCheckResourceAttr("selectel_mks_nodegroup_v1.nodegroup_tf_acc_test_1", "nodes.#", "2"),
					resource.TestCheckResourceAttr("selectel_mks_nodegroup_v1.nodegroup_tf_acc_test_1", "cpus", "1"),
					resource.TestCheckResourceAttr("selectel_mks_nodegroup_v1.nodegroup_tf_acc_test_1", "ram_mb", "1024"),
					resource.TestCheckResourceAttr("selectel_mks_nodegroup_v1.nodegroup_tf_acc_test_1", "volume_gb", "10"),
					resource.TestCheckResourceAttr("selectel_mks_nodegroup_v1.nodegroup_tf_acc_test_1", "volume_type", "fast.ru-3a"),
					resource.TestCheckResourceAttr("selectel_mks_nodegroup_v1.nodegroup_tf_acc_test_1", "enable_autoscale", "true"),
					resource.TestCheckResourceAttr("selectel_mks_nodegroup_v1.nodegroup_tf_acc_test_1", "autoscale_min_nodes", "2"),
					resource.TestCheckResourceAttr("selectel_mks_nodegroup_v1.nodegroup_tf_acc_test_1", "autoscale_max_nodes", "3"),
					resource.TestCheckResourceAttr("selectel_mks_nodegroup_v1.nodegroup_tf_acc_test_1", "labels.label-key0", "label-value0"),
					resource.TestCheckResourceAttr("selectel_mks_nodegroup_v1.nodegroup_tf_acc_test_1", "labels.label-key1", "label-value1"),
					resource.TestCheckResourceAttr("selectel_mks_nodegroup_v1.nodegroup_tf_acc_test_1", "labels.label-key2", "label-value2"),
					resource.TestCheckResourceAttr("selectel_mks_nodegroup_v1.nodegroup_tf_acc_test_1", "taints.#", "3"),
					resource.TestCheckResourceAttr("selectel_mks_nodegroup_v1.nodegroup_tf_acc_test_1", "taints.0.key", "test-key-0"),
					resource.TestCheckResourceAttr("selectel_mks_nodegroup_v1.nodegroup_tf_acc_test_1", "taints.0.value", "test-value-0"),
					resource.TestCheckResourceAttr("selectel_mks_nodegroup_v1.nodegroup_tf_acc_test_1", "taints.0.effect", "NoSchedule"),
					resource.TestCheckResourceAttr("selectel_mks_nodegroup_v1.nodegroup_tf_acc_test_1", "taints.1.key", "test-key-1"),
					resource.TestCheckResourceAttr("selectel_mks_nodegroup_v1.nodegroup_tf_acc_test_1", "taints.1.value", "test-value-1"),
					resource.TestCheckResourceAttr("selectel_mks_nodegroup_v1.nodegroup_tf_acc_test_1", "taints.1.effect", "NoExecute"),
					resource.TestCheckResourceAttr("selectel_mks_nodegroup_v1.nodegroup_tf_acc_test_1", "taints.2.key", "test-key-2"),
					resource.TestCheckResourceAttr("selectel_mks_nodegroup_v1.nodegroup_tf_acc_test_1", "taints.2.value", "test-value-2"),
					resource.TestCheckResourceAttr("selectel_mks_nodegroup_v1.nodegroup_tf_acc_test_1", "taints.2.effect", "PreferNoSchedule"),
				),
			},
			{
				Config: testAccMKSNodegroupV1Update(projectName, clusterName, kubeVersion, maintenanceWindowStart),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("selectel_mks_nodegroup_v1.nodegroup_tf_acc_test_1", "availability_zone", "ru-3a"),
					resource.TestCheckResourceAttr("selectel_mks_nodegroup_v1.nodegroup_tf_acc_test_1", "nodes_count", "3"),
					resource.TestCheckResourceAttr("selectel_mks_nodegroup_v1.nodegroup_tf_acc_test_1", "nodes.#", "3"),
					resource.TestCheckResourceAttr("selectel_mks_nodegroup_v1.nodegroup_tf_acc_test_1", "cpus", "1"),
					resource.TestCheckResourceAttr("selectel_mks_nodegroup_v1.nodegroup_tf_acc_test_1", "ram_mb", "1024"),
					resource.TestCheckResourceAttr("selectel_mks_nodegroup_v1.nodegroup_tf_acc_test_1", "volume_gb", "10"),
					resource.TestCheckResourceAttr("selectel_mks_nodegroup_v1.nodegroup_tf_acc_test_1", "volume_type", "fast.ru-3a"),
					resource.TestCheckResourceAttr("selectel_mks_nodegroup_v1.nodegroup_tf_acc_test_1", "enable_autoscale", "false"),
					resource.TestCheckResourceAttr("selectel_mks_nodegroup_v1.nodegroup_tf_acc_test_1", "autoscale_min_nodes", "1"),
					resource.TestCheckResourceAttr("selectel_mks_nodegroup_v1.nodegroup_tf_acc_test_1", "autoscale_max_nodes", "4"),
					resource.TestCheckResourceAttr("selectel_mks_nodegroup_v1.nodegroup_tf_acc_test_1", "labels.label-key3", "label-value3"),
					resource.TestCheckResourceAttr("selectel_mks_nodegroup_v1.nodegroup_tf_acc_test_1", "labels.label-key4", "label-value4"),
					resource.TestCheckResourceAttr("selectel_mks_nodegroup_v1.nodegroup_tf_acc_test_1", "taints.#", "3"),
					resource.TestCheckResourceAttr("selectel_mks_nodegroup_v1.nodegroup_tf_acc_test_1", "taints.0.key", "test-key-0"),
					resource.TestCheckResourceAttr("selectel_mks_nodegroup_v1.nodegroup_tf_acc_test_1", "taints.0.value", "test-value-0"),
					resource.TestCheckResourceAttr("selectel_mks_nodegroup_v1.nodegroup_tf_acc_test_1", "taints.0.effect", "NoSchedule"),
					resource.TestCheckResourceAttr("selectel_mks_nodegroup_v1.nodegroup_tf_acc_test_1", "taints.1.key", "test-key-1"),
					resource.TestCheckResourceAttr("selectel_mks_nodegroup_v1.nodegroup_tf_acc_test_1", "taints.1.value", "test-value-1"),
					resource.TestCheckResourceAttr("selectel_mks_nodegroup_v1.nodegroup_tf_acc_test_1", "taints.1.effect", "NoExecute"),
					resource.TestCheckResourceAttr("selectel_mks_nodegroup_v1.nodegroup_tf_acc_test_1", "taints.2.key", "test-key-3"),
					resource.TestCheckResourceAttr("selectel_mks_nodegroup_v1.nodegroup_tf_acc_test_1", "taints.2.value", "test-value-3"),
					resource.TestCheckResourceAttr("selectel_mks_nodegroup_v1.nodegroup_tf_acc_test_1", "taints.2.effect", "NoSchedule"),
				),
			},
		},
	})
}

func testAccCheckMKSNodegroupV1Exists(n string, mksNodegroup *nodegroup.View) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return errors.New("no ID is set")
		}

		clusterID, nodegroupID, err := mksNodegroupV1ParseID(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error parsing resource id: %s", err)
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
		foundNodegroup, _, err := nodegroup.Get(ctx, mksClient, clusterID, nodegroupID)
		if err != nil {
			return err
		}

		if foundNodegroup.ID != nodegroupID {
			return errors.New("nodegroup not found")
		}

		*mksNodegroup = *foundNodegroup

		return nil
	}
}

func testAccMKSNodegroupV1Basic(projectName, clusterName, kubeVersion, maintenanceWindowStart string) string {
	return fmt.Sprintf(`
resource "selectel_vpc_project_v2" "project_tf_acc_test_1" {
  name        = "%s"
  auto_quotas = true
}

resource "selectel_mks_cluster_v1" "cluster_tf_acc_test_1" {
  name                     = "%s"
  kube_version             = "%s"
  project_id               = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  region                   = "ru-3"
  maintenance_window_start = "%s"
}

resource "selectel_mks_nodegroup_v1" "nodegroup_tf_acc_test_1" {
  cluster_id          = "${selectel_mks_cluster_v1.cluster_tf_acc_test_1.id}"
  project_id          = "${selectel_mks_cluster_v1.cluster_tf_acc_test_1.project_id}"
  region              = "${selectel_mks_cluster_v1.cluster_tf_acc_test_1.region}"
  availability_zone   = "ru-3a"
  nodes_count         = 2
  cpus                = 1
  ram_mb              = 1024
  volume_gb           = 10
  volume_type         = "fast.ru-3a"
  enable_autoscale    = true
  autoscale_min_nodes = 2
  autoscale_max_nodes = 3
  labels = {
    label-key0 = "label-value0"
    label-key1 = "label-value1"
    label-key2 = "label-value2"
  }
  taints {
    key = "test-key-0"
    value = "test-value-0"
    effect = "NoSchedule"
  }
  taints {
    key = "test-key-1"
    value = "test-value-1"
    effect = "NoExecute"
  }
  taints {
    key = "test-key-3"
    value = "test-value-3"
    effect = "NoSchedule"
  }
}`, projectName, clusterName, kubeVersion, maintenanceWindowStart)
}

func testAccMKSNodegroupV1Update(projectName, clusterName, kubeVersion, maintenanceWindowStart string) string {
	return fmt.Sprintf(`
resource "selectel_vpc_project_v2" "project_tf_acc_test_1" {
  name        = "%s"
  auto_quotas = true
}

resource "selectel_mks_cluster_v1" "cluster_tf_acc_test_1" {
  name                     = "%s"
  kube_version             = "%s"
  project_id               = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  region                   = "ru-3"
  maintenance_window_start = "%s"
}

resource "selectel_mks_nodegroup_v1" "nodegroup_tf_acc_test_1" {
  cluster_id          = "${selectel_mks_cluster_v1.cluster_tf_acc_test_1.id}"
  project_id          = "${selectel_mks_cluster_v1.cluster_tf_acc_test_1.project_id}"
  region              = "${selectel_mks_cluster_v1.cluster_tf_acc_test_1.region}"
  availability_zone   = "ru-3a"
  nodes_count         = 3
  cpus                = 1
  ram_mb              = 1024
  volume_gb           = 10
  volume_type         = "fast.ru-3a"
  enable_autoscale    = false
  autoscale_min_nodes = 1
  autoscale_max_nodes = 4
  labels = {
    label-key3 = "label-value3"
    label-key4 = "label-value4"
  }
  taints {
    key = "test-key-0"
    value = "test-value-0"
    effect = "NoSchedule"
  }
  taints {
    key = "test-key-1"
    value = "test-value-1"
    effect = "NoExecute"
  }
  taints {
    key = "test-key-2"
    value = "test-value-2"
    effect = "PreferNoSchedule"
  }
}`, projectName, clusterName, kubeVersion, maintenanceWindowStart)
}
