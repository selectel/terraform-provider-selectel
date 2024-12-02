package selectel

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/selectel/dbaas-go"
	"github.com/selectel/go-selvpcclient/v4/selvpcclient/resell/v2/projects"
)

func TestAccDBaaSGrantV1Basic(t *testing.T) {
	var (
		dbaasGrant dbaas.Grant
		project    projects.Project
	)

	projectName := acctest.RandomWithPrefix("tf-acc")
	datastoreName := acctest.RandomWithPrefix("tf-acc-ds")
	userName := RandomWithPrefix("tf_acc_user")
	userPassword := acctest.RandomWithPrefix("tf-acc-pass")
	databaseName := RandomWithPrefix("tf_acc_db")
	nodeCount := 1

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccSelectelPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVPCV2ProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDBaaSGrantV1Basic(projectName, datastoreName, userName, userPassword, databaseName, nodeCount),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCV2ProjectExists("selectel_vpc_project_v2.project_tf_acc_test_1", &project),
					testAccCheckDBaaSGrantV1Exists("selectel_dbaas_grant_v1.grant_tf_acc_test_1", &dbaasGrant),
					resource.TestCheckResourceAttrSet("selectel_dbaas_grant_v1.grant_tf_acc_test_1", "datastore_id"),
					resource.TestCheckResourceAttrSet("selectel_dbaas_grant_v1.grant_tf_acc_test_1", "database_id"),
					resource.TestCheckResourceAttrSet("selectel_dbaas_grant_v1.grant_tf_acc_test_1", "user_id"),
					resource.TestCheckResourceAttr("selectel_dbaas_grant_v1.grant_tf_acc_test_1", "status", string(dbaas.StatusActive)),
				),
			},
		},
	})
}

func testAccCheckDBaaSGrantV1Exists(n string, dbaasGrant *dbaas.Grant) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		ctx := context.Background()

		dbaasClient, err := newTestDBaaSClient(ctx, rs, testAccProvider)
		if err != nil {
			return err
		}

		grant, err := dbaasClient.Grant(ctx, rs.Primary.ID)
		if err != nil {
			return err
		}

		if grant.ID != rs.Primary.ID {
			return errors.New("grant not found")
		}

		*dbaasGrant = grant

		return nil
	}
}

func testAccDBaaSGrantV1Basic(projectName, datastoreName, userName, userPassword, databaseName string, nodeCount int) string {
	return fmt.Sprintf(`
resource "selectel_vpc_project_v2" "project_tf_acc_test_1" {
  name        = "%s"
}

resource "selectel_vpc_subnet_v2" "subnet_tf_acc_test_1" {
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  region     = "ru-3"
}

data "selectel_dbaas_datastore_type_v1" "dt" {
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  region = "ru-3"
  filter {
    engine = "postgresql"
    version = "12"
  }
}

resource "selectel_dbaas_datastore_v1" "datastore_tf_acc_test_1" {
  name = "%s"
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  region = "ru-3"
  type_id = "${data.selectel_dbaas_datastore_type_v1.dt.datastore_types[0].id}"
  subnet_id = "${selectel_vpc_subnet_v2.subnet_tf_acc_test_1.subnet_id}"
  node_count = "%d"
  flavor {
    vcpus = 2
    ram = 4096
    disk = 32
  }
}

resource "selectel_dbaas_user_v1" "user_tf_acc_test_1" {
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  region = "ru-3"
  datastore_id = "${selectel_dbaas_datastore_v1.datastore_tf_acc_test_1.id}"
  name = "%s"
  password = "%s"
}

resource "selectel_dbaas_database_v1" "database_tf_acc_test_1" {
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  region = "ru-3"
  datastore_id = "${selectel_dbaas_datastore_v1.datastore_tf_acc_test_1.id}"
  name = "%s"
  owner_id = "${selectel_dbaas_user_v1.user_tf_acc_test_1.id}"
}

resource "selectel_dbaas_grant_v1" "grant_tf_acc_test_1" {
  project_id = "${selectel_vpc_project_v2.project_tf_acc_test_1.id}"
  region = "ru-3"
  datastore_id = "${selectel_dbaas_datastore_v1.datastore_tf_acc_test_1.id}"
  database_id = "${selectel_dbaas_database_v1.database_tf_acc_test_1.id}"
  user_id = "${selectel_dbaas_user_v1.user_tf_acc_test_1.id}"
}`, projectName, datastoreName, nodeCount, userName, userPassword, databaseName)
}
